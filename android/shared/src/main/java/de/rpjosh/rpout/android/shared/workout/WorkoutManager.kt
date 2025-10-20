package de.rpjosh.rpout.android.shared.workout

import android.annotation.SuppressLint
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.media.MediaPlayer
import android.os.VibrationEffect
import android.os.Vibrator
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.graphics.Color
import androidx.core.app.NotificationCompat
import androidx.health.services.client.ExerciseClient
import androidx.health.services.client.ExerciseUpdateCallback
import androidx.health.services.client.HealthServices
import androidx.health.services.client.data.Availability
import androidx.health.services.client.data.BatchingMode
import androidx.health.services.client.data.DataType
import androidx.health.services.client.data.DataTypeAvailability
import androidx.health.services.client.data.DeltaDataType
import androidx.health.services.client.data.ExerciseConfig
import androidx.health.services.client.data.ExerciseLapSummary
import androidx.health.services.client.data.ExerciseState
import androidx.health.services.client.data.ExerciseType
import androidx.health.services.client.data.ExerciseUpdate
import androidx.health.services.client.data.LocationAvailability
import androidx.health.services.client.data.SampleDataPoint
import androidx.health.services.client.data.WarmUpConfig
import androidx.health.services.client.endExercise
import androidx.health.services.client.getCapabilities
import androidx.health.services.client.getCurrentExerciseInfo
import androidx.health.services.client.pauseExercise
import androidx.health.services.client.prepareExercise
import androidx.health.services.client.resumeExercise
import androidx.health.services.client.startExercise
import de.rpjosh.rpout.android.shared.R
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.inject.InjectionFactory
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.Tr
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.selects.onTimeout
import kotlinx.coroutines.selects.select
import java.time.Duration
import java.time.Instant
import kotlin.math.abs
import kotlin.math.roundToInt
import androidx.core.graphics.toColorInt
import androidx.health.services.client.data.CumulativeDataPoint
import androidx.health.services.client.data.ExerciseTrackedStatus
import de.rpjosh.rpout.android.shared.models.ActivityType

/**
 * WorkoutManager contains all the logic for tracking a workout.
 *
 * It does also provide UI states that you can use in activities (by compose).
 *
 * Because it's different between the phone and watch,
 * you have to provide the device in which context it's called
 */
class WorkoutManager(val isWearOs: Boolean, private val typeId: Long) {

    @Inject(parameters = [ "WorkoutManager" ])
    private lateinit var logger: Logger
    @Inject private lateinit var workoutController: WorkoutController

    /** Workout type to track */
    lateinit var type: WorkoutType
        private set
    /** Foreground color of the type */
    var typeAccentColor = mutableStateOf(Color.White)

    /** State of the workout tracking */
    val state = mutableStateOf(State.NOT_INITIALIZED)

    /** Points of the workout */
    lateinit var gpsWorkout: GpsWorkout
    /** Points of the workout for the UI */
    var workoutData: Workout = Workout()
    /** Summary of the workout */
    var workoutSummary: WorkoutSummary = WorkoutSummary()

    /** Synchronized event used to interact with data points and the health API */
    val dataLock = Object()
    /** Activity that displays the workout data (StartActivity -> TrackingActivity) */
    var activityClass: Class<*>? = null
    private var healthExerciseClient: ExerciseClient? = null
    var healthSupportedCapabilities: SupportedCapabilities? = null
    private var healthExerciseType: ExerciseType? = null
    var lastGpsConnectedTime = 0L

    /** Chanel to send messages when the workout is ended and the last update was processed */
    val endChannel = Channel<String>(capacity = 5)

    /** Location manager to request one time locations */
    private lateinit var locationManagerOneTime: WorkoutLocation
    private lateinit var notifcationManager: NotificationManager

    companion object {

        /** Global state of the workout manager */
        var workoutManager: WorkoutManager? = null

        const val INTENT_NOTIFICATION_GET_ONETIME_LOCATION = "get_onetime_location"
        const val NOTIFICATION_NO_GPS_ID = -45

        /** Creates a new dummy instance used for composer preview generation */
        fun forPreview(isWearOs: Boolean, typeAccentColor: String = "#E37029", heartRate: Int = 132, totalKm: Double = 3.23): WorkoutManager {
            val rtc = WorkoutManager(isWearOs, -1)

            // Init type
            rtc.typeAccentColor.value = Color(typeAccentColor.toColorInt())
            rtc.type = WorkoutType(
                id = 0, nameEn = "Hiking", nameDe = "Gehen", tagDark = "#fff", tagWhite = "",
                icon = "<svg class=\"icon\" viewBox=\"0 0 16 21\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\"> <path transform=\"translate(-4,-2)\" fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M13 6C14.1046 6 15 5.10457 15 4C15 2.89543 14.1046 2 13 2C11.8955 2 11 2.89543 11 4C11 5.10457 11.8955 6 13 6ZM11.0528 6.60557C11.3841 6.43992 11.7799 6.47097 12.0813 6.68627L13.0813 7.40056C13.3994 7.6278 13.5559 8.01959 13.482 8.40348L12.4332 13.847L16.8321 20.4453C17.1384 20.9048 17.0143 21.5257 16.5547 21.8321C16.0952 22.1384 15.4743 22.0142 15.168 21.5547L10.5416 14.6152L9.72611 13.3919C9.58336 13.1778 9.52866 12.9169 9.57338 12.6634L10.1699 9.28309L8.38464 10.1757L7.81282 13.0334C7.70445 13.575 7.17759 13.9261 6.63604 13.8178C6.09449 13.7094 5.74333 13.1825 5.85169 12.641L6.51947 9.30379C6.58001 9.00123 6.77684 8.74356 7.05282 8.60557L11.0528 6.60557ZM16.6838 12.9487L13.8093 11.9905L14.1909 10.0096L17.3163 11.0513C17.8402 11.226 18.1234 11.7923 17.9487 12.3162C17.7741 12.8402 17.2078 13.1234 16.6838 12.9487ZM6.12844 20.5097L9.39637 14.7001L9.70958 15.1699L10.641 16.5669L7.87159 21.4903C7.60083 21.9716 6.99111 22.1423 6.50976 21.8716C6.0284 21.6008 5.85768 20.9911 6.12844 20.5097Z\" fill=\"currentColor\"/> </svg>",
            )

            // Init workout data for UI
            rtc.workoutData = Workout()
            rtc.workoutData.setHeartRate(SampleDataPoint(DataType.HEART_RATE_BPM, heartRate.toDouble(), Duration.ofMillis(0)))
            rtc.workoutData.setDistance(CumulativeDataPoint(DataType.DISTANCE_TOTAL, totalKm * 1000, Instant.now(), Instant.now()))

            return rtc
        }

    }

    /**
     * Initializes all internal variables that are required
     * for tracking a workout.
     *
     * It's an extra function to not block the UI thread but still provide
     * the states for initializing the UI
     */
    fun init() {

        // Initialize full workout type
        val t = workoutController.dao().getType(typeId)
        if (t == null) {
            logger.log("ee", "Failed to get workout type. Things won't work correctly")
            type = WorkoutType(0, "Unknown", "Unknown", "#000000", "#FFFFFF", "")
        } else {
            type = t
        }
        typeAccentColor.value = Color(type.tagDark.toColorInt())

    }

    /** Starts the (already prepared) workout with the exercise client */
    @SuppressLint("RestrictedApi")
    suspend fun start() {

        // Check if the workout is currently running
        healthExerciseClient?.let {
            try {
                if (it.getCurrentExerciseInfo().exerciseTrackedStatus == ExerciseTrackedStatus.OWNED_EXERCISE_IN_PROGRESS) {
                    when(state.value) {
                        State.TRACKED, State.TRACKED_GPS_CONNECTING -> {
                            logger.log("i", "Got a start request but found an already running workout. Not starting again")
                            return
                        }
                        State.PAUSED -> {
                            logger.log("i", "Got a start request but found an already running workout that's paused. Resuming it")
                            resume()
                            return
                        }
                        else -> { /* Not doing anything */ }
                    }
                }
            } catch (ex: Exception) {
                logger.log("w", ex, "Failed to get current exercise status")
            }
        }

        // Create workout struct to sync against the server
        synchronized(dataLock) {
            gpsWorkout = GpsWorkout(
                type = typeId
            )
            gpsWorkout.id = workoutController.dao().insertGpsWorkout(gpsWorkout)
            logger.log("i", "Started workout (#${gpsWorkout.id})")

            // Fill summary with values
            workoutSummary.typeId = type.id
            workoutSummary.name = type.getName(Tr.getUsedLanguage())
            workoutSummary.typeAccentColor = typeAccentColor.value
        }

        // Data types to support
        val dataTypes = mutableSetOf<DataType<*, *>>()
        if (healthSupportedCapabilities?.heartRate == true) {
            dataTypes.add(DataType.HEART_RATE_BPM)
            dataTypes.add(DataType.HEART_RATE_BPM_STATS)
            dataTypes.add(DataType.CALORIES_TOTAL)
        }
        if (healthSupportedCapabilities?.gps == true) dataTypes.add(DataType.LOCATION)
        if (healthSupportedCapabilities?.totalSteps == true) dataTypes.add(DataType.STEPS_TOTAL)
        if (healthSupportedCapabilities?.elevation == true) dataTypes.add(DataType.ABSOLUTE_ELEVATION)
        if (healthSupportedCapabilities?.elevationGain == true) dataTypes.add(DataType.ELEVATION_GAIN_TOTAL)
        if (healthSupportedCapabilities?.elevationLoss == true) dataTypes.add(DataType.ELEVATION_LOSS_TOTAL)
        if (healthSupportedCapabilities?.speed == true) dataTypes.apply { add(DataType.SPEED); add(DataType.SPEED_STATS) }
        if (healthSupportedCapabilities?.totalDistance == true) dataTypes.add(DataType.DISTANCE_TOTAL)

        // Batching mode override
        val batchOverrides = mutableSetOf<BatchingMode>()
        if (healthSupportedCapabilities?.heartRateLive == true) batchOverrides.add(BatchingMode.HEART_RATE_5_SECONDS)

        // Update exercise state and duration early so the time doesn't hang for ~2 seconds until the first values come in
        workoutData.exerciseState.value = ExerciseState.ACTIVE
        workoutData.activeDuration.value = ExerciseUpdate.ActiveDurationCheckpoint(Instant.now(), Duration.ofMillis(0))

        synchronized(dataLock) {
            if (state.value == State.READY) state.value = State.TRACKED
            else                            state.value = State.TRACKED_GPS_CONNECTING
        }

        healthExerciseClient?.startExercise(
            configuration = ExerciseConfig(
                exerciseType = healthExerciseType!!,
                dataTypes = dataTypes,
                isAutoPauseAndResumeEnabled = false,
                isGpsEnabled = healthSupportedCapabilities?.gps == true,
                batchingModeOverrides = batchOverrides
            )
        )
    }

    /** Pauses the currently running workout */
    suspend fun pause() {
        logger.log("i", "Paused workout (#${gpsWorkout.id})")

        healthExerciseClient?.pauseExercise()

        synchronized(dataLock) {
            state.value = State.PAUSED
        }
    }

    /** Resumes the workout from a paused workout */
    suspend fun resume() {
        logger.log("i", "Resumed workout (#${gpsWorkout.id})")

        healthExerciseClient?.resumeExercise()

        synchronized(dataLock) {
            // @TODO check current GPS connecting state
            state.value = State.TRACKED
        }
    }

    /** Stops the currently tracked workout */
    @OptIn(ExperimentalCoroutinesApi::class)
    suspend fun stop() {
        gpsWorkout.isFinished = true
        gpsWorkout.endTime = System.currentTimeMillis() / 1000
        gpsWorkout.speedAvg = workoutSummary.speedAv
        gpsWorkout.distanceTotal = workoutSummary.distance
        gpsWorkout.useDeviceData = healthSupportedCapabilities?.gps == false

        while (endChannel.tryReceive().isSuccess) {
            // Clear end channel so we have no messages
        }

        // Remove any pending notification
        notifcationManager.cancel(NOTIFICATION_NO_GPS_ID)

        // Wait until workout is completely processed
        healthExerciseClient?.endExercise()
        select {
            endChannel.onReceive {
               // Continue processing
            }
            onTimeout(1500) {
                logger.log("i", "Timed out waiting for all data to be processed. Some data my be missing")
            }
        }

        logger.log("i", "Stopped workout (#${gpsWorkout.id})")
    }

    /** Processes received data points from the exercise client */
    @Synchronized
    fun processDataPoints(update: ExerciseUpdate) {
        // logger.log("d", "Processing data point")
        val unixTime = System.currentTimeMillis() / 1000

        val latestMetrics = update.latestMetrics
        val activeDuration = update.activeDurationCheckpoint

        // Update active duration (only if it was changed to improve performance)
        if (activeDuration != null && ( workoutData.activeDuration.value.activeDuration.seconds != activeDuration.activeDuration.seconds || workoutData.activeDuration.value.time.epochSecond != activeDuration.time.epochSecond) ) {
            workoutData.activeDuration.value = activeDuration
        }
        if (workoutData.exerciseState.value != update.exerciseStateInfo.state) workoutData.exerciseState.value = update.exerciseStateInfo.state

        // If android batches workout data, we receive multiple values at once (but in different update callbacks).
        // Because the sensor interval times can be different, we have to create them all
        // before we can fill them with data
        var newPointsAdded: Boolean
        synchronized(dataLock) {
            val newPoints = arrayListOf<GpsWorkoutPoint>()

            // Workout already finished
            if (!::gpsWorkout.isInitialized || gpsWorkout.isFinished) return

            if (state.value == State.PAUSED) { /* Paused => don't add data points */ }
            // No previous points to compare available
            else if (gpsWorkout.points.isEmpty()) {
                val empty = GpsWorkoutPoint.emptyPoint(unixTime, gpsWorkout.id)

                // Add location from one time GPS fix
                if (healthSupportedCapabilities?.gps == false && workoutData.location.time.seconds != 0L) {
                    empty.longitude = workoutData.location.value.value.longitude.toFloat()
                    empty.latitude = workoutData.location.value.value.latitude.toFloat()
                }

                newPoints.add(empty)
            }
            else {
                // Initialize a new point every 6 seconds (sample rate of RPout)
                for (i in gpsWorkout.points.last().unixTime + 6 until unixTime + 1 step 6) {
                    newPoints.add(GpsWorkoutPoint.emptyPoint(i, gpsWorkout.id))
                }
            }

            // Add these (empty) points already to the global gps points
            newPointsAdded = newPoints.isNotEmpty()
            gpsWorkout.points.addAll(newPoints)
        }

        synchronized(dataLock) {
            // Workout already finished
            if (gpsWorkout.isFinished) return

            if (healthSupportedCapabilities?.heartRate == true) {
                val metrics = latestMetrics.getData(DataType.HEART_RATE_BPM)
                if (metrics.isNotEmpty()) workoutData.setHeartRate(metrics.last())
                gpsWorkout.points.forEachIndexed { i, it ->
                    if (it.heartRate == 0) {
                        val closest = getClosestPoint(metrics, it.unixTime, 2)
                        if (closest != null) gpsWorkout.points[i].heartRate = closest.value.toInt()
                        else if (workoutData.heartRate.isInLast(2, it.unixTime)) gpsWorkout.points[i].heartRate = workoutData.heartRate.value.value
                    }
                }
            }
            if (healthSupportedCapabilities?.gps == true) {
                val metrics = latestMetrics.getData(DataType.LOCATION)
                if (metrics.isNotEmpty()) {
                    workoutData.setLocation(metrics.last())
                    lastGpsConnectedTime = unixTime
                }

                gpsWorkout.points.forEachIndexed { i, it ->
                    if (it.latitude == 0f) {
                        val closest = getClosestPoint(metrics, it.unixTime, 3)
                        if (closest != null) {
                            gpsWorkout.points[i].latitude = closest.value.latitude.toFloat()
                            gpsWorkout.points[i].longitude = closest.value.longitude.toFloat()

                            val elevation = closest.value.altitude
                            if (!elevation.isNaN() && elevation < 10000) gpsWorkout.points[i].elevation = elevation.roundToInt()
                        } else if (workoutData.location.isInLast(1, it.unixTime)) {
                            gpsWorkout.points[i].latitude = workoutData.location.value.value.latitude.toFloat()
                            gpsWorkout.points[i].longitude = workoutData.location.value.value.longitude.toFloat()

                            val elevation = workoutData.location.value.value.altitude
                            if (!elevation.isNaN() && elevation < 10000) gpsWorkout.points[i].elevation = elevation.roundToInt()
                        }
                    }
                }
            }
            if (healthSupportedCapabilities?.elevation == true) {
                val metrics = latestMetrics.getData(DataType.ABSOLUTE_ELEVATION)
                if (metrics.isNotEmpty()) workoutData.setElevation(metrics.last())
                gpsWorkout.points.forEachIndexed { i, it ->
                    // Only apply calculated elevation from device (like barometer) if we don't have a GPS elevation value
                    if (it.elevation == 0) {
                        val closest = getClosestPoint(metrics, it.unixTime, 2)
                        if (closest != null) gpsWorkout.points[i].elevation = closest.value.roundToInt()
                        else if (workoutData.elevation.isInLast(3, it.unixTime)) gpsWorkout.points[i].elevation = workoutData.elevation.value.value
                    }
                }
            }
            if (healthSupportedCapabilities?.totalSteps == true && gpsWorkout.points.isNotEmpty()) {
                latestMetrics.getData(DataType.STEPS_TOTAL)?.let {
                    gpsWorkout.points[gpsWorkout.points.lastIndex].steps = it.total.toInt()
                    workoutSummary.steps = it.total.toInt()
                }
            }
            if (healthSupportedCapabilities?.totalDistance == true) {
                val latest = latestMetrics.getData(DataType.DISTANCE_TOTAL)
                latest?.let { workoutData.setDistance(it) }

                if (healthSupportedCapabilities?.gps === false) {
                    gpsWorkout.points.forEachIndexed { i, it ->
                        // We don't fill concrete data because we don't have a good way to track it (without summing individual values up)
                        it.totalDistance = latest?.total?.roundToInt() ?: workoutData.distance.value.value
                    }
                }
            }
            if (healthSupportedCapabilities?.speed == true) {
                val metrics = latestMetrics.getData(DataType.SPEED)
                if (metrics.isNotEmpty()) workoutData.setSpeed(metrics.last())

                if (healthSupportedCapabilities?.gps === false) {
                    gpsWorkout.points.forEachIndexed { i, it ->
                        if (it.speed == 0) {
                            val closest = getClosestPoint(metrics, it.unixTime, 3)
                            it.speed = closest?.value?.let { (1000 / it).roundToInt() } ?: 0
                        }
                    }
                }
            }

            // Process GPS points
            processGpsPoints(false)
        }

        // Update summary stats only every 6 seconds to save power.
        // We should always receive a value because they are based on totals
        if (newPointsAdded) {
            if (healthSupportedCapabilities?.heartRate == true) {
                latestMetrics.getData(DataType.HEART_RATE_BPM_STATS)?.let {
                    workoutSummary.heartRateMax = it.max.roundToInt()
                    workoutSummary.heartRateAv = it.average.roundToInt()
                }
                latestMetrics.getData(DataType.CALORIES_TOTAL)?.let {
                    workoutSummary.calories = it.total.roundToInt()
                }
            }
            if (healthSupportedCapabilities?.elevationGain == true) {
                latestMetrics.getData(DataType.ELEVATION_GAIN_TOTAL)?.let {
                    workoutSummary.elevationUp = it.total.roundToInt()
                }
            }
            if (healthSupportedCapabilities?.elevationLoss == true) {
                latestMetrics.getData(DataType.ELEVATION_LOSS_TOTAL)?.let {
                    workoutSummary.elevationDown = it.total.roundToInt()
                }
            }
            if (healthSupportedCapabilities?.speed == true) {
                latestMetrics.getData(DataType.SPEED_STATS)?.let {
                    workoutSummary.speedAv = (1000 / it.average).roundToInt()
                }
            }

            // Apply data from last point
            workoutSummary.distance = workoutData.distance.value.value
            workoutData.activeDuration.value.let { workoutSummary.duration = (unixTime - it.time.epochSecond + it.activeDuration.seconds).toInt() }

            // logger.log("d", "Stats: $workoutSummary")
        }
    }

    /**
     * Handles the processing and finishing of the previously added GPS points and stores them in the database.
     * You have to call this function while synchronizing over the data lock.
     *
     * If the force store option was provided, the callback is executed on a background task.
     *
     * This function returns whether points are stored in the db
     */
    @Synchronized
    private fun processGpsPoints(forceStore: Boolean, onStored: (() -> Unit)? = null): Boolean {
        if (!::gpsWorkout.isInitialized) return false

        // Only process if we have at least 50 data points
        if (gpsWorkout.points.size < 50 && !forceStore) return false

        // No data to process
        if (gpsWorkout.points.isEmpty()) return false

        // Log current stats
        logger.log("d", "Stats: $workoutSummary")

        // Get data points to process. We keep 10 points (at least 60 seconds) to still have these
        // points in update callback from exercise client when we got new data.
        // We always keep one "old" workout point to have default values for filling in empty points
        val defaultPoint = gpsWorkout.points.first()
        val points: List<GpsWorkoutPoint> = gpsWorkout.points.toList().subList(1, if(forceStore) gpsWorkout.points.size else 39)
        // Remove them from GPS workout
        gpsWorkout.points = gpsWorkout.points.toMutableList().subList(points.size, gpsWorkout.points.size)

        // Filter completely empty points that don't even have a single value.
        // We don't push them because it doesn't have any sense to push the same point
        // as the last one.
        // This could result into detecting a pause if the last point is more than a minute ago
        val filteredPoints = points.filter { !it.isEmpty() }

        // If we didn't received a value for a point, use the last available one
        filteredPoints.forEachIndexed{ i, v ->
            // No previous values are available
            val lastPoint = if (i == 0) {
                // Get default point we kept back
                defaultPoint
            } else {
                // Use last point of list
                filteredPoints[i-1]
            }

            // Fill last values
            if (v.latitude == 0f || v.longitude == 0f) {
                filteredPoints[i].latitude = lastPoint.latitude
                filteredPoints[i].longitude = lastPoint.longitude
            }
            if (v.elevation == 0) filteredPoints[i].elevation = lastPoint.elevation
            if (v.heartRate == 0) filteredPoints[i].heartRate = lastPoint.heartRate
            if (v.steps == 0) filteredPoints[i].steps = lastPoint.steps
        }

        // Set the first point to the last processed one
        if (gpsWorkout.points.isNotEmpty() && filteredPoints.isNotEmpty()) gpsWorkout.points[0] = filteredPoints.last()

        // Store points inside db
        Thread{
            workoutController.dao().insertGpsWorkoutPoints(filteredPoints)
            onStored?.let { it() }
        }.start()

        return true
    }

    /**
     * Returns the closest data point to the provided unix time stamp in seconds. Only a point with the
     * given "allowedOffsetSeconds" is used.
     * If no point was found, null is returned
     */
    private fun <T: Any> getClosestPoint(points: List<SampleDataPoint<T>>, toUnixTimeSec: Long, allowedOffsetSeconds: Int): SampleDataPoint<T>? {
        if (points.isEmpty()) return null
        val timeBoot = TimeHelper.getBootTimeFromUnixTime(toUnixTimeSec)

        var closest = points[0]
        points.forEach {
            // Get offsets
            val closestOffset = abs(timeBoot - closest.timeDurationFromBoot.toMillis())
            val itOffset = abs(timeBoot - it.timeDurationFromBoot.toMillis())

            if (itOffset < closestOffset) {
                closest = it
            }
        }

        // Check if in bounds
        if (abs(TimeHelper.getUnixTimeFromBootTime(closest.timeDurationFromBoot) - toUnixTimeSec) <= allowedOffsetSeconds) {
            return closest
        }

        return null
    }

    /**
     * Initializes the exercise client from the health API.
     *
     * This function should only be called from the foreground service!
     */
    @SuppressLint("RestrictedApi")
    suspend fun initExercise(context: Context, workoutActivityClass: Class<*>, inject: InjectionFactory) {
        synchronized(dataLock) {
            activityClass = workoutActivityClass
        }

        // Initialize dependencies
        locationManagerOneTime = WorkoutLocation(
            context = context,
            logger = inject.inject(Logger::class.java, arrayOf("OneTimeLocation"), false)
        )
        notifcationManager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

        val callback = object : ExerciseUpdateCallback {

            override fun onExerciseUpdateReceived(update: ExerciseUpdate) {
                val exerciseStateInfo = update.exerciseStateInfo
                val latestMetrics = update.latestMetrics

                when (exerciseStateInfo.state) {
                    ExerciseState.PREPARING -> {
                        // Update heart rate
                        if (healthSupportedCapabilities?.heartRate == true) {
                            if (latestMetrics.getData(DataType.HEART_RATE_BPM).isNotEmpty()) {
                                workoutData.setHeartRate(latestMetrics.getData(DataType.HEART_RATE_BPM).last())
                            }
                        }
                    }
                    ExerciseState.ENDED -> {
                        logger.log("i", "Received last update (state = ended) from workout client")
                        if (::notifcationManager.isInitialized) notifcationManager.cancel(NOTIFICATION_NO_GPS_ID)

                        try {
                            processDataPoints(update)

                            // Force storing of remaining data points
                            synchronized(dataLock) {
                                val storingDataPoints = processGpsPoints(true) {
                                    runBlocking { endChannel.send("") }
                                }
                                if (!storingDataPoints) {
                                    // Send end message manually because no data are processed
                                    Thread { runBlocking { endChannel.send("") } }.start()
                                }
                            }

                        } catch (ex: Exception) {
                            logger.log("e", ex, "Failed to process last data points (last one)")
                        }
                    }
                    else -> {
                        // Main processing
                        try {
                            processDataPoints(update)
                        } catch (ex: Exception) {
                            logger.log("e", ex, "Failed to process last data points")
                        }
                    }
                }
            }

            override fun onLapSummaryReceived(lapSummary: ExerciseLapSummary) {
                // No support for laps yet
            }

            override fun onRegistered() {
                logger.log("d", "Sensors registered successfully")

                // Request location (one time) if no location data is supported
                if (healthSupportedCapabilities?.gps == false) {
                    logger.log("d", "GPS is not supported by the workout type. Requesting one time location")
                    requestOneTimeLocation(context)
                }
            }

            override fun onRegistrationFailed(throwable: Throwable) {
                logger.log("e", "Registration of sensors failed: ${throwable.message}")
                synchronized(dataLock) {
                    state.value = State.ERROR
                }
            }

            @SuppressLint("MissingPermission")
            override fun onAvailabilityChanged(dataType: DataType<*, *>, availability: Availability) {
                when (availability) {
                    is LocationAvailability -> {
                        val isAvailable = availability == LocationAvailability.ACQUIRED_TETHERED || availability == LocationAvailability.ACQUIRED_UNTETHERED
                        val unixTime = System.currentTimeMillis() / 1000

                        synchronized(dataLock) {
                            // Play GPS connected sound
                            if (isAvailable && unixTime - lastGpsConnectedTime > 60 && isWearOs) {
                                notifyGPSConnected(context)
                            }
                            if (isAvailable) lastGpsConnectedTime = unixTime

                            // Update states
                            if (isAvailable && state.value == State.PRE_GPS_CONNECTING) {
                                logger.log("d", "Got GPS signal (pre)")
                                state.value = State.READY
                            } else if (!isAvailable && state.value == State.READY) {
                                logger.log("d", "Lost GPS signal (pre)")
                                state.value = State.PRE_GPS_CONNECTING
                            } else if (isAvailable && state.value == State.TRACKED_GPS_CONNECTING) {
                                logger.log("d", "Got GPS signal")
                                state.value = State.TRACKED
                            } else if (!isAvailable && state.value == State.TRACKED) {
                                logger.log("d", "Lost GPS signal")
                                state.value = State.TRACKED_GPS_CONNECTING
                            }
                        }
                    }

                    is DataTypeAvailability -> {}
                }
            }
        }

        // Initialize services
        val healthService = HealthServices.getClient(context)
        val exerciseClient = healthService.exerciseClient

        // Check if the device supports the workout type
        var exerciseType = getExerciseTypeFromRPout(type.id.toInt())
        val capabilities = exerciseClient.getCapabilities()
        if (exerciseType !in capabilities.supportedExerciseTypes) {
            logger.log("d", "Workout type (for exercise client) is not supported on the device (RPout type = ${type.nameEn}). Falling back to walking")
            exerciseType = ExerciseType.WALKING
        }
        val typeCapabilities = capabilities.getExerciseTypeCapabilities(exerciseType)
        val batchCapabilities = capabilities.supportedBatchingModeOverrides

        // Save supported capabilities
        val supCap = SupportedCapabilities(
            heartRate = DataType.HEART_RATE_BPM in typeCapabilities.supportedDataTypes,
            heartRateLive = BatchingMode.HEART_RATE_5_SECONDS in batchCapabilities,
            totalSteps = DataType.STEPS_TOTAL in typeCapabilities.supportedDataTypes,
            autoPause = typeCapabilities.supportsAutoPauseAndResume,
            gps = DataType.LOCATION in typeCapabilities.supportedDataTypes && type.shouldTrackGPS(),
            speed = DataType.SPEED in typeCapabilities.supportedDataTypes,
            elevation = DataType.ABSOLUTE_ELEVATION in typeCapabilities.supportedDataTypes,
            totalDistance = DataType.DISTANCE_TOTAL in typeCapabilities.supportedDataTypes,
            elevationGain = DataType.ELEVATION_GAIN_TOTAL in typeCapabilities.supportedDataTypes,
            elevationLoss = DataType.ELEVATION_LOSS_TOTAL in typeCapabilities.supportedDataTypes
        )
        logger.log("i", "Supported features for workout '${type.nameEn}': $supCap")

        // Check the current state
        val exerciseInfo = exerciseClient.getCurrentExerciseInfo()
        var isError = true
        when(exerciseInfo.exerciseTrackedStatus) {
            ExerciseTrackedStatus.OTHER_APP_IN_PROGRESS -> {
                logger.log("w", "Cannot start exercise because another app is tracking one already")
            }
            ExerciseTrackedStatus.OWNED_EXERCISE_IN_PROGRESS -> {
                logger.log("w", "An exercise in this app is already started. This should not happen. Stopping it")
                exerciseClient.endExercise()
                isError = false
            }
            ExerciseTrackedStatus.NO_EXERCISE_IN_PROGRESS -> {
                isError = false
            }
        }

        // Set variables
        exerciseClient.setUpdateCallback(callback)
        synchronized(dataLock) {
            if (healthExerciseClient != null) {
                logger.log("w", "Exercise client was already initialized. Doing nothing")
                return@synchronized
            }

            // Set variables
            healthExerciseClient = exerciseClient
            healthSupportedCapabilities = supCap
            healthExerciseType = exerciseType
            state.value = if (isError) State.ERROR else if (healthSupportedCapabilities?.gps == false) State.READY else State.PRE_GPS_CONNECTING
        }

        // Prepare exercise
        val warmUpData = mutableSetOf<DeltaDataType<*, *>>()
        if (healthSupportedCapabilities?.heartRate == true) warmUpData.add(DataType.HEART_RATE_BPM)
        if (healthSupportedCapabilities?.gps == true) warmUpData.add(DataType.LOCATION)
        if (healthSupportedCapabilities?.elevation == true) warmUpData.add(DataType.ABSOLUTE_ELEVATION)
        exerciseClient.prepareExercise(
            WarmUpConfig(exerciseType, warmUpData)
        )
    }


    /**
     * Changes and applies settings for the provided workout type
     */
    fun changeSettings(noGPS: Boolean? = null, liveData: Boolean? = null) {
        // Apply all new settings
        type.noGPS = noGPS ?: type.noGPS
        type.liveUpdates = liveData ?: type.liveUpdates

        // Store them in the db
        workoutController.dao().updateType(type)
    }

    /**
     * Shuts this instance down and removes all created dependencies
     */
    suspend fun shutdownExercise() {
        try {
            healthExerciseClient?.endExercise()
            healthExerciseClient = null
            if (::locationManagerOneTime.isInitialized) locationManagerOneTime.abort()
        } catch (ex: Exception) {
            logger.log("w", ex, "Failed to stop exercise")
        }

    }

    fun requestOneTimeLocation(context: Context) {
        if (::notifcationManager.isInitialized) notifcationManager.cancel(NOTIFICATION_NO_GPS_ID)

        if (workoutData.location.value.value.latitude != 0.0 || workoutData.location.value.value.longitude != 0.0) {
            // Location already known
            return
        }

        locationManagerOneTime.getCurrentLocation( onSuccess = {
            // Play GPS connected sound
            notifyGPSConnected(context)
            if (::notifcationManager.isInitialized) notifcationManager.cancel(NOTIFICATION_NO_GPS_ID)

            synchronized(dataLock) {
                // Insert into points (workout already started)
                if (::gpsWorkout.isInitialized && gpsWorkout.points.isNotEmpty()) gpsWorkout.points.last().let { i ->
                    i.latitude = it.latitude.toFloat()
                    i.longitude = it.longitude.toFloat()
                }

                // Store the values in last known GPS position. This will be set for all other points
                // (and also for already stored points by the go server)
                workoutData.setLocationOneTime(it)
            }
        }, onFailure = {
            logger.log("d", "Getting one time location failed. Pushing notification to retry")
            if (!::notifcationManager.isInitialized) {
                return@getCurrentLocation
            }

            val intent = Intent(context, activityClass).apply {
                putExtra(INTENT_NOTIFICATION_GET_ONETIME_LOCATION, true)
                addFlags(Intent.FLAG_ACTIVITY_SINGLE_TOP)
            }
            val pendingIntent = PendingIntent.getActivity(
                context, System.currentTimeMillis().toInt(), intent, PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
            )

            val channelId = "ActivityNotifications"
            val channel = NotificationChannel(
                channelId,
                Tr.get("workoutNotifications_channelText"),
                NotificationManager.IMPORTANCE_DEFAULT
            )
            notifcationManager.createNotificationChannel(channel)

            val action = NotificationCompat.Action.Builder(
                R.drawable.ic_launcher_foreground, Tr.get("retry"), pendingIntent
            ).build()

            val notification = NotificationCompat.Builder(context, channelId)
                .setSmallIcon(R.drawable.ic_launcher_foreground)
                .setContentTitle(Tr.get("workoutNotificationGPS_title"))
                .setContentText(Tr.get("workoutNotificationGPS_text"))
                .setContentIntent(pendingIntent)
                .setAutoCancel(true)
                .addAction(action)
                .build()

            notifcationManager.notify(NOTIFICATION_NO_GPS_ID, notification)
        })
    }

    /**
     * Transforms the generic workout type from RPout in a exercise
     * type that is supported and understand from the exercise client
     * from the health API.
     * It's mainly used to improve tracking accuracy for special workout types
     */
    private fun getExerciseTypeFromRPout(typeId: Int): ExerciseType {
        return when(typeId) {
            ActivityType.TYPE_HIKING.ordinal       -> ExerciseType.WALKING
            ActivityType.TYPE_RUNNING.ordinal      -> ExerciseType.RUNNING
            ActivityType.TYPE_SURFEN.ordinal       -> ExerciseType.SURFING
            ActivityType.TYPE_SAILING.ordinal      -> ExerciseType.SAILING
            ActivityType.TYPE_SNOWBOARDING.ordinal -> ExerciseType.SNOWBOARDING
            // No GPS support for swimming (SWIMMING_OPEN_WATER) on most watches. But it does work (kinda, based on swimming style)
            ActivityType.TYPE_SWIMMING.ordinal     -> ExerciseType.SURFING
            ActivityType.TYPE_CYCLING.ordinal      -> ExerciseType.MOUNTAIN_BIKING
            ActivityType.TYPE_SKATEBOARDING.ordinal-> ExerciseType.SKATING
            ActivityType.TYPE_VOLLEYBALL.ordinal   -> ExerciseType.VOLLEYBALL
            // Foil pumping doesn't use surfing because of missing steps
            ActivityType.TYPE_PUMP_FOILING.ordinal -> ExerciseType.WALKING
            // Strength training should also track steps. So we don't use the type STRENGTH_TRAINING
            ActivityType.TYPE_STRENGTH_TRAINING.ordinal -> ExerciseType.HIGH_INTENSITY_INTERVAL_TRAINING
            // Default to running for other (not explicitly supported) types because it supports all features
            else -> ExerciseType.RUNNING
        }
    }

    /** Plays a sound to notify the user about an established GPS connection. This method does not block */
    @SuppressLint("MissingPermission")
    fun notifyGPSConnected(context: Context) {
        if (type.isWaterActivity() && state.value in arrayListOf(State.TRACKED, State.TRACKED_GPS_CONNECTING)) {
            return
        }

        Thread {
            // Vibrate
            val vibrator = context.getSystemService(Vibrator::class.java)
            val pattern = longArrayOf(0,  170, 100, 170)
            val amplitude = intArrayOf(0, 255, 0,   255)
            val vibrationEffect = VibrationEffect.createWaveform(pattern, amplitude,-1)
            vibrator.vibrate(vibrationEffect)

            val mediaPlayer = MediaPlayer.create(context, R.raw.connected)
            mediaPlayer?.start()
            mediaPlayer?.setOnCompletionListener {
                mediaPlayer.release()
            }
        }.start()
    }

}

enum class State {
    /** Workout client has not been initialized already */
    NOT_INITIALIZED,
    /** Workout cannot be started (because another app tracks a workout already) */
    ERROR,
    /** GPS is still connecting (in Pre workout phase) */
    PRE_GPS_CONNECTING,
    /** Workout is ready to start */
    READY,
    /** Workout is tracked without an issue */
    TRACKED,
    /** GPS is connecting again (during workout) */
    TRACKED_GPS_CONNECTING,
    /** Workout is paused */
    PAUSED,
}

/**
 * SupportedCapabilities states which capabilities are supported on the current device
 */
data class SupportedCapabilities(
    val heartRate: Boolean,
    val heartRateLive: Boolean,
    val totalSteps: Boolean,
    val autoPause: Boolean,
    val gps: Boolean,
    val speed: Boolean,
    val elevation: Boolean,
    val totalDistance: Boolean,
    val elevationGain: Boolean,
    val elevationLoss: Boolean
) {
    override fun toString(): String {
        return "heartRate = $heartRate (live = $heartRateLive), steps = $totalSteps, autoPause = $autoPause, gps = $gps, speed = $speed, elevation = $elevation (up = $elevationGain, down = $elevationLoss)"
    }
}
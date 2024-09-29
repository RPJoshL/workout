package de.rpjosh.rpout.android.shared.workout

import android.annotation.SuppressLint
import android.content.Context
import android.media.MediaPlayer
import android.os.VibrationEffect
import android.os.Vibrator
import android.util.Log
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.graphics.Color
import androidx.health.services.client.ExerciseClient
import androidx.health.services.client.ExerciseUpdateCallback
import androidx.health.services.client.HealthServices
import androidx.health.services.client.data.Availability
import androidx.health.services.client.data.DataType
import androidx.health.services.client.data.DataTypeAvailability
import androidx.health.services.client.data.DeltaDataType
import androidx.health.services.client.data.ExerciseLapSummary
import androidx.health.services.client.data.ExerciseState
import androidx.health.services.client.data.ExerciseType
import androidx.health.services.client.data.ExerciseUpdate
import androidx.health.services.client.data.LocationAvailability
import androidx.health.services.client.data.WarmUpConfig
import androidx.health.services.client.endExercise
import androidx.health.services.client.getCapabilities
import androidx.health.services.client.getCurrentExerciseInfo
import androidx.health.services.client.prepareExercise
import de.rpjosh.rpout.android.shared.R
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Logger

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
    /** Current heart rate */
    var heartRate = mutableStateOf(0)
    /** Current color of the heart rate */
    var heartRateColor = mutableStateOf(Color.White)

    /** State of the workout tracking */
    val state = mutableStateOf(State.NOT_INITIALIZED)

    /** Synchronized event used to interact with data points and the health API */
    val dataLock = Object()
    /** Activity that displays the workout data (StartActivity -> TrackingActivity) */
    var activityClass: Class<*>? = null
    var healthExerciseClient: ExerciseClient? = null
    var healthSupportedCapabilities: SupportedCapabilities? = null
    var lastGpsConnectedTime = 0

    // UI states for p

    companion object {

        /** Global state of the workout manager */
        public var workoutManager: WorkoutManager? = null

        /** Creates a new dummy instance used for composer preview generation */
        fun forPreview(isWearOs: Boolean, typeAccentColor: String = "#E37029", heartRate: Int = 132): WorkoutManager {
            val rtc = WorkoutManager(isWearOs, -1)
            rtc.typeAccentColor.value = Color(android.graphics.Color.parseColor(typeAccentColor))
            rtc.type = WorkoutType(
                id = 0, nameEn = "Hiking", nameDe = "Gehen", tagDark = "#fff", tagWhite = "",
                icon = "<svg class=\"icon\" viewBox=\"0 0 16 21\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\"> <path transform=\"translate(-4,-2)\" fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M13 6C14.1046 6 15 5.10457 15 4C15 2.89543 14.1046 2 13 2C11.8955 2 11 2.89543 11 4C11 5.10457 11.8955 6 13 6ZM11.0528 6.60557C11.3841 6.43992 11.7799 6.47097 12.0813 6.68627L13.0813 7.40056C13.3994 7.6278 13.5559 8.01959 13.482 8.40348L12.4332 13.847L16.8321 20.4453C17.1384 20.9048 17.0143 21.5257 16.5547 21.8321C16.0952 22.1384 15.4743 22.0142 15.168 21.5547L10.5416 14.6152L9.72611 13.3919C9.58336 13.1778 9.52866 12.9169 9.57338 12.6634L10.1699 9.28309L8.38464 10.1757L7.81282 13.0334C7.70445 13.575 7.17759 13.9261 6.63604 13.8178C6.09449 13.7094 5.74333 13.1825 5.85169 12.641L6.51947 9.30379C6.58001 9.00123 6.77684 8.74356 7.05282 8.60557L11.0528 6.60557ZM16.6838 12.9487L13.8093 11.9905L14.1909 10.0096L17.3163 11.0513C17.8402 11.226 18.1234 11.7923 17.9487 12.3162C17.7741 12.8402 17.2078 13.1234 16.6838 12.9487ZM6.12844 20.5097L9.39637 14.7001L9.70958 15.1699L10.641 16.5669L7.87159 21.4903C7.60083 21.9716 6.99111 22.1423 6.50976 21.8716C6.0284 21.6008 5.85768 20.9911 6.12844 20.5097Z\" fill=\"currentColor\"/> </svg>",
            )

            rtc.heartRate.value = heartRate
            rtc.heartRateColor.value = rtc.getHeartRateZone(heartRate)

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
        typeAccentColor.value = Color(android.graphics.Color.parseColor(type.tagDark))

    }

    /** Stops all previously registered sensors like heartbeat, steps and GPS  */
    fun stopSensors() {

    }

    /**
     * Initializes the exercise client from the health API.
     *
     * This function should only be called from the foreground service!
     */
    suspend fun initExercise(context: Context, workoutActivityClass: Class<*>) {
        synchronized(dataLock) {
            activityClass = workoutActivityClass
        }

        val callback = object : ExerciseUpdateCallback {

            override fun onExerciseUpdateReceived(update: ExerciseUpdate) {
                val exerciseStateInfo = update.exerciseStateInfo
                val activeDuration = update.activeDurationCheckpoint
                val latestMetrics = update.latestMetrics
                val latestGoals = update.latestAchievedGoals

                if (exerciseStateInfo.state == ExerciseState.PREPARING) {
                    if (healthSupportedCapabilities?.heartRate == true) {
                        // Update heart rate
                        if (latestMetrics.getData(DataType.HEART_RATE_BPM).isNotEmpty()) {
                            val hr = latestMetrics.getData(DataType.HEART_RATE_BPM).last().value.toInt()
                            heartRate.value = hr
                            heartRateColor.value = getHeartRateZone(hr)
                        }
                    }
                } else {
                    // Main processing
                }
            }

            override fun onLapSummaryReceived(lapSummary: ExerciseLapSummary) {
                // No support for laps yet
            }

            override fun onRegistered() {
                logger.log("d", "Sensors registered successfully")
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
                            if (isAvailable) lastGpsConnectedTime = unixTime.toInt()

                            // Update states
                            if (isAvailable && state.value == State.PRE_GPS_CONNECTING) {
                                logger.log("d", "Got GPS signal (pre)")
                                state.value = State.READY
                            } else if (!isAvailable && state.value == State.READY) {
                                logger.log("d", "Lost GPS signal")
                                state.value = State.PRE_GPS_CONNECTING
                            }
                        }
                    }

                    is DataTypeAvailability -> {

                    }
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

        // Save supported capabilities
        val supCap = SupportedCapabilities(
            heartRate = DataType.HEART_RATE_BPM in typeCapabilities.supportedDataTypes,
            totalSteps = DataType.STEPS_TOTAL in typeCapabilities.supportedDataTypes,
            autoPause = typeCapabilities.supportsAutoPauseAndResume
        )
        logger.log("i", "Supported features for workout '${type.nameEn}': $supCap")

        // Check the current state
        val exerciseInfo = exerciseClient.getCurrentExerciseInfo()
        var isError = true
        when(exerciseInfo.exerciseTrackedStatus) {
            androidx.health.services.client.data.ExerciseTrackedStatus.OTHER_APP_IN_PROGRESS -> {
                logger.log("w", "Cannot start exercise because another app is tracking one already")
            }
            androidx.health.services.client.data.ExerciseTrackedStatus.OWNED_EXERCISE_IN_PROGRESS -> {
                logger.log("w", "An exercise in this app is already started. This should not happen")
            }
            androidx.health.services.client.data.ExerciseTrackedStatus.NO_EXERCISE_IN_PROGRESS -> {
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
            state.value = if (isError) State.ERROR else State.PRE_GPS_CONNECTING
        }

        // Prepare exercise
        val warmUpData = mutableSetOf<DeltaDataType<*, *>>(DataType.LOCATION)
        if (healthSupportedCapabilities?.heartRate == true) warmUpData.add(DataType.HEART_RATE_BPM)
        exerciseClient.prepareExercise(
            WarmUpConfig(exerciseType, warmUpData)
        )

    }


    /**
     * Changes and applies settings for the provided workout type
     */
    fun changeSettings(usePhoneGps: Boolean? = null) {
        // Apply all new settings
        type.preferSmartphoneGps = usePhoneGps ?: type.preferSmartphoneGps

        // Store them in the db
        workoutController.dao().updateType(type)
    }

    /**
     * Shuts this instance down and removes all created dependencies
     */
    suspend fun shutdownExercise() {
        try {
            healthExerciseClient?.endExercise()
        } catch (ex: Exception) {
            logger.log("w", ex, "Failed to stop exercise")
        }

    }

    /**
     * Transforms the generic workout type from RPout in a exercise
     * type that is supported and understand from the exercise client
     * from the health API.
     * It's mainly used to improve tracking accuracy for special workout types
     */
    private fun getExerciseTypeFromRPout(typeId: Int): ExerciseType {
        return when(typeId) {
            1 -> ExerciseType.WALKING
            2 -> ExerciseType.RUNNING
            // Surfing and Pump foiling
            3, 10 -> ExerciseType.SURFING
            4 -> ExerciseType.SAILING
            5 -> ExerciseType.SNOWBOARDING
            6 -> ExerciseType.SWIMMING_OPEN_WATER
            7 -> ExerciseType.MOUNTAIN_BIKING
            // Skateboarding
            8 -> ExerciseType.SKATING
            9 -> ExerciseType.VOLLEYBALL
            // Default to running for other (not explicitly supported) types
            else -> ExerciseType.RUNNING
        }
    }

    /**
     * Returns the color for the provided heart rate zone
     */
    fun getHeartRateZone(heartRate: Int): Color {
        val color = when {
            heartRate >= 174 -> "#aa00ff"
            heartRate >= 154 -> "#ff6d01"
            heartRate >= 135 -> "#65dd19"
            heartRate >= 116 -> "#00cee9"
            heartRate >= 97 -> "#2862ff"
            else -> "#ff8a80"
        }

        return Color(android.graphics.Color.parseColor(color))
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
    READY
}

/**
 * SupportedCapabilities states which capabilities are supported on the current device
 */
data class SupportedCapabilities(
    val heartRate: Boolean,
    val totalSteps: Boolean,
    val autoPause: Boolean
) {
    override fun toString(): String {
        return "heartRate = $heartRate, steps = $totalSteps, autoPause = $autoPause"
    }
}
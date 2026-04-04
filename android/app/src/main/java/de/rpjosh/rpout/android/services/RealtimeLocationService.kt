package de.rpjosh.rpout.android.services

import android.Manifest
import android.app.Notification
import android.app.Notification.FOREGROUND_SERVICE_IMMEDIATE
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.location.Location
import android.location.LocationListener
import android.location.LocationManager
import android.location.LocationRequest
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.compose.runtime.MutableFloatState
import androidx.compose.runtime.MutableIntState
import androidx.compose.runtime.MutableLongState
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.core.app.NotificationCompat
import androidx.core.content.ContextCompat
import com.google.android.gms.wearable.Node
import com.google.gson.Gson
import de.rpjosh.rpout.android.shared.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.workout.WorkoutTracking
import de.rpjosh.rpout.android.helper.VersionHelper
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.AndroidGpsData
import de.rpjosh.rpout.android.shared.models.AndroidGpsPoint
import de.rpjosh.rpout.android.shared.models.WorkoutStatus
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.Tr
import kotlin.math.abs
import kotlin.math.roundToInt

data class LocationServiceData(
    /** Unix timestamp the last action was received */
    val lastActionData: Long,
    /** Unix timestamp the last high sampling action was received */
    var lastHighSamplingData: Long,

    /** To get a distance correctly, we only calculate it in higher intervals */
    var lastDistancePoint: Location? = null,
    var lastDistanceTime: Long = 0,

    var lastPoint: AndroidGpsPoint? = null,
    /** The last point that was send to the watch */
    var lastWatchPoint: AndroidGpsPoint? = null,

    /** When the last bulk action of points were transmitted to WearOS */
    var lastTransmittedToWearOS: Long = 0,

    /** Weather the listener is already attached  */
    var registered: Boolean = false,

    /** Points that still has to be synced to WearOS */
    var unsyncedPoints: MutableList<AndroidGpsPoint> = mutableListOf(),
)

data class WorkoutUIState(
    var type: MutableState<WorkoutType?> = mutableStateOf(null),
    val state: MutableState<WorkoutStatus> = mutableStateOf(WorkoutStatus.STOP),
    val heartRate: MutableIntState = mutableIntStateOf(0),
    val heartRateAv: MutableIntState = mutableIntStateOf(0),
    val speed: MutableFloatState = mutableFloatStateOf(0f),
    val distance: MutableIntState = mutableIntStateOf(0),
    val elevation: MutableIntState = mutableIntStateOf(0),
    val duration: MutableLongState = mutableLongStateOf(0),
    val durationCheckpoint: MutableLongState = mutableLongStateOf(0),
)

class RealtimeLocationService: Service(), LocationListener {

    companion object {
        const val INTENT_HEART_RATE = "heartRate"
        const val INTENT_DURATION = "duration"
        const val INTENT_DURATION_CHECKPOINT = "durationCheckpoint"
        const val INTENT_HEART_RATE_AV = "heartRateAv"
        const val INTENT_ACTIVITY_ID = "activityId"

        /** How long a high sampling rate is valid */
        const val HIGH_SAMPLING_DURATION = 6_000

        /** Duration after which a distance is calculated */
        const val DISTANCE_CALCULATION_DURATION = 9_000

        /** Maximum accuracy of the GPS point for calculating the distance */
        const val DISTANCE_ACCURACY_THRESHOLD = 9

        /** Interval points are send to WearOS */
        const val DEFAULT_SYNC_INTERVAL = 11_000

        /** Timeout for stopping this service when no data points were received from WearOS */
        const val SERVICE_TIMEOUT = 5 * 60 * 1000

        /** The current status of the workout */
        val status = WorkoutUIState()
    }

    private val notificationId = Singleton.notificationId.getAndIncrement()
    @Inject(parameters = ["RealTimeLocationService"]) private lateinit var logger: Logger
    @Inject private lateinit var wearSynchronization: WearSynchronization
    @Inject private lateinit var workoutController: WorkoutController

    private var lastTrackingRestarted = getCurrentTime()
    private var lastForeground = ""
    private var data = LocationServiceData(
        lastActionData = 0,
        lastHighSamplingData = getCurrentTime(),
    )

    private lateinit var locationManager: LocationManager
    /** Timestamp when the last status was received from the watch */
    private var lastWatchStatusReceived = getCurrentTime()

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onCreate() {
        super.onCreate()

        Singleton.app()
        val app = Singleton.getAppSec()
        app.injection.inject(RealtimeLocationService::class.java, null, false, this)

        locationManager = getSystemService(LOCATION_SERVICE) as LocationManager
    }

    private fun loadWorkoutType(type: Long) {
        val types = workoutController.getWorkoutTypes(VersionHelper.getVersionName(), false)
        val typeDetails = types.find { it.id == type }
        if (typeDetails == null) {
            logger.log("w", "Failed to load details of workout type $type")
            return
        }

        status.type.value = typeDetails
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        val workoutStatus = WorkoutStatus.fromString(intent?.action ?: "")
        workoutStatus?.let {
            status.state.value = it

            val heartRate = intent?.getIntExtra(INTENT_HEART_RATE, 0) ?: 0
            val heartRateAv = intent?.getIntExtra(INTENT_HEART_RATE_AV, 0) ?: 0
            if (heartRate > 20) status.heartRate.intValue = heartRate
            if (heartRateAv > 20) status.heartRateAv.intValue = heartRateAv

            val activityType = intent?.getLongExtra(INTENT_ACTIVITY_ID, 0) ?: 0
            if (activityType != 0L) {
                Thread{
                    try {
                        loadWorkoutType(activityType)
                    } catch(ex: Exception) {
                        logger.log("e", ex, "Failed to load details of workout type")
                    }
                }.start()
            }

            val duration = intent?.getLongExtra(INTENT_DURATION, 0) ?: 0
            val durationCheckpoint = intent?.getLongExtra(INTENT_DURATION_CHECKPOINT, 0) ?: 0
            if (durationCheckpoint != 0L) {
                status.duration.longValue = duration
                status.durationCheckpoint.longValue = durationCheckpoint
            }

            lastWatchStatusReceived = getCurrentTime()

            // Log.d("RPout-Logger", "Received new watch status: $workoutStatus with heart rate = $heartRate | heart rate av = $heartRateAv")
        }

        when (workoutStatus) {

            WorkoutStatus.PREPARE -> {
                // Make sure to not use old data
                resetValues()

                startLocationManager()
            }

            WorkoutStatus.START, WorkoutStatus.RESUME -> {
                lastTrackingRestarted = getCurrentTime()
                startLocationManager()
            }

            WorkoutStatus.PAUSE -> {
                stopLocationManager()
            }

            WorkoutStatus.STOP -> {
                stopService()
            }

            WorkoutStatus.RUNNING -> {
                startLocationManager()
            }

            WorkoutStatus.HIGH_SAMPLING -> {
                startLocationManager()

                synchronized(this) {
                    data.lastHighSamplingData = getCurrentTime()
                }

                wearSynchronization.getNodes {
                    sendToWearOs(it)
                }
            }

            else -> {
                logger.log("d", "Received unknown action ${intent?.action}")
            }
        }

        return START_STICKY
    }

    private fun stopService() {
        // We keep the values so they can still be viewed by an opened activity
        status.state.value = WorkoutStatus.STOP
        stopLocationManager()

        stopForeground(STOP_FOREGROUND_REMOVE)
        stopSelf()
    }

    @Synchronized
    private fun startLocationManager() {
        foreground("service_location_running")

        // No rights to start it
        if (ContextCompat.checkSelfPermission(this, Manifest.permission.ACCESS_FINE_LOCATION) != PackageManager.PERMISSION_GRANTED) {
            logger.log("w", "Location permission not granted)")
            foreground("service_location_noPermission")
            return
        }

        if (data.registered) return
        data.registered = true

        locationManager.requestLocationUpdates(LocationManager.GPS_PROVIDER, 400, 0f, this)
    }

    @Synchronized
    private fun stopLocationManager() {
        foreground("service_location_paused")

        if (!data.registered) return
        data.registered = false
        locationManager.removeUpdates(this)
    }


    /**
     * Displays an foreground notification for the background service
     *
     * @param message       Message translation to display as the content of the notification
     */
    private fun foreground(message: String?) {
        lastForeground = message ?: "Error"

        val channelId = "de.rpjosh.rpout.android.location"
        val channelName = Tr.get("service_location", true)
        val contentTitle = Tr.get("service_locationTitle")

        val messageToShow =  Tr.get(message)

        val channel = NotificationChannel(
            channelId,
            channelName,
            NotificationManager.IMPORTANCE_NONE
        )
        channel.lockscreenVisibility = Notification.VISIBILITY_PRIVATE
        val manager = (getSystemService(NOTIFICATION_SERVICE) as NotificationManager)
        manager.createNotificationChannel(channel)

        // Activity to show when the user clicks on the notification
        val pendingIntent = PendingIntent.getActivity(
            this, 0, Intent(this, WorkoutTracking::class.java),
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) PendingIntent.FLAG_MUTABLE else 0
        )

        val notificationBuilder = NotificationCompat.Builder(this, channelId)

        // Conditional options for the notification per API level
        if (Build.VERSION.SDK_INT >= 31) {
            notificationBuilder.foregroundServiceBehavior = FOREGROUND_SERVICE_IMMEDIATE
        }

        val notification: Notification = notificationBuilder.setOngoing(true)
            .setSmallIcon(R.drawable.ic_launcher_foreground)
            .setContentTitle(contentTitle)
            .setContentIntent(pendingIntent)
            .setContentText(messageToShow)
            .setPriority(NotificationManager.IMPORTANCE_MIN)
            .setCategory(Notification.CATEGORY_SERVICE)
            .setShowWhen(false)
            .build()

        startForeground(notificationId, notification)
    }

    override fun onLocationChanged(location: Location) {
        // logger?.log("d", "Received new location: lat = ${location.latitude}, long = ${location.longitude}, accuracy = ${location.accuracy}, speed = ${location.speed}")

        val oldPoint = data.lastPoint
        var newPoint = data.lastPoint
        synchronized(this) {
            var distance = (data.lastPoint?.totalDistance ?: 0.0)

            val isRunning = status.state.value == WorkoutStatus.RUNNING || status.state.value == WorkoutStatus.HIGH_SAMPLING
            if (location.accuracy < DISTANCE_ACCURACY_THRESHOLD && timeoutReached(data.lastDistanceTime, DISTANCE_CALCULATION_DURATION) && isRunning) {
                distance += if (data.lastDistancePoint == null) 0.0 else location.distanceTo(data.lastDistancePoint!!).toDouble()
                data.lastDistancePoint = location
                data.lastDistanceTime = getCurrentTime()
            }

            if(!isPointAccurate(location)) {
                logger.log("d", "Location is not accurate enough: ${location.accuracy}. Skipping it")
                return
            }

            newPoint = AndroidGpsPoint(
                location.time,
                location.latitude,
                location.longitude,
                location.altitude,
                location.speed,
                distance,
            )
            data.lastPoint = newPoint
            data.unsyncedPoints.add(newPoint)
        }

        handleNewPoint(oldPoint, newPoint!!)
    }

    fun handleNewPoint(oldPoint: AndroidGpsPoint?, newPoint: AndroidGpsPoint) {
        val oldWatchPoint = data.lastWatchPoint

        val highSampling = withinTimeout(data.lastHighSamplingData, HIGH_SAMPLING_DURATION) && timeoutReached(data.lastTransmittedToWearOS, 400)
        val timeout = timeoutReached(data.lastTransmittedToWearOS, DEFAULT_SYNC_INTERVAL)
        val speedChanged = oldWatchPoint != null && abs(oldWatchPoint.speed - newPoint.speed) >= 2

        if (oldPoint == null || highSampling || timeout || speedChanged) {
            Log.d("RPout-Logger", "Sending new GPS point to WearOS. High sampling = $highSampling | Timeout = $timeout | Speed changed = $speedChanged")
            data.lastWatchPoint = newPoint

            wearSynchronization.getNodes {
                sendToWearOs(it)
            }
        }

        // Update UI
        status.speed.floatValue = newPoint.speed
        status.distance.intValue = newPoint.totalDistance.roundToInt()

        // Stop foreground service when WearOS probably stopped tracking
        if(timeoutReached(lastWatchStatusReceived, SERVICE_TIMEOUT)) {
            logger.log("w", "Didn't receive an watch status within the last 5 minutes. Terminating location tracking")
            stopService()
        }
    }

    fun sendToWearOs(nodes: Set<Node>) {
        val pointsToSync = mutableListOf<AndroidGpsPoint>()

        synchronized(this) {
            // To not burst data points to WearOS
            if (!timeoutReached(data.lastTransmittedToWearOS, 200)) {
                return
            }

            data.lastTransmittedToWearOS = getCurrentTime()

            pointsToSync.addAll(data.unsyncedPoints)
            data.unsyncedPoints.clear()
        }

        try {
            val json = Gson().toJson(AndroidGpsData(points = pointsToSync))
            wearSynchronization.sendTextMessageToNodes(nodes, MessageType.WORKOUT_GPS_DATA, json) {}
        } catch (ex: Exception) {
            logger.log("e", ex, "Failed to send points to WearOS")
        }
    }

    private fun getCurrentTime(): Long  = System.currentTimeMillis()
    private fun timeoutReached(check: Long, timeout: Int): Boolean {
        val now = getCurrentTime()

        return (now - check) > timeout
    }

    private fun withinTimeout(check: Long, timeout: Int): Boolean {
        val now = getCurrentTime()

        return (now - check) < timeout
    }

    /**
     * Checks if the received GPS point is accurate enough for tracking it
     */
    private fun isPointAccurate(point: Location): Boolean {
        if(point.accuracy < 12) {
            return true
        }

        // Allow a longer timeout when GPS tracking was lately started
        if(withinTimeout(lastTrackingRestarted, 15_000)) {
            return point.accuracy < 15
        }

        val lastPoint = data.lastPoint ?: return true
        return timeoutReached(lastPoint.unixTime, 8_000)
    }

    private fun resetValues() {
        status.durationCheckpoint.longValue = 0
        status.duration.longValue = 0

        status.distance.intValue = 0
        status.heartRate.intValue = 0
        status.heartRateAv.intValue = 0
        status.speed.floatValue = 0f

        synchronized(this) {
            data.lastPoint = null
            data.unsyncedPoints.clear()
            data.lastDistancePoint = null
            lastTrackingRestarted = getCurrentTime()
        }
    }

}
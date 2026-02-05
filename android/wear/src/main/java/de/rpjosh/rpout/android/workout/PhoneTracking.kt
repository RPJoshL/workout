package de.rpjosh.rpout.android.workout

import android.content.Context
import android.os.VibrationEffect
import android.os.Vibrator
import android.widget.Toast
import androidx.health.services.client.data.ExerciseUpdate
import com.google.android.gms.wearable.Node
import com.google.gson.Gson
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.services.AndroidSynchronization
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.models.AndroidGpsData
import de.rpjosh.rpout.android.shared.models.AndroidGpsPoint
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
import de.rpjosh.rpout.android.shared.models.WorkoutStatus
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.models.WorkoutWatchStatus
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.Tr
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import java.time.Duration
import kotlin.math.abs
import kotlin.math.roundToInt
import kotlin.math.roundToLong

class PhoneTracking(
    type: WorkoutType,
    private val manager: WorkoutManager
) {

    companion object {
        /** After which minimum duration (ms) a new point should be send to android */
        const val SEND_THRESHOLD = 1500
    }

    /** Weather GPS is tracked by phone */
    var enabled: Boolean = type.usePhoneGPS
        private set

    /** If the phone was connected while the workout was started */
    var initialPhoneConnected: Boolean = true

    private val androidSynchronization: AndroidSynchronization = Singleton.appController.injection.inject(AndroidSynchronization::class.java, null, false)
    private val logger = Singleton.appController.injection.inject(Logger::class.java, arrayOf("PhoneTracking"), false)

    private var lastState = WorkoutStatus.PREPARE
    private var lastDataSend = System.currentTimeMillis()

    /** GPS points received from android */
    private var lastPoint: AndroidGpsPoint? = null
    /** If the high sampling interval should be used */
    private var highSamplingInterval = false

    private var workoutType: WorkoutType? = null

    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    private var node: Node? = null

    /** Called when the user changes the settings in preparation phase */
    fun settingUpdates(usePhoneGPS: Boolean) {
        if (usePhoneGPS == enabled) {
            return
        }

        enabled = usePhoneGPS
        if (usePhoneGPS) {
            checkPhoneConnected(RPout.getAppContext())
            sendStatus(WorkoutWatchStatus(
                trackingStatus = WorkoutStatus.PREPARE,
            ))
        } else {
            // Stop in case its still started
            sendStatus(WorkoutWatchStatus(
                trackingStatus = WorkoutStatus.STOP,
            ))
        }

        // Restart workout preparing so watch has the correct state
        scope.launch {
            manager.reinitExercise(Singleton.appController.injection)
        }
    }

    /** Processes the receiving of a new watch data point */
    @Synchronized
    fun processNewPoint(duration: ExerciseUpdate.ActiveDurationCheckpoint, heartrate: Int, heartRateAv: Int) {
        if (!isEnabledForExercise()) return

        val now = System.currentTimeMillis()
        val diff = now - lastDataSend
        if (diff < SEND_THRESHOLD) {
            return
        }

        lastDataSend = now
        sendStatus(WorkoutWatchStatus(
            trackingStatus = lastState,
            duration = duration.activeDuration.toMillis(),
            durationTimestamp = duration.time.toEpochMilli(),
            heartRate = heartrate,
            heartRateAv = heartRateAv,
            activityType = workoutType?.id,
        ))
    }

    @Synchronized
    fun onAndroidDataReceived(data: AndroidGpsData) {
        val last = data.points.lastOrNull() ?: return
        lastPoint = last

        // Notify user that GPS of the phone is established
        if (lastState == WorkoutStatus.PREPARE) {
            manager.handleGpsAvailability(RPout.getAppContext(), true)
            return
        } else {
            manager.handleGpsAvailability(RPout.getAppContext(), true)
        }

        manager.modifyPoints { gpsPoints ->
            data.points.forEach {
                getClosestPoint(it.unixTime, -2000, gpsPoints)?.let { p ->
                    val closerDiff = abs((p.unixTime * 1000) - it.unixTime) < abs((p.unixTime * 1000) - p.refUnixTime)

                    if (p.latitude == 0.0f || closerDiff) {
                        p.speed =  (1000 / it.speed).roundToInt()
                        p.totalDistance = it.totalDistance.roundToInt()
                        p.latitude = it.latitude.toFloat()
                        p.longitude = it.longitude.toFloat()
                        p.refUnixTime = it.unixTime

                        // Prefer elevation from on device sensor (barometer). It's much more precise.
                        // But we cannot check for healthSupportedCapabilities here (barometer is disabled
                        // when GPS is disabled?). This can lead to a very high altitude mismatch when GPS points are missing
                        if (p.elevation == 0) {
                            p.elevation = it.altitude.roundToInt()
                        }
                    }
                }
            }
        }

        // Update UI elements
        manager.workoutData.distance.setValue(last.totalDistance.toInt(), Duration.ofSeconds(TimeHelper.getBootTimeFromUnixTime(last.unixTime)))
        manager.workoutData.speed.setValue(last.speed.toDouble(), Duration.ofSeconds(TimeHelper.getBootTimeFromUnixTime(last.unixTime)))
    }

    fun getClosestPoint(
        unixTimeMillis: Long, allowedNegativeOffset: Long, points: MutableList<GpsWorkoutPoint>,
        allowedPositiveOffset: Long = (allowedNegativeOffset.toFloat() * 0.4).roundToLong() * -1
    ): GpsWorkoutPoint? {
        var rtc: GpsWorkoutPoint? = null

        points.forEach {
            val diff = unixTimeMillis - (it.unixTime * 1000)
            val closestDiff = abs(unixTimeMillis - ((rtc?.unixTime ?: 0) * 1000))

            // We try to prefer points from the past to not have diffs between -2 and +2
            val earlier = diff in allowedNegativeOffset..0
            val later = diff in 0..allowedPositiveOffset
            if ((earlier || later) && abs(diff) < closestDiff) {
                rtc = it
            }
        }

        return rtc
    }

    fun prepareExercise(context: Context, type: WorkoutType) {
        // We also use the phones connection state. We don't expect that the phone is available yet
        if (!isEnabledForExercise()) return

        lastState = WorkoutStatus.PREPARE
        workoutType = type
        checkPhoneConnected(context)
        sendStatus(WorkoutWatchStatus(
            trackingStatus = WorkoutStatus.PREPARE,
            activityType = type.id
        ))
    }

    fun startExercise(type: WorkoutType) {
        if (!isEnabledForExercise()) return

        logger.log("d", "Using phones GPS for workout tracking")

        workoutType = type
        lastState = WorkoutStatus.RUNNING
        sendStatus(WorkoutWatchStatus(
            trackingStatus = WorkoutStatus.RUNNING,
            activityType = type.id
        ))
    }

    fun pauseExercise() {
        if (!isEnabledForExercise()) return

        lastState = WorkoutStatus.PAUSE
        sendStatus(WorkoutWatchStatus(
            trackingStatus = WorkoutStatus.PAUSE,
        ))
    }

    fun stopExercise() {
        if (!isEnabledForExercise()) return

        lastState = WorkoutStatus.STOP
        sendStatus(WorkoutWatchStatus(
            trackingStatus = WorkoutStatus.STOP,
        ))
    }

    fun resumeExercise() {
        if (!isEnabledForExercise()) return

        lastState = WorkoutStatus.RUNNING
        sendStatus(WorkoutWatchStatus(
            trackingStatus = WorkoutStatus.RUNNING,
        ))
    }

    /** Validates that the phone is connected. If not, the status is updated accordingly and a toast will be displayed to the user */
    private fun checkPhoneConnected(context: Context) {
        androidSynchronization.getNodes {
            if (it.isEmpty() || !it.first().isNearby) {
                // Vibrate so the user is looking at the device
                val vibrator = context.getSystemService(Vibrator::class.java)
                vibrator.vibrate(VibrationEffect.createWaveform(longArrayOf(0, 400, 100, 400, 100, 400), -1))

                Toast.makeText(context, Tr.get("workout_phoneNotConnected"), Toast.LENGTH_LONG).show()
                logger.log("w", "Phone is not connected for GPS tracking. Using device GPS")
                initialPhoneConnected = false

                // Restart workout preparing so watch has the correct state
                scope.launch {
                    manager.reinitExercise(Singleton.appController.injection)
                }
            } else {
                initialPhoneConnected = true
                node = it.first()
            }
        }
    }

    fun onAndroidStatusRequest(status: String) {
        when(status) {
            "resume" -> scope.launch { manager.resume() }
            "pause" -> scope.launch { manager.pause() }
            "stop" -> scope.launch { manager.stop(true) }
            else -> {
                logger.log("w", "Received unknown workout request from android: $status")
            }
        }
    }

    /**
     * Updates the workout summary with the last values received from android
     */
    fun updateWorkoutSummary(summary: WorkoutSummary) {
        if (!isEnabledForExercise()) return
        val last = lastPoint ?: return

        summary.distance = last.totalDistance.toInt()
        summary.speedAv = last.speed.roundToInt()
    }

    fun enableHighSamplingInterval() {
        if(!isEnabledForExercise()) return

        highSamplingInterval = true
        sendStatus(WorkoutWatchStatus(
            trackingStatus = lastState,
        ))
    }

    fun disableHighSamplingInterval() {
        if (!isEnabledForExercise()) return

        highSamplingInterval = false
    }

    private fun sendStatus(watchStatus: WorkoutWatchStatus) {
        // Modify status for high sampling interval
        if (highSamplingInterval && watchStatus.trackingStatus == WorkoutStatus.RUNNING) {
            watchStatus.trackingStatus = WorkoutStatus.HIGH_SAMPLING
        }

        node?.let {
            val statusStr = Gson().toJson(watchStatus)
            androidSynchronization.sendTextMessageToNode(it, MessageType.WORKOUT_STATUS_DATA, statusStr, onSuccess = {})
        }
    }

    fun isEnabledForExercise(): Boolean {
        return enabled && initialPhoneConnected
    }

}
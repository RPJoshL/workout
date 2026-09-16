package de.rpjosh.rpout.android.workout.types

import android.content.Context
import android.content.Intent
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import androidx.compose.runtime.MutableIntState
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.mutableDoubleStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.health.services.client.data.DataType
import androidx.health.services.client.data.ExerciseUpdate
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.main.WorkoutTrackingActivity
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.workout.WorkoutManager
import java.util.concurrent.ConcurrentLinkedQueue

// Raw data received from the sensor
data class RawAcceleration(
    val timestamp: Long, // nanoseconds since system boot
    val x: Float, // m/s²
    val y: Float, // m/s²
    val z: Float // m/s²
)

data class FoilingSessionUIData(
    /** Number of pump foil sessions within this workout */
    val count: MutableIntState = mutableIntStateOf(0),
    /** Weather a session is currently active */
    val isActive: MutableState<Boolean> = mutableStateOf(false),
    /** Distance travelled in meter for this session */
    val distance: MutableState<Double> = mutableDoubleStateOf(0.0),
    /** Checkpoint for duration tracking */
    val checkpoint: MutableState<ExerciseUpdate.ActiveDurationCheckpoint> = mutableStateOf(ExerciseUpdate.ActiveDurationCheckpoint(java.time.Instant.EPOCH, java.time.Duration.ZERO)),
    /** The duration (in seconds) of the last run */
    val lastSessionDuration: MutableState<Long> = mutableLongStateOf(0),
    /** Total active session duration */
    val totalDuration: MutableState<Long> = mutableLongStateOf(0),
)

class FoilingMetrics(
    context: Context,
    val type: WorkoutType,
    val manager: WorkoutManager
): TypeTracker, SensorEventListener {

    companion object {
        /** 40ms sampling interval. Each pump should give use ~8 raw values */
        const val ACCEL_HZ = 25
        /** Scale factor of 2048 (int16) == 1 g */
        const val ACCEL_SCALE = 2048.0
        const val G = 9.80665

        /** Minimum initial speed to count as foiling (m/s) */
        const val MIN_FOILING_START_SPEED = 3.3 // 12 km/h
        /** Duration (s) of minimum foiling speed after which a session is counted */
        const val MIN_FOILING_DURATION = 9
        /** Minimum continues speed to count as foiling (m/s) */
        const val MIN_FOILING_SPEED = 3.0 // 11 km/h
        /** Duration in seconds which may be below minimum foiling speed */
        const val FOILING_THRESHOLD_END = 6
    }

    @Inject(parameters = ["FolingMetrics"]) private lateinit var logger: Logger

    private var sensorManager = context.getSystemService(Context.SENSOR_SERVICE) as SensorManager
    private var accelerometer = sensorManager.getDefaultSensor(Sensor.TYPE_ACCELEROMETER)

    private val accelerationData = ConcurrentLinkedQueue<RawAcceleration>()

    @Volatile private var sensorRegistered = false

    /** Timestamp speed felt below minimum foiling speed / started to foil */
    private var thresholdTime: Long? = null
    /** Total distance in meters at which the session started */
    private var sessionStartDistance: Int = 0
    /** Unix time (seconds) the session started */
    private var sessionStartTime: Long = 0

    init {
        Singleton.appController.injection.inject(FoilingMetrics::class.java, null, false, this)
        if(accelerometer == null) {
            logger.log("w", "No accelerometer found for tracking detailed pump foiling metrics")
        }
    }

    @Synchronized
    private fun registerSensor() {
        if(!type.useHighSamplingInterval) return
        if(sensorRegistered) return

        accelerometer.let {
            sensorManager.registerListener(this, it, 1_000_000 / ACCEL_HZ)
            sensorRegistered = true
        }
    }

    @Synchronized
    private fun deregisterSensor() {
        if(!type.useHighSamplingInterval) return
        if(!sensorRegistered) return

        sensorManager.unregisterListener(this)
        sensorRegistered = false
    }

    override fun onStart(context: Context) {
        manager.foilingData = FoilingSessionUIData()
        registerSensor()
    }

    override fun onPause() {
        deregisterSensor()
    }

    override fun onResume() {
        registerSensor()
    }

    override fun onEnd() {
        deregisterSensor()
    }

    override fun onProcessMetrics(workout: GpsWorkout, summary: WorkoutSummary, update: ExerciseUpdate) {
        val speedMetrics = update.latestMetrics.getData(DataType.SPEED)

        if (speedMetrics.isNotEmpty()) {
            speedMetrics.forEach { sample ->
                val unixTime = TimeHelper.getUnixTimeFromBootTime(sample.timeDurationFromBoot)
                recognizeSession(sample.value, manager.workoutData.distance.value.value, unixTime)
            }
        }

        // 2. Acceleration Data Processing
        if (workout.points.isEmpty()) return

        val points = workout.points
        var pointIndex = 0

        var accel = accelerationData.peek()
        while (accel != null) {
            val accelUnixMs = TimeHelper.getUnixTimeFromBootTimeMillis(java.time.Duration.ofNanos(accel.timestamp))
            val accelUnixSec = accelUnixMs / 1000

            // 1. If the acceleration data is older than our first point, we can remove it.
            // It can never be matched
            if (accelUnixSec < points.first().unixTime) {
                accelerationData.poll()
                accel = accelerationData.peek()
                continue
            }

            // 2. Find the point where this acceleration data belongs to.
            // Because both data sets are sorted, we can use a linear scan (pointIndex)
            while (pointIndex < points.size - 1 && accelUnixSec >= points[pointIndex + 1].unixTime) {
                pointIndex++
            }

            if (pointIndex < points.size - 1) {
                val p1 = points[pointIndex]
                val p2 = points[pointIndex + 1]

                if (accelUnixSec >= p1.unixTime && accelUnixSec < p2.unixTime) {
                    val relTimeMs = (accelUnixMs - (p1.unixTime * 1000)).toShort()
                    val packed = shortArrayOf(
                        relTimeMs,
                        toAccelI16(accel.x),
                        toAccelI16(accel.y),
                        toAccelI16(accel.z)
                    )

                    p1.acceleration = if (p1.acceleration == null) packed else p1.acceleration!! + packed
                    accelerationData.poll()
                    accel = accelerationData.peek()
                } else {
                    // This sample is in the future relative to our currently processed points.
                    // We stop here and wait for more GPS points in the next update.
                    break
                }
            } else {
                // Also in the future relative to our last point.
                break
            }
        }

        // Buffer safety: If matching failed for a long time (e.g. GPS off), clear buffer to prevent memmory errors
        if (accelerationData.size > 5_000) {
            logger.log("w", "Clearing acceleration buffer because it grew too large (${accelerationData.size} samples)")
            accelerationData.clear()
        }
    }

    private fun recognizeSession(currentSpeed: Double, currentDistance: Int, now: Long) {
        val uiData = manager.foilingData

        if (!uiData.isActive.value) {
            // Not foiling
            if(currentSpeed < MIN_FOILING_START_SPEED) {
                return
            }

            if(thresholdTime == null) {
                thresholdTime = now
            }

            val duration = now - thresholdTime!!
            if (duration > MIN_FOILING_DURATION) {
                uiData.isActive.value = true
                uiData.count.intValue++
                uiData.checkpoint.value = ExerciseUpdate.ActiveDurationCheckpoint(java.time.Instant.ofEpochSecond(thresholdTime!!), java.time.Duration.ZERO)

                uiData.distance.value = 0.0
                sessionStartDistance = currentDistance
                sessionStartTime = thresholdTime!!

                thresholdTime = null
                switchTab(2)
            }
        } else {
            uiData.distance.value = (currentDistance - sessionStartDistance).toDouble()

            if(currentSpeed > MIN_FOILING_SPEED) {
                thresholdTime = null
                return
            }

            if (thresholdTime == null) {
                thresholdTime = now
            } else if (now - thresholdTime!! >= FOILING_THRESHOLD_END) {
                uiData.isActive.value = false
                thresholdTime = null
                switchTab(1)

                val sessionDuration = now - sessionStartTime - FOILING_THRESHOLD_END
                uiData.totalDuration.value += sessionDuration
                uiData.lastSessionDuration.value = sessionDuration
            }
        }
    }

    private fun switchTab(page: Int) {
        val intent = Intent(RPout.getAppContext(), WorkoutTrackingActivity::class.java).apply {
            putExtra(WorkoutTrackingActivity.INTENT_EXTRA_PAGE, page)
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_SINGLE_TOP)
        }
        RPout.getAppContext().startActivity(intent)
    }

    override fun onAccuracyChanged(sensor: Sensor?, accuracy: Int) {}

    override fun onSensorChanged(event: SensorEvent?) {
        if(event === null) return

        val values = event.values
        if (values.size < 3) {
            logger.log("w", "Got less than three data points from acceleration sensor: " + values.size)
            return
        }

        accelerationData.add(RawAcceleration(
            event.timestamp,
            x = values[0],
            y = values[1],
            z = values[2]
        ))
    }

    private fun toAccelI16(valueMs2: Float): Short {
        return (valueMs2 / G * ACCEL_SCALE)
            .coerceIn(Short.MIN_VALUE.toDouble(), Short.MAX_VALUE.toDouble())
            .toInt()
            .toShort()
    }

}
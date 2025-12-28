package de.rpjosh.rpout.android.shared.workout

import android.location.Location
import android.os.SystemClock
import androidx.compose.ui.graphics.Color
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.mutableDoubleStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.health.services.client.data.CumulativeDataPoint
import androidx.health.services.client.data.ExerciseState
import androidx.health.services.client.data.ExerciseUpdate
import androidx.health.services.client.data.LocationData
import androidx.health.services.client.data.SampleDataPoint
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.models.ActivityType
import de.rpjosh.rpout.android.shared.models.HeartRateZone
import java.time.Duration
import java.time.Instant
import java.util.Locale
import kotlin.math.abs
import kotlin.math.roundToInt

/** A single workout point */
open class WorkoutPoint<TypeValue> (
    /** Duration since the device was started */
    var time: Duration,
    val value: MutableState<TypeValue>
) {
    fun setValue(v: TypeValue) {
        value.value = v
        time = Duration.ofMillis(SystemClock.uptimeMillis())
    }

    fun setValue(v: TypeValue, time: Duration) {
        value.value = v
        this.time = time
    }

    /** Returns whether the currently tracked data point is within the specified offset seconds  */
    fun isInLast(seconds: Int, unixTimeSeconds: Long = System.currentTimeMillis() / 1000): Boolean {
        // Convert last sensor time to unix timestamp
        val sensorTime = TimeHelper.getUnixTimeFromBootTime(time)

        return abs(unixTimeSeconds - sensorTime) <= seconds
    }
}

/** A color workout type */
class ColoredWorkoutPoint<TypeValue>(
    time: Duration,
    value: MutableState<TypeValue>,

    /** Color of this data point */
    val color: MutableState<Color>
): WorkoutPoint<TypeValue>(time, value) {
    fun setValue(v: TypeValue, time: Duration, color: Color) {
        super.setValue(v, time)
        this.color.value = color
    }
}

/** Root workout struct that contains different data for a workout */
class Workout {

    companion object {

        /**
         * Formats the internal speed (in meters / second) into the speed unit
         * of the workout type. That's either "min / km" or "km / h"
         */
        fun getFormattedSpeed(typeId: Long, metersPerSecond: Double): String {
            // Return in minutes / minute for
            val minutesPerMinute = arrayListOf(ActivityType.TYPE_HIKING, ActivityType.TYPE_RUNNING)
            if (ActivityType.fromInt(typeId.toInt()) in minutesPerMinute) {
                if (metersPerSecond <= 0.1) return "--:--"

                val secondsPerKilometer = 1000 / metersPerSecond
                val minutesPerKilometer = secondsPerKilometer / 60
                val minutes = minutesPerKilometer.toInt()
                val seconds = ((minutesPerKilometer - minutes) * 60).toInt()

                return String.format(Locale.ENGLISH, "%d:%02d", minutes, seconds)
            }

            val kmPerHour = metersPerSecond * 3.6
            return String.format(Locale.ENGLISH, "%.2f", kmPerHour)
        }

    }

    val heartRate = ColoredWorkoutPoint(Duration.ofMillis(0), mutableIntStateOf(0), mutableStateOf(Color.White))
    val location = WorkoutPoint(Duration.ofMillis(0), mutableStateOf(LocationData(0.0, 0.0)))
    val elevation = WorkoutPoint(Duration.ofMillis(0), mutableIntStateOf(0))
    val distance = WorkoutPoint(Duration.ofMillis(0), mutableIntStateOf(0))
    val speed = WorkoutPoint(Duration.ofMillis(0), mutableDoubleStateOf(0.0))

    val activeDuration = mutableStateOf( ExerciseUpdate.ActiveDurationCheckpoint(time = Instant.now(), activeDuration = Duration.ofMillis(0)) )
    val exerciseState = mutableStateOf( ExerciseState.PREPARING )

    fun setHeartRate(v: SampleDataPoint<Double>) {
        heartRate.setValue(v.value.toInt(), v.timeDurationFromBoot, HeartRateZone.getZone(v.value.toInt()).color)
    }

    fun setLocation(v:  SampleDataPoint<LocationData>) {
        location.setValue(v.value, v.timeDurationFromBoot)
    }

    fun setLocationOneTime(v: Location) {
        // Transform to location data
        val locationData = LocationData(v.latitude, v.longitude, v.altitude, v.bearing.toDouble())
        location.setValue(locationData, Duration.ofMillis(TimeHelper.getBootTimeFromUnixTime(v.time / 1000)))
    }

    fun setElevation(v: SampleDataPoint<Double>) {
        elevation.setValue(v.value.roundToInt(), v.timeDurationFromBoot)
    }

    fun setDistance(v: CumulativeDataPoint<Double>) {
        distance.setValue(v.total.roundToInt(), Duration.ofMillis(v.end.toEpochMilli() - SystemClock.elapsedRealtime()))
    }

    fun setSpeed(v: SampleDataPoint<Double>) {
        speed.setValue(v.value, v.timeDurationFromBoot)
    }

    /**
     * Formats the internal speed (in meters / second) into the speed unit
     * of the workout type. That's either "min / km" or "km / h"
     */
    fun getFormattedSpeed(typeId: Long): String {
        return getFormattedSpeed(typeId, speed.value.value)
    }
}
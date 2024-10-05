package de.rpjosh.rpout.android.shared.models

import androidx.compose.ui.graphics.Color
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.workout.Workout
import java.time.Duration
import java.time.LocalDateTime
import java.util.Collections.addAll
import java.util.Locale

/** HeartRateZone is a single zone value starting by min (including)  */
open class HeartRateZone(
    val id: Int,
    val colorString: String,
    val min: Int
) {

    val color: Color = Color(android.graphics.Color.parseColor(colorString))

    companion object {

        val zones = arrayOf(
            HeartRateZone(0, "#ff8a80", 0),
            HeartRateZone(1, "#3370ff", 97),    // Original: #2862ff
            HeartRateZone(2, "#00cee9", 116),
            HeartRateZone(3, "#65dd19", 135),
            HeartRateZone(4, "#ff6d01", 154),
            HeartRateZone(5, "#a32bff", 174),   // Original: #aa00ff
        )

        /**
         * Returns the zone for the provided heart rate value
         */
        fun getZone(heartRate: Int): HeartRateZone {
            return zones.findLast { heartRate >= it.min } ?: zones[0]
        }

    }

    /**
     * Returns the translated name of this zone
     */
    fun getName(): String {
        return Tr.get("heartRateZone_$id")
    }
}

data class HeartRateZoneStat(
    var duration: Duration = Duration.ofSeconds(0),
    val zone: HeartRateZone
): HeartRateZone(zone.id, zone.colorString, zone.min) {

    /**
     * Returns the formatted duration for this zone duration
     */
    fun getDuration(): String {
        if (android.os.Build.VERSION.SDK_INT >= 31) {
            var rtc = ""
            if (duration.toHours() > 0) rtc += "${duration.toHours()}:"
            rtc += String.format(Locale.ENGLISH, "%02d:%02d", duration.toMinutesPart(), duration.toSecondsPart())

            return rtc
        } else {
            return duration.toString()
        }

    }
}

data class WorkoutSummary(
    var id: Long = 0,
    var name: String = "",
    var userId: Long = 0,
    var typeId: Long = 0,
    var start: String = TimeHelper.fromClientToServer(LocalDateTime.now()),
    var end: String = TimeHelper.fromClientToServer(LocalDateTime.now()),
    var country: String = "",
    var city: String = "",
    var cityId: String = "",
    var duration: Int = 0,
    var calories: Int = 0,
    var distance: Int = 0,
    /** Average traveling speed in sec/km */
    var speedAv: Int = 0,
    var elevationUp: Int = 0,
    var elevationDown: Int = 0,
    var heartRateAv: Int = 0,
    var heartRateMax: Int = 0,
    var pai: Int = 0,
    var steps: Int = 0,

    var typeAccentColor: Color = Color.White,
) {

    var heartRateZones = getHeartRateZoneStats(listOf())

    override fun toString(): String {
        return "Duration = $duration | Calories = $calories | Distance = $distance | Speed = $speedAv | Up = $elevationUp | Down = $elevationDown | HeartAvg = $heartRateAv | HeartMax  = $heartRateMax | Steps = $steps"
    }

    /**
     * Returns the formatted duration for this zone duration
     */
    fun getDuration(): String {
        val duration = Duration.ofSeconds(duration.toLong())
        if (android.os.Build.VERSION.SDK_INT >= 31) {
            var rtc = ""
            if (duration.toHours() > 0) rtc += "${duration.toHours()}:"
            rtc += String.format(Locale.ENGLISH, "%02d:%02d", duration.toMinutesPart(), duration.toSecondsPart())

            return rtc
        } else {
            return duration.toString()
        }

    }

    /**
     * Formats the average traveling speed into "km/h" or "min/km" based on the
     * provided workout type
     */
    fun getFormattedSpeed(typeId: Long): String {
        return Workout.getFormattedSpeed(typeId, if(speedAv == 0) 0.0 else 1000.0 / speedAv)
    }

    /** Calculates the times spend in a single heart rate zone from the provided points  */
    fun getHeartRateZoneStats(points: List<GpsWorkoutPoint>): List<HeartRateZoneStat> {
        val zones = arrayListOf<HeartRateZoneStat>()
        zones.addAll(HeartRateZone.zones.map { HeartRateZoneStat(zone = it) })

        // No "real" data received
        if (points.size < 2) return zones

        // Get durations
        var lastPoint = points.first()
        for(i in 1 until points.size step 1) {
            val p = points[i]

            // Do not track pauses
            val diff = p.unixTime - lastPoint.unixTime
            if (diff > 36) {
                lastPoint = p
                continue
            }

            // Get zone of last point and increment with it
            zones[HeartRateZone.getZone(lastPoint.heartRate).id].duration = zones[HeartRateZone.getZone(lastPoint.heartRate).id].duration.plusSeconds(diff)
            lastPoint = p
        }

        return zones
    }

}
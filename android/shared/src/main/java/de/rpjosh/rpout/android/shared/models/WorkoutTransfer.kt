package de.rpjosh.rpout.android.shared.models

enum class WorkoutStatus(val status: String) {
    PREPARE("prepare"),
    START("start"),
    STOP("stop"),
    PAUSE("pause"),
    RESUME("resume"),
    RUNNING("running"),

    /** Indicates that the workout is started the screen of the watch is turned on -> high sampling interval for data receivement.
     * This status automatically times out after 10 seconds */
    HIGH_SAMPLING("high-sampling");

    companion object {
        fun fromString(status: String): WorkoutStatus? {
            return entries.find { it.status == status }
        }
    }
}

/** Data that is tracked from the watch. Note: the watch owns the workout! */
data class WorkoutWatchStatus(
    /** The activity type of the workout (ID) */
    val activityType: Long? = null,
    /** The current tracking status */
    var trackingStatus: WorkoutStatus,
    /** Unix timestamp in seconds this workout status was created */
    val timestamp: Long = System.currentTimeMillis() / 1000,
    /** Workout duration (without any pauses) at tha durationTimestamp */
    val duration: Long = 0,
    /** Timestamp the specified duration was */
    val durationTimestamp: Long = 0,
    /** Currently measured heart rate */
    val heartRate: Int = 0,
    /** Average heart rate */
    val heartRateAv: Int = 0,
)

data class AndroidGpsPoint(
    /** In milliseconds */
    val unixTime: Long,
    val latitude: Double,
    val longitude: Double,
    val altitude: Double,
    /** Current travelling speed in meters per second */
    val speed: Float,
    /** Total distance in meters */
    val totalDistance: Double,
)

/** GPS data that is send from android to the watch */
data class AndroidGpsData(
    /** Data points that were tracked since the last message. When the watch is not available or unable to process this points,
     * they are silently dropped. We don't implement an acknowledge system! */
    val points: List<AndroidGpsPoint>,
)
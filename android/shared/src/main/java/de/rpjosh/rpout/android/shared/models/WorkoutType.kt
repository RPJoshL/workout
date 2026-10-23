package de.rpjosh.rpout.android.shared.models

import androidx.compose.runtime.Immutable
import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey
import de.rpjosh.rpout.android.shared.services.TranslationService.Language

/** Action to perform when the primary gesture (double pinch) was detected while tracking a workout */
enum class PrimaryGestureActionTracking(val id: Int) {
    DISABLED(0),
    PAUSE_RESUME(1),
    CHANGE_PAGE(2);

    companion object {
        fun fromId(id: Int): PrimaryGestureActionTracking =
            entries.firstOrNull { it.id == id } ?: DISABLED
    }
}

/** Action to perform when the dismiss gesture was detected while tracking a workout */
enum class DismissGestureActionTracking(val id: Int) {
    DISABLED(0),
    PAUSE_RESUME(1),
    STOP(2);

    companion object {
        fun fromId(id: Int): DismissGestureActionTracking =
            entries.firstOrNull { it.id == id } ?: DISABLED
    }
}

enum class GPSMode(val id: Int) {
    DEFAULT(0),
    USE_PHONE(1),
    DISABLED(2);

    companion object {
        fun fromId(id: Int): GPSMode =
            GPSMode.entries.firstOrNull { it.id == id } ?: DISABLED
    }
}

@Immutable
@Entity(tableName = "workout_type")
data class WorkoutType(
    // Properties from the server
    @PrimaryKey val id: Long,
    val nameDe: String = "",
    val nameEn: String = "",
    val tagDark: String = "",
    val tagWhite: String = "",
    val icon: String,

    /** Whether to update data showed in ambient mode directly and every second (this will drain the battery more)  */
    @ColumnInfo(defaultValue = "0")
    var liveUpdates: Boolean = false,
    /** Whether to track more accurate data (3s vs 6s sampling interval) and enable extended tracking for specific types */
    @ColumnInfo(defaultValue = "0")
    var useHighSamplingInterval: Boolean = false,
    /** Whether to register a tilt to wake sensor for waking up the screen */
    @ColumnInfo(defaultValue = "0")
    var useTiltToWake: Boolean = false,
    @ColumnInfo(defaultValue = "0")
    var primaryGestureAction: Int = 0,
    @ColumnInfo(defaultValue = "0")
    var dismissGestureAction: Int = 0,
    @ColumnInfo(defaultValue = "0")
    var gpsMode: Int = 0,
) {
    /** Copies / applies all application settings to the provided workout type  */
    fun copySettingsTo(a: WorkoutType) {
        a.liveUpdates = liveUpdates
        a.useHighSamplingInterval = useHighSamplingInterval
        a.useTiltToWake = useTiltToWake
        a.primaryGestureAction = primaryGestureAction
        a.dismissGestureAction = dismissGestureAction
        a.gpsMode = gpsMode
    }

    /** Returns the translated name for this type */
    fun getName(language: Language): String {
        return when(language) {
            Language.GERMAN -> nameDe
            Language.ENGLISH -> nameEn
            else -> nameEn
        }
    }

    /** Whether GPS tracking is useful (and enabled) for this exercise type */
    fun shouldTrackGPS(): Boolean {
        return isGPSTrackingSupported() && getGPSMode() != GPSMode.DISABLED
    }

    /** Weather GPS tracking is supported / useful for this workout type */
    fun isGPSTrackingSupported(): Boolean {
        val noGps = arrayListOf(ActivityType.TYPE_VOLLEYBALL, ActivityType.TYPE_STRENGTH_TRAINING)
        return id.toInt() !in noGps.map { it.ordinal }
    }

    /** Whether this workout type is a water sport activity. Features like reconnect sounds are disabled */
    fun isWaterActivity(): Boolean {
        val types = arrayListOf(ActivityType.TYPE_PUMP_FOILING, ActivityType.TYPE_SAILING, ActivityType.TYPE_SWIMMING, ActivityType.TYPE_SURFEN)

        return id.toInt() in types.map { it.ordinal }
    }

    fun getPrimaryGestureAction(): PrimaryGestureActionTracking =
        PrimaryGestureActionTracking.fromId(primaryGestureAction)

    fun getDismissGestureAction(): DismissGestureActionTracking =
        DismissGestureActionTracking.fromId(dismissGestureAction)

    fun getGPSMode(): GPSMode = GPSMode.fromId(gpsMode)
}

enum class ActivityType {
    TYPE_UNKNOWN,
    TYPE_HIKING,
    TYPE_RUNNING,
    TYPE_SURFEN,
    TYPE_SAILING,
    TYPE_SNOWBOARDING,
    TYPE_SWIMMING,
    TYPE_CYCLING,
    TYPE_SKATEBOARDING,
    TYPE_VOLLEYBALL,
    TYPE_PUMP_FOILING,
    TYPE_STRENGTH_TRAINING;

    companion object {
        fun fromInt(value: Int) = entries.first { it.ordinal == value }
    }
}
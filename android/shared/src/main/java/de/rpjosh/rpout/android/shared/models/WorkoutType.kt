package de.rpjosh.rpout.android.shared.models

import androidx.compose.runtime.Immutable
import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey
import de.rpjosh.rpout.android.shared.services.TranslationService.Language

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

    /** Whether to not track the workout with GPS */
    @ColumnInfo(defaultValue = "0")
    var noGPS: Boolean = false,
    /** Whether to update data showed in ambient mode directly and every second (this will drain the battery more)  */
    @ColumnInfo(defaultValue = "0")
    var liveUpdates: Boolean = false

) {
    /** Copies / applies all application settings to the provided workout type  */
    fun copySettingsTo(a: WorkoutType) {
        a.noGPS = a.noGPS
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
        return isGPSTrackingSupported() && !noGPS
    }

    /** Weather GPS tracking is supported / useful for this workout type */
    fun isGPSTrackingSupported(): Boolean {
        val noGps = arrayListOf(ActivityType.TYPE_VOLLEYBALL, ActivityType.TYPE_STRENGTH_TRAINING)
        return id.toInt() !in noGps.map { it.ordinal }
    }

    /** Weather this workout type is a water sports. Features like reconnect sounds are disabled */
    fun isWaterActivity(): Boolean {
        val types = arrayListOf(ActivityType.TYPE_PUMP_FOILING, ActivityType.TYPE_SAILING, ActivityType.TYPE_SWIMMING, ActivityType.TYPE_SURFEN)

        return id.toInt() in types.map { it.ordinal }
    }
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
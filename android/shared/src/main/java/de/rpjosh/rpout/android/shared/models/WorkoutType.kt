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

    /** Whether to prefer / use the GPS signal from a connected smartphone */
    var preferSmartphoneGps: Boolean = false,
    /** Whether to update data showed in ambient mode directly and every second (this will drain the battery more)  */
    @ColumnInfo(defaultValue = "0")
    var liveUpdates: Boolean = false

) {
    /** Copies / applies all application settings to the provided workout type  */
    fun copySettingsTo(a: WorkoutType) {
        a.preferSmartphoneGps = preferSmartphoneGps
    }

    /** Returns the translated name for this type */
    fun getName(language: Language): String {
        return when(language) {
            Language.GERMAN -> nameDe
            Language.ENGLISH -> nameEn
            else -> nameEn
        }
    }

    /** Whether GPS tracking is useful for this exercise type */
    fun shouldTrackGPS(): Boolean {
        val noGps = arrayListOf(ActivityType.TYPE_VOLLEYBALL)

        return id.toInt() !in noGps.map { it.ordinal }
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
    TYPE_PUMP_FOILING;

    companion object {
        fun fromInt(value: Int) = entries.first { it.ordinal == value }
    }
}
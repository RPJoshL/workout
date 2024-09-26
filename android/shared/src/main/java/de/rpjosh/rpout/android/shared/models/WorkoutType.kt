package de.rpjosh.rpout.android.shared.models

import androidx.compose.runtime.Immutable
import androidx.room.Entity
import androidx.room.PrimaryKey
import de.rpjosh.rpout.android.shared.services.TranslationService.Language

@Immutable
@Entity(tableName = "workout_type")
data class WorkoutType(

    // Properties from the server
    @PrimaryKey val id: Long,
    val nameDe: String,
    val nameEn: String,
    val tagDark: String,
    val tagWhite: String,
    val icon: String,

    /** Whether to prefer / use the GPS signal from a connected smartphone */
    var preferSmartphoneGps: Boolean = false

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
}
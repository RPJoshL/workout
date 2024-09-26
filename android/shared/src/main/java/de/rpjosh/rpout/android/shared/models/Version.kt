package de.rpjosh.rpout.android.shared.models

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "version")
data class Version(

    /** This ID should ALWAYS be null */
    @PrimaryKey val id: Int = 0,

    /** Last app version when the workout types were updated */
    var typeVersion: String

)

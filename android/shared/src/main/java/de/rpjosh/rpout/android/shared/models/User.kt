package de.rpjosh.rpout.android.shared.models

import androidx.room.Entity
import androidx.room.PrimaryKey
import de.rpjosh.rpout.android.shared.services.Logger

@Entity(tableName = "user")
data class User(

    /** Unique ID of the user that belongs to the API key */
    @PrimaryKey val id: Long = 0,

    /** Full base URL (without /api/v1) of the server */
    val serverUrl: String,

    /** Username for the provided API-Key */
    val username: String,
    /** Raw API key value for authentication */
    val apikey: String,
    /** Unique ID of the API key */
    val apiKeyId: Long,

    /** Log level to log messages with */
    var logLevel: Int = Logger.LEVEL.INFO.value
)

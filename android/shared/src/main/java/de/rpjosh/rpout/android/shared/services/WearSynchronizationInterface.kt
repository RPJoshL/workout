package de.rpjosh.rpout.android.shared.services

import de.rpjosh.rpout.android.shared.inject.Inject

abstract class WearSynchronizationInterface {

    @Inject protected lateinit var responseViewInterface: ResponseViewInterface

    /**
     * Sends the provided text message to the other device (wearable -> android or android -> wearable)
     */
    abstract fun sendTextMessage(type: MessageType, message: String, onSuccess: () -> Unit)

}


/** Message types that are send / received from the android or wearable site */
enum class MessageType(var path: String) {
    /** Synchronizes the whole user model to the wearable site (including API-Keys, ...). A message of "DELETE" */
    SETTINGS("/settings"),

    /** Requests log messages from Wear OS. The app will receive a message of LOG_RESPONSE */
    LOG_REQUEST("/log/request"),
    /** Content of the log after a "LOG_REQUEST" */
    LOG_RESPONSE("/log/response"),

    /** Sync all data (like steps, metrics, workout data) on the WearOS side */
    SYNC_DATA("/sync/allData"),
    /** Sync all workout types on the WearOS side */
    SYNC_DATA_WORKOUT("/sync/workoutTypes"),

    /** The current workout status with the current heart rate  */
    WORKOUT_STATUS_DATA("/workout/status"),
    /** Is send from Android to indicate the current speed, distance and GPS data */
    WORKOUT_GPS_DATA("/workout/gps/data"),
    /** Update the workout status (resume, pause, stop) */
    WORKOUT_STATUS_UPDATE("/workout/status/update"),
}
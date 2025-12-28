package de.rpjosh.rpout.android.services

import android.content.Intent
import android.util.Log
import com.google.android.gms.wearable.MessageEvent
import com.google.android.gms.wearable.WearableListenerService
import com.google.gson.Gson
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.models.WorkoutWatchStatus
import de.rpjosh.rpout.android.shared.services.MessageType

class DataSyncListener: WearableListenerService() {

    companion object {
        const val TAG = "RPout-Logger"
    }

    override fun onMessageReceived(ev: MessageEvent) {
        super.onMessageReceived(ev)

        val type = MessageType.entries.find { m -> m.path == ev.path }
        Log.i(TAG, "Received message from wear os with path ${ev.path} and type ${type?.name ?: "Unknown"}")
        val data = String(ev.data)
        when (type) {
            null -> {}
            MessageType.WORKOUT_STATUS_DATA -> {
                try {
                    val status = Gson().fromJson(data, WorkoutWatchStatus::class.java)
                    distributeWorkoutStatus(status)
                } catch (ex: Exception) {
                    Singleton.getApp()?.sharedLogger?.log("w", ex,"Failed to parse status from watch")
                }
            }
            else -> {
                // Send message to all listener
                Singleton.sendMessageToWearMessageReceiver(type, data)
            }
        }
    }

    /** Distributes the watch status messages to all activities / foreground services */
    private fun distributeWorkoutStatus(status: WorkoutWatchStatus) {
        val serviceIntent = Intent(this, RealtimeLocationService::class.java).apply {
            action = status.trackingStatus.status
            putExtra(RealtimeLocationService.INTENT_HEART_RATE, status.heartRate)
            putExtra(RealtimeLocationService.INTENT_HEART_RATE_AV, status.heartRateAv)
            putExtra(RealtimeLocationService.INTENT_DURATION, status.duration)
            putExtra(RealtimeLocationService.INTENT_DURATION_CHECKPOINT, status.durationTimestamp)
            putExtra(RealtimeLocationService.INTENT_ACTIVITY_ID, status.activityType)
        }
        startForegroundService(serviceIntent)
    }

    override fun onDestroy() {
        super.onDestroy()

        Log.d(TAG, "Destroyed data sync listener")
    }

}
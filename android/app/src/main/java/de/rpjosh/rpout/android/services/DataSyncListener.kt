package de.rpjosh.rpout.android.services

import android.Manifest
import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.pm.PackageManager
import android.util.Log
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import com.google.android.gms.wearable.MessageEvent
import com.google.android.gms.wearable.WearableListenerService
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.models.User
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
            else -> {
                // Send message to all listener
                Singleton.sendMessageTOWearMessageReceiver(type, data)
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()

        Log.d(TAG, "Destroyed data sync listener")
    }

}
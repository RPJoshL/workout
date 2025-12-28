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
import com.google.gson.Gson
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.helper.VersionHelper
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.models.AndroidGpsData
import de.rpjosh.rpout.android.shared.models.User
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.tiles.PaiTile
import de.rpjosh.rpout.android.workout.WorkoutManager

class DataSyncListener: WearableListenerService() {

    companion object {
        const val TAG = "RPout-Logger"
    }

    override fun onMessageReceived(ev: MessageEvent) {
        super.onMessageReceived(ev)

        val type = MessageType.entries.find { m -> m.path == ev.path }
        Log.i(TAG, "Received message from android with path ${ev.path} and type ${type?.name ?: "Unknown"}")
        when (type) {

            MessageType.SETTINGS -> {
                // Initialize app controller
                val app = Singleton.getAppSec(true)
                val userController = app.injection.inject(UserController::class.java, null, false)
                app.sharedLogger.log("i", "Received update for user settings")

                // Delete or update / insert user settings
                app.stopAndroidServices()
                val data = String(ev.data)
                if (data == "DELETE") userController.setSettingsWearOs(null)
                else {
                    // Convert to user settings
                    try {
                        val user = Gson().fromJson(data, User::class.java)
                        userController.setSettingsWearOs(user)
                    } catch (ex: Exception) {
                        app.sharedLogger.log("e", ex, "Failed to convert message to user")
                        app.sharedLogger.log("d", "Message: $data")
                    }
                }
                app.startAndroidServices()
            }

            MessageType.LOG_REQUEST -> {
                val app = Singleton.getAppSec(true)
                val syncClient = app.injection.inject(AndroidSynchronization::class.java, null, false)
                app.sharedLogger.log("i", "Received request to send log messages")

                // Get log file
                val logFile = app.sharedLogger.getLogFile(600)
                val text = logFile.readText()
                logFile.delete()

                // Send log content to android
                syncClient.sendTextMessage(MessageType.LOG_RESPONSE, text) {}
            }

            MessageType.SYNC_DATA -> {
                val app = Singleton.getAppSec(true)
                val metricController = app.injection.inject(MetricController::class.java, null, false)
                val workoutController = app.injection.inject(WorkoutController::class.java, null, false)
                app.sharedLogger.log("i", "Received request to sync all data")

                // Sync all entities
                metricController.synchronizeSteps()
                workoutController.synchronizeWorkouts()
                metricController.synchronizePai()

                // Request update of PAI tile
                androidx.wear.tiles.TileService.getUpdater(this).requestUpdate(PaiTile::class.java)
            }

            MessageType.SYNC_DATA_WORKOUT -> {
                val app = Singleton.getAppSec(true)
                val workoutController = app.injection.inject(WorkoutController::class.java, null, false)
                app.sharedLogger.log("i", "Received request to sync all workout types")

                // Sync workout types
                workoutController.getWorkoutTypes(VersionHelper.getVersionName(), true)
                Singleton.sendMessageTOWearMessageReceiver(type, "")
            }

            MessageType.WORKOUT_GPS_DATA -> {
                val manager = WorkoutManager.workoutManager
                if (manager == null) {
                    Log.d("RPdb-Logger", "Received GPS data but no workout manager available")
                    return
                }

                try {
                    val data = Gson().fromJson(String(ev.data), AndroidGpsData::class.java)
                    manager.phoneTracking.onAndroidDataReceived(data)
                } catch (ex: Exception) {
                    Log.w(TAG, "Failed to parse GPS data", ex)
                    Log.d(TAG, "Message: ${String(ev.data)}")
                }
            }

            MessageType.WORKOUT_STATUS_UPDATE -> {
                val manager = WorkoutManager.workoutManager
                if (manager == null) {
                    Log.d("RPdb-Logger", "Received GPS data but no workout manager available")
                    return
                }

                manager.phoneTracking.onAndroidStatusRequest(String(ev.data))
            }

            else -> {

            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()

        Log.d(TAG, "Destroyed data sync listener")
    }

    private fun sendNotificationWithMessage(message: String) {
        val channelId = "TestChannel"
        val channel = NotificationChannel(
            channelId,
            getString(R.string.service_steps_title),
            NotificationManager.IMPORTANCE_DEFAULT
        )
        val manager = getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(channel)

        val builder = NotificationCompat.Builder(this, channelId)
            .setSmallIcon(R.drawable.ic_launcher_foreground)
            .setContentTitle("Hinweis")
            .setContentText(message)
            .setPriority(NotificationCompat.PRIORITY_DEFAULT)
            .setAutoCancel(true)

        // Show notification
        if (ActivityCompat.checkSelfPermission( RPout.getAppContext(),
                Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED) {
            manager.notify(1, builder.build())
        }
    }
}
package de.rpjosh.rpout.android

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.util.Log

class Autostart : BroadcastReceiver() {

    override fun onReceive(context: Context, intent: Intent) {
        Log.d(Singleton.TAG, "RPout: Device boot complete")

        // Initialize app
        if (Singleton.app()) return

        if (Singleton.appController.globalConfiguration.user == null) {
            Log.d(Singleton.TAG, "Shutting RPout down again because of disabled step service")
            Singleton.appController.sharedLogger.log("d", "Shutting RPout down again because of disabled step service")
        } else {
            Singleton.appController.sharedLogger.log("d", "Starting application")

            // Start all android services
            Singleton.appController.startAndroidServices()
        }

    }
}
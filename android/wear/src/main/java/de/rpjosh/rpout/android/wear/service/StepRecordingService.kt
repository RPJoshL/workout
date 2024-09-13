package de.rpjosh.rpout.android.wear.service

import android.Manifest
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.content.pm.ServiceInfo
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import android.os.Build
import android.os.IBinder
import android.util.Log
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import androidx.core.app.NotificationManagerCompat
import de.rpjosh.rpout.android.wear.R
import de.rpjosh.rpout.android.wear.RPout
import java.time.LocalDateTime

class StepRecordingService: Service(), SensorEventListener {

    private lateinit var sensorManager: SensorManager
    private lateinit var stepCounterSensor: Sensor

    private var lastStepNotification = LocalDateTime.now().minusHours(2)
    private var lastStepCount = 0f

    override fun onCreate() {
        super.onCreate()

        // Initialize sensors
        sensorManager = getSystemService(Context.SENSOR_SERVICE) as SensorManager
        sensorManager.getDefaultSensor(Sensor.TYPE_STEP_COUNTER).let {
            if (it == null) {
                Log.d("Workout", "Recevied no step counter sensor")
                return
            }
            stepCounterSensor = it
        }

        // Register sensor listener
        sensorManager.registerListener(this, stepCounterSensor, SensorManager.SENSOR_DELAY_NORMAL)

        // Start the foreground service
        startForeground(1, createNotification(),
            if (Build.VERSION.SDK_INT >= 34) ServiceInfo.FOREGROUND_SERVICE_TYPE_HEALTH else 0
        )
    }

    /**
     * Creates a notification for the foreground service
     */
    private fun createNotification(): Notification {
        val channelId = "StepActivity"
        val channel = NotificationChannel(
            channelId,
            getString(R.string.service_steps_title),
            NotificationManager.IMPORTANCE_DEFAULT
        )
        val manager = getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(channel)

        return Notification.Builder(this, channelId)
            .setContentTitle(getString(R.string.service_steps_title))
            .setContentText(getString(R.string.service_steps_text))
            .setSmallIcon(R.drawable.splash_icon)
            .build()
    }

    override fun onBind(intent: Intent?): IBinder? {
        // Don't allow binding
        return null
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        // Return sticky that the service is restarted if the system kills the service
        return START_STICKY
    }

    override fun onSensorChanged(event: SensorEvent?) {
        event?.values?.get(0)?.let {
            if (lastStepCount == 0f) {
                lastStepCount = it
                lastStepNotification = LocalDateTime.now()
                sendNotificationWithMessage("Step counter initialised")
            } else if (lastStepNotification.plusMinutes(55).isBefore(LocalDateTime.now())) {
                val stepsInHour = it - lastStepCount
                lastStepCount = it
                lastStepNotification = LocalDateTime.now()
                sendNotificationWithMessage("$stepsInHour Schritte in der letzten Stunde gemacht")
            }
        }
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
        if (ActivityCompat.checkSelfPermission( RPout.getAppContext(),Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED) {
            manager.notify(1, builder.build())
        }
    }

    override fun onAccuracyChanged(sensor: Sensor?, accuracy: Int) {
        Log.d("Workout", "Sensor accuracy changed: $accuracy")
    }

}
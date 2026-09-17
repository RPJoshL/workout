package de.rpjosh.rpout.android.services

import android.Manifest
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.Service
import android.content.Intent
import android.content.pm.PackageManager
import android.content.pm.ServiceInfo
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import android.os.IBinder
import android.util.Log
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import java.time.Duration

class StepRecordingService: Service(), SensorEventListener, StepRecorderCallback {

    private lateinit var sensorManager: SensorManager
    private lateinit var stepCounterSensor: Sensor

    private lateinit var recorder: StepRecorder

    @Volatile var isSensorManagerRegistered = false

    private val serviceJob = SupervisorJob()
    private val serviceScope = CoroutineScope(Dispatchers.IO + serviceJob)

    override fun onCreate() {
        super.onCreate()

        // Initialize dependencies
        recorder = StepRecorder(this, this)

        startForeground(1, recorder.createNotification(), ServiceInfo.FOREGROUND_SERVICE_TYPE_HEALTH)

        // Initialize sensor manager
        val attributionContext = createAttributionContext("step-recording")
        sensorManager = attributionContext.getSystemService(SENSOR_SERVICE) as SensorManager

        sensorManager.getDefaultSensor(Sensor.TYPE_STEP_COUNTER).let {
            if (it == null) {
                recorder.logger.log("e", "Received no step counter sensor")
                stopSelf()
                return
            }
            stepCounterSensor = it
        }

        recorder.logger.log("i", "Using sensor manager for step tracking")
        sensorManager.registerListener(this, stepCounterSensor, SensorManager.SENSOR_DELAY_NORMAL, 120_000_000)
        isSensorManagerRegistered = true
    }

    override fun onBind(intent: Intent?): IBinder? {
        // Don't allow binding
        return null
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        // Stop service if we received a stop command
        when (intent?.action?.uppercase()) {
            "STOP" -> {
                stop()
                stopSelf()
            }

            ActivityChecker.TAG_ACTIVITY_CHECK -> {
                serviceScope.launch { recorder.checkActivity() }
            }
        }

        // Return sticky that the service is restarted if the system kills the service
        return START_STICKY
    }

    private fun stop() {
        if (::sensorManager.isInitialized) {
            if (isSensorManagerRegistered) sensorManager.unregisterListener(this)
            isSensorManagerRegistered = false
        }

        recorder.stop()

        stopForeground(STOP_FOREGROUND_REMOVE)
        serviceJob.cancel()
    }

    override fun onSensorChanged(event: SensorEvent?) {
        event?.values?.get(0)?.let {
            val unixTime = TimeHelper.getUnixTimeFromBootTime(Duration.ofNanos(event.timestamp))

            Log.d("RPout-Logger", "Got new data from step sensor: $it at $unixTime")
            recorder.processNewStepCount(it, unixTime)
        }
    }

    override suspend fun flushMetrics() {
        // Nothing to flush
    }

    override fun sendNotification(message: String) {
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
        Log.d("RPout", "Sensor accuracy changed: $accuracy")
    }

}

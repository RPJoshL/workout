package de.rpjosh.rpout.android.services

import android.Manifest
import android.app.NotificationChannel
import android.app.NotificationManager
import android.content.Intent
import android.content.pm.PackageManager
import android.content.pm.ServiceInfo
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import androidx.health.services.client.HealthServices
import androidx.health.services.client.PassiveListenerCallback
import androidx.health.services.client.PassiveListenerService
import androidx.health.services.client.PassiveMonitoringClient
import androidx.health.services.client.clearPassiveListenerService
import androidx.health.services.client.data.DataPointContainer
import androidx.health.services.client.data.DataType
import androidx.health.services.client.data.PassiveListenerConfig
import androidx.health.services.client.flush
import androidx.health.services.client.setPassiveListenerService
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class PassiveStepRecordingService: PassiveListenerService(), StepRecorderCallback {

    private lateinit var healthClient: PassiveMonitoringClient
    @Volatile private var healthClientStepCounter = 0L

    private lateinit var recorder: StepRecorder

    private val serviceJob = SupervisorJob()
    private val serviceScope = CoroutineScope(Dispatchers.IO + serviceJob)

    override fun onCreate() {
        super.onCreate()

        // Initialize dependencies
        recorder = StepRecorder(this, this)

        startForeground(1, recorder.createNotification(), ServiceInfo.FOREGROUND_SERVICE_TYPE_HEALTH)
        recorder.logger.log("i", "Using battery efficient monitoring client for step tracking")

        val healthService = HealthServices.getClient(createAttributionContext("step-recording"))
        healthClient = healthService.passiveMonitoringClient

        val listener = PassiveListenerConfig.builder()
            .setDataTypes(setOf(DataType.STEPS_TOTAL))
            .build()

        val passiveListenerCallback: PassiveListenerCallback = object : PassiveListenerCallback {
            override fun onNewDataPointsReceived(dataPoints: DataPointContainer) {
                onNewDataPointsReceived(dataPoints)
            }

            override fun onRegistrationFailed(throwable: Throwable) {
                recorder.logger.log("w",  "Registration for health services failed: " + throwable.message)
            }
        }

        // Callback is much faster as the service but will drain more battery
        //healthClient.setPassiveListenerCallback(listener, passiveListenerCallback)

        serviceScope.launch {
            healthClient.setPassiveListenerService(PassiveStepRecordingService::class.java, listener)
        }
    }


    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        // Stop service if we received a stop command
        when (intent?.action?.uppercase()) {
            "STOP" -> {
                stop()
            }

            ActivityChecker.TAG_ACTIVITY_CHECK -> {
                serviceScope.launch { recorder.checkActivity() }
            }
        }

        // Return sticky that the service is restarted if the system kills the service
        return START_STICKY
    }

    private fun stop() {
        recorder.stop()

        serviceScope.launch {
            try {
                healthClient.clearPassiveListenerService()
            } catch(_: Exception) {
                stopForeground(STOP_FOREGROUND_REMOVE)
                stopSelf()
            }
        }
    }

    override fun onDestroy() {
        super.onDestroy()

        // Just to make sure the listener is always removed
        serviceScope.launch {
            try {
                healthClient.clearPassiveListenerService()
            } catch(_: Exception) {}
        }
    }

    /**
     * Processes the received step count from the passive monitor client
     */
    @Synchronized
    override fun onNewDataPointsReceived(dataPoints: DataPointContainer) {
        val originalRebootCount = healthClientStepCounter

        dataPoints.intervalDataPoints.forEach {
            val stepIncrement = it.value as Long

            if (stepIncrement > 0) {
                healthClientStepCounter += stepIncrement
                val unixTimestamp = TimeHelper.getUnixTimeFromBootTime(it.endDurationFromBoot)
                recorder.processNewStepCount(healthClientStepCounter.toFloat(), unixTimestamp)
            }

            // logger.log("d", "New data points received: " + it.value + " -> from " + it.startDurationFromBoot.seconds + " to " + it.endDurationFromBoot.seconds)
        }

        recorder.logger.log("d", "Received {0} steps within {1} data points from health client", healthClientStepCounter - originalRebootCount, dataPoints.intervalDataPoints.size)
    }

    override suspend fun flushMetrics() {
        healthClient.flush()
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
        if (ActivityCompat.checkSelfPermission( RPout.getAppContext(), Manifest.permission.POST_NOTIFICATIONS) == PackageManager.PERMISSION_GRANTED) {
            manager.notify(1, builder.build())
        }
    }

}
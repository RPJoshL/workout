package de.rpjosh.rpout.android.services

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
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.services.Logger
import java.time.Instant
import java.time.LocalDateTime
import java.util.TimeZone

class StepRecordingService: Service(), SensorEventListener {

    private lateinit var logger: Logger
    private lateinit var metricController: MetricController

    private lateinit var sensorManager: SensorManager
    private lateinit var stepCounterSensor: Sensor

    // Synchronize object used for access the step variables
    private val syncObject = Object()

    // The current step entry that is tracked and should be saved in the local SQLite database
    @Volatile var currentStep: Step? = null

    /* The last received sensor value */
    @Volatile var lastSensorTime: Long = 0
    @Volatile var lastSensorValue: Float = 0f

    override fun onCreate() {
        super.onCreate()

        // Initialize dependencies
        val app = Singleton.getAppSec()
        logger = app.injection.inject(Logger::class.java, arrayOf("StepRecordingService"), false)
        metricController = app.injection.inject(MetricController::class.java, null, false)

        // Initialize sensor manager
        sensorManager = getSystemService(Context.SENSOR_SERVICE) as SensorManager
        sensorManager.getDefaultSensor(Sensor.TYPE_STEP_COUNTER).let {
            if (it == null) {
                Singleton.getAppSec().sharedLogger.log("e", "Received no step counter sensor")
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
        // Stop service if we received a stop command
        if (intent?.action?.uppercase() == "STOP") {
            stop()
            stopSelf()
        }

        // Return sticky that the service is restarted if the system kills the service
        return START_STICKY
    }

    private fun stop() {
        if (::sensorManager.isInitialized) {
            sensorManager.unregisterListener(this)

            // Process the last step count to not loose any process
            val sensorEvent = sensorManager.getSensorList(Sensor.TYPE_STEP_COUNTER)
            if (sensorEvent.isNotEmpty()) {
                processNewStepCount(sensorEvent[0].resolution)
            }
        }

        stopForeground(STOP_FOREGROUND_REMOVE)
    }

    @Synchronized
    private fun processNewStepCount(rebootCounter: Float) {
        val unixTime = System.currentTimeMillis() / 1000

        // Initialize tracking
        if (currentStep == null) {
            currentStep = Step.Empty()
            currentStep!!.stepsSinceLastReboot = rebootCounter.toLong()

            sendNotificationWithMessage("Started to count steps")
        } else if (unixTime - currentStep!!.startUnix > 300) {
            // Push a value every five minutes
            currentStep!!.endNow(rebootCounter - currentStep!!.stepsSinceLastReboot)

            // Construct new step value
            val newStep = Step.Empty()
            newStep.stepsSinceLastReboot = rebootCounter.toLong()

            // We don't receive an update if the user doesn't do any steps.
            // To make sensor events more exact, we check the last sensor value (we received)
            // and compare it to this one
            if (unixTime - currentStep!!.startUnix > 900 && unixTime - lastSensorTime > 600) {
                // If the current step count equals almost the last sensor time, we use the time of the last sensor
                if (rebootCounter - lastSensorValue < 20 && currentStep!!.count > 20) {
                    // Add at least a minute because auf time truncating and add a few steps so calculation is "correct"
                    currentStep!!.endUnix = lastSensorTime + 60
                    val modifiedEnd = LocalDateTime.ofInstant(Instant.ofEpochMilli(newStep.endUnix), TimeZone.getDefault().toZoneId())
                    currentStep!!.end = TimeHelper.fromClientToServer(modifiedEnd)
                    currentStep!!.count += 10
                }
            }

            // Save step count in db and overwrite current step instance
            logger.log("d", "Having ${currentStep!!.count} steps to sync")
            if (currentStep!!.count > 5) {
                val _currentStep = currentStep!!
                Thread {
                    metricController.addStep(_currentStep)
                }.start()
            } else {
                // Add the old steps to the new steps
                newStep.stepsSinceLastReboot -= currentStep!!.count
            }
            currentStep = newStep
        }

        // Update last sensor value
        lastSensorTime = unixTime
        lastSensorValue = rebootCounter
    }

    override fun onSensorChanged(event: SensorEvent?) {
        event?.values?.get(0)?.let {
            processNewStepCount(it)
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
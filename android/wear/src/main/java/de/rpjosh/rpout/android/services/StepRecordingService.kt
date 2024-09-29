package de.rpjosh.rpout.android.services

import android.Manifest
import android.annotation.SuppressLint
import android.app.AlarmManager
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.Intent.getIntent
import android.content.pm.PackageManager
import android.content.pm.ServiceInfo
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import android.os.Build
import android.os.IBinder
import android.provider.Settings
import android.util.Log
import androidx.core.app.ActivityCompat
import androidx.core.app.NotificationCompat
import androidx.core.content.ContextCompat
import androidx.work.ExistingWorkPolicy
import androidx.work.ListenableWorker.Result
import androidx.work.OneTimeWorkRequest
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.Worker
import androidx.work.WorkerParameters
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.services.NotActiveActivity
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.services.Logger
import java.time.Instant
import java.time.LocalDateTime
import java.time.LocalTime
import java.util.Calendar
import java.util.TimeZone
import java.util.concurrent.TimeUnit

class StepRecordingService: Service(), SensorEventListener {

    private lateinit var logger: Logger
    private lateinit var metricController: MetricController

    private lateinit var sensorManager: SensorManager
    private lateinit var stepCounterSensor: Sensor

    // The current step entry that is tracked and should be saved in the local SQLite database
    @Volatile var currentStep: Step? = null

    /** The next scheduled activity check */
    @Volatile var activityCheckTask: OneTimeWorkRequest? = null

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
        when (intent?.action?.uppercase()) {
            "STOP" -> {
                Thread {
                    stop()
                    stopSelf()
                }.start()
            }

            "ACTIVITY_CHECK" -> {
                Thread { checkActivity() }.start()
            }
        }

        // Return sticky that the service is restarted if the system kills the service
        return START_STICKY
    }

    private fun stop() {
        if (::sensorManager.isInitialized) {
            sensorManager.unregisterListener(this)

            // Process the last step count to not lose any process
            processNewStepCount(lastSensorValue)

            // Stop any work manager
            activityCheckTask?.id?.let { WorkManager.getInstance(RPout.getAppContext()).cancelWorkById(it) }
        }

        stopForeground(STOP_FOREGROUND_REMOVE)
    }

    @SuppressLint("WearRecents")
    @Synchronized
    private fun checkActivity() {
        val currentTime = LocalTime.now()

        if (currentTime.hour >= 20 || currentTime.hour < 7) {
            // Exclude time range from 20:00 Uhr - 07:00 Uhr
            val scheduleIn = if (currentTime.hour >= 20) 24 - currentTime.hour + 7 else 7 - currentTime.hour
            scheduleActivityCheck(scheduleIn.toLong(), TimeUnit.HOURS)
        } else if(Settings.Global.getInt(contentResolver, "zen_mode") != 0) {
            // DND mode enabled => don't show any activity
            scheduleActivityCheck(90, TimeUnit.MINUTES)
        } else {
            // Store steps now if they didn't change
            val unixTime = System.currentTimeMillis() / 1000
            if (unixTime - currentStep!!.startUnix > 900 && unixTime - lastSensorTime > 600) {
                processNewStepCount(lastSensorValue)
            }

            // Get the last time the user was active
            val lastActiveTime = metricController.dao().getLastTimeGoalReached(150)

            // Activity in the last 60 minutes required (activity counts step in the last 58 minutes)
            if (lastActiveTime == null || (unixTime - lastActiveTime) > (60 * 60) ) {
                logger.log("i", "Last activity was ${lastActiveTime?.let { unixTime - lastActiveTime } ?: "?"} seconds ago")
                scheduleActivityCheck(65, TimeUnit.MINUTES)

                // Start activity to notify the user about being active
                val intent = Intent(RPout.getAppContext(), NotActiveActivity::class.java).apply {
                    addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                    addFlags(Intent.FLAG_ACTIVITY_EXCLUDE_FROM_RECENTS);
                }
                RPout.getAppContext().startActivity(intent)
            } else {
                // Schedule task when 60 minutes since the last activity are left
                var scheduleIn = (65 * 60) - unixTime - lastActiveTime
                if (scheduleIn < 10 * 60) scheduleIn = 10 * 60

                logger.log("d", "Found activity within last 60 minutes (${unixTime - lastActiveTime} seconds ago)")
                scheduleActivityCheck(scheduleIn, TimeUnit.SECONDS)
            }
        }
    }

    /**
     * Schedules the next activity check within the provided duration
     */
    @Synchronized
    private fun scheduleActivityCheck(duration: Long, timeUnit: TimeUnit) {
        var scheduleIn = timeUnit.toSeconds(duration)
        val alarmManager = RPout.getAppContext().getSystemService(Context.ALARM_SERVICE) as AlarmManager

        if (!alarmManager.canScheduleExactAlarms()) {
            logger.log("w", "Cannot schedule task via alarm manager because \"SCHEDULE_EXACT\" permission is not granted. Falling back to Work manager")

            // Minimum schedule time are 15 minutes
            if (scheduleIn <  15 * 60) scheduleIn = 15 * 60
            activityCheckTask = OneTimeWorkRequestBuilder<ActivityChecker>().setInitialDelay(scheduleIn, TimeUnit.SECONDS).addTag(ActivityChecker.TAG_ACTIVITY_CHECK).build()
            WorkManager.getInstance(RPout.getAppContext()).enqueueUniqueWork(ActivityChecker.TAG_ACTIVITY_CHECK, ExistingWorkPolicy.REPLACE, activityCheckTask!!)
        } else {
            logger.log("d", "Scheduled activity check (with alarm manager) in $scheduleIn seconds")
            // Stop any pending work manager
            WorkManager.getInstance(RPout.getAppContext()).cancelAllWorkByTag(ActivityChecker.TAG_ACTIVITY_CHECK)

            // Set alarm
            val wakeUpTime = Calendar.getInstance().also{ it.add(Calendar.SECOND, scheduleIn.toInt()) }.timeInMillis
            alarmManager.setExact(AlarmManager.RTC_WAKEUP, wakeUpTime,getIntent(ActivityCheckerAlarm::class.java))
        }
    }

    private fun getIntent(cls: Class<*>): PendingIntent {
        val intent = Intent(RPout.getAppContext(), cls)
        return PendingIntent.getBroadcast(
            RPout.getAppContext(), 0, intent, PendingIntent.FLAG_MUTABLE
        )
    }

    @Synchronized
    private fun processNewStepCount(rebootCounter: Float) {
        val unixTime = System.currentTimeMillis() / 1000

        // Initialize tracking
        if (currentStep == null) {
            currentStep = Step.Empty()
            currentStep!!.stepsSinceLastReboot = rebootCounter.toLong()

            // Initialize activity checker
            scheduleActivityCheck(50, TimeUnit.MINUTES)
            // Thread {  checkActivity() }.start()

            // Send simple info notification
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
                    // Add at least a minute because of time truncating and add a few steps so calculation is "correct"
                    currentStep!!.endUnix = lastSensorTime + 60
                    logger.log("d", "Modifying end time to improve accuracy (now -> ${currentStep!!.endUnix})")
                    val modifiedEnd = LocalDateTime.ofInstant(Instant.ofEpochSecond(currentStep!!.endUnix), TimeZone.getDefault().toZoneId())
                    currentStep!!.end = TimeHelper.fromClientToServer(modifiedEnd)
                    currentStep!!.count += 10
                }
            }

            // Save step count in db and overwrite current step instance
            logger.log("d", "Having ${currentStep!!.count} steps to sync")
            if (currentStep!!.count > 5) {
                val _currentStep = currentStep!!.copy()
                metricController.addStep(_currentStep)
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
            Thread { processNewStepCount(it) }.start()
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

/**
 * ActivityChecker checks the activity score of the user within the
 * last 60 minutes and displays an activity that the user should move
 * in order to stay active.
 *
 * Because this state depends on the tracked steps, it's execution is
 * managed from the "StepRecordingService"
 */
public class ActivityChecker(appContext: Context, workerParams: WorkerParameters): Worker(appContext, workerParams) {

    companion object {
        const val TAG_ACTIVITY_CHECK = "ACTIVITY_CHECK"
    }

    override fun doWork(): Result {
        // We should have a app reference
        val app = Singleton.getApp() ?: return Result.failure()
        app.sharedLogger.log("d", "Executing activity check scheduled from Work manager")

        // Send request to foreground service
        val serviceIntent = Intent(RPout.getAppContext(), StepRecordingService::class.java)
        serviceIntent.action = "ACTIVITY_CHECK"
        ContextCompat.startForegroundService(RPout.getAppContext(), serviceIntent)

        return Result.success()
    }

}

class ActivityCheckerAlarm: BroadcastReceiver() {

    override fun onReceive(context: Context, intent: Intent) {
        // We should have a app reference
        val app = Singleton.getApp() ?: return
        app.sharedLogger.log("d", "Executing activity check from alarm manager")

        // Send request to foreground service
        val serviceIntent = Intent(RPout.getAppContext(), StepRecordingService::class.java)
        serviceIntent.action = "ACTIVITY_CHECK"
        ContextCompat.startForegroundService(RPout.getAppContext(), serviceIntent)
    }

}
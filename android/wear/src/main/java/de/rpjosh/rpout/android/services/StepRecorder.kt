package de.rpjosh.rpout.android.services

import android.annotation.SuppressLint
import android.app.AlarmManager
import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.provider.Settings
import androidx.core.content.ContextCompat
import androidx.health.services.client.ExerciseClient
import androidx.health.services.client.HealthServices
import androidx.health.services.client.data.ExerciseTrackedStatus
import androidx.health.services.client.getCurrentExerciseInfo
import androidx.work.ExistingWorkPolicy
import androidx.work.OneTimeWorkRequest
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.Worker
import androidx.work.WorkerParameters
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.services.NotActiveActivity
import de.rpjosh.rpout.android.complication.StepComplicationDataService
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.tiles.PaiTile
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import java.time.Instant
import java.time.LocalDateTime
import java.time.LocalTime
import java.util.Calendar
import java.util.TimeZone
import java.util.concurrent.TimeUnit
import kotlin.math.roundToLong

interface StepRecorderCallback {
    suspend fun flushMetrics()
    fun sendNotification(message: String)
}

class StepRecorder(
    val context: Context,
    val client: StepRecorderCallback
) {

    companion object {
        /** Threshold for showing the Not active activity in minutes */
        const val NOT_ACTIVE_TIMEOUT = 65
        /** Whether to enable the activity check */
        const val ENABLE_NOT_ACTIVE = false
    }

    var logger: Logger
    private var metricController: MetricController

    private var notActiveTimeout = NOT_ACTIVE_TIMEOUT
    private var enableNotActive = ENABLE_NOT_ACTIVE

    // The current step entry that is tracked and should be saved in the local SQLite database
    @Volatile var currentStep: Step? = null

    /** The next scheduled activity check */
    @Volatile var activityCheckTask: OneTimeWorkRequest? = null

    /* The last received sensor value */
    @Volatile var lastSensorTime: Long = 0
    @Volatile var lastSensorValue: Float = 0f

    /** The last day when the PAI tile was updated */
    @Volatile var lastPaiUpdate = 0

    private val serviceJob = SupervisorJob()
    private val serviceScope = CoroutineScope(Dispatchers.IO + serviceJob)

    private var workoutClient: ExerciseClient

    init {
        // Initialize dependencies
        val app = Singleton.getAppSec()
        logger = app.injection.inject(Logger::class.java, arrayOf("StepRecordingService"), false)
        metricController = app.injection.inject(MetricController::class.java, null, false)

        // Initialize health client
        val healthService = HealthServices.getClient(context.createAttributionContext("step-recording"))
        workoutClient = healthService.exerciseClient
    }

    @SuppressLint("WearRecents", "RestrictedApi")
    suspend fun checkActivity() {
        if (!enableNotActive) {
            return
        }

        val currentTime = LocalTime.now()

        if (currentTime.hour !in 7..<20) {
            // Exclude time range from 20:00 Uhr - 07:00 Uhr
            val scheduleIn = if (currentTime.hour >= 20) 24 - currentTime.hour + 7 else 7 - currentTime.hour
            scheduleActivityCheck(scheduleIn.toLong(), TimeUnit.HOURS)
        } else if(Settings.Global.getInt(context.contentResolver, "zen_mode") != 0) {
            // DND mode enabled => don't show any activity
            scheduleActivityCheck(120, TimeUnit.MINUTES)
        } else if(workoutClient.getCurrentExerciseInfo().exerciseTrackedStatus in listOf(ExerciseTrackedStatus.OTHER_APP_IN_PROGRESS, ExerciseTrackedStatus.OWNED_EXERCISE_IN_PROGRESS)) {
            logger.log("d", "Not executing activity check because a workout is currently tracked")
            scheduleActivityCheck((notActiveTimeout * 1.5).roundToLong(), TimeUnit.MINUTES)
        } else {
            client.flushMetrics()

            // Store steps now if they didn't change
            val unixTime = System.currentTimeMillis() / 1000
            if (unixTime - currentStep!!.startUnix > 900 && unixTime - lastSensorTime > 600) {
                processNewStepCount(lastSensorValue)
            }

            // Get the last time the user was active
            val lastActiveTime = metricController.dao().getLastTimeGoalReached(150)

            // Activity in the last 60 minutes required (activity counts step in the last x - 2 minutes)
            if (lastActiveTime == null || (unixTime - lastActiveTime) > (notActiveTimeout * 60) ) {
                logger.log("i", "Last activity was ${lastActiveTime?.let { unixTime - lastActiveTime } ?: "?"} seconds ago")
                scheduleActivityCheck(notActiveTimeout.toLong() + 5, TimeUnit.MINUTES)

                // Start activity to notify the user about being active
                val intent = Intent(RPout.getAppContext(), NotActiveActivity::class.java).apply {
                    addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                    addFlags(Intent.FLAG_ACTIVITY_EXCLUDE_FROM_RECENTS)
                }
                RPout.getAppContext().startActivity(intent)
            } else {
                // Schedule task when x minutes since the last activity are left
                var scheduleIn = ( (notActiveTimeout + 5) * 60) - unixTime - lastActiveTime
                if (scheduleIn < 12 * 60) scheduleIn = 12 * 60

                logger.log("d", "Found activity within last $notActiveTimeout minutes (${unixTime - lastActiveTime} seconds ago)")
                scheduleActivityCheck(scheduleIn, TimeUnit.SECONDS)
            }
        }
    }

    /**
     * Schedules the next activity check within the provided duration
     */
    @Synchronized
    private fun scheduleActivityCheck(duration: Long, timeUnit: TimeUnit) {
        if (!enableNotActive) {
            return
        }

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
    fun processNewStepCount(rebootCounter: Float, unixTimeForData: Long? = null) {
        val unixTime = unixTimeForData ?: (System.currentTimeMillis() / 1000)

        // Initialize tracking
        if (currentStep == null) {
            currentStep = Step.Empty()
            currentStep!!.stepsSinceLastReboot = rebootCounter.toLong()

            // Initialize activity checker
            scheduleActivityCheck(50, TimeUnit.MINUTES)

            // Send simple info notification
            client.sendNotification("Started to count steps")
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
                Thread{
                    metricController.addStep(_currentStep)
                    StepComplicationDataService.requestUpdate(context)
                }.start()
            } else {
                // Add the old steps to the new steps
                newStep.stepsSinceLastReboot -= currentStep!!.count
            }
            currentStep = newStep

            // Check if PAI tile has to be updated (new day)
            val currentDay = LocalDateTime.now().dayOfMonth
            if (currentDay != lastPaiUpdate) {
                lastPaiUpdate = currentDay
                logger.log("d", "Detected day transition")

                androidx.wear.tiles.TileService.getUpdater(context).requestUpdate(PaiTile::class.java)
            }
        }

        // Update last sensor value
        lastSensorTime = unixTime
        lastSensorValue = rebootCounter
    }

    fun stop() {
        // Stop any work manager
        activityCheckTask?.id?.let { WorkManager.getInstance(context).cancelWorkById(it) }

        // Process the last step count to not lose any process
        processNewStepCount(lastSensorValue)
    }

    /**
     * Creates a notification for the foreground service
     */
    fun createNotification(): Notification {
        val channelId = "StepActivity"
        val channel = NotificationChannel(
            channelId,
            context.getString(R.string.service_steps_title),
            NotificationManager.IMPORTANCE_DEFAULT
        )
        val manager = context.getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(channel)

        return Notification.Builder(context, channelId)
            .setContentTitle(context.getString(R.string.service_steps_title))
            .setContentText(context.getString(R.string.service_steps_text))
            .setSmallIcon(R.drawable.splash_icon)
            .build()
    }

}

/**
 * ActivityChecker checks the activity score of the user within the
 * last x minutes and displays an activity that the user should move
 * in order to stay active.
 *
 * Because this state depends on the tracked steps, it's execution is
 * managed from the "StepRecordingService"
 */
class ActivityChecker(appContext: Context, workerParams: WorkerParameters): Worker(appContext, workerParams) {

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
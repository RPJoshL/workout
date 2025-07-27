package de.rpjosh.rpout.android.workout

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Intent
import android.content.pm.ServiceInfo
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.PorterDuff
import android.graphics.drawable.Icon
import android.graphics.drawable.PictureDrawable
import android.os.Build
import android.os.IBinder
import android.os.SystemClock
import android.util.Log
import androidx.compose.ui.platform.LocalDensity
import androidx.core.app.NotificationCompat
import androidx.core.graphics.drawable.IconCompat
import androidx.wear.ongoing.OngoingActivity
import de.rpjosh.rpout.android.shared.services.Tr
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import androidx.wear.ongoing.Status
import com.caverock.androidsvg.SVG
import de.rpjosh.rpout.android.activities.main.WorkoutTrackingActivity
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.workout.State
import de.rpjosh.rpout.android.shared.workout.WorkoutManager
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import androidx.core.graphics.createBitmap
import de.rpjosh.rpout.android.activities.main.WorkoutStartActivity

/**
 * WorkoutTrackService is a foreground service that tracks a workout in the
 * background and updates the matching data in the WorkoutManager
 */
class WorkoutTrackService: Service() {

    private val job = SupervisorJob()
    private val scope = CoroutineScope(Dispatchers.IO + job)

    override fun onCreate() {
        super.onCreate()

        // Workout manager is required. If we don't have one, the service was probably restarted by the system
        if (WorkoutManager.workoutManager == null) {
            stopSelf()
            return
        }

        // Start the foreground service
        startForeground(141, buildNotification(),
            if (Build.VERSION.SDK_INT >= 34) ServiceInfo.FOREGROUND_SERVICE_TYPE_HEALTH or ServiceInfo.FOREGROUND_SERVICE_TYPE_LOCATION else 0
        )

        // Initialize workout
        val t = this
        scope.launch {
            WorkoutManager.workoutManager?.initExercise(t, WorkoutTrackingActivity::class.java, Singleton.appController.injection)
            sendStartIntentToActivity()
        }
    }

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        // Stop service if we received a stop command
        when (intent?.action?.uppercase()) {
            "STOP" -> {
                scope.launch {
                    WorkoutManager.workoutManager?.shutdownExercise()
                }
                stopSelf()
            }
            "RESTART" -> {
                val t = this
                scope.launch {
                    WorkoutManager.workoutManager?.shutdownExercise()
                    WorkoutManager.workoutManager?.initExercise(t, WorkoutTrackingActivity::class.java, Singleton.appController.injection)
                    sendStartIntentToActivity()
                }
            }
            "NOTIFICATION" -> {
                if (WorkoutManager.workoutManager == null) {
                    return START_STICKY
                }

                // Update the displayed foreground notification
                val manager = getSystemService(NotificationManager::class.java)
                manager.notify(141, buildNotification())
            }
        }

        // Return sticky that the service is restarted if the system kills the service
        return START_STICKY
    }

    private fun sendStartIntentToActivity() {
        val intent = Intent(WorkoutStartActivity.BROADCAST_FILTER_ACTION).apply {
            putExtra(WorkoutStartActivity.INTENT_WORKOUT_INITED, true)
        }
        sendBroadcast(intent)
    }

    /**
     * Creates a notification for the foreground service
     */
    private fun buildNotification(): Notification {
        val channelId = "Workout"
        val notificationId = 12
        val channel = NotificationChannel(
            channelId,
            Tr.get("workoutService_channel"),
            NotificationManager.IMPORTANCE_DEFAULT
        )
        val manager = getSystemService(NotificationManager::class.java)
        manager.createNotificationChannel(channel)

        // On tap action
        val notificationIntent = Intent(applicationContext, WorkoutManager.workoutManager?.activityClass ?: this.javaClass)
        val pendingIntent = PendingIntent.getActivity(
            applicationContext,
            0,
            notificationIntent,
            PendingIntent.FLAG_IMMUTABLE or PendingIntent.FLAG_UPDATE_CURRENT
        )

        // Get workout manager
        val workoutManager = WorkoutManager.workoutManager!!

        // Get icon to render
        val svgString = workoutManager.type.icon.replace("\"currentColor\"", "\"" + workoutManager.type.tagDark + "\"")
        val svg = SVG.getFromString(svgString)
        svg.documentWidth = 64f
        svg.documentHeight = 64f
        // Convert it into a drawable
        val bitmap = createBitmap(svg.documentWidth.toInt(), svg.documentHeight.toInt())
        val canvas = Canvas(bitmap)
        svg.renderToCanvas(canvas)

        val builder = NotificationCompat.Builder(this, channelId)
            .setContentTitle(Tr.get("workoutService_title"))
            .setContentText(Tr.get("workoutService_text"))
            .setSmallIcon(IconCompat.createWithBitmap(bitmap))
            .setLargeIcon(bitmap)
            // Ongoing notification
            .setOngoing(true)
            .setCategory(NotificationCompat.CATEGORY_WORKOUT)
            .setVisibility(NotificationCompat.VISIBILITY_PUBLIC)

        // Get duration of workout
        val currentMillis = SystemClock.elapsedRealtime()
        val isWorkoutPaused = workoutManager.state.value in arrayOf(State.PAUSED, State.READY, State.PRE_GPS_CONNECTING, State.ERROR)
        val isWorkoutPreparing = workoutManager.state.value in arrayOf(State.READY, State.PRE_GPS_CONNECTING, State.ERROR)

        // Add ongoing activity text
        val type = WorkoutManager.workoutManager!!.type
        val onGoingStatus = Status.Builder()
            .addTemplate(type.getName(Tr.getUsedLanguage()) + " #duration#")
            .addPart("duration", Status.StopwatchPart(
                currentMillis - workoutManager.workoutData.activeDuration.value.activeDuration.toMillis(),
                if (isWorkoutPreparing) SystemClock.elapsedRealtime() else if (isWorkoutPaused) SystemClock.elapsedRealtime() - TimeHelper.getBootTimeFromUnixTime(workoutManager.workoutData.activeDuration.value.time.epochSecond) else -1L)
            )
            .build()

        // Cannot use dynamic SVG icon because of this error: "The interactive icon is not resource type. Ignore it.".
        // Therefore (and only because of this reason), we've added all types redundant to this app...
        val iconArray = arrayOf(R.drawable.type_1, R.drawable.type_2, R.drawable.type_3, R.drawable.type_4, R.drawable.type_5, R.drawable.type_6, R.drawable.type_7, R.drawable.type_8, R.drawable.type_9, R.drawable.type_10)

        val onGoingActivity = OngoingActivity.Builder(applicationContext, notificationId, builder)
            .setStaticIcon(if (iconArray.size >= type.id) iconArray[(type.id-1).toInt()] else iconArray.first())
            .setTouchIntent(pendingIntent)
            .setStatus(onGoingStatus)
            .build()
        onGoingActivity.apply(applicationContext)

        return builder.build()
    }

    override fun onBind(intent: Intent?): IBinder? {
        // Don't allow binding
        return null
    }

}
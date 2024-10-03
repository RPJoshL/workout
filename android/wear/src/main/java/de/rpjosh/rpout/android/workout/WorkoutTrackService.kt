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
import de.rpjosh.rpout.android.shared.R
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
            WorkoutManager.workoutManager?.initExercise(t, WorkoutTrackingActivity::class.java)
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
            "NOTIFICATION" -> {
                // Update the displayed foreground notification
                val manager = getSystemService(NotificationManager::class.java)
                manager.notify(141, buildNotification())
            }
        }

        // Return sticky that the service is restarted if the system kills the service
        return START_STICKY
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
        val bitmap = Bitmap.createBitmap(svg.documentWidth.toInt(), svg.documentHeight.toInt(), Bitmap.Config.ARGB_8888)
        val canvas = Canvas(bitmap)
        svg.renderToCanvas(canvas)

        val builder = NotificationCompat.Builder(this, channelId)
            .setContentTitle(Tr.get("workoutService_title"))
            .setContentText(Tr.get("workoutService_text"))
            .setSmallIcon(IconCompat.createWithBitmap(bitmap))
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
        val onGoingActivity = OngoingActivity.Builder(applicationContext, notificationId, builder)
            // This display won't be show because of this error: "The interactive icon is not resource type. Ignore it."
            .setStaticIcon(Icon.createWithBitmap(bitmap))
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
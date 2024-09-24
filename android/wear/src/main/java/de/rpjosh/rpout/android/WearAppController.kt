package de.rpjosh.rpout.android

import android.app.Activity
import android.content.Context
import android.content.Intent
import android.util.Log
import androidx.core.content.ContextCompat
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.NetworkType
import androidx.work.PeriodicWorkRequest
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import de.rpjosh.rpout.android.services.AndroidSynchronization
import de.rpjosh.rpout.android.shared.controller.AppController
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.services.ResponseView
import de.rpjosh.rpout.android.services.StepRecordingService
import de.rpjosh.rpout.android.services.Uploader
import de.rpjosh.rpout.android.services.WearUtils
import java.util.concurrent.TimeUnit

class WearAppController: AppController(
    RPout.getAppContext(),
    ResponseView::class,
    WearUtils::class,
    AndroidSynchronization::class
) {

    companion object {
        private var isAndroidTranslationAdded = false

        fun addAndroidTranslation() {
            if (!isAndroidTranslationAdded) {
                addAdditionalPropsTranslations("translation.wear")
                isAndroidTranslationAdded = true
            }
        }
    }

    private val responseA: ResponseView

    @Volatile
    private var isMainStarted = false
    private var firstStartOfMain = true

    val sharedLogger: Logger

    init {
        Log.d(Singleton.TAG, "RPout started")

        responseA = injection.inject(ResponseView::class.java, null, true)
        sharedLogger = injection.inject(Logger::class.java, arrayOf("shared"), false)

        Log.d(Singleton.TAG, "RPout startup completed (injection)")
        Singleton.setApp(this)

        // Start any services
        if (globalConfiguration.user != null) startAndroidServices()
    }

    override fun beforeInjection() {
        addAndroidTranslation()

        // Add self as a concrete class
        injection.addConcreteDependency(WearAppController::class.java, this)

        Singleton.setAppSilent(this)
    }

    /**
     * This method will be called from the main activity, when it was created.
     * All necessary actions for the application will be handled inside this
     */
    fun activityCreated(context: Context, activity: Activity, toolbar: Int) {
        responseA.setActivity(activity.baseContext, activity)
        Log.d(Singleton.TAG, "Activity created")
    }

    /**
     * This method will be called from the main activity, when it was paused.
     * All necessary actions for the application will be handled inside this
     */
    fun activityPaused(activity: Activity) {
        responseA.removeActivity(activity)
        Log.d(Singleton.TAG, "Activity paused")
    }

    /**
     * This method will be called from the main activity, when it was started.
     * All necessary actions for the application will be handled inside this
     */
    fun activityStarted(context: Context, activity: Activity) {
        responseA.setActivity(context, activity)
        Log.d(Singleton.TAG, "Activity started")
    }

    /**
     * This method will be called from the main activity, when it was destroyed.
     * All necessary actions for the application will be handled inside this
     */
    fun activityDestroyed(activity: Activity) {
        responseA.removeActivity(activity)
        Log.d(Singleton.TAG, "Activity destroyed")
    }

    fun startAndroidServices() {
        // Start step foreground service
        if (globalConfiguration.user != null) {
            val serviceIntent = Intent(RPout.getAppContext(), StepRecordingService::class.java)
            ContextCompat.startForegroundService(RPout.getAppContext(), serviceIntent)

            // Start Work manager to sync data
            val constraint = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .build()
            val worker = PeriodicWorkRequestBuilder<Uploader>(120, TimeUnit.MINUTES)
                .setConstraints(constraint)
                .addTag(Uploader.TAG_UPLOADER)
                .build()
            WorkManager.getInstance(RPout.getAppContext()).enqueueUniquePeriodicWork(Uploader.TAG_UPLOADER, ExistingPeriodicWorkPolicy.UPDATE, worker)
        }
    }

    fun stopAndroidServices() {
        // Stop the foreground service
        val serviceIntent = Intent(RPout.getAppContext(), StepRecordingService::class.java)
        serviceIntent.action = "STOP"
        ContextCompat.startForegroundService(RPout.getAppContext(), serviceIntent)

        // Stop synchronize service
        WorkManager.getInstance(RPout.getAppContext()).cancelAllWorkByTag(Uploader.TAG_UPLOADER)
    }

}
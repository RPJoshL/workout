package de.rpjosh.rpout.android

import android.app.Activity
import android.content.Context
import android.util.Log
import de.rpjosh.rpout.android.services.AndroidUtils
import de.rpjosh.rpout.android.services.ResponseView
import de.rpjosh.rpout.android.services.WearSynchronization
import de.rpjosh.rpout.android.shared.controller.AppController
import de.rpjosh.rpout.android.shared.services.Logger

class AndroidAppController
    () : AppController(
    RPout.getAppContext(),
    ResponseView::class,
    AndroidUtils::class,
    WearSynchronization::class
) {

    companion object {
        private var isAndroidTranslationAdded = false

        fun addAndroidTranslation() {
            if (!isAndroidTranslationAdded) {
                addAdditionalPropsTranslations("translation.android")
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
        injection.addConcreteDependency(AndroidAppController::class.java, this)

        Singleton.setAppSilent(this)
    }

    /**
     * This method will be called from the main activity, when it was created.
     * All necessary actions for the application will be handled inside this
     */
    fun activityCreated(context: Context, activity: Activity) {
        responseA.setActivity(activity.baseContext, activity)
        Log.d(Singleton.TAG, "Activity created: " + activity.javaClass.simpleName)
    }

    /**
     * This method will be called from the main activity, when it was paused.
     * All necessary actions for the application will be handled inside this
     */
    fun activityPaused(activity: Activity) {
        responseA.removeActivity(activity)
        Log.d(Singleton.TAG, "Activity paused: " + activity.javaClass.simpleName)
    }

    /**
     * This method will be called from the main activity, when it was started.
     * All necessary actions for the application will be handled inside this
     */
    fun activityStarted(context: Context, activity: Activity) {
        responseA.setActivity(context, activity)
        Log.d(Singleton.TAG, "Activity started: " + activity.javaClass.simpleName)
    }

    /**
     * This method will be called from the main activity, when it was destroyed.
     * All necessary actions for the application will be handled inside this
     */
    fun activityDestroyed(activity: Activity) {
        responseA.removeActivity(activity)
        Log.d(Singleton.TAG, "Activity destroyed: " + activity.javaClass.simpleName)
    }

    private fun startAndroidServices() {}

}
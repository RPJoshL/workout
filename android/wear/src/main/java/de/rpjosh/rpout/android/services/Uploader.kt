package de.rpjosh.rpout.android.services

import android.content.Context
import android.content.Intent
import androidx.core.content.ContextCompat
import androidx.work.Worker
import androidx.work.WorkerParameters
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.controller.MetricController

/**
 * Uploader uploads and synchronizes metrics like steps and workout data
 */
public class Uploader(appContext: Context, workerParams: WorkerParameters): Worker(appContext, workerParams) {

    companion object {
        const val TAG_UPLOADER = "UPLOADER"
    }

    override fun doWork(): Result {
        // We should have a app reference
        val app = Singleton.getApp() ?: return Result.failure()

        // Inject controller
        val metricController = app.injection.inject(MetricController::class.java, null,  false)

        // Synchronize everything
        if (!metricController.synchronizeSteps()) Result.retry()

        return Result.success()
    }

}
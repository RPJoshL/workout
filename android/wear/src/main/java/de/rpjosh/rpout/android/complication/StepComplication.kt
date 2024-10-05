package de.rpjosh.rpout.android.complication

import android.util.Log
import androidx.wear.watchface.complications.data.ComplicationType
import androidx.wear.watchface.complications.data.ShortTextComplicationData
import androidx.wear.watchface.complications.datasource.ComplicationRequest
import androidx.wear.watchface.complications.data.ComplicationText
import androidx.wear.watchface.complications.datasource.SuspendingComplicationDataSourceService
import androidx.wear.watchface.complications.data.PlainComplicationText
import androidx.wear.watchface.complications.data.ComplicationData
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.controller.MetricController
import java.util.Locale

class StepComplicationDataService: SuspendingComplicationDataSourceService() {

    companion object {
        val TAG = "RPout-Logger"
    }

    override fun getPreviewData(type: ComplicationType): ComplicationData? {
        return when (type) {
            ComplicationType.SHORT_TEXT ->
                ShortTextComplicationData.Builder(
                    text = PlainComplicationText.Builder(text = "1.456").build(),
                    contentDescription = PlainComplicationText.Builder(text = "Daily step count").build()
                ).setTapAction(null).build()
            else -> null
        }
    }

    override suspend fun onComplicationRequest(request: ComplicationRequest): ComplicationData? {

        // Initialize app
        val app = Singleton.getAppSec()
        val metricController = app.injection.inject(MetricController::class.java, null, false)
        val stepsToday = metricController.getStepCountToday()

        return when (request.complicationType) {

            ComplicationType.SHORT_TEXT -> ShortTextComplicationData.Builder(
                text = PlainComplicationText.Builder(text = String.format(Locale.getDefault(), "%,d", stepsToday)).build(),
                contentDescription = PlainComplicationText.Builder(text = "Daily step count").build()
            ).build()

            else -> {
                if (Log.isLoggable(TAG, Log.WARN)) {
                    Log.w(TAG, "Unexpected complication type ${request.complicationType}")
                }
                null
            }
        }
    }

}
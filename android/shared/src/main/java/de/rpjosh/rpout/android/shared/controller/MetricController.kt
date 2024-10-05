package de.rpjosh.rpout.android.shared.controller

import de.rpjosh.rpout.android.shared.api.RPoutAPI
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.persistence.Database
import de.rpjosh.rpout.android.shared.persistence.MetricDao
import de.rpjosh.rpout.android.shared.persistence.UserDao
import de.rpjosh.rpout.android.shared.services.Logger
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.time.LocalDateTime

class MetricController: BaseDataController() {

    @Inject( parameters = ["MetricController"])
    private lateinit var logger: Logger

    @Inject private lateinit var db: Database

    fun dao(): MetricDao {
        return db.metricDao()
    }

    /** Adds the provided step value into the local database */
    fun addStep(step: Step) {
        db.metricDao().insert(step)
    }

    /** Synchronizes all locally cached steps to the server */
    @Synchronized
    fun synchronizeSteps(): Boolean {

        // Get all unsynced steps
        val unsyncedSteps = db.metricDao().getUnsyncedSteps()
        if (unsyncedSteps.isEmpty()) {
            logger.log("d", "No steps to sync are available")
            return true
        }

        // Push them to the server
        try {
            val call = apiClient.getRetrofitService(RPoutAPI::class.java).postSteps(unsyncedSteps)
            val response = getResponse(call)
            val result = response.body()
            logger.log("d", "Pushed ${result?.storedCount} steps (dropped ${result?.droppedCount} steps)")
        } catch (ex: Exception) {
            logger.log("e", ex, "Failed to push steps")
            return false
        }

        // Update all uploaded steps
        unsyncedSteps.forEach { it.wasSynchronized = true }
        db.metricDao().updateSteps(unsyncedSteps)

        return true
    }

    /** Returns the current (per day) step count */
    suspend fun getStepCountToday(): Int {
        val now =  LocalDateTime.now()
        val secondsSinceMidnight = now.hour * 60 * 60 + now.minute * 60 + now.second

        return withContext(Dispatchers.IO) {
            dao().getStepsSince(secondsSinceMidnight)
        }
    }

}
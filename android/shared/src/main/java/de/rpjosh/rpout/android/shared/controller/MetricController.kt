package de.rpjosh.rpout.android.shared.controller

import de.rpjosh.rpout.android.shared.api.RPoutAPI
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.PaiDay
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.persistence.Database
import de.rpjosh.rpout.android.shared.persistence.MetricDao
import de.rpjosh.rpout.android.shared.persistence.UserDao
import de.rpjosh.rpout.android.shared.services.Logger
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZonedDateTime
import java.util.TimeZone

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

    /** Synchronizes and updates the PAI values earned in the last seven days with the server */
    @Synchronized
    fun synchronizePai(): Boolean {
        try {
            val call = apiClient.getRetrofitService(RPoutAPI::class.java).getPaiValues()
            val response = getResponse(call)
            val result = response.body() ?: throw Exception("No response body received")

            // Update (merge) values into database
            db.metricDao().insertPaiProgression(result.progression)

            logger.log("d", "Updated PAI values of the last seven days (current score = ${result.score})")
            return true
        } catch (ex: Exception) {
            logger.log("e", ex, "Failed to synchronize PAI values")
            return false
        }
    }

    /** Returns the current (per day) step count */
    suspend fun getStepCountToday(): Int {
        val now =  LocalDateTime.now()
        val secondsSinceMidnight = now.hour * 60 * 60 + now.minute * 60 + now.second

        return withContext(Dispatchers.IO) {
            dao().getStepsSince(secondsSinceMidnight)
        }
    }

    /**
     * Returns the PAI progression of the last seven days. Missing values
     * are automatically replaced by "0", when the last received PAI value is
     * not mor than five days ago. Otherwise an empty list is returend
     */
    fun getPaiProgression(): List<PaiDay> {
        // Get current day index
        var currentTime = LocalDateTime.now()
        currentTime = LocalDateTime.of(currentTime.year, currentTime.month, currentTime.dayOfMonth + 1, 0, 0, 0)
        val currentDayIndex = (currentTime.toEpochSecond(ZonedDateTime.now().offset) + ZonedDateTime.now().offset.totalSeconds) / (24*60*60)

        // Get PAI data from DB
        val db = dao().getPaiProgression()

        // We return an empty PAI value if the last fetched value is more than five days ago
        if (db.isEmpty() || db.size < 7) return emptyList()
        val missingDays = currentDayIndex - db.last().dayIndex
        if (missingDays >= 5 ) return emptyList()

        // Build own list with empty values inserted
        val rtc = mutableListOf<PaiDay>()
        for (i in missingDays.toInt() until missingDays.toInt() + 7 step 1) {
            if (i > db.size) {
                // We don't have a value for this one (condition should always be i >= 7)
                // => add empty one

                // Calculate the PAI value at that day
                var sum = 0
                for (a in i - 6 until i step 1) {
                    sum += db[a].earned
                }

                rtc.add(PaiDay(
                    dayIndex = (i + 7 - i),
                    value = sum, earned = 0,
                    weekdayIndex = db[i-7].weekdayIndex,
                    weekdayAbbrevation = db[i-7].weekdayAbbrevation
                ))
            } else {
                rtc.add(db[i])
            }
        }

        return rtc
    }

}
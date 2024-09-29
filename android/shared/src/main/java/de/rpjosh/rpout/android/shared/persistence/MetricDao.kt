package de.rpjosh.rpout.android.shared.persistence

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.Query
import androidx.room.Update
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.models.User

@Dao
interface MetricDao {

    /**
     * Creates a new step entry in the database
     */
    @Insert
    fun insert(step: Step)

    @Query("SELECT * FROM steps WHERE wasSynchronized = 0")
    fun getUnsyncedSteps(): List<Step>

    /**
     * Returns the last timestamp (unix, seconds), when the user
     * reached the provided step goal.
     *
     * The latest starting time for a series is returned
     */
    @Query(
        value = """
            SELECT
                -- This is by design! Only select the last active time from the first data point
                t1.endUnix
                -- MAX(t1.startUnix) AS endUnix,
                -- dateTime(t1.startUnix, 'unixepoch') AS "start_date",
                -- dateTime(MAX(t2.startUnix), 'unixepoch') AS "end_date",
                -- SUM(t2.count) + t1.count AS total_steps
            FROM steps t1
            -- Join itself with all other step values 
            JOIN steps t2 ON t2.startUnix > t1.startUnix
            WHERE t1.startUnix > strftime('%s', 'now') - 60 * 60 * 4
            GROUP BY t1.startUnix
             -- Only select datasets which do have more than 150 steps
            HAVING SUM(t2.count) + t1.count > :threshold
            ORDER BY t1.startUnix DESC
            -- Only the last time is relevant (when the user reached the target rate)
            LIMIT 1
            """
    )
    fun getLastTimeGoalReached(threshold: Int): Long?

    /**
     * Returns the tracked steps in the last x seconds
     */
    @Query("SELECT SUM(count) from steps WHERE startUnix > strftime('%s', 'now') - :offsetSeconds")
    fun getStepsSince(offsetSeconds: Int): Int

    @Update
    fun updateSteps(steps: List<Step>)

}
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

    @Update
    fun updateSteps(steps: List<Step>)

}
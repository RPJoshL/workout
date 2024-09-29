package de.rpjosh.rpout.android.shared.persistence

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import de.rpjosh.rpout.android.shared.models.Version
import de.rpjosh.rpout.android.shared.models.WorkoutType

@Dao
interface WorkoutDao {

    // Workout types
    @Query("SELECT * FROM workout_type")
    fun getAllTypes(): List<WorkoutType>
    @Query("SELECT * FROM workout_type WHERE id = :id")
    fun getType(id: Long): WorkoutType?

    @Query("DELETE FROM workout_type")
    fun deleteAllTypes()

    @Insert
    fun insertTypes(types: List<WorkoutType>)

    @Update
    fun updateType(type: WorkoutType)

    @Query("SELECT * FROM version LIMIT 1")
    fun getVersions(): Version?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    fun insertTypeVersion(version: Version)
}
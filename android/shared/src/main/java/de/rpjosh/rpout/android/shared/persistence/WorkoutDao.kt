package de.rpjosh.rpout.android.shared.persistence

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Update
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
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

    // Workout
    @Query("SELECT * FROM gpsworkout WHERE wasSynchronized = 0 AND isFinished = 1")
    fun getUnsyncedWorkouts(): List<GpsWorkout>
    @Query("SELECT * FROM gpsWorkoutStep WHERE workoutId = :workoutId")
    fun getWorkoutPoints(workoutId: Long): List<GpsWorkoutPoint>
    @Query("SELECT type FROM ( SELECT type, MAX(startTime) as \"time\" FROM gpsWorkout group by type ORDER BY \"time\" DESC LIMIT 6)")
    fun getLastWorkoutTypes(): List<Long>
    @Query("SELECT * FROM gpsWorkout WHERE wasSynchronized = 1 AND startTime > strftime('%s', 'now') - 60 * 60 * 12 ORDER BY startTime DESC LIMIT 3")
    fun getMergableWorkouts(): List<GpsWorkout>
    @Query("SELECT * FROM gpsWorkout WHERE id = :id")
    fun getWorkout(id: Long): GpsWorkout
    @Query("SELECT * FROM gpsWorkout WHERE serverId = :id")
    fun getWorkoutByServerId(id: Long): GpsWorkout
    @Query("DELETE FROM gpsWorkout WHERE id = :id")
    fun deleteWorkout(id: Long)

    @Insert
    fun insertGpsWorkout(gpsWorkout: GpsWorkout): Long
    @Update
    fun updateWorkout(gpsWorkout: GpsWorkout)

    @Insert
    fun insertGpsWorkoutPoints(gpsWorkoutPoint: List<GpsWorkoutPoint>)
}
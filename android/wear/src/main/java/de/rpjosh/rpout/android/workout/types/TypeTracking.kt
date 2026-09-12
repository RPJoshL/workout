package de.rpjosh.rpout.android.workout.types

import android.content.Context
import androidx.health.services.client.data.ExerciseUpdate
import de.rpjosh.rpout.android.shared.models.ActivityType
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.workout.WorkoutManager

// TypeTracker defines a common interface to calculate custom, type specific data
interface TypeTracker {
    fun onStart(context: Context, type: WorkoutType)

    fun onPause()
    fun onResume()
    fun onEnd()

    /** Called when GPS points are processed. The tracker should write its data directly into the workout */
    fun onProcessMetrics(workout: GpsWorkout, summary: WorkoutSummary, update: ExerciseUpdate)
}

class TypeTracking {
    companion object {
        fun getTracker(context: Context, type: WorkoutType, manager: WorkoutManager): TypeTracker? {
            return when(ActivityType.fromInt(type.id.toInt())) {
                ActivityType.TYPE_PUMP_FOILING -> FoilingMetrics(context, type, manager)
                else -> {
                    null
                }
            }
        }
    }
}
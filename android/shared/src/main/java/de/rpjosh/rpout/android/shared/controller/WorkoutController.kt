package de.rpjosh.rpout.android.shared.controller

import de.rpjosh.rpout.android.shared.api.RPoutAPI
import de.rpjosh.rpout.android.shared.exceptions.ServerException
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
import de.rpjosh.rpout.android.shared.models.Version
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.persistence.Database
import de.rpjosh.rpout.android.shared.persistence.WorkoutDao
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.Tr

class WorkoutController: BaseDataController() {

    @Inject( parameters = ["WorkoutController"])
    private lateinit var logger: Logger

    @Inject
    private lateinit var db: Database

    fun dao(): WorkoutDao {
        return db.WorkoutDao()
    }

    /**
     * Returns all available workout types of RPout.
     *
     * The provided version name (like '1.0.3') is used to identify
     * if the workout types has to be fetched again from the server to
     * add new ones
     */
    fun getWorkoutTypes(versionName: String, forceUpdate: Boolean = false): List<WorkoutType> {
        // Get existing types / version code
        val existingTypes = db.WorkoutDao().getAllTypes()
        val existingVersion = db.WorkoutDao().getVersions()

        // Check if new workout types has to be fetched
        if (existingVersion == null || existingTypes.isEmpty() || existingVersion.typeVersion != versionName || forceUpdate) {
            logger.log("i", "Downloading new workout types from API")

            try {
                val call = apiClient.getRetrofitService(RPoutAPI::class.java).getWorkoutTypes()
                val response = getResponse(call)
                val result = response.body()

                // Copy settings from existing types
                if (existingTypes.isNotEmpty()) {
                    result?.forEach { newType ->
                        val oldSettings = existingTypes.find { it.id == newType.id }
                        oldSettings?.copySettingsTo(newType)
                    }
                }

                // Update types in db
                db.WorkoutDao().deleteAllTypes()
                db.WorkoutDao().insertTypes(result ?: emptyList())

                // Update version
                val version = existingVersion ?: Version(typeVersion = versionName)
                version.typeVersion = versionName
                db.WorkoutDao().insertTypeVersion(version)

                return result ?: emptyList()
            } catch (ex: Exception) {
                logger.log("d", ex, "Failed to get workout types")
                if (existingTypes.isEmpty()) {
                    responseView.displayStatic(Tr.get("workout_noTypesAvailable"))
                } else {
                    responseView.displayError(Tr.get("workout_noTypesAvailable"))
                }

                return emptyList()
            }
        }

        // Return existing types
        return existingTypes
    }

    /** Synchronizes all locally cached workouts to the server */
    @Synchronized
    fun synchronizeWorkouts(): Boolean {

        // Get all unsynced workouts
        val unsyncedWorkouts = db.WorkoutDao().getUnsyncedWorkouts()
        unsyncedWorkouts.forEachIndexed{i, w ->
            unsyncedWorkouts[i].points = db.WorkoutDao().getWorkoutPoints(w.id).toMutableList()
        }
        if (unsyncedWorkouts.isEmpty()) {
            logger.log("d", "No workouts to sync are available")
            return true
        }

        // Push them to the server. We don't have a bulk endpoint because there shouldn't be so much workouts...
        var allSuccess = true
        unsyncedWorkouts.forEach {
            if (pushWorkout(it) == null) allSuccess = false
        }
        return allSuccess
    }

    /**
     * Pushes a single workout to the server and returns the calculated workout
     * summary from the server.
     *
     * If there is an error, null is returned. Otherwise the synchronized flag within the db is updated
     */
    fun pushWorkout(workout: GpsWorkout): WorkoutSummary? {
        logger.log("d", "Starting to push workout (#${workout.id}) with ${workout.points.size} points")

        val setWorkoutSynced = { serverId: Long ->
            workout.wasSynchronized = true
            workout.serverId = serverId
            db.WorkoutDao().updateWorkout(workout)
        }

        // 3 Points are at least required for the server
        if (workout.points.size < 3) {
            logger.log("w", "Received a workout with less than 3 GPS points. Deleting it")
            db.WorkoutDao().deleteWorkout(workout.id)
            return WorkoutSummary()
        }

        try {
            ensureConnection(false)

            val call = apiClient.getRetrofitService(RPoutAPI::class.java).postWorkout(workout)
            val response = getResponse(call)
            val result = response.body()

            // Update synchronized flag
            if (result != null) {
                logger.log("d", "Pushed workout (${workout.id} -> ${result.id})")
                setWorkoutSynced(result.id)
            } else {
                logger.log("d", "Failed to push workout ${workout.id}. Received no response")
            }

            return result
        } catch (ex: ServerException) {
            logger.log("e", ex, "Failed to push workout (internal id = ${workout.id})")

            // Also mark already synced workouts as "synced"
            if (ex.response.code == 409) {
                ex.response.headers.get("Existing-Workout-Id")?.let {
                    try {
                        val serverId = it.toLong()
                        logger.log("d", "Workout with internal id ${workout.id} was already synced as $serverId")

                        // Don't mark it as an error
                        setWorkoutSynced(serverId)
                        return WorkoutSummary()
                    } catch (ex: Exception) {
                        logger.log("e", ex, "Failed to convert '$it' to a number")
                    }
                }
            }

            return null
        } catch (ex: Exception) {
            logger.log("e", ex, "Failed to push workout (internal id = ${workout.id})")
            return null
        }
    }

    /**
     * Merges two separate workouts into a single one. The baseID has to be before
     * newId in time
     */
    fun mergeWorkout(baseId: Long, newId: Long): Boolean {
        try {
            val call = apiClient.getRetrofitService(RPoutAPI::class.java, false).mergeWorkouts(baseId, newId)
            val response = getResponse(call)

            // Update server ID within internal database so it's correct if the user
            // want's to merge more workouts
            val existing = dao().getWorkoutByServerId(newId)
            existing.serverId = baseId
            dao().updateWorkout(existing)

            return true
        } catch (ex: Exception) {
            logger.log("e", ex, "Failed to merge workouts ($baseId with $newId)")
            return false
        }
    }

}
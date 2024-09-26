package de.rpjosh.rpout.android.shared.controller

import de.rpjosh.rpout.android.shared.api.RPoutAPI
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.Version
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.persistence.Database
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.Tr

class WorkoutController: BaseDataController() {

    @Inject( parameters = ["WorkoutController"])
    private lateinit var logger: Logger

    @Inject
    private lateinit var db: Database

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

}
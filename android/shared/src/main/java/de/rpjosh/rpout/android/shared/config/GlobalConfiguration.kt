package de.rpjosh.rpout.android.shared.config

import android.util.Log
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.User
import de.rpjosh.rpout.android.shared.services.Logger
import java.io.File

/**
 * GlobalConfiguration contains global and generic configuration options
 * that are used / needed across the whole application
 */
class GlobalConfiguration {

    /** The currently authorized user with all settings, or null if the user is not logged in */
    var user: User? = null

    /** Short name of this application */
    val applicationName: String = "RPout"

    /** The full path of the applications file directory */
    var appDir: String? = null

    @Inject( parameters = ["GlobalConfiguration"])
    private var logger: Logger? = null

    /** Returns the log level to use for all log messages */
    fun getLogLevel(): Logger.LEVEL {
        // Log with debug level by default
        if (user == null) return Logger.LEVEL.DEBUG

        val lvl = user!!.logLevel
        return when (lvl) {
            Logger.LEVEL.DEBUG.value -> Logger.LEVEL.DEBUG
            Logger.LEVEL.INFO.value -> Logger.LEVEL.INFO
            Logger.LEVEL.WARNING.value -> Logger.LEVEL.WARNING
            Logger.LEVEL.ERROR.value -> Logger.LEVEL.ERROR
            Logger.LEVEL.ERROR_PRINT.value -> Logger.LEVEL.ERROR_PRINT
            else -> {
                Log.d("Workout", "Received unknown log level")
                return Logger.LEVEL.INFO
            }
        }
    }

    /** Returns this Apps configuration directory with the provided sub path (like '/root/sub/') */
    fun getAppDir(subPath: String?): String? {
        if (appDir === null) return null

        var basePath = appDir ?: ""
        if (!basePath.endsWith("/")) basePath += "/"

        if (subPath != null) {
            basePath += subPath
            createDirectory(basePath)
            if (!basePath.endsWith("/")) basePath +="/"
        }

        return basePath
    }

    /** Creates the specified directory (if it does not exist) */
    private fun createDirectory(directory: String) {
        try {
            val currentDirectory = File(directory)
            if (!currentDirectory.exists()) File(directory).mkdirs()
        }
        catch (ex: Exception) {
            logger?.log("e", "Failed to create directory '${directory}'")
        }

    }

}
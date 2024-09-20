package de.rpjosh.rpout.android.shared.controller

import com.google.gson.Gson
import de.rpjosh.rpout.android.shared.api.RPoutAPI
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration
import de.rpjosh.rpout.android.shared.exceptions.AuthenticationException
import de.rpjosh.rpout.android.shared.exceptions.OfflineException
import de.rpjosh.rpout.android.shared.exceptions.UnknownServerException
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.ApiKey
import de.rpjosh.rpout.android.shared.models.User
import de.rpjosh.rpout.android.shared.persistence.Database
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.ResponseViewInterface
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.services.Validator
import de.rpjosh.rpout.android.shared.services.WearSynchronizationInterface

class UserController: BaseDataController() {

    @Inject( parameters = ["UserController"])
    private lateinit var logger: Logger

    @Inject private lateinit var db: Database
    @Inject private lateinit var response: ResponseViewInterface
    @Inject private lateinit var config: GlobalConfiguration
    @Inject private lateinit var deviceSync: WearSynchronizationInterface

    /**
     * Obtains a new API key for the provided user and updates the settings accordingly
     *
     * @param   serverURL       Base path of the server (like 'https://workout.rpjosh.de/')
     * @param   username        Username to login with
     * @param   password        Password for "username"
     *
     * @return  Whether the login was successfully
     */
    fun getApiKey(serverURL: String, username: String, password: String, key: ApiKey): Boolean {
        // Throw an error if user is still logged in
        if (config.user !== null) {
            logger.log("ee", Tr.get("user_stillLoggedIn"))
            return false
        }

        // Validate values
        Validator()
            .notBlank(serverURL, "user_login_userPassRequired")
            .notBlank(username, "user_login_userPassRequired")
            .notBlank(password, "user_login_userPassRequired")
            .url(serverURL, "user_login_urlInvalid")
            .validate()?.let {
                response.displayError(it)
                return false
            }

        // Force some (default) values
        key.darkTheme = 1

        // Reset any errors (because the user can provide different credentials or a URL)
        resetUserStatus()

        // Execute request to get a new key
        var gotKey: ApiKey? = null
        try {
            val call = apiClient.getRetrofitService(RPoutAPI::class.java, serverURL, username, password).createApiKey(key)
            val response = getResponse(call)
            gotKey = response.body()
        } catch (ex: Exception) {
            logger.log("d", ex)
            response.displayError(ex.message)
            return false
        }

        // Store the create API key in the database
        val usr = User(gotKey!!.userId, serverURL, username, gotKey.key, gotKey.id)
        db.userDao().login(usr)

        // Also update user reference
        globalConfig.user = usr

        // Sync settings to wearable device
        synchronizeSettings()

        return true
    }

    /**
     * Deletes the currently used API key and removes the user from the database and app
     */
    fun logout(onSuccess: () -> Unit) {
        try {
            ensureConnection(true)
        } catch (ex: Exception) {
            response.displayError(ex.message)
            return
        }

        // Wear OS device has to be connected (to delete the user reference also for it)
        deviceSync.sendTextMessage(MessageType.SETTINGS, "DELETE") {
            Thread {
                try {
                    val call = apiClient.getRetrofitService(RPoutAPI::class.java, false).deleteApiKey(-1)
                    val response = getResponse(call)
                } catch (ex: OfflineException) {
                    response.displayError(ex.message)
                    return@Thread
                } catch (ex: UnknownServerException) {
                    logger.log("d", ex)
                    response.displayError(ex.message)
                    return@Thread
                } catch (ex: Exception) {
                    // Ignore other error types like authentication exception, ...
                }

                // Delete from database and config
                db.userDao().logout()
                globalConfig.user = null
                onSuccess()
            }.start()
        }
    }

    /**
     * Updates "updatable" settings of the user
     */
    fun updateSettings(logLevel: Int = config.user?.logLevel ?: Logger.LEVEL.INFO.value) {
        if (globalConfig.user == null) return

        // Update settings
        val newUser = globalConfig.user!!
        newUser.logLevel = logLevel

        // Store update in db
        db.userDao().update(newUser)

        // Sync settings to wearable device
        synchronizeSettings()
    }

    /**
     * Fetches the full details of the API key with which the user is currently logged in.
     *
     * It throws the underlying exception to get the reason why a request failed.
     * You can use the exceptions message for displaying a status to the user
     */
    fun getDetailsOfLogin(): ApiKey {
        if (config.user == null) {
            throw Exception(Tr.get("status_notLoggedIn"))
        }

        // Execute request to get the current API key
        try {
            val call = apiClient.getRetrofitService(RPoutAPI::class.java).getApiKey()
            val response = getResponse(call)
            return response.body()!!
        } catch (ex: UnknownServerException) {
            logger.log("d", ex)
            throw Exception(Tr.get("status_unknown"))
        } catch (ex: AuthenticationException) {
            throw Exception(Tr.get("status_authFailed"))
        } catch (ex: OfflineException) {
            throw Exception(ex.message)
        } catch (ex: Exception) {
            logger.log("d", ex)
            throw Exception(Tr.get("status_unknown"))
        }
    }

    /** Synchronizes the current settings (of the globalConfig) to the wearable device */
    public fun synchronizeSettings() {
        var message = "DELETE"
        if (globalConfig.user != null){
            message = Gson().toJson(globalConfig.user)
        }

        deviceSync.sendTextMessage(
            MessageType.SETTINGS, message
        ) {}
    }

    /**
     * Updates the settings with the provided one. If the provided configuration is
     * null, the settings are deleted.
     *
     * This function should only be called from WearOS!
     */
    public fun setSettingsWearOs(user: User?) {
        // Update configuration in db
        if (user == null) {
            db.userDao().logout()
        } else {
            db.userDao().logout()
            db.userDao().login(user)
        }

        // Update global configuration
        config.user = user
    }

}
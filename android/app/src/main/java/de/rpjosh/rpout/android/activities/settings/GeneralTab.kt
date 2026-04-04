package de.rpjosh.rpout.android.activities.settings

import android.app.Activity
import android.content.Intent
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.core.content.FileProvider
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.WearMessageReceiver
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.SelectOption
import de.rpjosh.rpout.android.activities.theme.Spinner
import de.rpjosh.rpout.android.helper.VersionHelper
import de.rpjosh.rpout.android.services.ResponseView
import de.rpjosh.rpout.android.services.WearSynchronization
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.Logger.LEVEL
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.Tr
import java.io.File
import java.nio.file.Files

class GeneralTab: Tab, WearMessageReceiver {

    override lateinit var activity: Activity
    @Inject lateinit var userController: UserController
    @Inject lateinit var responseView: ResponseView
    @Inject lateinit var deviceSync: WearSynchronization
    @Inject lateinit var globalConfiguration: GlobalConfiguration
    @Inject lateinit var workoutController: WorkoutController
    @Inject(parameters = ["SettingsGeneralTab"]) lateinit var logger: Logger

    override fun getLabel(): String {
        return Tr.get("settings_general_label")
    }

    override fun loadData() {
    }

    override fun onPause() {
        Singleton.deRegisterOnWearMessageReceived(this)
    }
    override fun onDestroy() {
        Singleton.deRegisterOnWearMessageReceived(this)
    }

    private fun downloadLogsAndroid() {
        Thread{
            try {
                val logFile = logger.logFile
                logFile.setReadable(true)
                val fileUri = FileProvider.getUriForFile(activity.baseContext, "de.rpjosh.rpout.android.fileprovider", logFile)
                val intent = Intent(Intent.ACTION_VIEW).apply {
                    setDataAndType(fileUri, "text/plain")
                    addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                }

                activity.startActivity(intent)
            } catch (ex: Exception) {
                logger.log("e", ex)
            }
        }.start()
    }

    private fun downloadLogsWearable() {
        // Register receiver
        Singleton.registerOnWearMessageReceived(this)
        deviceSync.sendTextMessage(
            type = MessageType.LOG_REQUEST,
            message ="REQUEST",
            onSuccess = {},
            onError = { deviceSync.showNotConnectedMessage() }
        )
    }

    private fun syncSettings() {
        userController.synchronizeSettings()
    }
    private fun syncData() {
        deviceSync.sendTextMessage(
            MessageType.SYNC_DATA, "",
            onError = { deviceSync.showNotConnectedMessage() }
        ) {
            responseView.displaySuccess(Tr.get("sync_successfully"))
        }
    }
    private fun syncWorkoutType() {
        deviceSync.sendTextMessage(
            MessageType.SYNC_DATA_WORKOUT, "",
            onError = { deviceSync.showNotConnectedMessage() }
        ){
            responseView.displaySuccess(Tr.get("sync_successfully"))
        }

        Thread{
            workoutController.getWorkoutTypes(VersionHelper.getVersionName(), true)
        }.start()
    }

    @Composable
    override fun Content() {
        return ContentInternal(
            onLogWearable = { downloadLogsWearable() },
            onLogAndroid = { downloadLogsAndroid() },
            onSettingsSync = { syncSettings() },
            onDataSync = { syncData() },
            onWorkoutTypeSync = { syncWorkoutType() },
            onLogLevelChanged = {
                Thread{ userController.updateSettings(logLevel = it) }.start()
            }
        )
    }

    @Composable
    fun ContentInternal(onLogWearable: () -> Unit, onLogAndroid: () -> Unit, onSettingsSync: () -> Unit, onDataSync: () -> Unit, onWorkoutTypeSync: () -> Unit, onLogLevelChanged: (id: Int) -> Unit) {

        // Options for available log levels
        val logLevels = arrayListOf(
            SelectOption(LEVEL.DEBUG.value, "Debug"),
            SelectOption(LEVEL.INFO.value, "Info"),
            SelectOption(LEVEL.WARNING.value, "Warning"),
            SelectOption(LEVEL.ERROR.value, "Error")
        )
        val defaultLogLevel = if(!::globalConfiguration.isInitialized) logLevels[1]
            else logLevels.find { it.id == (globalConfiguration.user?.logLevel ?: 20) } ?: logLevels[1]

        Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {

            // Log section
            SectionText(stringResource(R.string.settings_general_logging_section), noExtraPaddingTop = true)
            Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(10.dp)) {
                Text(text = stringResource(R.string.settings_general_log_level) + ":", color = RPoutTheme.colors.text)
                Spinner(
                    options = logLevels,
                    preselected = defaultLogLevel,
                    onSelectionChanged = {onLogLevelChanged(it.id)},
                )
            }
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text(
                    text = stringResource(R.string.settings_general_logging_download) + ":",
                    color = RPoutTheme.colors.text, textAlign = TextAlign.Left,
                )
                Row(verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.SpaceEvenly, modifier = Modifier.fillMaxWidth()) {
                    Button(onClick = {onLogWearable()}) { Text("Wearable") }
                    Button(onClick = {onLogAndroid()}) { Text("Android") }
                }
            }

            // Synchronization section
            SectionText(stringResource(R.string.settings_general_synchronisation_section))
            Row(verticalAlignment = Alignment.CenterVertically) {
                Row(horizontalArrangement = Arrangement.SpaceEvenly, modifier = Modifier.fillMaxWidth().padding(4.dp)) {
                    Button(onClick = { onDataSync() }) { Text(stringResource(R.string.settings_general_sync_data)) }
                    Button(onClick = { onSettingsSync() }) { Text(stringResource(R.string.settings_general_sync_settings)) }
                    Button(onClick = { onWorkoutTypeSync() }) { Text(stringResource(R.string.settings_general_sync_workout_types)) }
                }
            }

        }

    }

    override fun onWearMessageReceived(type: MessageType, data: String) {
        when(type) {
            MessageType.LOG_RESPONSE -> {
                Singleton.deRegisterOnWearMessageReceived(this)

                Thread{
                    try {
                        // Write received messages to a temp file
                        val logFile = File.createTempFile("log-file", ".txt")
                        logFile.setReadable(true)

                        // Write received logs to it
                        logFile.writeText(data)

                        // Open file to open
                        val fileUri = FileProvider.getUriForFile(activity.baseContext, "de.rpjosh.rpout.android.fileprovider", logFile)
                        val intent = Intent(Intent.ACTION_VIEW).apply {
                            setDataAndType(fileUri, "text/plain")
                            addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                        }
                        activity.startActivity(intent)
                    } catch (ex: Exception) {
                        logger.log("e", ex)
                    }
                }.start()
            }

            else -> {}
        }
    }
}

@Preview(showBackground = true, device = "id:pixel_7", showSystemUi = true)
@Composable
fun GeneralTabPreview() {
    RPoutTheme(false) {
        GeneralTab().Content()
    }
}
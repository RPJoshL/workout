package de.rpjosh.rpout.android.helper

import android.Manifest
import android.app.Activity
import android.app.AlarmManager
import android.app.AlertDialog
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.os.PowerManager
import android.provider.Settings
import android.util.Log
import androidx.activity.result.ActivityResultLauncher
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.core.content.IntentCompat
import androidx.core.content.PackageManagerCompat
import androidx.core.content.UnusedAppRestrictionsConstants.DISABLED
import androidx.core.content.UnusedAppRestrictionsConstants.ERROR
import com.google.common.util.concurrent.ListenableFuture
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.services.Tr

class PermissionHelper(
    val context: Context
) {

    fun askForBatteryOptimization() {
        showDialog(
            "permissions_batteryInformation_title",
            "permissions_batteryInformation_description",
            Settings.ACTION_IGNORE_BATTERY_OPTIMIZATION_SETTINGS
        )
    }

    fun askForScheduleExact() {
        showDialog(
            "permissions_schedule_exact_permission_title",
            "permissions_schedule_exact_permission_description",
            Settings.ACTION_REQUEST_SCHEDULE_EXACT_ALARM
        )
    }

    fun askForDrawOverlay() {
        // WearOS has no "Settings.ACTION_MANAGE_OVERLAY_PERMISSION" screen...
        val intent = Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS, Uri.parse("package:${context.packageName}"))

        showDialog(
            "permissions_drawOverlay_title",
            "permissions_drawOverlay_description",
            "",
            intent
        )
    }

    fun askForDisableUnusedAppRestrictions() {
        val future: ListenableFuture<Int> = PackageManagerCompat.getUnusedAppRestrictionsStatus(context)
        future.addListener(
            {
                val appRestrictionsStatus = future.get()
                when (appRestrictionsStatus) {
                    // Status could not be fetched
                    ERROR -> { }
                    // Restrictions have been disabled by the user
                    DISABLED -> { }

                    // Restriction is enabled
                    else -> {
                        val intent = IntentCompat.createManageUnusedAppRestrictionsIntent(context, context.packageName)
                        showDialog(
                            "permissions_unused_title",
                            "permissions_unused_description",
                            "",
                            intent
                        )
                    }
                }
            },
            ContextCompat.getMainExecutor(context)
        )
    }

    /**
     * Shows an alert dialog asking for the given permission by opening the settings dialog
     * when the user clicks on the "Ok" Button.
     *
     * @param   title   Translation id of the title
     * @param   message Translation id of the message
     * @param   setting Setting dialog to open
     */
    private fun showDialog(title: String, message: String, setting: String, intent: Intent? = null) {
        val alertDialogBuilder = AlertDialog.Builder(context)
        alertDialogBuilder.setTitle(Tr.get(title))
        alertDialogBuilder.setMessage(Tr.get(message))
        alertDialogBuilder.setPositiveButton(android.R.string.ok) { _, _ ->
            try {
                context.startActivity(intent ?: Intent(setting))
            } catch (ex: Exception) {
                Singleton.appController.sharedLogger.log(
                    "e",
                    ex,
                    "Unable to prompt a dialog"
                )
            }
        }
        alertDialogBuilder.show()
    }

    /**
     * Checks if the battery optimization is ignored for this app
     */
    fun isBatterOptimizationIgnored(): Boolean {
        val packageName = context.packageName
        val pm = context.getSystemService(Context.POWER_SERVICE) as PowerManager
        return pm.isIgnoringBatteryOptimizations(packageName)
    }

    /**
     * Checks if alarms can be scheduled exactly.
     * A special permission is required for android >= 14
     */
    fun canScheduleExact(): Boolean {
        return (context.getSystemService(Context.ALARM_SERVICE) as AlarmManager).canScheduleExactAlarms()
    }

    fun canDrawOverlays(): Boolean {
        return Settings.canDrawOverlays(context)
    }

}
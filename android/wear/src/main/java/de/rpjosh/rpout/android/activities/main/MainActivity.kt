package de.rpjosh.rpout.android.activities.main

import android.Manifest
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.media.MediaPlayer
import android.net.Uri
import android.os.Bundle
import android.os.VibrationEffect
import android.os.Vibrator
import android.provider.Settings
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.core.content.IntentCompat
import androidx.core.content.PackageManagerCompat
import androidx.core.content.UnusedAppRestrictionsConstants.DISABLED
import androidx.core.content.UnusedAppRestrictionsConstants.ERROR
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.compose.material.Button
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.Text
import androidx.wear.compose.material.TimeText
import androidx.wear.tooling.preview.devices.WearDevices
import com.google.common.util.concurrent.ListenableFuture
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.helper.PermissionHelper
import de.rpjosh.rpout.android.services.StepRecordingService
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration


class MainActivity : ComponentActivity() {

    private lateinit var globalConfig: GlobalConfiguration
    private lateinit var permissionHelper: PermissionHelper

    // Androids permission contract helper to ask for permissions easily
    private val requestPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted: Boolean ->
            if (isGranted) {
                Toast.makeText(this, "Permission granted", Toast.LENGTH_SHORT).show()
                // Ask for other permissions / start service
                checkAndRequestPermission()
            } else {
                Toast.makeText(this, "Permission not granted! The app will not work as expected", Toast.LENGTH_SHORT).show()
                Singleton.appController.sharedLogger.log("w", "User did not grant all rights. The app won't work correctly")
            }
        }

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()
        super.onCreate(savedInstanceState)

        setTheme(android.R.style.Theme_DeviceDefault)

        // Initialize app
        initApp()
        permissionHelper = PermissionHelper(this)

        setContent {
            WearApp("Android")
        }

        // Ask for permission
        checkAndRequestPermission()
    }

    /**
     * Checks all required permissions for this apps and ask for it
     * if they were not granted already
     */
    private fun checkAndRequestPermission() {
        val permissions = arrayListOf(
            Manifest.permission.ACTIVITY_RECOGNITION,
            Manifest.permission.POST_NOTIFICATIONS
        )

        // Check and request all permission
        permissions.forEach { p ->
            if (baseContext.checkSelfPermission(p) == PackageManager.PERMISSION_DENIED) {
                // We don't show any additional infos to the user. The permissions are required
                // for the app to work correctly
                if (shouldShowRequestPermissionRationale(p)) {
                    requestPermissionLauncher.launch(p)
                } else {
                    requestPermissionLauncher.launch(p)
                }

                // Stop checking permissions. This function is called again when the
                // user granted us the permissions
                return
            }
        }

        // Special permission we have to ask. The log will still be false
        if (!permissionHelper.canDrawOverlays()) permissionHelper.askForDrawOverlay()
        if (!permissionHelper.canScheduleExact()) permissionHelper.askForScheduleExact()
        // if (!permissionHelper.isBatterOptimizationIgnored()) permissionHelper.askForBatteryOptimization()
        permissionHelper.askForDisableUnusedAppRestrictions()

        Singleton.appController.sharedLogger.log("d", "All permissions are granted")

        // Start login activity if we don't have a user context
        if (globalConfig.user == null) {
            startActivity(Intent(this, NotLoggedInActivity::class.java))
            finish()
        }
    }

    /**
     * Initializes the app controller with all dependencies. This method
     * does block until the APP is fully loaded
     */
    private fun initApp() {
        if (Singleton.getApp() == null) Singleton.app()

        // Inject dependencies
        globalConfig = Singleton.appController.injection.inject(GlobalConfiguration::class.java, null,  false)
    }

}


@Composable
fun WearApp(greetingName: String) {

    val context = LocalContext.current
    RPoutTheme {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colors.background),
            contentAlignment = Alignment.Center
        ) {
            TimeText()
            ActionButton("Starte service") {
                val serviceIntent = Intent(context, StepRecordingService::class.java)
                ContextCompat.startForegroundService(context, serviceIntent)
            }
        }
    }
}

@Composable
fun ActionButton(title: String, onClick: () -> Unit) {
    Button(
        onClick = { onClick() },
        modifier = Modifier
            .fillMaxWidth(0.8f)
            .height(28.dp)
    ) {
        Text(title)
    }
}

@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun DefaultPreview() {
    WearApp("Preview Android")
}
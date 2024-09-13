package de.rpjosh.rpout.android.wear.presentation

import android.Manifest
import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.os.Bundle
import android.util.Log
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.core.app.ActivityCompat
import androidx.core.content.ContextCompat
import androidx.wear.compose.material.Button
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.Text
import androidx.wear.compose.material.TimeText
import androidx.wear.tooling.preview.devices.WearDevices
import com.google.android.gms.fitness.FitnessLocal
import com.google.android.gms.fitness.data.LocalDataType
import de.rpjosh.rpout.android.wear.R
import de.rpjosh.rpout.android.wear.RPout
import de.rpjosh.rpout.android.wear.presentation.theme.RPoutTheme
import de.rpjosh.rpout.android.wear.service.StepRecordingService

class MainActivity : ComponentActivity() {

    // Androids permission contract helper to ask for permissions easily
    private val requestPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted: Boolean ->
            if (isGranted) {
                Toast.makeText(this, "Permission granted", Toast.LENGTH_SHORT).show()
                // Ask for other permissions / start service
                checkAndRequestPermission()
            } else {
                Toast.makeText(this, "Permission not granted! The app will not work as expected", Toast.LENGTH_SHORT).show()
            }
        }

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()

        super.onCreate(savedInstanceState)

        setTheme(android.R.style.Theme_DeviceDefault)

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

        Log.d("Workout", "All permissions were granted")
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
fun Greeting(greetingName: String) {
    Text(
        modifier = Modifier.fillMaxWidth(),
        textAlign = TextAlign.Center,
        color = MaterialTheme.colors.primary,
        text = stringResource(R.string.hello_world, greetingName)
    )
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
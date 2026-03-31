package de.rpjosh.rpout.android.activities.main

import android.content.Intent
import android.content.pm.PackageManager
import android.os.Bundle
import android.util.Log
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.lifecycle.lifecycleScope
import com.google.android.gms.wearable.CapabilityClient
import com.google.android.gms.wearable.PutDataMapRequest
import com.google.android.gms.wearable.Wearable
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.login.LoginActivity
import de.rpjosh.rpout.android.activities.settings.SettingsActivity
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.workout.WorkoutTracking
import de.rpjosh.rpout.android.helper.VersionHelper
import de.rpjosh.rpout.android.services.RealtimeLocationService
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.WorkoutStatus
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {

    @Inject private lateinit var globalConfig: GlobalConfiguration
    @Inject private lateinit var workoutController: WorkoutController

    // Androids permission contract helper to ask for permissions easily
    private val requestPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted: Boolean ->
            if (isGranted) {
                Toast.makeText(this, "Permission granted", Toast.LENGTH_SHORT).show()
                // Ask for other permissions / start service
                checkPermissions()
            } else {
                Toast.makeText(this, "Permission not granted! The app will not work as expected", Toast.LENGTH_SHORT).show()
                Singleton.appController.sharedLogger.log("w", "User did not grant all rights. The app won't work correctly")
            }
        }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Block app until it's initialized
        if (!initApp()) {
            return
        }
        Singleton.appController.activityCreated(this, this)

        // Validate required permissions
        lifecycleScope.launch {
            delay(2000L)
            checkPermissions()
        }

        enableEdgeToEdge()
        setContent {
            RPoutTheme {
                Scaffold(modifier = Modifier.fillMaxSize()) { innerPadding ->
                    Greeting(
                        name = "Android",
                        modifier = Modifier.padding(innerPadding),
                        { sendMessage() },
                        {
                            startActivity(Intent(this, SettingsActivity::class.java))
                        },
                        onOpenWorkout = {
                            startActivity(Intent(this, WorkoutTracking::class.java))
                        }
                    )
                }
            }
        }
    }

    fun sendMessage() {
        // Get message client
        Thread {
            val dataClient = Wearable.getDataClient(this)
            val messageClient = Wearable.getMessageClient(this)
            val capatibilityClient = Wearable.getCapabilityClient(this)

            val putRequest = PutDataMapRequest.create("/auth").apply {
                dataMap.putString("token", "Hello from android")
            }.asPutDataRequest().setUrgent()

            val nodes = capatibilityClient.getCapability("wear", CapabilityClient.FILTER_REACHABLE)
            nodes.addOnSuccessListener { it ->
                Log.d("RPout-Logger", "Got ${it.nodes.size} available nodes")

                it.nodes.forEach{ node ->
                    messageClient.sendMessage(node.id, "/auth", "Token".toByteArray())
                }
            }

            val task = dataClient.putDataItem(putRequest)
            task.addOnCompleteListener {
                if (it.isSuccessful) {
                    Log.d("RPout-Logger", "Successfully transferred")
                } else {
                    Log.d("RPout-Logger", "Transfer failed")
                }
            }

        }.start()
    }

    /**
     * Initializes the app controller with all dependencies. This method
     * does block until the APP is fully loaded
     */
    private fun initApp(): Boolean {
        if (Singleton.getApp() == null) Singleton.app()

        // Inject dependencies
        Singleton.appController.injection.inject(MainActivity::class.java, null,  false, this)

        // Start login activity if we don't have a user context
        if (globalConfig.user == null) {
            startActivity(Intent(this, LoginActivity::class.java))
            finish()
            return false
        }

        // Synchronize additional data
        Thread{
            workoutController.getWorkoutTypes(VersionHelper.getVersionName(), false)
        }.start()

        return true
    }

    override fun onPause() {
        Singleton.getApp()?.activityPaused(this)
        super.onPause()
    }

    override fun onStart() {
        Singleton.getApp()?.activityStarted(this, this)
        super.onStart()
    }

    override fun onDestroy() {
        Singleton.getApp()?.activityDestroyed(this)
        super.onDestroy()
    }

    private fun checkPermissions() {
        val permissions = arrayListOf(
            android.Manifest.permission.ACCESS_FINE_LOCATION,
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

    }

}

@Composable
fun Greeting(
    name: String, modifier: Modifier = Modifier,
    onSendMessage: () -> Unit,
    onSettingsClick: () -> Unit,
    onOpenWorkout: () -> Unit
) {

    val workoutStatus = remember { RealtimeLocationService.status }

    Column {
        Text(
            text = "Hello $name!",
            modifier = modifier
        )
        Button(onClick = onSendMessage) {
            Text("Send message to wear")
        }
        Button(onClick = onSettingsClick) {
            Text("Open settings")
        }

        if (workoutStatus.state.value in listOf(WorkoutStatus.RUNNING, WorkoutStatus.HIGH_SAMPLING, WorkoutStatus.PREPARE)) {
            Button(onClick = onOpenWorkout) {
                Text("Open workout")
            }
        }
    }
}

@Preview(showBackground = true, device = "id:pixel_7", showSystemUi = true)
@Composable
fun GreetingPreview() {
    RPoutTheme {
        Greeting("Android", onSendMessage = {}, onSettingsClick = {}, onOpenWorkout = {})
    }
}

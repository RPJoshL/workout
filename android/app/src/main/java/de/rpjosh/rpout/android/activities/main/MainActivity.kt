package de.rpjosh.rpout.android.activities.main

import android.content.Intent
import android.os.Bundle
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.material3.Button
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import androidx.lifecycle.lifecycleScope
import com.google.android.gms.wearable.CapabilityClient
import com.google.android.gms.wearable.PutDataMapRequest
import com.google.android.gms.wearable.Wearable
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.login.LoginActivity
import de.rpjosh.rpout.android.activities.settings.SettingsActivity
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {

    private lateinit var globalConfig: GlobalConfiguration

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Block app until it's initialized
        initApp()
        Singleton.appController.activityCreated(this, this)


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
    private fun initApp() {
        if (Singleton.getApp() == null) Singleton.app()

        // Inject dependencies
        globalConfig = Singleton.appController.injection.inject(GlobalConfiguration::class.java, null,  false)

        // Start login activity if we don't have a user context
        if (globalConfig.user == null) {
            startActivity(Intent(this, LoginActivity::class.java))
            finish()
        }
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

}

@Composable
fun Greeting(name: String, modifier: Modifier = Modifier, onSendMessage: () -> Unit, onSettingsClick: () -> Unit) {
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
    }
}

@Preview(showBackground = true, device = "id:pixel_7")
@Composable
fun GreetingPreview() {
    RPoutTheme {
        Greeting("Android", onSendMessage = {}, onSettingsClick = {})
    }
}
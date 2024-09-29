package de.rpjosh.rpout.android.activities.main

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import android.util.Log
import android.window.OnBackInvokedDispatcher
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.ambient.AmbientLifecycleObserver
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.Text
import androidx.wear.compose.material.TimeText
import androidx.wear.tooling.preview.devices.WearDevices
import com.google.android.gms.wearable.Wearable
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme

class WorkoutTrackingActivity: ComponentActivity(), AmbientLifecycleObserver.AmbientLifecycleCallback {

    private val ambientObserver = AmbientLifecycleObserver(this, this)
    private val isAmbient = mutableStateOf(false)

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()

        super.onCreate(savedInstanceState)
        lifecycle.addObserver(ambientObserver)
        setTheme(android.R.style.Theme_DeviceDefault)

        // Disallow user to go back
        onBackInvokedDispatcher.registerOnBackInvokedCallback(
            OnBackInvokedDispatcher.PRIORITY_DEFAULT
        ) {
            // Do nothing
        }

        setContent {
            WorkoutTrackingScreen(isAmbient.value)
        }
    }

    override fun onEnterAmbient(ambientDetails: AmbientLifecycleObserver.AmbientDetails) {
        isAmbient.value = true
        super.onEnterAmbient(ambientDetails)
    }

    override fun onExitAmbient() {
        isAmbient.value = false
        super.onExitAmbient()
    }

    override fun onUpdateAmbient() {
        super.onUpdateAmbient()
    }

    override fun onDestroy() {
        super.onDestroy()
        lifecycle.removeObserver(ambientObserver)
    }

}


@Composable
fun WorkoutTrackingScreen(isAmbient: Boolean) {
    val context = LocalContext.current

    RPoutTheme {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colors.background)
                .padding(start = 5.dp, end = 5.dp),
            contentAlignment = Alignment.Center
        ) {
            TimeText()
            Text(
                text = if (isAmbient) "Ambient" else "Normal",
                textAlign = TextAlign.Center
            )
        }
    }
}

@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutTrackingPreview() {
    WorkoutTrackingScreen(false)
}
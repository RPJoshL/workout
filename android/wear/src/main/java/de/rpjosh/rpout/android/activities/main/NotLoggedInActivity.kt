package de.rpjosh.rpout.android.activities.main

import android.content.Intent
import android.os.Bundle
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.compose.material3.MaterialTheme
import androidx.wear.compose.material3.Text
import androidx.wear.compose.material3.TimeText
import com.google.android.gms.wearable.Wearable
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.WearPreview
import de.rpjosh.rpout.android.activities.theme.RPoutTheme

class NotLoggedInActivity : ComponentActivity() {

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()

        super.onCreate(savedInstanceState)
        setTheme(android.R.style.Theme_DeviceDefault)

        Thread {
            val dataClient = Wearable.getDataClient(RPout.getAppContext())
            dataClient.addListener { events ->
                events.forEach {
                    Log.d("RPout-Logger", "Received data event: ${it.dataItem.uri.path}")
                }
            }
        }.start()

        setContent {
            NotLoggedInScreen()
        }
    }

    override fun onResume() {
        super.onResume()

        // Check if user is now logged in is now initialized
        if (Singleton.getAppSec(this).globalConfiguration.user != null) {
            startActivity(Intent(this, MainActivity::class.java))
            finish()
        }
    }
}


@Composable
fun NotLoggedInScreen() {
    val context = LocalContext.current

    RPoutTheme {
        Box(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colorScheme.background)
                .padding(start = 5.dp, end = 5.dp),
            contentAlignment = Alignment.Center
        ) {
            TimeText()
            Text(
                text = stringResource(R.string.activity_notLoggedIn_Banner),
                textAlign = TextAlign.Center
            )
        }
    }
}

@WearPreview
@Composable
fun NotLoggedInPreview() {
    NotLoggedInScreen()
}
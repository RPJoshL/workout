package de.rpjosh.rpout.android.activities.login

import android.content.Intent
import android.os.Build
import android.os.Bundle
import androidx.compose.foundation.layout.WindowInsets
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import androidx.compose.runtime.getValue
import androidx.compose.runtime.setValue
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.autofill.ContentType
import androidx.compose.ui.autofill.contentType
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.main.MainActivity
import de.rpjosh.rpout.android.activities.theme.OutlinedTextField
import de.rpjosh.rpout.android.activities.theme.OutlinedTextFieldWithLabel
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.ApiKey
import java.time.LocalDateTime

class LoginActivity : ComponentActivity() {

    @Inject private lateinit var userController: UserController

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        initApp()

        // Configure elements
        enableEdgeToEdge()
        setContent {
            RPoutTheme {
                Login { u, p, url -> doLogin(u, p, url) }
            }
        }
    }

    private fun initApp() {
        if (Singleton.getApp() == null) Singleton.app()

        // Inject dependencies
        Singleton.appController.injection.inject(LoginActivity::class.java, null,  false, this)
    }

    private fun doLogin(username: String, password: String, serverURL: String) {
        val key = ApiKey(
            alias = Build.MODEL,
            validUntil = LocalDateTime.now().plusYears(2),
        )

        Thread {
            if (userController.getApiKey(serverURL, username, password, key)) {
                startActivity(Intent(this, MainActivity::class.java))
            }
        }.start()
    }

    override fun onPause() {
        Singleton.getApp()?.activityPaused(this)
        super.onPause()
    }
    override fun onDestroy() {
        Singleton.appController.activityDestroyed(this)
        super.onDestroy()
    }

    override fun onStart() {
        Singleton.appController.activityStarted(applicationContext, this)
        super.onStart()
    }
    override fun onResume() {
        Singleton.appController.activityStarted(applicationContext, this)
        super.onResume()
    }

}

@OptIn(ExperimentalComposeUiApi::class)
@Composable
fun Login(
    onLogin: (username: String, password: String, url: String) -> Unit
) {
    val uriHandler = LocalUriHandler.current

    // User input
    var username by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var serverUrl by remember { mutableStateOf("") }

    Column(modifier = Modifier
        .fillMaxSize()
        .background(RPoutTheme.colors.defaultBackground)
        .padding(15.dp),
    ) {
        // Don't draw onto status bar
        Spacer(Modifier.windowInsetsPadding(WindowInsets.statusBars))

        Column(modifier = Modifier.padding(top = 20.dp)) {
            Image(
                painter = painterResource(R.drawable.banner),
                contentDescription = "Logo",
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(15.dp),
                contentScale = ContentScale.FillWidth
            )

            // Server name
            OutlinedTextFieldWithLabel(
                placeholder = "https://yourDomain.de",
                label = stringResource(R.string.serverURL),
                value = serverUrl,
                onValueChange = { serverUrl = it },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 10.dp)
            )

            // Username
            OutlinedTextField(
                placeholder = stringResource(R.string.username),
                value = username,
                onValueChange = { username = it },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 30.dp)
                    .contentType(ContentType.EmailAddress),
            )
            // Password
            OutlinedTextField(
                placeholder = stringResource(R.string.password),
                value = password, onValueChange = { password = it },
                password= true,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 15.dp)
                    .contentType(ContentType.Password)
            )

            // Login password
            Button(
                onClick = { onLogin(username, password, serverUrl)},
                modifier = Modifier.fillMaxWidth().padding(top = 16.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = RPoutTheme.colors.secondary,
                    contentColor = RPoutTheme.colors.text
                )
            ) {
                Text(stringResource(R.string.doLogin))
            }

            // Self host server
            Text(
                text = stringResource(R.string.noServer),
                textAlign = TextAlign.Center,
                color = RPoutTheme.colors.textBlue,
                modifier = Modifier
                    .padding(top = 16.dp)
                    .clickable {
                        uriHandler.openUri("https://github.com/RPJoshL/workout")
                    }
                    .fillMaxWidth()
            )
        }

    }
}



@Preview(showBackground = true, device = "id:pixel_7", showSystemUi = true)
@Composable
fun LoginPreview() {
    RPoutTheme {
        Login(onLogin = { _, _, _ -> })
    }
}

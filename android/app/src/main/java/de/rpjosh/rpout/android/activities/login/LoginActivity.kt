package de.rpjosh.rpout.android.activities.login

import android.content.Intent
import android.os.Build
import android.os.Bundle
import androidx.compose.foundation.layout.WindowInsets
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.animation.Animatable
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.safeDrawingPadding
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.layout.systemBars
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.foundation.layout.windowInsetsTopHeight
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.tooling.preview.Preview
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import androidx.compose.runtime.getValue
import androidx.compose.runtime.key
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.autofill.AutofillType
import androidx.compose.ui.focus.onFocusEvent
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.SolidColor
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.text.input.PasswordVisualTransformation
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.view.WindowCompat
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.main.MainActivity
import de.rpjosh.rpout.android.activities.theme.OutlinedTextField
import de.rpjosh.rpout.android.activities.theme.OutlinedTextFieldWithLabel
import de.rpjosh.rpout.android.activities.theme.accentBlueBorder
import de.rpjosh.rpout.android.activities.theme.autofill
import de.rpjosh.rpout.android.activities.theme.backgroundDarker
import de.rpjosh.rpout.android.activities.theme.backgroundDisabledDarker
import de.rpjosh.rpout.android.activities.theme.backgroundError
import de.rpjosh.rpout.android.activities.theme.backgroundSuccess
import de.rpjosh.rpout.android.activities.theme.defaultBackground
import de.rpjosh.rpout.android.activities.theme.secondary
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textBlue
import de.rpjosh.rpout.android.activities.theme.textDarker
import de.rpjosh.rpout.android.activities.theme.textHint
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.ApiKey
import kotlinx.coroutines.launch
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

    // User input
    var username by remember { mutableStateOf("") }
    var password by remember { mutableStateOf("") }
    var serverUrl by remember { mutableStateOf("") }

    Column(modifier = Modifier
        .fillMaxSize()
        .background(defaultBackground)
        .padding(15.dp),
    ) {
        // Don't draw onto status bar
        Spacer(Modifier.windowInsetsPadding(WindowInsets.statusBars))

        // Main App content (a little bit offseted)
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
                value = serverUrl, onValueChange = { serverUrl = it },
                modifier = Modifier.fillMaxWidth().padding(top = 10.dp)
            )

            // Username
            OutlinedTextField(
                placeholder = stringResource(R.string.username),
                value = username, onValueChange = { username = it },
                modifier = Modifier.fillMaxWidth().padding(top = 30.dp)
                    .autofill(listOf(AutofillType.EmailAddress)) { username = it }
            )
            // Password
            OutlinedTextField(
                placeholder = stringResource(R.string.password),
                value = password, onValueChange = { password = it },
                password= true,
                modifier = Modifier.fillMaxWidth().padding(top = 15.dp)
                    .autofill(listOf(AutofillType.Password)) { password = it }
            )

            // Login password
            Button(
                onClick = { onLogin(username, password, serverUrl)},
                modifier = Modifier.fillMaxWidth().padding(top = 16.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = secondary,
                    contentColor = text
                )
            ) {
                Text(stringResource(R.string.doLogin))
            }

            // Self host server
            Text(
                text = stringResource(R.string.noServer),
                textAlign = TextAlign.Center,
                color = textBlue,
                modifier = Modifier.fillMaxWidth().padding(top = 16.dp)
            )
        }

    }
}



@Preview(showBackground = true, device = "id:pixel_7", showSystemUi = true)
@Composable
fun LoginPreview() {
    RPoutTheme {
        Login(onLogin = { _, _, _ -> Unit})
    }
}
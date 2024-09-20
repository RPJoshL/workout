package de.rpjosh.rpout.android.activities.settings

import android.app.Activity
import android.content.Context
import android.content.Intent
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.List
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.activities.login.LoginActivity
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundDisabled
import de.rpjosh.rpout.android.activities.theme.backgroundError
import de.rpjosh.rpout.android.activities.theme.backgroundSuccess
import de.rpjosh.rpout.android.activities.theme.defaultBackground
import de.rpjosh.rpout.android.activities.theme.textBlue
import de.rpjosh.rpout.android.activities.theme.textGreen
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.services.TranslationService
import java.time.format.DateTimeFormatter

class UserTab: Tab {

    override lateinit var activity: Activity
    @Inject lateinit var userController: UserController
    @Inject lateinit var globalConfiguration: GlobalConfiguration
    @Inject(parameters = ["UserTab"]) lateinit var logger: Logger

    // Current login state
    private var status = mutableStateOf(LoginStatus(
        Tr.get("settings_user_unknownStatus"),
        " -",
        " -",
        " -",
        " -",
        0
    ))

    override fun getLabel(): String {
        return Tr.get("settings_user_label")
    }

    override fun loadData() {
        // Fetch user status on loading
        Thread {
            try {
                val key = userController.getDetailsOfLogin()
                status.value = LoginStatus(
                    Tr.get("settings_user_loggedInStatus"),
                    globalConfiguration.user?.serverUrl ?: "",
                    globalConfiguration.user?.username ?: "",
                    key.obfuscated + " (" + key.alias + ")",
                    TimeHelper.fromServerToClient(key.validUntil).format(DateTimeFormatter.ofPattern("dd.MM.yyyy")),
                    1
                )
            } catch (ex: Exception) {
                logger.log("d", ex)
                status.value = LoginStatus(
                    ex.message ?: Tr.get("settings_user_unknownStatus"),
                    globalConfiguration.user?.serverUrl ?: "",
                    globalConfiguration.user?.username ?: "",
                    " -", " -",
                    2
                )
            }
        }.start()
    }

    override fun onPause() {}
    override fun onDestroy() {}

    @Composable
    override fun Content() {
        return ContentInternal(
            status = status.value,
            onLogout = {
                Thread {
                    userController.logout {
                        // Go back to login activity
                        activity.startActivity(Intent(activity, LoginActivity::class.java))
                        activity.finish()
                    }
                }.start()
            }
        )
    }

    @Composable
    fun ContentInternal(status: LoginStatus, onLogout: () -> Unit) {
        Column (modifier = Modifier.fillMaxWidth()) {

            // Current login status
            Row(
                modifier = Modifier.fillMaxWidth()
                    .padding(8.dp)
                    .background(color = if (status.background == 0) backgroundDisabled else if (status.background == 1) backgroundSuccess else backgroundError, shape = RoundedCornerShape(5.dp))
            ) {
                // Left side status
                Column(verticalArrangement = Arrangement.spacedBy(8.dp), modifier = Modifier.weight(3f).padding(10.dp)) {
                    // Overall status
                    Text(
                        text = status.status,
                        textAlign = TextAlign.Center, fontWeight = FontWeight.Bold,
                        color = if (status.background == 1) textGreen else de.rpjosh.rpout.android.activities.theme.error, fontSize = 17.sp,
                        // To align in "center", we have to use somme offset padding according to the logout button
                        modifier = Modifier.padding(start = 40.dp, top = 10.dp).fillMaxWidth()
                    )

                    // User and server URL
                    Text(
                        text = stringResource(R.string.settings_user_server) + ": " + status.server,
                        modifier = Modifier.padding(top = 10.dp),
                        color = textBlue, fontSize = 14.sp
                    )
                    Text(
                        text = stringResource(R.string.settings_user_user) + ": " + status.user,
                        color = textBlue, fontSize = 14.sp
                    )

                    // API key
                    Text(
                        text = stringResource(R.string.settings_user_apiKey) + ": " + status.apiKey,
                        modifier = Modifier.padding(top = 10.dp),
                        color = textGreen, fontSize = 14.sp
                    )
                    Text(
                        text = stringResource(R.string.settings_user_validUntil) + ": " + status.validUntil,
                        color = textGreen, fontSize = 14.sp
                    )
                }

                // Right side (logout button)
                Column(modifier = Modifier.width(70.dp).align(Alignment.CenterVertically)) {
                    IconButton(
                        onClick = { onLogout() },
                        modifier = Modifier.width(70.dp).padding(end = 20.dp, start = 8.dp).background(
                            defaultBackground, shape = RoundedCornerShape(16.dp)
                        )
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.ic_logout),
                            contentDescription = "Logout",
                            modifier = Modifier.size(70.dp).padding(2.dp)
                        )
                    }
                }
            }
        }
    }

}

data class LoginStatus(
    val status: String,
    val server: String,
    val user: String,
    val apiKey: String,
    val validUntil: String,
    // Background color: 0 = gray, 1 = green, 2 = error
    val background: Int
)


@Preview(showBackground = true, device = "id:pixel_7", showSystemUi = true)
@Composable
fun UserTabPreview() {
    Tr.addTranslationService(TranslationService("translation.shared"))
    RPoutTheme {
        UserTab().ContentInternal(
            status = LoginStatus(
                "Unknown status", "https://workout.rpjosh.de",
                "myUser@myDomain.de","123 ... 456 (Pixel 7)",
                "22.07.2024", 0
            ),
            onLogout = {}
        )
    }
}
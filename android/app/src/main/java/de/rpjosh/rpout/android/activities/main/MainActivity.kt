package de.rpjosh.rpout.android.activities.main

import android.content.Intent
import android.content.pm.PackageManager
import android.os.Bundle
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Icon
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.os.ConfigurationCompat
import androidx.lifecycle.lifecycleScope
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.components.SvgIcon
import de.rpjosh.rpout.android.activities.components.dummyWorkoutIcon
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
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.TranslationService
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class MainActivity : ComponentActivity() {

    @Inject private lateinit var globalConfig: GlobalConfiguration
    @Inject private lateinit var workoutController: WorkoutController

    private var webview: Webview? = null

    val connectionError = mutableStateOf(false)

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
        globalConfig.user?.let {
            webview = Webview(
                context = this,
                user = it,
                onFinish = { finish() },
                onConnectionError = { isErr -> connectionError.value = isErr }
            )
        }

        // Validate required permissions
        lifecycleScope.launch {
            delay(2000L)
            checkPermissions()
        }

        enableEdgeToEdge()
        setContent {
            RPoutTheme {
                MainActivityScreen(
                    webview = webview,
                    onSettingsClick = {
                        startActivity(Intent(this, SettingsActivity::class.java))
                    },
                    onOpenWorkout = {
                        startActivity(Intent(this, WorkoutTracking::class.java))
                    },
                    connectionError = connectionError.value,
                )
            }
        }
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
fun MainActivityScreen(
    webview: Webview?,
    onSettingsClick: () -> Unit,
    onOpenWorkout: () -> Unit,
    connectionError: Boolean,
    renderSvgIcon: Boolean = true
) {
    val showWebview = webview != null && !connectionError

    Scaffold(
        modifier = Modifier.fillMaxSize(),
        containerColor = RPoutTheme.colors.defaultBackground
    ) { innerPadding ->
        Column {
            Spacer(
                modifier = Modifier
                    .fillMaxWidth().height(innerPadding.calculateTopPadding())
                    .background(
                        if(showWebview) RPoutTheme.colors.webviewHeaderColor
                        else RPoutTheme.colors.defaultBackground
                    )
            )

            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(bottom = innerPadding.calculateBottomPadding())
            ) {
                if(showWebview) WebViewScreen(webview)
                else {
                    NoWebViewScreen(
                        webview = webview,
                        onSettingsClick = onSettingsClick,
                    )
                }

                WorkoutTrackingOverlay(
                    modifier = Modifier.align(Alignment.BottomCenter),
                    onOpenWorkout = onOpenWorkout,
                    renderSvgIcon = renderSvgIcon,
                )
            }
        }
    }
}

@Composable
fun NoWebViewScreen(
    webview: Webview?,
    onSettingsClick: () -> Unit
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(24.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.Center
    ) {
        Icon(
            painter = painterResource(id = R.drawable.hide_image),
            contentDescription = null,
            modifier = Modifier.size(64.dp),
            tint = RPoutTheme.colors.textHint
        )
        
        Spacer(modifier = Modifier.height(16.dp))
        
        Text(
            text = stringResource(R.string.main_noConnection),
            fontSize = 20.sp,
            fontWeight = FontWeight.Bold,
            color = RPoutTheme.colors.text,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(8.dp))
        
        Text(
            text = stringResource(R.string.main_noConnectionHint),
            fontSize = 14.sp,
            color = RPoutTheme.colors.textDarker,
            textAlign = TextAlign.Center
        )
        
        Spacer(modifier = Modifier.height(32.dp))
        
        Button(
            onClick = { webview?.reload() },
            modifier = Modifier.fillMaxWidth(0.7f),
            colors = ButtonDefaults.buttonColors(
                containerColor = RPoutTheme.colors.secondary
            )
        ) {
            Text(stringResource(R.string.main_retry), color = Color.White)
        }
        
        Spacer(modifier = Modifier.height(8.dp))
        
        OutlinedButton(
            onClick = onSettingsClick,
            modifier = Modifier.fillMaxWidth(0.7f)
        ) {
            Text(stringResource(R.string.main_settings), color = RPoutTheme.colors.text)
        }
    }
}

 @Composable
fun WorkoutTrackingOverlay(
    modifier: Modifier = Modifier,
    onOpenWorkout: () -> Unit,
    renderSvgIcon: Boolean = true,
) {
    val workoutStatus = remember { RealtimeLocationService.status }
     val locale = ConfigurationCompat.getLocales(LocalConfiguration.current).get(0)

     val isRunning =  workoutStatus.state.value in listOf(WorkoutStatus.RUNNING, WorkoutStatus.HIGH_SAMPLING)
     val isPaused =  workoutStatus.state.value in listOf(WorkoutStatus.PREPARE, WorkoutStatus.PAUSE)

     if (!isRunning && !isPaused) {
         return
     }

     val type = workoutStatus.type.value
     val color = if(isSystemInDarkTheme()) type?.tagDark else type?.tagWhite

     Column(
         modifier = modifier
             .fillMaxWidth()
             .clickable(true) {
                 onOpenWorkout()
             },
     ) {
         Box(
             modifier = Modifier
                 .fillMaxWidth()
                 .height(2.dp)
                 .background(RPoutTheme.colors.backgroundLighter)
         ) { }

         Row(
             modifier = Modifier
                 .fillMaxWidth()
                 .background(RPoutTheme.colors.defaultBackground)
                 .padding(6.dp),
             verticalAlignment = Alignment.CenterVertically,
             horizontalArrangement = Arrangement.spacedBy(12.dp)
         ) {
             if(renderSvgIcon) {
                 SvgIcon(
                     svgString = type?.icon ?: dummyWorkoutIcon,
                     size = 40.dp,
                     hexTint = color,
                 )
             } else {
                 Icon(
                     modifier = Modifier.size(40.dp),
                     contentDescription = "",
                     painter = painterResource(de.rpjosh.rpout.android.shared.R.drawable.walking),
                     tint = RPoutTheme.colors.text
                 )
             }

             Row(
                 modifier = Modifier.weight(1f),
                 horizontalArrangement = Arrangement.Center,
             ) {
                 Text(
                     text = (type?.getName(TranslationService.Language.fromAndroidLocale(locale)) ?: "") + " ",
                     fontWeight = FontWeight.Bold,
                     fontSize = 16.sp,
                     color = RPoutTheme.colors.text
                 )
                 Text(
                     text = "(${if(isPaused) stringResource(R.string.paused) else stringResource(R.string.running)})",
                     fontStyle = FontStyle.Italic,
                     fontSize = 16.sp,
                     color = RPoutTheme.colors.text
                 )
             }

             Icon(
                 modifier = Modifier.size(40.dp),
                 contentDescription = "Workout status indicator",
                 painter = painterResource(if(isPaused) R.drawable.pause else R.drawable.start),
                 tint = RPoutTheme.colors.text
             )
         }
     }
}


@Preview(showBackground = true, device = "id:pixel_7", showSystemUi = true)
@Composable
fun MainActivityPreview() {
    RealtimeLocationService.status.state.value = WorkoutStatus.RUNNING
    RealtimeLocationService.status.type.value = WorkoutType(
        id = 0,
        icon = "",
        nameEn = "Walking",
    )

    RPoutTheme(darkTheme = false) {
        MainActivityScreen(
            webview = null,
            onSettingsClick = {},
            onOpenWorkout = {},
            connectionError = false,
            renderSvgIcon = false,
        )
    }
}

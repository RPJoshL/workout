package de.rpjosh.rpout.android.activities.main.workoutstart

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.media.MediaPlayer
import android.os.Bundle
import android.os.VibrationEffect
import android.os.Vibrator
import android.view.WindowManager
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.wear.compose.foundation.pager.HorizontalPager
import androidx.wear.compose.foundation.pager.rememberPagerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import de.rpjosh.rpout.android.activities.WearPreview
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.compose.material3.Button
import androidx.wear.compose.material3.ButtonDefaults
import androidx.wear.compose.material3.HorizontalPageIndicator
import androidx.wear.compose.material3.Icon
import androidx.wear.compose.material3.ScreenScaffold
import androidx.wear.compose.material3.Text
import androidx.compose.foundation.interaction.MutableInteractionSource
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.runtime.MutableState
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalInspectionMode
import androidx.compose.ui.text.font.FontWeight
import androidx.wear.compose.material3.AnimatedPage
import androidx.wear.compose.material3.AppScaffold
import androidx.wear.compose.material3.onehandedgesture.OneHandedGestureAction
import androidx.wear.compose.material3.onehandedgesture.OneHandedGestureClickIndicator
import androidx.wear.compose.material3.onehandedgesture.OneHandedGestureClickIndicatorState
import androidx.wear.compose.material3.onehandedgesture.oneHandedGesture
import androidx.wear.compose.material3.onehandedgesture.rememberOneHandedGestureConfiguration
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.components.GpsStatusIndicator
import de.rpjosh.rpout.android.activities.main.workouttracking.WorkoutTrackingActivity
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.textHintDarker
import de.rpjosh.rpout.android.shared.models.GPSMode
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.workout.State
import de.rpjosh.rpout.android.workout.WorkoutManager
import de.rpjosh.rpout.android.workout.WorkoutTrackService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlin.time.Duration.Companion.milliseconds

class WorkoutStartActivity : ComponentActivity() {

    private lateinit var workoutManager: WorkoutManager

    /** Whether the workout is already started -> don't stop exercise */
    @Volatile private var isStarted = false

    companion object {
        const val KEY_TYPE_ID = "TYPE_ID"
        const val INTENT_WORKOUT_INITED = "WORKOUT_STARTED"
        const val BROADCAST_FILTER_ACTION = "WORKOUT_START_BROADCAST_ACTION"
    }

    private val broadcastReceiver = object : BroadcastReceiver() {
        override fun onReceive(context: Context, intent: Intent) {
            if (intent.getBooleanExtra(INTENT_WORKOUT_INITED, false)) {
                // Nothing to handle currently
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()

        super.onCreate(savedInstanceState)
        setTheme(android.R.style.Theme_DeviceDefault)

        // Do not turn display off
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)

        // Get provided workout type ID
        var typeId = 1L
        intent.extras?.let {
            typeId = it.getLong(KEY_TYPE_ID)
        }

        // Register broadcast listener
        registerReceiver(broadcastReceiver, IntentFilter(BROADCAST_FILTER_ACTION), RECEIVER_EXPORTED)

        // Initialize workout manager
        workoutManager = WorkoutManager(typeId)
        Singleton.appController.injection.inject(WorkoutManager::class.java, null,  false, workoutManager)
        Thread{
            workoutManager.init()
            WorkoutManager.workoutManager = workoutManager

            // Start the tracking
            val serviceIntent = Intent(RPout.getAppContext(), WorkoutTrackService::class.java)
            serviceIntent.action = "START"
            ContextCompat.startForegroundService(this, serviceIntent)
        }.start()

        setContent {
            RPoutTheme {
                WorkoutStartScreen(workoutManager) { onStartExercise() }
            }
        }
    }

    private fun onStartExercise() {
        isStarted = true
        val intent = Intent(this, WorkoutTrackingActivity::class.java).apply {
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK)
        }
        startActivity(intent)
        finish()
    }

    override fun onPause() {
        super.onPause()

        // Stop the tracking
        if (!isStarted) {
            val serviceIntent = Intent(RPout.getAppContext(), WorkoutTrackService::class.java)
            serviceIntent.action = "STOP"
            ContextCompat.startForegroundService(this, serviceIntent)
        }
    }
    override fun onDestroy() {
        super.onDestroy()

        unregisterReceiver(broadcastReceiver)

        // Workout tracker is already stopped in onPause
    }

    override fun onResume() {
        super.onResume()

        // Start the tracking (again)
        val serviceIntent = Intent(RPout.getAppContext(), WorkoutTrackService::class.java)
        serviceIntent.action = "START"
        ContextCompat.startForegroundService(this, serviceIntent)
    }
}

@Composable
fun WorkoutStartScreen(manager: WorkoutManager, onStart: () -> Unit) {
    val scope = rememberCoroutineScope()

    val gpsMode = remember { mutableStateOf(manager.type.getGPSMode()) }
    val pagerState = rememberPagerState { 2 }

    AppScaffold(
        timeText = {}
    ) {
        ScreenScaffold {
            HorizontalPager(
                state = pagerState,
                modifier = Modifier.fillMaxSize()
            ) { index ->
                AnimatedPage(index, pagerState) {
                    when(index) {
                        0 -> StartPage(
                            manager, gpsMode,
                            onGoToSettings = { scope.launch { pagerState.animateScrollToPage(1) } },
                            onStart = onStart
                        )
                        1 -> SettingsPage(manager, gpsMode, isVisible = pagerState.currentPage == 1)
                    }
                }
            }

            HorizontalPageIndicator(
                pagerState = pagerState,
                modifier = Modifier.align(Alignment.BottomCenter).padding(bottom = 2.dp, start = 1.dp)
            )
        }
    }
}

@Composable
fun StartPage(manager: WorkoutManager, gpsMode: MutableState<GPSMode>, onGoToSettings: () -> Unit, onStart: () -> Unit) {
    val coroutineScope = rememberCoroutineScope()
    val context = LocalContext.current

    /** Whether the GPS signal is already acquired */
    val gpsConnected by remember { derivedStateOf { manager.state.value == State.READY } }
    /** Whether the start is disabled */
    val startDisabled = remember { derivedStateOf { manager.state.value == State.ERROR} }

    /** Whether the "start" mode is currently active */
    val isStarting = remember { mutableStateOf(false) }
    val startingProgress = remember { Animatable(0f) }

    val gestureConfig = rememberOneHandedGestureConfiguration(
        action = OneHandedGestureAction.Primary,
        gestureId = "workout_start"
    )
    val gestureIndicatorState = remember { OneHandedGestureClickIndicatorState() }
    val startInteractionSource = remember { MutableInteractionSource() }

    val configuration = LocalConfiguration.current
    val scale = (configuration.screenWidthDp / 192f).coerceAtLeast(1.0f)

    val isPreview = LocalInspectionMode.current
    val onClickStart: () -> Unit = {
        coroutineScope.launch {
            if (startDisabled.value || isStarting.value) {
                return@launch
            }

            if (isPreview) {
                isStarting.value = true
                startingProgress.animateTo(
                    targetValue = 1f,
                    animationSpec = tween(durationMillis = 3000, easing = LinearEasing)
                )
                return@launch
            }

            coroutineScope.launch(Dispatchers.IO) {
                try {
                    val vibrator = context.getSystemService(Vibrator::class.java)
                    val mediaPlayer = MediaPlayer.create(context, R.raw.start_tick)
                    mediaPlayer?.setVolume(0.15f, 0.15f)
                    mediaPlayer?.start()
                    mediaPlayer?.setOnCompletionListener { it.release() }

                    val pattern = longArrayOf(150, 1000, 120, 1000, 120, 1000, 120, 120, 120)
                    val amplitude = intArrayOf(255, 0, 230, 0, 230, 0, 255, 0, 255)
                    val vibrationEffect = VibrationEffect.createWaveform(pattern, amplitude, -1)
                    vibrator?.vibrate(vibrationEffect)
                } catch (e: Exception) {
                    e.printStackTrace()
                }
            }

            // Give UI some time to hide elements
            delay(100.milliseconds)
            isStarting.value = true
            startingProgress.snapTo(0f)
            delay(100.milliseconds)

            // Real-time progress timer (independent of system animation scale / developer options).
            // Just because I'm using it and the experience is otherwise bad :)
            val startTime = System.currentTimeMillis()
            val durationMs = 3200L

            while (isActive) {
                val elapsed = System.currentTimeMillis() - startTime
                val progress = (elapsed.toFloat() / durationMs).coerceIn(0f, 1f)
                startingProgress.snapTo(progress)

                if (progress >= 1f) break
                delay(16.milliseconds)
            }

            // Play start confirmation sound in application context so the sound
            // survives activity change
            CoroutineScope(Dispatchers.IO).launch {
                try {
                    val mediaPlayer2 = MediaPlayer.create(context, R.raw.start_start)
                    mediaPlayer2?.setVolume(0.15f, 0.15f)
                    mediaPlayer2?.start()
                    mediaPlayer2?.setOnCompletionListener { it.release() }
                } catch (e: Exception) {
                    e.printStackTrace()
                }
            }

            delay(50.milliseconds)
            onStart()
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .padding(start = 2.dp, end = 2.dp),
        contentAlignment = Alignment.Center
    ) {
        GpsStatusIndicator(
            available = gpsConnected,
            supported = manager.healthSupportedCapabilities?.gps != false,
        )

        Column(
            modifier = Modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = manager.type.getName(Tr.getUsedLanguage()),
                textAlign = TextAlign.Center,
                modifier = Modifier
                    .align(Alignment.CenterHorizontally)
                    .padding(top = if(manager.type.getName(Tr.getUsedLanguage()).length > 10) 24.dp else 14.dp),
                fontWeight = FontWeight.Bold,
                color = manager.typeAccentColor.value
            )

            Row(
                horizontalArrangement = Arrangement.spacedBy(10.dp),
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.align(Alignment.CenterHorizontally)
                    .padding(start = 8.dp, end = 8.dp, top = 12.dp)
            ) {
                if (!isStarting.value) {
                    val showGPSToggle = manager.type.isGPSTrackingSupported() && gpsMode.value != GPSMode.DISABLED
                    val usePhone = gpsMode.value == GPSMode.USE_PHONE

                    Button(
                        modifier = Modifier
                            .size(43.dp * scale)
                            .border(2.dp, if(usePhone) manager.typeAccentColor.value else textHintDarker, CircleShape),
                        onClick = {
                            if(!showGPSToggle) {
                                onGoToSettings()
                                return@Button
                            }

                            val newMode = if(usePhone) GPSMode.DEFAULT else GPSMode.USE_PHONE
                            gpsMode.value = newMode
                            Thread{ manager.changeSettings(context, gpsMode = newMode) }.start()
                        },
                        shape = CircleShape,
                        colors = ButtonDefaults.buttonColors(containerColor = backgroundLighter),
                    ) {
                        Icon(
                            painter = painterResource(
                                if(showGPSToggle) R.drawable.ic_phone else R.drawable.settings
                            ),
                            contentDescription = "Settings",
                            modifier = Modifier.size(22.dp * scale),
                            tint = manager.typeAccentColor.value.copy(alpha = if(usePhone) 1.0f else 0.6f)
                        )
                    }
                }

                val buttonSize = 54.dp * scale
                val progressStrokeWidth = 3.dp * scale
                val buttonPadding = 3.dp * scale

                Box(
                    modifier = Modifier.size(buttonSize + (progressStrokeWidth * 2) + (buttonPadding * 2)),
                    contentAlignment = Alignment.Center
                ) {
                    if (isStarting.value) {
                        // CircularProgressIndicater lagged a lot
                        Canvas(
                            modifier = Modifier.fillMaxSize()
                        ) {
                            val stroke = progressStrokeWidth.toPx()
                            val radius = (size.minDimension - stroke) / 2f

                            drawCircle(
                                color = backgroundLighter,
                                radius = radius,
                                style = Stroke(width = stroke)
                            )

                            drawArc(
                                color = manager.typeAccentColor.value,
                                startAngle = -90f,
                                sweepAngle = 360f * startingProgress.value.coerceIn(0f, 1f),
                                useCenter = false,
                                style = Stroke(
                                    width = stroke,
                                    cap = StrokeCap.Round
                                )
                            )
                        }
                    }

                    Button(
                        shape = CircleShape,
                        interactionSource = startInteractionSource,
                        modifier = Modifier
                            .size(buttonSize)
                            .oneHandedGesture(
                                gestureConfiguration = gestureConfig,
                                interactionSource = startInteractionSource,
                                onGestureLabel = "start workout",
                                onGestureAvailable = {
                                    coroutineScope.launch { gestureIndicatorState.showIndicator() }
                                },
                                onGesture = onClickStart,
                            ),
                        onClick = onClickStart,
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (startDisabled.value || isStarting.value) backgroundLighter else manager.typeAccentColor.value
                        ),
                    ) {
                        OneHandedGestureClickIndicator(gestureConfig, gestureIndicatorState) {
                            Icon(
                                painter = painterResource(R.drawable.start),
                                contentDescription = "Start",
                                modifier = Modifier.size(28.dp * scale),
                                tint = if (isStarting.value) manager.typeAccentColor.value else Color.Black,
                            )
                        }
                    }
                }

                if (!isStarting.value) {
                    Button(
                        onClick = { onGoToSettings() },
                        shape = CircleShape,
                        colors = ButtonDefaults.buttonColors(containerColor = backgroundLighter),
                        modifier = Modifier.size(43.dp * scale)
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.settings),
                            contentDescription = "Settings",
                            modifier = Modifier.size(22.dp * scale),
                            tint = manager.typeAccentColor.value
                        )
                    }
                }
            }

            Row(
                horizontalArrangement = Arrangement.spacedBy(6.dp),
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.align(Alignment.CenterHorizontally).padding(top = 16.dp)
            ) {
                Icon(
                    painter = painterResource(R.drawable.heart),
                    contentDescription = "Heart rate",
                    modifier = Modifier.size(18.dp),
                    tint = manager.workoutData.heartRate.color.value
                )
                Text(
                    text = manager.workoutData.heartRate.value.value.toString(),
                    fontSize = 15.sp,
                    color = manager.workoutData.heartRate.color.value
                )
            }
        }
    }
}

@WearPreview
@Composable
fun WorkoutStartPreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview()

    RPoutTheme {
        WorkoutStartScreen(manager, {})
    }
}

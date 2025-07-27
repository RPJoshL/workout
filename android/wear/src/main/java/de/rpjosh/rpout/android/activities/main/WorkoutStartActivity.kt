package de.rpjosh.rpout.android.activities.main

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
import androidx.compose.animation.Animatable
import androidx.compose.animation.core.Easing
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.compose.foundation.ArcPaddingValues
import androidx.wear.compose.foundation.CurvedLayout
import androidx.wear.compose.foundation.CurvedModifier
import androidx.wear.compose.foundation.curvedComposable
import androidx.wear.compose.foundation.curvedRow
import androidx.wear.compose.foundation.lazy.ScalingLazyColumn
import androidx.wear.compose.foundation.lazy.ScalingLazyColumnDefaults
import androidx.wear.compose.foundation.lazy.ScalingLazyListAnchorType
import androidx.wear.compose.foundation.lazy.ScalingLazyListState
import androidx.wear.compose.foundation.padding
import androidx.wear.compose.material.Button
import androidx.wear.compose.material.ButtonDefaults
import androidx.wear.compose.material.CircularProgressIndicator
import androidx.wear.compose.material.HorizontalPageIndicator
import androidx.wear.compose.material.Icon
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.PageIndicatorState
import androidx.wear.compose.material.PositionIndicator
import androidx.wear.compose.material.Scaffold
import androidx.wear.compose.material.Switch
import androidx.wear.compose.material.Text
import androidx.wear.compose.material.TimeText
import androidx.wear.compose.material.ToggleChip
import androidx.wear.compose.material.ToggleChipDefaults
import androidx.wear.compose.material.curvedText
import androidx.wear.tooling.preview.devices.WearDevices
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.backgroundMoreLighter
import de.rpjosh.rpout.android.activities.theme.backgroundSelection
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textHint
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.workout.State
import de.rpjosh.rpout.android.shared.workout.WorkoutManager
import de.rpjosh.rpout.android.workout.WorkoutTrackService
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

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
        workoutManager = WorkoutManager(true, typeId)
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
                WorkoutStartScreen(workoutManager, { onStartExercise() })
            }
        }
    }

    private fun onStartExercise() {
        isStarted = true
        val intent = Intent(this,  WorkoutTrackingActivity::class.java).apply {
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

/** All (vertical) pages of this activity */
val startPages: List<@Composable (manager: WorkoutManager, onGoToSettings: () -> Unit, onStart: () -> Unit) -> Unit> = listOf(
    { manager, onGoToSettings, onStart -> StartPage(manager, onGoToSettings, onStart = onStart)  },
    { manager, _, _  -> SettingsPage(manager) }
)

@Composable
fun WorkoutStartScreen(manager: WorkoutManager, onStart: () -> Unit) {
    val scope = rememberCoroutineScope()

    // Page state
    val pagerState = rememberPagerState { startPages.size }
    val pageIndicatorState: PageIndicatorState = remember {
        object : PageIndicatorState {
            override val pageOffset: Float
                get() = 0f
            override val selectedPage: Int
                get() = pagerState.currentPage
            override val pageCount: Int
                get() = pagerState.pageCount
        }
    }

    Scaffold(
        // @BUG we need a 2.dp padding at the bottom to not cut off the circles (tested on Pixel Watch 2)
        pageIndicator = { HorizontalPageIndicator(pageIndicatorState, modifier = Modifier.padding(bottom = 2.dp, start = 1.dp)) }
    ) {
        HorizontalPager(
            state = pagerState,
            modifier = Modifier.fillMaxSize()
        ) { index ->
            startPages[index](
                manager,
                { scope.launch { pagerState.animateScrollToPage(2) } },
                onStart
            )
        }
    }

}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutStartPreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true)

    RPoutTheme {
        WorkoutStartScreen(manager, {})
    }
}

@Composable
fun StartPage(manager: WorkoutManager, onGoToSettings: () -> Unit, onStart: () -> Unit) {
    val coroutineScope = rememberCoroutineScope()
    val context = LocalContext.current

    /** Whether the GPS signal is already acquired */
    val gpsConnected = remember { derivedStateOf {manager.state.value == State.READY } }
    /** (Animated) colors of the GPS signal icon and the text */
    val gpsColor = remember { Animatable(text) }
    /** Rotation value of the GPS icon */
    val gpsRotate = remember { mutableFloatStateOf(0f) }
    /** (Animated) rotation value of the GPS icon */
    val gpsRotating by animateFloatAsState(gpsRotate.floatValue, tween(3000), label = "GPS rotating icon")
    /** Whether the start is disabled */
    val startDisabled = remember { derivedStateOf { manager.state.value == State.ERROR} }

    /** Whether the "start" mode is currently active */
    val isStarting = remember { mutableStateOf(false) }
    val startingProgress = remember { mutableFloatStateOf(0f) }
    val startingProgressAnimated by animateFloatAsState(startingProgress.floatValue, tween(3100, easing = LinearEasing), label = "Start Progress indicator")

    // Texts cannot be obtained inside curvedText
    val txtGpsConnected = stringResource(R.string.main_gpsConnected)
    val txtGpsConnecting = stringResource(R.string.main_gpsConnecting) + "..."

    LaunchedEffect(gpsConnected.value) {
        var isDark = false
        while (!gpsConnected.value) {
            gpsColor.animateTo(if(isDark) text else textHint, animationSpec = tween(1000))
            isDark = !isDark
            delay(1000L)

            if (isDark) gpsRotate.floatValue += 360f
        }

        // Reset color if gps is connected
        gpsColor.animateTo(text, animationSpec = tween(500))
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colors.background)
            .padding(start = 2.dp, end = 2.dp),
        contentAlignment = Alignment.Center
    ) {
        if (manager.healthSupportedCapabilities?.gps == false) {
            TimeText()
        } else {
            CurvedLayout {
                curvedRow() {
                    curvedComposable(modifier = CurvedModifier.padding(ArcPaddingValues(after = 6.dp))) {
                        Icon(
                            painter = painterResource(if(gpsConnected.value) R.drawable.gps_connected else R.drawable.gps_connecting),
                            contentDescription = "GPS status",
                            modifier = Modifier.size(20.dp).rotate(gpsRotating),
                            tint = gpsColor.value
                        )
                    }
                    curvedText(
                        text = if (gpsConnected.value) txtGpsConnected else txtGpsConnecting,
                        fontSize = 13.sp,
                        color = gpsColor.value
                    )
                }
            }
        }

        Column(
            modifier = Modifier.fillMaxWidth(),
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = manager.type.getName(Tr.getUsedLanguage()),
                textAlign = TextAlign.Center,
                modifier = Modifier.align(Alignment.CenterHorizontally)
                    .padding(top = if(manager.type.getName(Tr.getUsedLanguage()).length > 10) 24.dp else 14.dp),
                color = manager.typeAccentColor.value
            )

            Row(
                horizontalArrangement = Arrangement.spacedBy(10.dp),
                verticalAlignment = Alignment.CenterVertically,
                modifier = Modifier.align(Alignment.CenterHorizontally)
                    .padding(start = 8.dp, end = 8.dp, top = 12.dp)
            ) {
                if (!isStarting.value) {
                    Button(
                        onClick = { onGoToSettings() },
                        colors = ButtonDefaults.primaryButtonColors(backgroundColor = backgroundLighter),
                        modifier = Modifier.width(43.dp).height(43.dp)
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.settings),
                            contentDescription = "Settings",
                            modifier = Modifier.size(20.dp),
                            tint = manager.typeAccentColor.value
                        )
                    }
                }
                Box(
                    modifier = Modifier.width(60.dp).height(60.dp)
                ) {
                    Button(
                        onClick = {
                            if (!startDisabled.value) {
                                coroutineScope.launch {
                                    val vibrator = context.getSystemService(Vibrator::class.java)

                                    // Play start sound
                                    val mediaPlayer = MediaPlayer.create(context, R.raw.start_tick)
                                    mediaPlayer?.start()
                                    mediaPlayer?.setOnCompletionListener {
                                        mediaPlayer.release()
                                    }
                                    delay(100)

                                    isStarting.value = true
                                    startingProgress.floatValue = 1f

                                    // Vibrate the device
                                    val pattern = longArrayOf(150,  1000, 120, 1000, 120, 1000, 120, 120, 120)
                                    val amplitude = intArrayOf(255, 0,    230, 0,    230,    0, 255,  0,  255)
                                    val vibrationEffect = VibrationEffect.createWaveform(pattern, amplitude,-1)
                                    vibrator.vibrate(vibrationEffect)

                                    delay(3000)

                                    // Play start sound
                                    val mediaPlayer2 = MediaPlayer.create(context, R.raw.start_start)
                                    mediaPlayer2?.start()
                                    mediaPlayer2?.setOnCompletionListener {
                                        mediaPlayer2.release()
                                    }

                                    delay(100)
                                    onStart()
                                }
                            }
                        },
                        colors = ButtonDefaults.primaryButtonColors(
                            backgroundColor = if (startDisabled.value || isStarting.value) backgroundLighter else manager.typeAccentColor.value
                        ),
                        modifier = Modifier.width(54.dp).height(54.dp).align(Alignment.Center),
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.start),
                            contentDescription = "Start",
                            modifier = Modifier.size(28.dp).align(Alignment.Center),
                            tint = if (isStarting.value) manager.typeAccentColor.value else Color.Black,
                        )
                    }

                    // Progress indicator
                    if(isStarting.value) {
                        CircularProgressIndicator(
                            modifier = Modifier.fillMaxSize(),
                            trackColor = Color.Transparent,
                            progress = startingProgressAnimated
                        )
                    }
                }

                if (!isStarting.value) {
                    Button(
                        onClick = { onGoToSettings() },
                        colors = ButtonDefaults.primaryButtonColors(backgroundColor = backgroundLighter),
                        modifier = Modifier.width(43.dp).height(43.dp)
                    ) {
                        Icon(
                            painter = painterResource(R.drawable.settings),
                            contentDescription = "Settings",
                            modifier = Modifier.size(20.dp),
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
                        modifier = Modifier.size(16.dp),
                        tint = manager.workoutData.heartRate.color.value
                    )
                    Text(
                        text = manager.workoutData.heartRate.value.value.toString(),
                        fontSize = 14.sp,
                        color = manager.workoutData.heartRate.color.value
                    )
            }

        }
    }
}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun StartPagePreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true)

    RPoutTheme {
        StartPage(manager, {}, {})
    }
}

@Composable
fun SettingsPage(manager: WorkoutManager) {
    val context = LocalContext.current

    val listState = remember { ScalingLazyListState(initialCenterItemIndex = 0) }

    val noGPS = remember { mutableStateOf(manager.type.noGPS) }
    val liveUpdates = remember { mutableStateOf(manager.type.liveUpdates) }

    val txtSettings = stringResource(R.string.main_settings)

    Scaffold(
        positionIndicator = { PositionIndicator(scalingLazyListState = listState) },
        timeText = {
            CurvedLayout {
                curvedText(txtSettings, fontSize = 13.sp)
            }
        }
    ) {
        ScalingLazyColumn(
            modifier = Modifier.fillMaxWidth(),
            state = listState,
            flingBehavior = ScalingLazyColumnDefaults.snapFlingBehavior(state = listState),
            contentPadding = PaddingValues(
                top = 34.dp,
                start = 12.dp,
                end = 12.dp,
                bottom = 35.dp
            ),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(4.dp),
            anchorType = ScalingLazyListAnchorType.ItemCenter,
            // Do not center the first elements index => use contentPadding or AutoCenteringParams(itemIndex = 3)
            autoCentering = null
        ) {
            item(key = "gps") {
                SettingsToggle(
                    text = stringResource(R.string.main_noGPS),
                    checked = noGPS.value,
                    isVisible = manager.type.isGPSTrackingSupported()
                ) {
                    noGPS.value = it
                    Thread {
                        manager.changeSettings(noGPS = it)

                        // Restart the tracking
                        val serviceIntent = Intent(RPout.getAppContext(), WorkoutTrackService::class.java)
                        serviceIntent.action = "RESTART"
                        ContextCompat.startForegroundService(context, serviceIntent)
                    }.start()
                }
            }
            item(key = "live-data") {
                SettingsToggle(
                    text = stringResource(R.string.main_liveData),
                    checked = liveUpdates.value
                ) {
                    Thread{ manager.changeSettings(liveData = it) }.start()
                    liveUpdates.value = it
                }
            }
        }
    }
}
/** Display a toggle with the provided text and value the user can check- and uncheck */
@Composable
fun SettingsToggle(text: String, checked: Boolean, isVisible: Boolean = true, onCheckChanged: (isChecked: Boolean) -> Unit) {
    if (!isVisible) {
        return
    }

    ToggleChip(
        label = {
            Text(text, maxLines = 1, overflow = TextOverflow.Ellipsis, fontSize = 14.sp)
        },
        checked = checked,
        colors = ToggleChipDefaults.toggleChipColors(
            uncheckedStartBackgroundColor = backgroundLighter,
            checkedStartBackgroundColor = backgroundMoreLighter,
            checkedEndBackgroundColor = backgroundSelection
        ),
        toggleControl = {
            Switch(checked = checked, enabled = true)
        },
        onCheckedChange = { onCheckChanged(it) },
        enabled = true,
        modifier = Modifier.fillMaxWidth(),
    )
}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun SettingsPagePreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true)

    RPoutTheme {
        SettingsPage(manager)
    }
}

package de.rpjosh.rpout.android.activities.main

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import android.util.Log
import android.view.WindowManager
import android.window.OnBackInvokedDispatcher
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.RoundedCornerShape
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
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.health.services.client.data.ExerciseState
import androidx.health.services.client.proto.DataProto.ExerciseUpdate.ActiveDurationCheckpoint
import androidx.wear.ambient.AmbientLifecycleObserver
import androidx.wear.compose.foundation.ArcPaddingValues
import androidx.wear.compose.foundation.CurvedLayout
import androidx.wear.compose.foundation.CurvedModifier
import androidx.wear.compose.foundation.curvedComposable
import androidx.wear.compose.foundation.curvedRow
import androidx.wear.compose.foundation.padding
import androidx.wear.compose.material.Button
import androidx.wear.compose.material.ButtonDefaults
import androidx.wear.compose.material.HorizontalPageIndicator
import androidx.wear.compose.material.Icon
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.PageIndicatorState
import androidx.wear.compose.material.Scaffold
import androidx.wear.compose.material.Text
import androidx.wear.compose.material.TimeText
import androidx.wear.compose.material.curvedText
import androidx.wear.tooling.preview.devices.WearDevices
import com.google.android.gms.wearable.Wearable
import com.google.android.horologist.annotations.ExperimentalHorologistApi
import com.google.android.horologist.health.composables.ActiveDurationText
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textDarker
import de.rpjosh.rpout.android.activities.theme.textHint
import de.rpjosh.rpout.android.shared.workout.State
import de.rpjosh.rpout.android.shared.workout.WorkoutManager
import de.rpjosh.rpout.android.workout.WorkoutTrackService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.DelicateCoroutinesApi
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.newSingleThreadContext
import java.util.Locale

class WorkoutTrackingActivity: ComponentActivity(), AmbientLifecycleObserver.AmbientLifecycleCallback {

    private val ambientObserver = AmbientLifecycleObserver(this, this)
    private val isAmbient = mutableStateOf(false)
    private lateinit var manager: WorkoutManager

    @OptIn(ExperimentalCoroutinesApi::class, DelicateCoroutinesApi::class)
    private val scope = CoroutineScope(newSingleThreadContext("pauseResumeWorkout"))

    fun onTrackingStop() {
        val intent = Intent(this,  WorkoutFinishedActivity::class.java).apply {
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK)
        }
        startActivity(intent)
        finish()
    }

    fun onLockScreen() {
        window.addFlags(WindowManager.LayoutParams.FLAG_NOT_TOUCHABLE)
    }

    fun onPauseResum() {
        scope.launch {
            if (manager.state.value == State.PAUSED) manager.resume()
            else manager.pause()

            // Notify foreground service to update notification duration
            val serviceIntent = Intent(this@WorkoutTrackingActivity, WorkoutTrackService::class.java)
            serviceIntent.action = "NOTIFICATION"
            ContextCompat.startForegroundService(this@WorkoutTrackingActivity, serviceIntent)
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        // No workout manager available
        val manager = WorkoutManager.workoutManager
        if (manager == null) {
            finish()
            return
        } else {
            this.manager = manager
            scope.launch {
                manager.start()

                // Notify foreground service to update notification duration
                val serviceIntent = Intent(this@WorkoutTrackingActivity, WorkoutTrackService::class.java)
                serviceIntent.action = "NOTIFICATION"
                ContextCompat.startForegroundService(this@WorkoutTrackingActivity, serviceIntent)
            }
        }

        super.onCreate(savedInstanceState)
        lifecycle.addObserver(ambientObserver)
        setTheme(android.R.style.Theme_DeviceDefault)

        // Disallow user to go back
        onBackInvokedDispatcher.registerOnBackInvokedCallback(OnBackInvokedDispatcher.PRIORITY_DEFAULT) {
            // Do nothing
        }

        setContent {
            RPoutTheme {
                WorkoutTrackingScreen(
                   isAmbient = isAmbient.value,
                    manager = manager,
                    onStop = { onTrackingStop() },
                    onScreenLock = { onLockScreen() },
                    onPauseResume = { onPauseResum() }
                )
            }
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

/** All (vertical) pages of this activity */
val trackingPages: List<@Composable (manager: WorkoutManager, onStop: () -> Unit, onScreenLock: () -> Unit, onPauseResume: () -> Unit) -> Unit> = listOf(
    { manager, onStop, onScreenLock, onPauseResume -> WorkoutTrackActionTab(manager, onStop, onScreenLock, onPauseResume)  },
    { manager, _, _, _ -> WorkoutTrackMainTab(manager) },
    { manager, _, _, _  -> WorkoutTrackExtraTab(manager) },
)

@Composable
fun WorkoutTrackingScreen(isAmbient: Boolean, manager: WorkoutManager, onStop: () -> Unit, onScreenLock: () -> Unit, onPauseResume: () -> Unit) {

    /** Whether the GPS signal is already acquired */
    val gpsConnected = remember { derivedStateOf {manager.state.value == State.TRACKED || manager.state.value == State.PAUSED } }
    /** Rotation value of the GPS icon */
    val gpsRotate = remember { mutableFloatStateOf(0f) }
    /** (Animated) rotation value of the GPS icon */
    val gpsRotating by animateFloatAsState(gpsRotate.floatValue, tween(3000), label = "GPS rotating icon")

    // Texts cannot be obtained inside curvedText
    val txtPaused = stringResource(R.string.main_paused)
    val txtGpsConnecting = stringResource(R.string.main_gpsConnecting)

    // Animate GPS icon
    LaunchedEffect(gpsConnected.value.toString() + "-" + isAmbient) {
        var isDark = false
        while (!gpsConnected.value && !isAmbient) {
            isDark = !isDark
            delay(1100L)

            if (isDark) gpsRotate.floatValue += 360f
        }
    }

    // Page state
    val pagerState = rememberPagerState(initialPage = 1) { trackingPages.size }
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
        pageIndicator = { if (isAmbient) null else HorizontalPageIndicator(pageIndicatorState, modifier = Modifier.padding(bottom = 2.dp, start = 1.dp)) },
        timeText = {
            if (manager.state.value == State.PAUSED) {
                CurvedLayout {
                    curvedRow() {
                        curvedText(
                            text = txtPaused,
                            fontWeight = FontWeight.SemiBold,
                            fontSize = 15.sp,
                            color = if (isAmbient) textDarker else manager.typeAccentColor.value
                        )
                    }
                }
            } else if (!gpsConnected.value) {
                CurvedLayout {
                    curvedRow() {
                        curvedComposable(modifier = CurvedModifier.padding(ArcPaddingValues(after = 6.dp))) {
                            Icon(
                                painter = painterResource(if(gpsConnected.value) R.drawable.gps_connected else R.drawable.gps_connecting),
                                contentDescription = "GPS status",
                                modifier = Modifier.size(20.dp).rotate(gpsRotating),
                                tint = if (isAmbient) textDarker else manager.typeAccentColor.value
                            )
                        }
                        curvedText(
                            text = txtGpsConnecting,
                            fontSize = 12.sp,
                            color = if (isAmbient) textDarker else manager.typeAccentColor.value
                        )
                    }
                }
            } else {
                TimeText()
            }
        },
    ) {
        HorizontalPager(
            state = pagerState,
            modifier = Modifier.fillMaxSize()
        ) { index ->
            trackingPages[index](
                manager, onStop,
                onScreenLock, onPauseResume
            )
        }
    }
}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutTrackingPreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true)

    RPoutTheme {
        WorkoutTrackingScreen(false, manager, {}, {}, {})
    }
}

@Composable
fun WorkoutTrackActionTab(manager: WorkoutManager, onStop: () -> Unit, onScreenLock: () -> Unit, onPauseResume: () -> Unit) {
    val scope = rememberCoroutineScope()

    Box(modifier = Modifier.fillMaxSize().padding(25.dp)) {
        Column(
            modifier = Modifier.fillMaxWidth().align(Alignment.Center),
            verticalArrangement = Arrangement.spacedBy(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {

            Row {
                Button(
                    onClick = { onScreenLock() },
                    colors = ButtonDefaults.primaryButtonColors(backgroundColor = backgroundLighter),
                    modifier = Modifier.width(65.dp).height(50.dp),
                    shape = RoundedCornerShape(20.dp)
                ) {
                    Icon(
                        painter = painterResource(R.drawable.lock),
                        contentDescription = "Settings",
                        modifier = Modifier.size(28.dp),
                        tint = manager.typeAccentColor.value
                    )
                }
            }

            Row(horizontalArrangement = Arrangement.SpaceBetween, modifier = Modifier.fillMaxWidth()) {
                Button(
                    onClick = { onStop() },
                    colors = ButtonDefaults.primaryButtonColors(backgroundColor = backgroundLighter),
                    modifier = Modifier.width(65.dp).height(50.dp),
                    shape = RoundedCornerShape(20.dp)
                ) {
                    Text(
                        text = "X",
                        fontSize = 24.sp,
                        color = de.rpjosh.rpout.android.activities.theme.error,
                        fontFamily = FontFamily.Monospace,
                        fontWeight = FontWeight.SemiBold
                    )
                }

                Button(
                    onClick = { onPauseResume() },
                    colors = ButtonDefaults.primaryButtonColors(backgroundColor = backgroundLighter),
                    modifier = Modifier.width(65.dp).height(50.dp),
                    shape = RoundedCornerShape(20.dp)
                ) {
                    Icon(
                        painter = painterResource(if(manager.state.value == State.PAUSED) R.drawable.start else R.drawable.pause),
                        contentDescription = "Settings",
                        modifier = Modifier.size(26.dp),
                        tint = manager.typeAccentColor.value
                    )
                }
            }
        }
    }
}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutTrackActionTabPreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true)

    RPoutTheme {
        WorkoutTrackActionTab(manager, {}, {}, {})
    }
}

@OptIn(ExperimentalHorologistApi::class)
@Composable
fun WorkoutTrackMainTab(manager: WorkoutManager) {

    val speedText =  remember { derivedStateOf { manager.workoutData.getFormattedSpeed(manager.type.id) } }

    Column(
        modifier = Modifier
            .padding(top = 40.dp, bottom = 20.dp, start = 3.dp, end = 3.dp)
            .fillMaxSize(),
        verticalArrangement = Arrangement.SpaceBetween
    ) {
        ActiveDurationText(
            checkpoint = manager.workoutData.activeDuration.value,
            state = manager.workoutData.exerciseState.value,
            content = {
                var durationFormatted = ""
                if (it.toHours() > 0) durationFormatted += "${it.toHours()}:"
                durationFormatted += String.format(Locale.ENGLISH, "%02d:%02d", it.toMinutesPart(), it.toSecondsPart())

                Text(
                    text = durationFormatted,
                    fontSize = 30.sp,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
            }
        )

        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.SpaceAround,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = String.format(Locale.ENGLISH, "%.2f", manager.workoutData.distance.value.value / 1000.0 ),
                fontSize = 30.sp,
                textAlign = TextAlign.Start
            )

            Text(
                text = manager.workoutData.heartRate.value.value.toString(),
                color = manager.workoutData.heartRate.color.value,
                fontSize = 30.sp,
                textAlign = TextAlign.End,
                //modifier = Modifier.padding(if(manager.workoutData.heartRate.value.value < 100) 6.dp else 0.dp)
            )
        }

        Text(
            text = speedText.value,
            fontSize = 30.sp,
            textAlign = TextAlign.Center,
            modifier = Modifier.fillMaxWidth()
        )
    }
}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutTrackMainTabPreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true)

    RPoutTheme {
        WorkoutTrackMainTab(manager)
    }
}

@Composable
fun WorkoutTrackExtraTab(manager: WorkoutManager) {
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(MaterialTheme.colors.background)
            .padding(start = 5.dp, end = 5.dp),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = "Not implemented",
            textAlign = TextAlign.Center
        )
    }
}

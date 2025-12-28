package de.rpjosh.rpout.android.activities.workout

import android.os.Bundle
import android.os.VibrationEffect
import android.os.Vibrator
import android.util.Log
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.animation.animateColorAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.FloatingActionButton
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.graphics.toColorInt
import androidx.core.os.ConfigurationCompat
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.lifecycle.repeatOnLifecycle
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundDarker
import de.rpjosh.rpout.android.activities.theme.error
import de.rpjosh.rpout.android.activities.theme.errorStatic
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textDarker
import de.rpjosh.rpout.android.activities.theme.textHint
import de.rpjosh.rpout.android.services.RealtimeLocationService
import de.rpjosh.rpout.android.services.WearSynchronization
import de.rpjosh.rpout.android.services.WorkoutUIState
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.HeartRateZone
import de.rpjosh.rpout.android.shared.models.WorkoutStatus
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.TranslationService
import kotlinx.coroutines.delay

class WorkoutTracking: ComponentActivity() {

    @Inject private lateinit var workoutController: WorkoutController
    @Inject(parameters = ["WorkoutTracking"]) private lateinit var logger: Logger
    @Inject private lateinit var wearSynchronization: WearSynchronization

    private var workoutType = mutableStateOf<WorkoutType?>(null)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        initData()

        enableEdgeToEdge()
        setContent {
            RPoutTheme {
                WorkoutTrackingScreen(
                    state = RealtimeLocationService.status,
                    type = workoutType.value,
                    onGoBack = {
                        finish()
                    },
                    onPause = { isResume ->
                        if(isResume) requestWorkoutStatus("resume")
                        else requestWorkoutStatus("pause")
                    },
                    onStop = {
                        requestWorkoutStatus("stop")
                    }
                )
            }
        }
    }

    private fun requestWorkoutStatus(status: String) {
        Thread{
            wearSynchronization.sendTextMessage(MessageType.WORKOUT_STATUS_UPDATE, status, onSuccess = {})
        }.start()
    }

    private fun initData() {
        Singleton.appController.injection.inject(WorkoutTracking::class.java, null, false, this)

        Thread{
            try {
                val typ = workoutController.dao().getType(RealtimeLocationService.status.type ?: 0)
                if (typ == null) {
                    logger.log("w", "Found no workout type for ID ${RealtimeLocationService.status.type}")
                    finish()
                    return@Thread
                }

                workoutType.value = typ
            } catch (ex: Exception) {
                logger.log("e", ex, "Failed to get workout type")
            }
        }.start()
    }

}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun WorkoutTrackingScreen(
    state: WorkoutUIState, type: WorkoutType?,
    onGoBack: () -> Unit,
    onPause: (resume: Boolean) -> Unit, onStop: () -> Unit,
) {
    val locale = ConfigurationCompat.getLocales(LocalConfiguration.current).get(0)
    val typeColor = type?.let{Color(it.tagDark.toColorInt())} ?: text

    val hideNavigationButtons = remember { mutableStateOf(false) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(backgroundDarker),
    ) {
        // Don't draw onto status bar
        with(LocalDensity.current) {
            val paddingTop = WindowInsets.statusBars.getTop(LocalDensity.current) - 24.dp.toPx()
            Spacer(Modifier.height(if (paddingTop < 0) 0.dp else paddingTop.toDp()))
        }

        Scaffold(
            topBar = {
                TopAppBar(
                    colors = TopAppBarDefaults.topAppBarColors(
                        containerColor = backgroundDarker,
                        titleContentColor = typeColor,
                    ),
                    windowInsets = WindowInsets(
                        top = 0.dp,
                        bottom = 0.dp,
                    ),
                    title = { Text(
                        text = type?.getName(TranslationService.Language.fromAndroidLocale(locale)) ?: "Workout"
                    )},
                    navigationIcon = {
                        IconButton(onClick = onGoBack) {
                            Icon(
                                painter = painterResource(R.drawable.arrow_back),
                                contentDescription = "Go back to main view",
                                tint = typeColor,
                            )
                        }
                    },
                    actions = {
                        IconButton(
                            modifier = Modifier.padding(top = 4.dp),
                            onClick = {
                                hideNavigationButtons.value = !hideNavigationButtons.value
                            }
                        ) {
                            Icon(
                                painter = painterResource(R.drawable.hide_image),
                                contentDescription = "Hide/show buttons",
                                tint = if(hideNavigationButtons.value) textDarker else textHint,
                            )
                        }
                    }
                )
            }
        ) { innerPadding ->
            Box(
                modifier = Modifier.fillMaxSize()
            ) {
                Column(
                    modifier = Modifier.padding(top = innerPadding.calculateTopPadding() + 4.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    StatusIndicator(state.state.value, type)
                    WorkoutData(state)
                }

                if (!hideNavigationButtons.value) {
                    ControlButtons(state, typeColor, onPause = onPause, onStop = onStop)
                }
            }

        }
    }
}

@Composable
fun ControlButtons(
    state: WorkoutUIState, typeColor: Color,
    onPause: (resume: Boolean) -> Unit, onStop: () -> Unit
) {
    val vibrator = LocalContext.current.getSystemService(Vibrator::class.java)

    var isStopActive by remember { mutableStateOf(false) }
    val animatedStopColor by animateColorAsState(
        targetValue = if (isStopActive) error else errorStatic,
        animationSpec = tween(durationMillis = 200),
        label = "StopButtonColor"
    )

    val isResume = state.state.value == WorkoutStatus.PAUSE

    LaunchedEffect(isStopActive) {
        if (isStopActive) {
            delay(2000L)
            isStopActive = false
        }
    }

    val onStopTap = {
        if (isStopActive) {
            onStop()
            vibrator.vibrate(VibrationEffect.createOneShot(80, 255))
            isStopActive = false
        } else {
            isStopActive = true
            vibrator.vibrate(VibrationEffect.createOneShot(60, 150))
        }
    }

    // Do not show buttons when workout is not running
    if (state.state.value in listOf(WorkoutStatus.PREPARE, WorkoutStatus.STOP)) {
        return Box{}
    }

    Box(
        modifier = Modifier.fillMaxSize(),
        contentAlignment = Alignment.BottomCenter
    ) {
        Row(
            modifier = Modifier
                .padding(bottom = 20.dp)
                .fillMaxWidth(0.9f),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            FloatingActionButton(
                onClick = { vibrator.vibrate(VibrationEffect.createOneShot(80, 255)); onPause(isResume) },
                containerColor = typeColor,
                contentColor = Color.White
            ) {
                Icon(
                    modifier = Modifier.size(36.dp),
                    painter = painterResource(
                        if (isResume) R.drawable.start else R.drawable.pause
                    ),
                    contentDescription = "Pause or Resume Workout"
                )
            }

            FloatingActionButton(
                onClick = onStopTap,
                containerColor = animatedStopColor,
                contentColor = Color.White,
            ) {
                Icon(
                    modifier = Modifier.size(36.dp),
                    painter = painterResource(R.drawable.stop),
                    contentDescription = "Stop Workout",
                )
            }
        }
    }
}

@Composable
fun WorkoutData(state: WorkoutUIState) {
    val locale = ConfigurationCompat.getLocales(LocalConfiguration.current).get(0)

    Column(
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(2.dp)
    ) {
        WorkoutSection(
            topLabel = stringResource(R.string.workout_speed),
            bottomLabel = ""
        ) {
            val speed = state.speed.floatValue * 3.6
            Text(
                text = String.format(locale, "%.2f",if(speed < 0.05) 0.0 else speed),
                fontSize = 66.sp,
                fontWeight = FontWeight.Bold,
            )
        }

        Row {
            WorkoutSection(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
                topLabel = stringResource(R.string.workout_duration),
                bottomLabel = ""
            ) {
                ActiveDurationText(state.durationCheckpoint.longValue, state.duration.longValue, state.state.value) {
                    val totalSeconds = it
                    val hours = totalSeconds / 3600
                    val minutes = (totalSeconds % 3600) / 60
                    val seconds = totalSeconds % 60

                    val durationFormatted = if (hours > 0) {
                        String.format(locale, "%d:%02d", hours, minutes, seconds)
                    } else {
                        String.format(locale, "%02d:%02d", minutes, seconds)
                    }

                    Text(
                        text = durationFormatted,
                        fontSize = 40.sp,
                        fontWeight = FontWeight.Bold,
                    )
                }
            }

            WorkoutSection(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
                topLabel = stringResource(R.string.workout_distance),
                bottomLabel = ""
            ) {
                Text(
                    text = String.format(
                        locale,
                        if(state.distance.intValue > 100_000) "%.1f" else "%.2f",
                        state.distance.intValue / 1000.0
                    ),
                    fontSize = 40.sp,
                    fontWeight = FontWeight.Bold,
                )
            }

            WorkoutSection(
                modifier = Modifier
                    .fillMaxWidth()
                    .weight(1f),
                topLabel = stringResource(R.string.workout_heartRate),
                bottomLabel = "Ø ${state.heartRateAv.intValue}"
            ) {
                HeartRateIndicator(state.heartRate.intValue)
            }
        }

    }
}

@Composable
fun StatusIndicator(status: WorkoutStatus, type: WorkoutType?) {
    val text = when (status) {
        WorkoutStatus.PAUSE -> stringResource(R.string.workout_status_paused)
        WorkoutStatus.PREPARE -> stringResource(R.string.workout_status_preparing)
        WorkoutStatus.STOP -> stringResource(R.string.workout_status_stopped)
        else -> ""
    }

    if (text == "") {
        return Box{}
    }

    Text(
        modifier = Modifier.fillMaxWidth(),
        text = text,
        fontSize = 16.sp,
        fontWeight = FontWeight.Bold,
        textAlign = TextAlign.Center,
        color = type?.let{Color(it.tagDark.toColorInt())} ?: de.rpjosh.rpout.android.activities.theme.text
    )
}

@Composable
fun HeartRateIndicator(heartRate: Int) {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .background(HeartRateZone.getZone(heartRate).color, shape = RoundedCornerShape(2.dp)),
        contentAlignment = Alignment.Center
    ) {
        Text(
            text = heartRate.toString(),
            textAlign = TextAlign.Center,
            fontSize = 40.sp,
            fontWeight = FontWeight.Bold,
        )
    }
}

@Composable
fun WorkoutSection(modifier: Modifier = Modifier, topLabel: String = "", bottomLabel: String = "", content: @Composable () -> Unit) {
    Column(
        modifier = modifier,
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = topLabel,
            fontSize = 12.sp,
            color = textHint,
        )

        content()

        Text(
            text = bottomLabel,
            fontSize = 12.sp,
            color = textHint,
        )
    }
}

@Composable
fun ActiveDurationText(
    checkpoint: Long,
    duration: Long,
    state: WorkoutStatus,
    content: @Composable (duration: Long) -> Unit
) {
    var activeMillis by remember { mutableLongStateOf(calculateDurationMillis(checkpoint, duration, state)) }

    val lifecycleOwner = LocalLifecycleOwner.current
    LaunchedEffect(state, checkpoint) {
        lifecycleOwner.repeatOnLifecycle(Lifecycle.State.RESUMED) {
            if (checkpoint == 0L) {
                activeMillis = 0
            } else if (state in listOf(WorkoutStatus.RUNNING, WorkoutStatus.HIGH_SAMPLING)) {
                val absoluteOffset = absoluteTimeOffsetMillis(checkpoint, duration)

                while (true) {
                    val now = System.currentTimeMillis()
                    // Delay until the next active duration second boundary
                    val delayInterval = nextTimeForOffset(now, absoluteOffset) - now
                    // Delay should delay for _at least_ the interval specified.
                    delay(delayInterval)
                    activeMillis = calculateDurationMillis(checkpoint, duration, state)
                }

            }
        }
    }

    content(activeMillis / 1000)
}

/** Calculates the total duration of the workout in milliseconds */
private fun calculateDurationMillis(
    checkpoint: Long,
    duration: Long,
    state: WorkoutStatus,
): Long {
    val delta = if (state in listOf(WorkoutStatus.RUNNING, WorkoutStatus.HIGH_SAMPLING)) {
        System.currentTimeMillis() - checkpoint
    } else {
        0L
    }

    return duration + delta
}

private fun absoluteTimeOffsetMillis(checkpoint: Long, duration: Long) = (checkpoint - duration) % 1000
private fun nextTimeForOffset(timeMillis: Long, offsetMillis: Long) = ((timeMillis - offsetMillis) / 1000) * 1000 + 1000 + offsetMillis

@Preview(showBackground = true, device = "id:pixel_7", showSystemUi = true, locale = "DE")
@Composable
fun WorkoutTrackingPreview() {
    val state = remember {
        WorkoutUIState(
            state = mutableStateOf(WorkoutStatus.PAUSE),
            distance = mutableIntStateOf(10425),
            speed = mutableFloatStateOf(10f),
            heartRate = mutableIntStateOf(120),
            heartRateAv = mutableIntStateOf(110),
            elevation = mutableIntStateOf(400),
            duration = mutableLongStateOf(29 * 1000 + 800),
            durationCheckpoint = mutableLongStateOf(System.currentTimeMillis()),
        )
    }

    RPoutTheme(darkTheme = true) {
        WorkoutTrackingScreen(state, WorkoutType(
            id = 0,
            nameDe = "Joggen",
            nameEn = "Running",
            icon = "",
            tagDark = "#E37029"
        ), onGoBack = {}, onPause = {}, onStop = {})
    }
}
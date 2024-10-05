package de.rpjosh.rpout.android.activities.main

import android.annotation.SuppressLint
import android.content.Intent
import android.os.Bundle
import android.os.VibrationEffect
import android.os.Vibrator
import android.util.Log
import android.view.WindowManager
import android.window.OnBackInvokedDispatcher
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
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
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.rotary.onPreRotaryScrollEvent
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.ambient.AmbientLifecycleObserver
import androidx.wear.compose.foundation.lazy.ScalingLazyColumn
import androidx.wear.compose.foundation.lazy.ScalingLazyColumnDefaults
import androidx.wear.compose.foundation.lazy.ScalingLazyListAnchorType
import androidx.wear.compose.foundation.lazy.ScalingLazyListState
import androidx.wear.compose.foundation.lazy.itemsIndexed
import androidx.wear.compose.material.Button
import androidx.wear.compose.material.ButtonDefaults
import androidx.wear.compose.material.Chip
import androidx.wear.compose.material.ChipDefaults
import androidx.wear.compose.material.CircularProgressIndicator
import androidx.wear.compose.material.Icon
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.PositionIndicator
import androidx.wear.compose.material.Scaffold
import androidx.wear.compose.material.Text
import androidx.wear.compose.material.TimeText
import androidx.wear.tooling.preview.devices.WearDevices
import androidx.work.Constraints
import androidx.work.ExistingPeriodicWorkPolicy
import androidx.work.ExistingWorkPolicy
import androidx.work.NetworkType
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.PeriodicWorkRequestBuilder
import androidx.work.WorkManager
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundLightDarker
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.defaultBackground
import de.rpjosh.rpout.android.activities.theme.overlayAmbient
import de.rpjosh.rpout.android.activities.theme.success
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textBlue
import de.rpjosh.rpout.android.activities.theme.textDarker
import de.rpjosh.rpout.android.services.Uploader
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
import de.rpjosh.rpout.android.shared.models.HeartRateZone
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.shared.workout.Workout
import de.rpjosh.rpout.android.shared.workout.WorkoutManager
import de.rpjosh.rpout.android.workout.WorkoutTrackService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.DelicateCoroutinesApi
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.newSingleThreadContext
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext
import java.time.Duration
import java.time.Instant
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter
import java.util.Locale
import java.util.TimeZone
import java.util.concurrent.TimeUnit
import kotlin.math.abs

/**
 * Simple data class that is used to display a success / error indicated by a circle
 * around the smart watch display
 */
data class OperationState(

    /** Whether an error or an success should be displayed */
    var isError: Boolean = false,

    /** State to indicate a recomposition of the animation */
    val trigger: MutableState<Boolean> = mutableStateOf(false),

    /** Whether the animation should really be displayed */
    var showAnimation: Boolean = false
) {

    /** Starts the animation for the provided error / success */
    fun animate(isError: Boolean) {
        this.isError = isError
        showAnimation = true
        trigger.value = !trigger.value
    }

}

class WorkoutFinishedActivity: ComponentActivity() {

    @OptIn(ExperimentalCoroutinesApi::class, DelicateCoroutinesApi::class)
    private val scope = CoroutineScope(newSingleThreadContext("uploadWorkout"))

    private lateinit var workoutController: WorkoutController

    private val workoutSummary = mutableStateOf(WorkoutSummary())
    private val lastWorkouts = mutableStateOf( listOf<GpsWorkout>() )
    private val workoutTypes = mutableStateOf( listOf<WorkoutType>() )

    private val operationState = OperationState()

    /** Whether the workout sync job has to be scheduled when leaving the activity */
    @Volatile private var pushSyncJobOnExit = true

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()

        super.onCreate(savedInstanceState)

        // No workout manager available
        val manager = WorkoutManager.workoutManager
        if (manager == null) {
            finish()
            return
        }
        workoutController = Singleton.appController.injection.inject(WorkoutController::class.java, null, false)

        // Stop foreground service
        val serviceIntent = Intent(RPout.getAppContext(), WorkoutTrackService::class.java)
        serviceIntent.action = "STOP"
        ContextCompat.startForegroundService(this, serviceIntent)

        // Remove reference (workouts are finished)
        WorkoutManager.workoutManager = null

        // Do not turn display off
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)

        // Vibrate the device
        val vibrator = baseContext.getSystemService(Vibrator::class.java)
        val pattern = longArrayOf(90,  55, 75, 40, 90, 55, 75, 40, 90,  55, 75, 40, 90, 55, 75, 40, 90, 55, 75, 50)
        val amplitude = intArrayOf(255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0)
        val vibrationEffect = VibrationEffect.createWaveform(pattern, amplitude,-1)
        vibrator.vibrate(vibrationEffect)

        // Handle finish
        scope.launch {
            manager.stop()

            // Finish workout, update state and sync workout to server
            var workout: GpsWorkout
            synchronized(manager.dataLock) {
                workout = manager.gpsWorkout.copy()
                workoutController.dao().updateWorkout(workout)
                workout.points = workoutController.dao().getWorkoutPoints(manager.gpsWorkout.id).toMutableList()

                // Calculate heart rate zones
                manager.workoutSummary.heartRateZones = manager.workoutSummary.getHeartRateZoneStats(workout.points)

                // Set value for UI
                workoutSummary.value = manager.workoutSummary
                lastWorkouts.value = workoutController.dao().getMergableWorkouts()
                workoutTypes.value = workoutController.dao().getAllTypes()
            }

            Singleton.appController.sharedLogger.log("d", "Trying to push finished workout")
            val serverSummary = workoutController.pushWorkout(workout)
            if (serverSummary == null) {
                // Upload failed => schedule work manager task to retry it (if it was not done previously)
                if (pushSyncJobOnExit) {
                    pushSyncJobOnExit = false
                    val constraint = Constraints.Builder()
                        .setRequiredNetworkType(NetworkType.CONNECTED)
                        .build()
                    val worker = OneTimeWorkRequestBuilder<Uploader>()
                        .setConstraints(constraint)
                        .addTag(Uploader.TAG_UPLOADER)
                        .build()
                    WorkManager.getInstance(RPout.getAppContext()).enqueueUniqueWork(Uploader.TAG_UPLOADER_PRIO, ExistingWorkPolicy.REPLACE, worker)
                }
            } else {
                // Merge workout summary data
                serverSummary.heartRateZones = workoutSummary.value.heartRateZones
                serverSummary.typeAccentColor = workoutSummary.value.typeAccentColor

                // Update summary
                workoutSummary.value = serverSummary
            }
        }

        // Finish activity if user is going back
        onBackInvokedDispatcher.registerOnBackInvokedCallback(OnBackInvokedDispatcher.PRIORITY_DEFAULT) { onExit() }

        setContent {
            RPoutTheme {
                Box {
                    WorkoutEndScreen(
                        summary = workoutSummary.value,
                        lastWorkouts = lastWorkouts.value,
                        activityTypes = workoutTypes.value,
                        onOk = { onExit() },
                        onWorkoutMerge = { onMerge(it) }
                    )
                    PulsatingCircle(operationState)
                }
            }
        }
    }

    private fun onExit() {
        // Open main UI (don't show start activity screen)
        val intent = Intent().apply {
            setAction(Intent.ACTION_MAIN)
            addCategory(Intent.CATEGORY_HOME)
        }
        startActivity(intent)

        // Push a sync job if it wasn't done already (activity was exited immediately before the initial push wasn't even tried)
        if (pushSyncJobOnExit) {
            pushSyncJobOnExit = false
            val constraint = Constraints.Builder()
                .setRequiredNetworkType(NetworkType.CONNECTED)
                .build()
            val worker = OneTimeWorkRequestBuilder<Uploader>()
                .setConstraints(constraint)
                .addTag(Uploader.TAG_UPLOADER)
                .build()
            WorkManager.getInstance(RPout.getAppContext()).enqueueUniqueWork(Uploader.TAG_UPLOADER_PRIO, ExistingWorkPolicy.REPLACE, worker)
        }

        finish()
    }

    private fun onMerge(withId: Long) {
        Thread{
            operationState.animate(!workoutController.mergeWorkout(withId, workoutSummary.value.id))

            // Vibrate device
            val vibrator = baseContext.getSystemService(Vibrator::class.java)
            val pattern = longArrayOf(150,  100, 65)
            val amplitude = intArrayOf(255, 0, 255)
            vibrator.vibrate(VibrationEffect.createWaveform(pattern, amplitude,-1))

        }.start()
    }
}

@SuppressLint("DefaultLocale")
@Composable
fun WorkoutEndScreen(
    summary: WorkoutSummary,
    lastWorkouts: List<GpsWorkout>, activityTypes: List<WorkoutType>,
    onOk: () -> Unit, onWorkoutMerge: (id: Long) -> Unit
) {
    val listState = remember { ScalingLazyListState(initialCenterItemIndex = 0) }

    val duration = Duration.ofSeconds(summary.duration.toLong())
    var durationFormatted = ""
    if (duration.toHours() > 0) durationFormatted += "${duration.toHours()}:"
    durationFormatted += String.format(Locale.ENGLISH, "%02d:%02d", duration.toMinutesPart(), duration.toSecondsPart())

    val rowWidth = remember { mutableIntStateOf(0) }

    Scaffold(
        positionIndicator = { PositionIndicator(scalingLazyListState = listState) },
    ) {
        ScalingLazyColumn(
            modifier = Modifier.fillMaxWidth(),
            state = listState,
            flingBehavior = ScalingLazyColumnDefaults.snapFlingBehavior(state = listState),
            contentPadding = PaddingValues(
                top = 4.dp,
                start = 12.dp,
                end = 12.dp,
                bottom = 35.dp
            ),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(2.dp),
            anchorType = ScalingLazyListAnchorType.ItemCenter,
            // Do not center the first elements index => use contentPadding or AutoCenteringParams(itemIndex = 3)
            autoCentering = null
        ) {

            // Earned PAI score
            item(key = "pai") {
                Column(horizontalAlignment = Alignment.CenterHorizontally, verticalArrangement = Arrangement.spacedBy(8.dp)) {
                    Box {
                        Image(
                            painter = painterResource(R.drawable.pai),
                            contentDescription = "PAI",
                            modifier = Modifier.width(64.dp)
                        )
                        Text(
                            text = if (summary.id == 0L) "?" else summary.pai.toString(),
                            textAlign = TextAlign.Center,
                            fontWeight = FontWeight.Bold,
                            fontSize = 24.sp,
                            modifier = Modifier.align(Alignment.Center),
                            fontFamily = FontFamily.SansSerif
                        )
                    }
                    Text(stringResource(R.string.main_paiEarned), fontSize = 16.sp, textAlign = TextAlign.Center)
                }
            }

            // Speed
            item(key = "speed") {
                Row(
                    verticalAlignment = Alignment.CenterVertically, horizontalArrangement = Arrangement.spacedBy(4.dp),
                    modifier = Modifier.padding(top = 10.dp, bottom = 6.dp)
                ) {
                    Icon(
                        painter = painterResource(R.drawable.speed),
                        contentDescription = "Speed",
                        modifier = Modifier.width(24.dp),
                        tint = textDarker
                    )
                    Text(
                        text = summary.getFormattedSpeed(summary.typeId),
                        fontSize = 17.sp,
                        modifier = Modifier.padding(start = 5.dp)
                    )
                    Text(
                        text = if(summary.getFormattedSpeed(summary.typeId).contains(":")) "min/km" else "km/h",
                        fontSize = 15.sp,
                        color = textDarker
                    )
                }
            }

            // Heart rate
            item(key = "heartrate") {
                Column {
                    Text(
                        text = stringResource(R.string.main_heartRate),
                        fontSize = 17.sp,
                        color = summary.typeAccentColor,
                        modifier = Modifier.padding(top= 12.dp, bottom = 10.dp).fillMaxWidth(),
                        textAlign = TextAlign.Center,
                        fontWeight = FontWeight.SemiBold
                    )

                    DataInfoRow(
                        icon = R.drawable.heart,
                        value = summary.heartRateAv.toString(), unit = "bpm",
                        accentColor = HeartRateZone.getZone(summary.heartRateAv).color
                    )
                    DataInfoRow(
                        icon = R.drawable.heart_max, iconOverlayText = "M",
                        value = summary.heartRateMax.toString(), unit = "bpm",
                        accentColor = HeartRateZone.getZone(summary.heartRateMax).color
                    )
                    Spacer(Modifier.height(6.dp))
                }
            }
            items(5, key = { "zone-" + it }) { index ->
                val it = summary.heartRateZones[index+1]

                var percent = if (it.duration.toSeconds() == 0L) 100.0 else summary.duration.toDouble() / it.duration.toSeconds()
                if (percent > 20.0) percent = 20.0
                else percent -= 0.3

                Box(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(start = 5.dp, end = 5.dp, top = 2.dp, bottom = 2.dp)
                        .background(
                            backgroundLightDarker,
                            shape = RoundedCornerShape(12.dp)
                        )
                        .onGloballyPositioned {
                            rowWidth.intValue = it.size.width
                        }
                ) {
                    Box(
                        modifier = Modifier
                            .height(32.dp)
                            .width(with(LocalDensity.current) { (rowWidth.intValue / percent).toInt().toDp() + 12.dp } )
                            .background(Color(red = it.color.red, green = it.color.green, blue = it.color.blue, alpha = 0.6f), RoundedCornerShape(12.dp))
                    ) {}

                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(start = 24.dp, end = 6.dp, top = 4.dp, bottom = 4.dp )
                            .height(24.dp),
                        horizontalArrangement = Arrangement.spacedBy(2.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        Text(it.getName(), fontSize = 12.sp, fontWeight = FontWeight.SemiBold)
                        Text(
                            text = it.getDuration(),
                            textAlign = TextAlign.End,
                            modifier = Modifier.fillMaxWidth(),
                            fontSize = 13.sp,
                            fontWeight = FontWeight.SemiBold
                        )
                    }
                }
            }

            // General stats
            item(key = "general-stats") {
                Text(
                    text = stringResource(R.string.main_stats),
                    fontSize = 17.sp,
                    color = summary.typeAccentColor,
                    modifier = Modifier.padding(top= 12.dp, bottom = 10.dp).fillMaxWidth(),
                    textAlign = TextAlign.Center,
                    fontWeight = FontWeight.SemiBold
                )
            }
            item(key = "duration") {
                DataInfoRow(
                    icon = R.drawable.stopwatch,
                    value = summary.getDuration(), unit = "",
                )
            }
            item(key = "distance") {
                DataInfoRow(
                    icon = R.drawable.distance,
                    value = String.format(Locale.ENGLISH, "%.2f", summary.distance / 1000.0), unit = "km",
                )
            }
            item(key = "speedAv") {
                DataInfoRow(
                    icon = R.drawable.stopwatch,
                    value = summary.getFormattedSpeed(summary.typeId),
                    unit = if(summary.getFormattedSpeed(summary.typeId).contains(":")) "min/km" else "km/h",
                )
            }
            item(key = "steps") {
                DataInfoRow(
                    icon = R.drawable.steps,
                    value = summary.steps.toString(),
                )
            }
            item(key = "calories") {
                DataInfoRow(
                    icon = R.drawable.calories,
                    value = summary.calories.toString(), unit = "calories",
                )
            }
            item(key = "elevation-up") {
                DataInfoRow(
                    icon = R.drawable.upstairs,
                    value = summary.elevationUp.toString(), unit = "m",
                )
            }
            item(key = "elevation-down") {
                DataInfoRow(
                    icon = R.drawable.downstairs,
                    value = abs(summary.elevationDown).toString(), unit = "m",
                )
            }

            // Merge with existing workouts
            if (lastWorkouts.isNotEmpty() && summary.id != 0L && activityTypes.isNotEmpty()) {
                item(key = "merge-workouts") {
                    Text(
                        text = stringResource(R.string.main_mergeWith),
                        fontSize = 17.sp,
                        color = summary.typeAccentColor,
                        modifier = Modifier.padding(top= 12.dp, bottom = 10.dp).fillMaxWidth(),
                        textAlign = TextAlign.Center,
                        fontWeight = FontWeight.SemiBold
                    )
                }

                // Formatter for date
                val formatter = DateTimeFormatter.ofPattern("dd.MM HH:mm")
                itemsIndexed(lastWorkouts, key = { _, item -> "merge-${item.id}" }) { _, item ->
                    // Get the type for the workout
                    val type = activityTypes.find { it.id == item.type } ?: activityTypes[0]

                    Chip(
                        modifier = Modifier.fillMaxWidth().padding(top = 2.dp, bottom = 2.dp),
                        colors = ChipDefaults.primaryChipColors(
                            backgroundColor = backgroundLighter
                        ),
                        icon = {
                            SvgIcon(
                                svgString = type.icon,
                                size = 28.dp,
                                hexTint = type.tagDark,
                            )
                        },
                        label = {
                            Text(
                                modifier = Modifier.fillMaxWidth(),
                                color = text,
                                text = type.getName(Tr.getUsedLanguage())
                            )
                        },
                        secondaryLabel = {
                            Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {

                            }
                            Text(
                                text = LocalDateTime.ofInstant(Instant.ofEpochSecond(item.startTime), ZoneId.systemDefault()).format(formatter)
                            )
                        },
                        onClick = { onWorkoutMerge(item.serverId) }
                    )
                }
            }

            item(key = "ok") {
                Button(
                    onClick = { onOk() },
                    colors = ButtonDefaults.primaryButtonColors(
                        backgroundColor = summary.typeAccentColor
                    ),
                    modifier = Modifier.padding(top = 10.dp)
                ) {
                    Icon(
                        painter = painterResource(R.drawable.check),
                        modifier = Modifier.size(22.dp),
                        contentDescription = "Ok",
                        tint = Color.Black
                    )
                }
            }
        }
    }
}

@Composable
/** Displays a simple row with an icon on the provided value (with unit) */
fun DataInfoRow(icon: Int, value: String, unit: String = "", accentColor: Color = text, iconOverlayText: String = "") {
    Box(
        modifier = Modifier
            .fillMaxWidth()
            .padding(start = 2.dp, end = 2.dp, top = 2.dp, bottom = 2.dp)
            .background(backgroundLightDarker, shape = RoundedCornerShape(14.dp))
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(start = 10.dp, end = 8.dp, top = 7.dp, bottom = 7.dp),
            horizontalArrangement = Arrangement.spacedBy(2.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Box(
                modifier = Modifier.width(18.dp).height(18.dp)
            ) {
                Icon(
                    painter = painterResource(icon),
                    contentDescription = "Icon",
                    modifier = Modifier.height(20.dp),
                    tint = accentColor
                )
                if (iconOverlayText != "") {
                    Text(
                        text = iconOverlayText,
                        fontSize = 10.sp,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.align(Alignment.Center).padding(bottom = 2.dp),
                        color = accentColor
                    )
                }
            }
            Text(
                text = value,
                fontSize = 14.sp,
                modifier = Modifier.padding(start = 6.dp),
                color = accentColor
            )
            Text(
                text = unit,
                fontSize = 13.sp,
                color = Color(red = accentColor.red, green = accentColor.green, blue = accentColor.blue, alpha = 0.8f),
                modifier = Modifier.padding(top = 1.dp)
            )
        }
    }
}

@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutEndPreview() {
    val lastWorkouts = arrayListOf(
        GpsWorkout(type = 1, startTime = (System.currentTimeMillis() / 1000) - 60 * 24 ),
        GpsWorkout(type = 4, startTime = (System.currentTimeMillis() / 1000) - 26 * 60 * 60)
    )

    RPoutTheme {
        Box(modifier = Modifier.fillMaxSize().background(defaultBackground)) {
            WorkoutEndScreen(
                WorkoutSummary(id = 1, typeId = 1, speedAv = 306, steps = 1345, heartRateMax = 167, heartRateAv = 144, duration = 203, typeAccentColor = Color.Blue),
                lastWorkouts,
                sampleActivityTypes,
                {}, {}
            )
        }
    }
}

@Composable
fun PulsatingCircle(state: OperationState) {

    /* The current animation state. 0 = gone, 1 = visible */
    val animationState = remember { mutableIntStateOf(0) }

    // Animate the alpha from 0 to 1 and then back to 0 when the variable changes
    val animatedAlpha by animateFloatAsState(
        targetValue = if (animationState.intValue == 1) 1f else 0f,
        animationSpec = tween(
            durationMillis = 400,
            easing = LinearEasing
        ), label = "errorSuccessIndicator"
    )

    LaunchedEffect(state.trigger.value) {
        // Do not show animation for initial render
        if (!state.showAnimation) return@LaunchedEffect

        // Make circle visible whenever variable was changed
        animationState.intValue = 1

        // Reset color after 2 seconds
        delay(1500)

        animationState.intValue = 0
    }

    Canvas(modifier = Modifier.fillMaxSize()) {
        // Draw a circle around the edge of the screen (for a round watch face)
        drawCircle(
            color = if (state.isError) de.rpjosh.rpout.android.activities.theme.error.copy(alpha = animatedAlpha) else success.copy(alpha = animatedAlpha),
            //color = success,
            style = Stroke(
                width = 3.dp.toPx(),
                cap = StrokeCap.Round
            ),
            radius = size.minDimension / 2 - 2.dp.toPx(),
        )
    }
}

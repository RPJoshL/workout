package de.rpjosh.rpout.android.activities.main

import android.content.Intent
import android.os.Bundle
import android.util.Log
import android.view.WindowManager
import android.window.OnBackInvokedDispatcher
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableLongStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.rotary.onPreRotaryScrollEvent
import androidx.compose.ui.layout.onGloballyPositioned
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.painterResource
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
import androidx.wear.compose.material.Chip
import androidx.wear.compose.material.ChipDefaults
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
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textBlue
import de.rpjosh.rpout.android.activities.theme.textDarker
import de.rpjosh.rpout.android.services.Uploader
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
import de.rpjosh.rpout.android.shared.models.HeartRateZone
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
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
import java.util.Locale
import java.util.concurrent.TimeUnit

class WorkoutFinishedActivity: ComponentActivity() {

    @OptIn(ExperimentalCoroutinesApi::class, DelicateCoroutinesApi::class)
    private val scope = CoroutineScope(newSingleThreadContext("uploadWorkout"))

    private lateinit var workoutController: WorkoutController

    private val createdWorkoutId = mutableLongStateOf(0)
    private val workoutSummary = mutableStateOf(WorkoutSummary())

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
            }

            val workoutSummary = workoutController.pushWorkout(workout)
            if (workoutSummary == null) {
                // Upload failed => schedule work manager task to retry it
                val constraint = Constraints.Builder()
                    .setRequiredNetworkType(NetworkType.CONNECTED)
                    .build()
                val worker = OneTimeWorkRequestBuilder<Uploader>()
                    .setConstraints(constraint)
                    .addTag(Uploader.TAG_UPLOADER)
                    .build()
                WorkManager.getInstance(RPout.getAppContext()).enqueueUniqueWork(Uploader.TAG_UPLOADER_PRIO, ExistingWorkPolicy.REPLACE, worker)
            } else {
                createdWorkoutId.longValue = workoutSummary.id
            }
        }

        // Finish activity if user is going back
        onBackInvokedDispatcher.registerOnBackInvokedCallback(OnBackInvokedDispatcher.PRIORITY_DEFAULT) {
            finish()
        }

        setContent {
            RPoutTheme {
                WorkoutEndScreen(createdWorkoutId.longValue, workoutSummary.value)
            }
        }
    }

}


@Composable
fun WorkoutEndScreen(syncId: Long, summary: WorkoutSummary) {
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
                            text = if (summary.id == 0L) "22" else summary.pai.toString(),
                            textAlign = TextAlign.Center,
                            fontWeight = FontWeight.Bold,
                            fontSize = 24.sp,
                            modifier = Modifier.align(Alignment.Center),
                            fontFamily = FontFamily.SansSerif
                        )
                    }
                    Text("PAIs verdient!", fontSize = 16.sp, textAlign = TextAlign.Center)
                }
            }

            // Heart rate
            item(key = "heartrate") {
                Column {
                    Text(
                        text = "Herzfrequenz",
                        fontSize = 16.sp,
                        color = summary.typeAccentColor,
                        modifier = Modifier.padding(top= 12.dp, bottom = 10.dp).fillMaxWidth(),
                        textAlign = TextAlign.Center
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
            items(5) { index ->
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
                    text = "Statistiken",
                    fontSize = 16.sp,
                    color = summary.typeAccentColor,
                    modifier = Modifier.padding(top= 12.dp, bottom = 10.dp).fillMaxWidth(),
                    textAlign = TextAlign.Center
                )
            }

            item {
                Text(
                    text = "End screen = $syncId",
                    textAlign = TextAlign.Center,
                    fontSize = 10.sp
                )
            }
            item {
                Text( text = "Kalorien: ${summary.calories}", fontSize = 10.sp )
            }
            item {
                Text(text = "Herzrate Avg: ${summary.heartRateAv}", fontSize = 10.sp)
            }
            item {
                Text ( text = "Herzrate Max: ${summary.heartRateMax}", fontSize = 10.sp )
            }
            item {
                Text ( text = "Schritte: ${summary.steps}", fontSize = 10.sp )
            }
            item {
                Text ( text = "Speed: ${summary.getFormattedSpeed(summary.typeId)}", fontSize = 10.sp )
            }
            item {
                Text ( text = "Elevation (Up): ${summary.elevationUp}", fontSize = 10.sp )
            }
            item {
                Text ( text = "Elevation (Down): ${summary.elevationDown}", fontSize = 10.sp )
            }
            item {
                Text ( text = "Distanz: " + (String.format(Locale.ENGLISH, "%.2f", summary.distance / 1000.0 )), fontSize = 10.sp )
            }
            item {
                Text ( text = "Dauer: $durationFormatted", fontSize = 10.sp )
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
    RPoutTheme {
        Box(modifier = Modifier.fillMaxSize().background(defaultBackground)) {
            WorkoutEndScreen(0, WorkoutSummary(typeId = 1, speedAv = 306, steps = 1345, heartRateMax = 167, heartRateAv = 144, duration = 203))
        }
    }
}
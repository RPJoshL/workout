package de.rpjosh.rpout.android.activities.main.workouttracking

import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.health.services.client.data.ExerciseState
import androidx.health.services.client.data.ExerciseUpdate
import androidx.wear.compose.material3.Text
import com.google.android.horologist.annotations.ExperimentalHorologistApi
import de.rpjosh.rpout.android.activities.WearPreview
import com.google.android.horologist.health.composables.ActiveDurationText
import de.rpjosh.rpout.android.activities.theme.FontSourceSanseProSemibold
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.shared.models.ActivityType
import de.rpjosh.rpout.android.workout.WorkoutManager
import java.time.Duration
import java.time.Instant
import java.util.Locale

@OptIn(ExperimentalHorologistApi::class)
@Composable
fun FoilingMetricsScreen(manager: WorkoutManager) {
    val mainData = manager.workoutData
    val data = manager.foilingData


    Box(modifier = Modifier.padding(end = 3.dp, bottom = 2.dp)) {
        HeartRateIndicator(manager)

        Column(
            modifier = Modifier
                .padding(top = 24.dp, bottom = 15.dp, start = 8.dp, end = 8.dp)
                .fillMaxSize(),
            verticalArrangement = Arrangement.SpaceBetween,
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            if(data.isActive.value) {
                ActiveDurationText(
                    checkpoint = data.checkpoint.value,
                    state = ExerciseState.ACTIVE,
                    content = {
                        Text(
                            text = formatDuration(it),
                            fontSize = 35.sp,
                            textAlign = TextAlign.Center,
                            modifier = Modifier.fillMaxWidth(),
                            fontFamily = FontFamily(FontSourceSanseProSemibold),
                            color = Color.Cyan
                        )
                    }
                )
            } else {
                Text(
                    text = formatDuration(Duration.ofSeconds(data.totalDuration.value)),
                    fontSize = 35.sp,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth(),
                    fontFamily = FontFamily(FontSourceSanseProSemibold),
                    color = text
                )
            }

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceAround,
                verticalAlignment = Alignment.CenterVertically
            ) {
                TextWithHint(
                    txt = String.format(Locale.ENGLISH, "%.2f", data.distance.value / 1000.0),
                    hint = if(data.isActive.value) "km" else "last"
                )

                if(data.isActive.value) {
                    TextWithHint(
                        txt = mainData.heartRate.value.value.toString(),
                        hint = "bpm",
                        color = mainData.heartRate.color.value
                    )
                } else {
                    TextWithHint(
                        txt = String.format(Locale.ENGLISH, "#%d", data.count.value),
                        hint = "count"
                    )
                }
            }

            if(data.isActive.value) {
                TextWithHint(
                    txt = String.format(Locale.ENGLISH, "%.1f", mainData.speed.value.value * 3.6),
                    hint = "km/h"
                )
            } else {
                TextWithHint(
                    txt = formatDuration(Duration.ofSeconds(data.lastSessionDuration.value)),
                    hint = "last",
                    fontSize = 32.sp
                )
            }
        }
    }
}

@WearPreview
@Composable
fun FoilingMetricsScreenPreview() {
    val manager = WorkoutManager.forPreview(
        totalKm = 5.67, heartRate = 145,
        typeId = ActivityType.TYPE_PUMP_FOILING.ordinal.toLong()
    )

    manager.foilingData.isActive.value = true
    manager.foilingData.distance.value = 1234.0
    manager.foilingData.checkpoint.value = ExerciseUpdate.ActiveDurationCheckpoint(Instant.now(), Duration.ZERO)
    manager.foilingData.totalDuration.value = 37 * 60 + 35
    manager.foilingData.lastSessionDuration.value = 16 * 60 + 12
    manager.foilingData.count.intValue = 8

    RPoutTheme {
        WorkoutTrackingScreen(
            false, manager, null, 2, {},
            {}, {}, {}
        )
    }
}

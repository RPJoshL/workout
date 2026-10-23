package de.rpjosh.rpout.android.activities.main.workoutstart

import android.content.Intent
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.runtime.Composable
import androidx.compose.runtime.MutableState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import de.rpjosh.rpout.android.activities.WearPreview
import de.rpjosh.rpout.android.shared.models.GPSMode
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import androidx.wear.compose.foundation.lazy.ScalingLazyColumn
import androidx.wear.compose.foundation.lazy.ScalingLazyListState
import androidx.wear.compose.material3.AlertDialog
import androidx.wear.compose.material3.AlertDialogDefaults
import androidx.wear.compose.material3.Icon
import androidx.wear.compose.material3.ScreenScaffold
import androidx.wear.compose.material3.SwitchButton
import androidx.wear.compose.material3.SwitchButtonDefaults
import androidx.wear.compose.material3.Text
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.lerp
import androidx.compose.ui.platform.LocalInspectionMode
import androidx.compose.ui.text.font.FontWeight
import androidx.wear.compose.material3.onehandedgesture.OneHandedGestureAction
import androidx.wear.compose.material3.onehandedgesture.OneHandedGestureClickIndicator
import androidx.wear.compose.material3.onehandedgesture.OneHandedGestureClickIndicatorState
import androidx.wear.compose.material3.onehandedgesture.oneHandedGesture
import androidx.wear.compose.material3.onehandedgesture.rememberOneHandedGestureConfiguration
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.backgroundSelection
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textDarker
import de.rpjosh.rpout.android.shared.models.DismissGestureActionTracking
import de.rpjosh.rpout.android.shared.models.PrimaryGestureActionTracking
import de.rpjosh.rpout.android.workout.WorkoutManager
import de.rpjosh.rpout.android.workout.WorkoutTrackService
import kotlinx.coroutines.delay
import kotlinx.coroutines.isActive
import kotlinx.coroutines.launch
import kotlin.time.Duration.Companion.milliseconds

@Composable
fun SettingsPage(manager: WorkoutManager, gpsMode: MutableState<GPSMode>, isVisible: Boolean = true) {
    val context = LocalContext.current
    val scope = rememberCoroutineScope()

    val listState = remember { ScalingLazyListState() }

    val liveUpdates = remember { mutableStateOf(manager.type.liveUpdates) }
    val highSamplingInterval = remember { mutableStateOf(manager.type.useHighSamplingInterval) }
    val primaryGestureAction = remember { mutableStateOf(manager.type.getPrimaryGestureAction()) }
    val dismissGestureAction = remember { mutableStateOf(manager.type.getDismissGestureAction()) }
    val useTiltToWake = remember { mutableStateOf(manager.type.useTiltToWake) }

    val color = manager.typeAccentColor.value

    val gestureConfig = rememberOneHandedGestureConfiguration(
        action = OneHandedGestureAction.Primary,
        gestureId = "workout_type_settings_demo"
    )
    val gestureIndicationState = remember {
        OneHandedGestureClickIndicatorState()
    }

    val dismissGestureConfig = rememberOneHandedGestureConfiguration(
        action = OneHandedGestureAction.Dismiss,
        gestureId = "workout_type_settings_demo_dismiss",
    )
    val dismissGestureIndicationState = remember {
        OneHandedGestureClickIndicatorState()
    }

    val isPreview = LocalInspectionMode.current
    LaunchedEffect(isVisible, primaryGestureAction.value, dismissGestureAction.value) {
        if (isPreview || !isVisible) return@LaunchedEffect

        while (isActive) {
            try {
                gestureIndicationState.showIndicator()
                dismissGestureIndicationState.showIndicator()
                delay(2000.milliseconds)
            } catch (_: Exception) {
                delay(100.milliseconds)
            }
        }
    }

    ScreenScaffold(
        scrollState = listState,
        timeText = {}
    ) {
        ScalingLazyColumn(
            modifier = Modifier.fillMaxWidth(),
            state = listState,
            contentPadding = PaddingValues(
                top = 16.dp,
                start = 14.dp,
                end = 14.dp,
                bottom = 50.dp
            ),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(4.dp),
            autoCentering = null
        ) {
            item(key = "dataHeader") {
                SectionHeader(R.string.main_sectionData, color)
            }

            if(manager.type.isGPSTrackingSupported()) {
                item(key = "gps-toggle") {
                    ButtonToggle(
                        iconRes = R.drawable.gps_connected,
                        label = stringResource(R.string.main_gps),
                        hintText = stringResource(R.string.main_gps_hint),
                        accentColor = color,
                        selectedIndex = gpsMode.value.id,
                        onSelectionChanged = {
                            val newMode = GPSMode.fromId(it)
                            gpsMode.value = newMode

                            Thread { manager.changeSettings(context, gpsMode = newMode) }.start()
                        },
                        buttons = listOf(
                            squareButton(R.drawable.ic_phone),
                            squareButton(R.drawable.outline_location_disabled_24),
                        )
                    )
                }
            }

            item(key = "accurate-data") {
                SettingsToggle(
                    text = stringResource(R.string.main_accurateData),
                    checked = highSamplingInterval.value,
                    color = color,
                ) {
                    Thread { manager.changeSettings(context, highSampling = it) }.start()
                    highSamplingInterval.value = it
                }
            }

            item(key = "live-data") {
                SettingsToggle(
                    text = stringResource(R.string.main_liveData),
                    checked = liveUpdates.value,
                    color = color,
                ) {
                    Thread { manager.changeSettings(context, liveData = it) }.start()
                    liveUpdates.value = it
                }
            }

            // Sektion Gesten
            item(key = "gesturesHeader") {
                SectionHeader(R.string.main_sectionGestures, color)
            }

            item(key = "gestures-toggle") {
                ButtonToggle(
                    modifier = Modifier.oneHandedGesture(
                        gestureConfiguration = gestureConfig,
                        onGestureLabel = "change primary gesture",
                        onGestureAvailable = {
                            scope.launch {
                                try { gestureIndicationState.showIndicator() } catch (_: Exception) {}
                            }
                        },
                        onGesture = {
                            val nextId = (primaryGestureAction.value.id + 1) % 3
                            val newAction = PrimaryGestureActionTracking.fromId(nextId)
                            primaryGestureAction.value = newAction
                            Thread { manager.changeSettings(context, primaryGestureAction = newAction) }.start()
                        }
                    ),
                    icon = {
                        OneHandedGestureClickIndicator(
                            modifier = Modifier.size(22.dp),
                            gestureConfiguration = gestureConfig,
                            state = gestureIndicationState,
                            gestureIndicatorTint = if (primaryGestureAction.value != PrimaryGestureActionTracking.DISABLED) color else textDarker
                        ){
                            // No fallback icon is not the best way to indicate that gestures are not supported.
                            // I would need to debug what happens if this is the case
                        }
                    },
                    label = stringResource(R.string.main_gesture),
                    hintText = stringResource(R.string.main_gestures_hint),
                    accentColor = color,
                    selectedIndex = primaryGestureAction.value.id,
                    onSelectionChanged = {
                        val newAction = PrimaryGestureActionTracking.fromId(it)
                        primaryGestureAction.value = newAction
                        Thread { manager.changeSettings(context, primaryGestureAction = newAction) }.start()
                    },
                    buttons = listOf(
                        squareButton(R.drawable.pause),
                        squareButton(R.drawable.outline_arrows_outward_24),
                    ),
                )
            }

            item(key = "dismiss-gesture-toggle") {
                ButtonToggle(
                    modifier = Modifier.oneHandedGesture(
                        gestureConfiguration = dismissGestureConfig,
                        onGestureLabel = "change dismiss gesture",
                        onGestureAvailable = {
                            scope.launch {
                                try { dismissGestureIndicationState.showIndicator() } catch (_: Exception) {}
                            }
                        },
                        onGesture = {
                            val nextId = (dismissGestureAction.value.id + 1) % 3
                            val newAction = DismissGestureActionTracking.fromId(nextId)
                            dismissGestureAction.value = newAction
                            Thread { manager.changeSettings(context, dismissGestureAction = newAction) }.start()
                        }
                    ),
                    icon = {
                        OneHandedGestureClickIndicator(
                            modifier = Modifier.size(22.dp),
                            gestureConfiguration = dismissGestureConfig,
                            state = dismissGestureIndicationState,
                            gestureIndicatorTint = if (dismissGestureAction.value != DismissGestureActionTracking.DISABLED) color else textDarker
                        ){
                            // No fallback icon is not the best way to indicate that gestures are not supported.
                            // I would need to debug what happens if this is the case
                        }
                    },
                    label = stringResource(R.string.main_gesture_dismiss),
                    hintText = stringResource(R.string.main_gestures_dismiss_hint),
                    accentColor = color,
                    selectedIndex = dismissGestureAction.value.id,
                    onSelectionChanged = {
                        val newAction = DismissGestureActionTracking.fromId(it)
                        dismissGestureAction.value = newAction
                        Thread { manager.changeSettings(context, dismissGestureAction = newAction) }.start()
                    },
                    buttons = listOf(
                        squareButton(R.drawable.pause),
                        squareButton(R.drawable.stop),
                    ),
                )
            }

            item(key = "tilt-to-wake") {
                SettingsToggle(
                    text = stringResource(R.string.main_tiltToWake),
                    checked = useTiltToWake.value,
                    color = color,
                ) {
                    Thread { manager.changeSettings(context, tiltToWake = it) }.start()
                    useTiltToWake.value = it
                }
            }
        }
    }
}

@Composable
fun SectionHeader(id: Int, color: Color) {
    Text(
        text = stringResource(id),
        color = color,
        fontSize = 16.sp,
        fontWeight = FontWeight.Bold,
        modifier = Modifier.padding(top = 8.dp)
    )
}

fun squareButton(
    icon: Int,
): @Composable (selected: Boolean, onClick: () -> Unit, accentColor: Color) -> Unit = { selected, onClick, accentColor ->
    Box(
        modifier = Modifier
            .size(36.dp)
            .clip(RoundedCornerShape(8.dp))
            .background(
                if (selected) backgroundLighter else Color.Transparent
            )
            .border(
                width = if (selected) 1.5.dp else 1.dp,
                color = if (selected) accentColor else Color(0xFF676A70),
                shape = RoundedCornerShape(8.dp)
            )
            .clickable { onClick() },
        contentAlignment = Alignment.Center
    ) {
        Icon(
            painter = painterResource(icon),
            contentDescription = null,
            modifier = Modifier.size(20.dp),
            tint = if (selected) accentColor else textDarker
        )
    }
}

/** Display a toggle box with rounded corners, an icon, label on the left,
 *  and multiple icon buttons on the right.
 *  Shows an AlertDialog with [hintText] on long press. */
@Composable
fun ButtonToggle(
    modifier: Modifier = Modifier,
    iconRes: Int? = null,
    icon: (@Composable () -> Unit)? = null,
    label: String,
    hintText: String,
    accentColor: Color,
    selectedIndex: Int,
    onSelectionChanged: (index: Int) -> Unit,
    buttons: List<@Composable (selected: Boolean, onClick: () -> Unit, accentColor: Color) -> Unit>,
) {
    var showHintDialog by remember { mutableStateOf(false) }

    if (showHintDialog) {
        AlertDialog(
            visible = showHintDialog,
            onDismissRequest = { showHintDialog = false },
            title = { Text(label) },
            text = { Text(hintText) },
            edgeButton = {
                AlertDialogDefaults.EdgeButton(onClick = { showHintDialog = false })
            }
        )
    }

    Box(
        modifier = modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(20.dp))
            .background(lerp(backgroundLighter, accentColor, 0.15f))
            .combinedClickable(
                onClick = {},
                onLongClick = {
                    showHintDialog = true
                }
            )
            .padding(horizontal = 10.dp, vertical = 5.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth(),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.SpaceBetween
        ) {
            Row(
                modifier = Modifier.weight(1f),
                verticalAlignment = Alignment.CenterVertically
            ) {
                if(icon != null) {
                    icon()
                } else if (iconRes != null) {
                    Icon(
                        painter = painterResource(iconRes),
                        contentDescription = label,
                        modifier = Modifier.size(18.dp),
                        tint = text
                    )
                }

                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = label,
                    fontSize = 13.sp,
                    color = text,
                    maxLines = 1,
                    overflow = TextOverflow.Ellipsis
                )
            }

            Row(
                horizontalArrangement = Arrangement.spacedBy(4.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                buttons.forEachIndexed { idx, it ->
                    it(
                        idx+1 == selectedIndex,
                        {
                            if(selectedIndex == idx + 1) {
                                // No button selected
                                onSelectionChanged(0)
                            } else {
                                onSelectionChanged(idx+1)
                            }
                        },
                        accentColor,
                    )
                }
            }
        }
    }
}

/** Display a toggle with the provided text and value the user can check- and uncheck */
@Composable
fun SettingsToggle(
    text: String, checked: Boolean, isVisible: Boolean = true, color: Color,
    onCheckChanged: (isChecked: Boolean) -> Unit
) {
    if (!isVisible) {
        return
    }

    SwitchButton(
        label = {
            Text(text, maxLines = 1, overflow = TextOverflow.Ellipsis, fontSize = 14.sp)
        },
        checked = checked,
        colors = SwitchButtonDefaults.switchButtonColors(
            uncheckedContainerColor = lerp(backgroundLighter, color, 0.15f),
            checkedContainerColor = lerp(backgroundSelection, color, 0.4f),
            checkedThumbColor = color.copy(alpha = 0.85f),
            checkedTrackColor = color.copy(alpha = 0.35f),

            uncheckedThumbColor = Color.LightGray,
            uncheckedTrackColor = Color.DarkGray
        ),
        onCheckedChange = { onCheckChanged(it) },
        enabled = true,
        modifier = Modifier.fillMaxWidth(),
    )
}

@WearPreview
@Composable
fun SettingsPagePreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview()
    val gpsMode = remember { mutableStateOf(GPSMode.USE_PHONE) }

    RPoutTheme {
        SettingsPage(manager, gpsMode)
    }
}

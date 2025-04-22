package de.rpjosh.rpout.android.activities.main

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import android.content.IntentFilter
import android.os.BatteryManager
import android.os.Bundle
import android.os.PowerManager
import android.os.VibrationEffect
import android.os.Vibrator
import android.os.VibratorManager
import android.window.OnBackInvokedDispatcher
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.annotation.DrawableRes
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.wrapContentSize
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.TextUnitType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
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
import androidx.wear.compose.material.TimeTextDefaults
import androidx.wear.compose.material.curvedText
import androidx.wear.tooling.preview.devices.WearDevices
import com.google.android.horologist.annotations.ExperimentalHorologistApi
import com.google.android.horologist.health.composables.ActiveDurationText
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.FontSourceSanseProSemibold
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.overlayAmbient
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textDarker
import de.rpjosh.rpout.android.activities.theme.textHint
import de.rpjosh.rpout.android.shared.models.HeartRateZone
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.workout.State
import de.rpjosh.rpout.android.shared.workout.WorkoutManager
import de.rpjosh.rpout.android.workout.WorkoutTrackService
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.DelicateCoroutinesApi
import kotlinx.coroutines.ExperimentalCoroutinesApi
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.newSingleThreadContext
import java.time.Duration
import java.util.Locale
import java.util.concurrent.atomic.AtomicInteger
import java.util.concurrent.atomic.AtomicLong

data class Notification(
    val id: Long,
    val read: Boolean,
    @DrawableRes val icon: Int,
    val category: String = "",
    val lastNotified: Int = 0,
    val notifiedCount: AtomicInteger = AtomicInteger(0),
)

class WorkoutTrackingActivity: ComponentActivity(), AmbientLifecycleObserver.AmbientLifecycleCallback {

    companion object {
        /** How often the watch should vibrate when an unread notification is pending */
        const val NOTIFICATION_REMEMBER_INTERVAL = 4 * 60

        const val BROADCAST_NOTIFICATION = "de.rpjosh.rpout.android.workout.NOTIFICATION"
        const val BROADCAST_NOTIFICATION_ID = "notification_id"
        const val BROADCAST_ACTION = "notification_action"
        const val BROADCAST_CATEGORY = "notification_category"
    }

    private val ambientObserver = AmbientLifecycleObserver(this, this)
    private val isAmbient = mutableStateOf(false)
    private lateinit var manager: WorkoutManager
    private lateinit var logger: Logger

    @OptIn(ExperimentalCoroutinesApi::class, DelicateCoroutinesApi::class)
    private val scope = CoroutineScope(newSingleThreadContext("pauseResumeWorkout"))

    // Optional wake lock
    private lateinit var powerManager: PowerManager
    @Volatile private var wakeLock: PowerManager.WakeLock? = null

    // Tilt to wake sensor
    private lateinit var tiltSensor: TiltToWake

    private val lastNotification = mutableStateOf<Notification?>(null)

    private lateinit var batteryManager: BatteryManager
    private lateinit var vibrator: VibratorManager
    @Volatile private var lastVibratedForNotification = AtomicLong(0)

    private val notificationBroadcastReceiver = object : BroadcastReceiver() {
        override fun onReceive(context: Context, intent: Intent) {
            val id = intent.getLongExtra(BROADCAST_NOTIFICATION_ID, -1)
            val action = intent.getStringExtra(BROADCAST_ACTION)
            val category = intent.getStringExtra(BROADCAST_CATEGORY)

            if (::logger.isInitialized) logger.log("d", "Received notification with action = $action | id = $id | category = $category")

            // Battery icon mustn't be changed
            if (lastNotification.value?.category == "battery") {
                logger.log("d", "Notification cannot be applied because battery is low")
                return
            }

            if(action == "CREATE") {
                vibrateForNotification()
                lastNotification.value = Notification(id, false, R.drawable.rpdb, category = category ?: "")
            } else if (action == "DELETE" && lastNotification.value?.id == id) {
                lastNotification.value = null
            }
        }
    }

    override fun onCreate(savedInstanceState: Bundle?) {

        // Init dependencies
        batteryManager = getSystemService(Context.BATTERY_SERVICE) as BatteryManager
        vibrator = getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as VibratorManager

        // Register broadcast receiver
        registerReceiver(notificationBroadcastReceiver, IntentFilter(BROADCAST_NOTIFICATION), Context.RECEIVER_EXPORTED)

        // Check if workout manager is available
        val manager = WorkoutManager.workoutManager
        if (manager == null) {
            finish()
            return
        } else {
            this.manager = manager
            this.logger = Singleton.appController.injection.inject(Logger::class.java, arrayOf("WorkoutTrackingActivity"), false)
            this.powerManager = getSystemService(Context.POWER_SERVICE) as PowerManager
            tiltSensor = TiltToWake(baseContext, this.manager.type, { onTilted() }, Singleton.appController.injection.inject(Logger::class.java, arrayOf("TiltToWake"), false) )

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

        // Use wakelock (if required)
        setWakelock()

        setContent {
            RPoutTheme {
                WorkoutTrackingScreen(
                   isAmbient = isAmbient.value,
                    manager = manager,
                    lastNotification = lastNotification.value,
                    onStop = { onTrackingStop() },
                    onScreenLock = { onLockScreen() },
                    onPauseResume = { onPauseResume() },
                    onNotificationRead = {
                        if (lastNotification.value?.read == false) {
                            vibrator.defaultVibrator.vibrate(VibrationEffect.createOneShot(120, 255))
                            lastNotification.value = lastNotification.value?.copy(read = true)
                        }
                    }
                )
            }
        }
    }

    private fun onTrackingStop() {
        val intent = Intent(this,  WorkoutFinishedActivity::class.java).apply {
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK)
        }
        startActivity(intent)
        finish()
    }

    private fun onLockScreen() {
        // To disable: END_WET_MODE | Intent filter: com.google.android.clockwork.actions.WET_MODE_STARTED (_ENDED).
        // It's not publicly documented. Thanks google
        val intent = Intent("com.google.android.wearable.action.ENABLE_WET_MODE")
        intent.putExtra("relaunch_component_name", componentName.flattenToString())
        sendBroadcast(intent)
    }

    private fun onPauseResume() {
        scope.launch {
            if (manager.state.value == State.PAUSED) manager.resume()
            else manager.pause()

            // Delete any previous notifications. It's expected the user paused the activity to read it.
            // Additionally, it's a workaround if an external app doesn't delete notifications correctly
            lastNotification.value = null

            // Notify foreground service to update notification duration
            val serviceIntent = Intent(this@WorkoutTrackingActivity, WorkoutTrackService::class.java)
            serviceIntent.action = "NOTIFICATION"
            ContextCompat.startForegroundService(this@WorkoutTrackingActivity, serviceIntent)
        }
    }

    override fun onEnterAmbient(ambientDetails: AmbientLifecycleObserver.AmbientDetails) {
        isAmbient.value = true
        tiltSensor.register()
        super.onEnterAmbient(ambientDetails)
    }

    override fun onExitAmbient() {
        isAmbient.value = false
        tiltSensor.deRegister()
        super.onExitAmbient()
    }

    override fun onDestroy() {
        releaseWakelock()
        tiltSensor.deRegister()
        lifecycle.removeObserver(ambientObserver)
        unregisterReceiver(notificationBroadcastReceiver)

        super.onDestroy()
    }

    override fun onPause() {
        releaseWakelock()

        super.onPause()
    }

    override fun onResume() {
        setWakelock()

        super.onResume()
    }

    private fun setWakelock() {
        if (manager.type.liveUpdates && wakeLock == null) {
            logger.log("d", "Aquired a wake lock to keep UI updating")
            wakeLock = powerManager.newWakeLock(PowerManager.PARTIAL_WAKE_LOCK, "RPout:exercise")
            wakeLock?.acquire(Duration.ofHours(4).toMillis())
        }
    }

    private fun releaseWakelock() {
        if (wakeLock != null && wakeLock?.isHeld == true) {
            logger.log("d", "Release wake lock")
            wakeLock!!.release()
            wakeLock = null
        }
    }

    private fun onTilted() {
        // I really didn't found another possible way to wake up the screen without relaunching the activity
        val wakeLock = powerManager.newWakeLock(
            PowerManager.ACQUIRE_CAUSES_WAKEUP or PowerManager.SCREEN_BRIGHT_WAKE_LOCK,
            "RPout::WakeScreenUp"
        )
        wakeLock.acquire(2000L)

        // TurnScreenOn alone without wakelock doesn't work
        setTurnScreenOn(true)
    }

    override fun onUpdateAmbient() {
        super.onUpdateAmbient()

        // Check battery level
        if (batteryManager.getIntProperty(BatteryManager.BATTERY_PROPERTY_CAPACITY) <= 15 && lastNotification.value?.category != "battery") {
            lastNotification.value = Notification(-111, false, R.drawable.battery_low, "battery")
            vibrateForNotification()
        }

        // Vibrate again after threshold is over
        if (lastNotification.value?.read == false && (System.currentTimeMillis() / 1000) - lastVibratedForNotification.get() > NOTIFICATION_REMEMBER_INTERVAL) {
            vibrateForNotification()
        }
    }

    @Synchronized
    private fun vibrateForNotification() {
        lastVibratedForNotification.set(System.currentTimeMillis() / 1000)

        val pattern = longArrayOf(50, 220, 150, 280,  300, 220, 150, 280,  300, 400)
        vibrator.defaultVibrator.vibrate(VibrationEffect.createWaveform(pattern, -1))
    }

}

/** All (vertical) pages of this activity */
val trackingPages: List<@Composable (
    manager: WorkoutManager, lastNotification: Notification?,
    onStop: () -> Unit, onScreenLock: () -> Unit, onPauseResume: () -> Unit,
    onNotificationRead: (id: Long) -> Unit,
) -> Unit> = listOf(
    { manager, _, onStop, onScreenLock, onPauseResume, _ -> WorkoutTrackActionTab(manager, onStop, onScreenLock, onPauseResume)  },
    { manager, notification, _, _, _, onRead -> WorkoutTrackMainTab(manager, notification, onRead) },
    { _, _, _, _, _, _  -> WorkoutTrackExtraTab() },
)

@Composable
fun WorkoutTrackingScreen(
    isAmbient: Boolean, manager: WorkoutManager, lastNotification: Notification?,
    onStop: () -> Unit, onScreenLock: () -> Unit, onPauseResume: () -> Unit,
    onNotificationRead: (id: Long) -> Unit,
) {
    val scope = rememberCoroutineScope()

    /** Whether the GPS signal is already acquired */
    val gpsConnected = remember { derivedStateOf {manager.state.value == State.TRACKED || manager.state.value == State.PAUSED || manager.healthSupportedCapabilities?.gps == false } }
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

    // Go back to main page when locking screen with touch lock
    val enableTouchLock = {
        scope.launch {
            pagerState.animateScrollToPage(1)
            onScreenLock()
        }
    }

    Scaffold(
        // @BUG we need a 2.dp padding at the bottom to not cut off the circles (tested on Pixel Watch 2)
        pageIndicator = { if (!isAmbient) HorizontalPageIndicator(pageIndicatorState, modifier = Modifier.padding(bottom = 2.dp, start = 1.dp)) },
        timeText = {
            if (manager.state.value == State.PAUSED) {
                CurvedLayout {
                    curvedRow {
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
                    curvedRow {
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
                TimeText(
                    timeTextStyle = TimeTextDefaults.timeTextStyle(
                        color = if (isAmbient) textDarker else text
                    ),
                )
            }
        },
    ) {
        HorizontalPager(
            state = pagerState,
            modifier = Modifier.fillMaxSize()
        ) { index ->
            Box {
                trackingPages[index](
                    manager, lastNotification,
                    onStop, { enableTouchLock() }, onPauseResume,
                    onNotificationRead,
                )

                if (isAmbient) {
                    Box(modifier = Modifier.background(overlayAmbient).fillMaxSize()) {  }
                }
            }
        }
    }
}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutTrackingPreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true, totalKm = 22.56, heartRate = 122)
    val notification = remember { mutableStateOf<Notification?>(Notification(0, false, R.drawable.battery_low, category = "batter")) }

    RPoutTheme {
        WorkoutTrackingScreen(
            false, manager, notification.value, {}, {}, {},
            onNotificationRead = { notification.value = notification.value?.copy(read = !(notification.value?.read ?: false)) }
        )
    }
}

@Composable
fun WorkoutTrackActionTab(manager: WorkoutManager, onStop: () -> Unit, onScreenLock: () -> Unit, onPauseResume: () -> Unit) {
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
fun WorkoutTrackMainTab(manager: WorkoutManager, notification: Notification?, onNotificationRead: (id: Long) -> Unit) {

    val speedText =  remember { derivedStateOf { manager.workoutData.getFormattedSpeed(manager.type.id) } }

    // Right side heart rate indicator
    Box(modifier = Modifier.padding(end = 3.dp, bottom = 2.dp)) {
        CurvedLayout(modifier = Modifier.align(Alignment.Center).rotate(90f)) {
            curvedRow {
                for(i in 5 downTo 1 step 1) {
                    val col = HeartRateZone.zones[i].color
                    val currentZone = HeartRateZone.getZone(manager.workoutData.heartRate.value.value)

                    curvedComposable(modifier = CurvedModifier.padding(ArcPaddingValues(after = 3.dp)) ) {
                        Box(modifier = Modifier
                            .height(7.dp).width(15.dp)
                        ) {

                            // Highlight the currently active zone
                            if (i == currentZone.id) {
                                Box(modifier = Modifier.fillMaxSize().background(color = col, shape = RoundedCornerShape(1.dp)))
                            }

                            Box(
                                modifier = Modifier
                                    .height(6.dp).width(14.dp).align(Alignment.Center)
                            ) {
                                // Default background indicator
                                Box(
                                    modifier = Modifier.fillMaxSize()
                                        .background(
                                            color = HeartRateZone.zones[i].color,
                                            shape = RoundedCornerShape(1.dp)
                                        )
                                ) {}
                                // Overlay that is always visible to make the colors less bright
                                Box(
                                    modifier = Modifier.background(
                                        color = Color(0x30000000),
                                        shape = RoundedCornerShape(1.dp)
                                    ).fillMaxSize()
                                )

                                if (currentZone.id <= i) {
                                    // Overlay to indicate progress
                                    val next = if (i == 5) 190 else HeartRateZone.zones[i + 1].min
                                    var percentage = (next - manager.workoutData.heartRate.value.value) / (next.toDouble() - HeartRateZone.zones[i].min)
                                    if (currentZone.id != i) percentage = 1.0

                                    if (percentage > 0.2) {
                                        Box(
                                            modifier = Modifier.fillMaxSize()
                                        ) {
                                            Box(
                                                modifier = Modifier
                                                    .background(color = Color(0xA0000000))
                                                    .align(Alignment.CenterStart)
                                                    .width(14.dp * percentage.toFloat())
                                                    .fillMaxHeight()
                                            )
                                        }
                                    }

                                }
                            }
                        }
                    }
                }

            }

        }

        Column(
            modifier = Modifier
                .padding(top = 24.dp, bottom = 15.dp, start = 3.dp, end = 8.dp)
                .fillMaxSize(),
            verticalArrangement = Arrangement.SpaceBetween
        ) {
            ActiveDurationText(
                checkpoint = manager.workoutData.activeDuration.value,
                state = manager.workoutData.exerciseState.value,
                content = {
                    var durationFormatted = ""
                    if (it.toHours() > 0) durationFormatted += "${it.toHours()}:"
                    durationFormatted += String.format(
                        Locale.ENGLISH,
                        "%02d:%02d",
                        it.toMinutesPart(),
                        it.toSecondsPart()
                    )

                    Text(
                        text = durationFormatted,
                        fontSize = 35.sp,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.fillMaxWidth().padding(start = 6.dp),
                        fontFamily = FontFamily(FontSourceSanseProSemibold)
                    )
                }
            )

            Column(verticalArrangement = Arrangement.spacedBy((-8).dp)) {
                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceAround,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy((-6).dp)
                    ) {
                        Text(
                            text = String.format(
                                Locale.ENGLISH,
                                if(notification != null && manager.workoutData.distance.value.value > 10000) "%.1f" else "%.2f",
                                manager.workoutData.distance.value.value / 1000.0
                            ),
                            fontSize = 35.sp,
                            textAlign = TextAlign.Start,
                            fontFamily = FontFamily(FontSourceSanseProSemibold),
                        )
                        Text(
                            text = "km",
                            fontSize = 14.sp,
                            color = textHint
                        )
                    }

                    notification?.let { NotificationBox(it, onNotificationRead) }


                    Column(
                        horizontalAlignment = Alignment.CenterHorizontally,
                        verticalArrangement = Arrangement.spacedBy((-6).dp),
                        modifier = Modifier.padding(end = if(manager.workoutData.heartRate.value.value < 100) 5.dp else 0.dp)
                    ) {
                        Text(
                            text = manager.workoutData.heartRate.value.value.toString(),
                            color = manager.workoutData.heartRate.color.value,
                            fontSize = 35.sp,
                            textAlign = TextAlign.End,
                            fontFamily = FontFamily(FontSourceSanseProSemibold),
                            letterSpacing = TextUnit(0.2f, TextUnitType.Sp),
                        )

                        Text(
                            text = "bpm",
                            fontSize = 14.sp,
                            color = textHint
                        )
                    }
                }
            }

            Column(
                modifier = Modifier.align(Alignment.CenterHorizontally).padding(start = 6.dp),
                verticalArrangement = Arrangement.spacedBy((-6).dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = speedText.value,
                    fontSize = 35.sp,
                    textAlign = TextAlign.Center,
                    fontFamily = FontFamily(FontSourceSanseProSemibold)
                )

                Text(
                    text = if (speedText.value.contains(":")) "min/km" else "km/h",
                    fontSize = 14.sp,
                    color = textHint
                )
            }
        }
    }
}
@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun WorkoutTrackMainTabPreview() {
    // Initialize dummy workout manager for tests
    val manager = WorkoutManager.forPreview(true, heartRate = 122, totalKm = 12.56)

    RPoutTheme {
        WorkoutTrackMainTab(manager, null, {})
    }
}

@Composable
fun NotificationBox(notification: Notification, onNotificationClick: (id: Long) -> Unit) {
    val color = remember(notification) {
        var baseColor = Color(0xFF00FFB7)
        var readColor = Color(0x9900FFB7)

        // Handling of extra categories
        if (notification.category == "battery") {
            baseColor = Color(0xFFFF4D4D)
            readColor = Color(0x90FF4D4D)
        }

        if (notification.read) readColor else baseColor
    }

    Box(
        modifier = Modifier.size(22.dp).padding(top = 6.dp),
        contentAlignment = Alignment.Center
    ) {
        Canvas(modifier = Modifier.size(8.dp)) {
            if(!notification.read) {
                drawCircle(
                    color = color.copy(alpha = color.alpha - 0.2f),
                    radius = 16.dp.toPx(),
                    style = androidx.compose.ui.graphics.drawscope.Stroke(width = 2.5.dp.toPx())
                )
            }
        }

        Box(
            contentAlignment = Alignment.Center,
            modifier = Modifier
                .wrapContentSize(unbounded = true)
                .size(if(notification.read) 27.dp else 20.dp)
                .clickable {
                    onNotificationClick(notification.id)
                }
        ) {
            Icon(
                modifier = Modifier.padding(2.dp),
                painter = painterResource(notification.icon),
                contentDescription = "Notification icon",
                tint = color
            )
        }
    }
}

@Composable
fun WorkoutTrackExtraTab() {
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

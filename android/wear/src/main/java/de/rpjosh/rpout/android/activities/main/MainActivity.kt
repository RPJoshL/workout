package de.rpjosh.rpout.android.activities.main

import android.Manifest
import android.annotation.SuppressLint
import android.content.Intent
import android.content.pm.PackageManager
import android.graphics.Bitmap
import android.graphics.Canvas
import android.os.Bundle
import android.os.VibrationEffect
import android.os.Vibrator
import android.provider.Settings
import android.util.Log
import android.widget.Toast
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.painter.BitmapPainter
import androidx.compose.ui.input.rotary.onPreRotaryScrollEvent
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.compose.foundation.lazy.ScalingLazyColumn
import androidx.wear.compose.foundation.lazy.ScalingLazyColumnDefaults
import androidx.wear.compose.foundation.lazy.ScalingLazyListAnchorType
import androidx.wear.compose.foundation.lazy.ScalingLazyListState
import androidx.wear.compose.foundation.lazy.items
import androidx.wear.compose.foundation.lazy.itemsIndexed
import androidx.wear.compose.material.Button
import androidx.wear.compose.material.ButtonDefaults
import androidx.wear.compose.material.Chip
import androidx.wear.compose.material.ChipDefaults
import androidx.wear.compose.material.Icon
import androidx.wear.compose.material.PositionIndicator
import androidx.wear.compose.material.Scaffold
import androidx.wear.compose.material.Text
import androidx.wear.tooling.preview.devices.WearDevices
import com.caverock.androidsvg.SVG
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.WearMessageReceiver
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.accentBlueBorder
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.defaultBackground
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textBlue
import de.rpjosh.rpout.android.helper.PermissionHelper
import de.rpjosh.rpout.android.helper.VersionHelper
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.controller.WorkoutController
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.MessageType
import de.rpjosh.rpout.android.shared.services.Tr
import de.rpjosh.rpout.android.tiles.PaiTile
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch
import kotlinx.coroutines.newSingleThreadContext
import kotlin.math.ceil


class MainActivity : ComponentActivity(), WearMessageReceiver {

    companion object {
        /** Key of the workout type to indicate that a sync should be done */
        const val TYPE_ID_SYNC = -6L
    }

    private lateinit var logger: Logger
    private lateinit var globalConfig: GlobalConfiguration
    private lateinit var permissionHelper: PermissionHelper
    private lateinit var workoutController: WorkoutController
    private lateinit var metricController: MetricController

    private val activityTypes = mutableStateListOf<WorkoutType>()
    private val lastActivityTypes = mutableStateListOf<Long>()

    // Androids permission contract helper to ask for permissions easily
    private val requestPermissionLauncher =
        registerForActivityResult(ActivityResultContracts.RequestPermission()) { isGranted: Boolean ->
            if (isGranted) {
                Toast.makeText(this, "Permission granted", Toast.LENGTH_SHORT).show()
                // Ask for other permissions / start service
                checkAndRequestPermission()
            } else {
                Toast.makeText(this, "Permission not granted! The app will not work as expected", Toast.LENGTH_SHORT).show()
                Singleton.appController.sharedLogger.log("w", "User did not grant all rights. The app won't work correctly")
            }
        }

    override fun onCreate(savedInstanceState: Bundle?) {
        installSplashScreen()
        super.onCreate(savedInstanceState)

        setTheme(android.R.style.Theme_DeviceDefault)

        // Initialize app
        initApp()
        Singleton.registerOnWearMessageReceived(this)
        permissionHelper = PermissionHelper(this)

        setContent {
            RPoutTheme {
                ActivityList(activityTypes, lastActivityTypes) { onActivityClicked(it) }
            }
        }

        // Ask for permission
        checkAndRequestPermission()
    }

    /**
     * Checks all required permissions for this apps and ask for it
     * if they were not granted already
     */
    private fun checkAndRequestPermission() {
        val permissions = arrayListOf(
            Manifest.permission.ACTIVITY_RECOGNITION,
            Manifest.permission.BODY_SENSORS,
            Manifest.permission.POST_NOTIFICATIONS,
            Manifest.permission.ACCESS_FINE_LOCATION,
            "com.google.android.clockwork.settings.WATCH_TOUCH"
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

        // Special permission we have to ask. The log will still be false
        if (!permissionHelper.canDrawOverlays()) permissionHelper.askForDrawOverlay()
        if (!permissionHelper.canScheduleExact()) permissionHelper.askForScheduleExact()
        // if (!permissionHelper.isBatterOptimizationIgnored()) permissionHelper.askForBatteryOptimization()
        permissionHelper.askForDisableUnusedAppRestrictions()

        Singleton.appController.sharedLogger.log("d", "All permissions are granted")

        // Start login activity if we don't have a user context
        if (globalConfig.user == null) {
            startActivity(Intent(this, NotLoggedInActivity::class.java))
            finish()
        }
    }

    /**
     * Initializes the app controller with all dependencies. This method
     * does block until the APP is fully loaded
     */
    private fun initApp() {
        if (Singleton.getApp() == null) Singleton.app()

        // Inject dependencies
        globalConfig = Singleton.appController.injection.inject(GlobalConfiguration::class.java, null,  false)
        logger = Singleton.appController.injection.inject(Logger::class.java, arrayOf("MainActivity"), false)
        workoutController = Singleton.appController.injection.inject(WorkoutController::class.java, null, false)
        metricController = Singleton.appController.injection.inject(MetricController::class.java, null, false)
    }

    override fun onPause() {
        Singleton.getApp()?.activityPaused(this)
        Singleton.deRegisterOnWearMessageReceived(this)
        super.onPause()
    }
    override fun onDestroy() {
        Singleton.appController.activityDestroyed(this)
        Singleton.deRegisterOnWearMessageReceived(this)
        super.onDestroy()
    }

    override fun onStart() {
        Singleton.appController.activityStarted(applicationContext, this)
        super.onStart()
    }
    override fun onResume() {
        Singleton.appController.activityStarted(applicationContext, this)
        Singleton.registerOnWearMessageReceived(this)
        super.onResume()

        // Get workout types if no one are loaded already
        Thread { setWorkoutTypes() }.start()

        // Get last activity types (again)
        Thread { setLastActivityTypes() }.start()
    }

    @Synchronized
    fun setWorkoutTypes() {
        if (activityTypes.isNotEmpty()) return

        // Get the current version name of the app
        activityTypes.addAll(workoutController.getWorkoutTypes(VersionHelper.getVersionName()))
        // Add the dummy sync icon
        activityTypes.add(
            WorkoutType(
                id = TYPE_ID_SYNC, tagDark = "#FFFFFF",
                icon = "<svg width=\"800px\" height=\"800px\" viewBox=\"0 0 24 24\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\"> <path d=\"M3 11.9998C3 7.02925 7.02944 2.99982 12 2.99982C14.8273 2.99982 17.35 4.30348 19 6.34248\" stroke=\"currentColor\" stroke-width=\"2.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"/> <path d=\"M19.5 2.99982L19.5 6.99982L15.5 6.99982\" stroke=\"currentColor\" stroke-width=\"2.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"/> <path d=\"M21 11.9998C21 16.9704 16.9706 20.9998 12 20.9998C9.17273 20.9998 6.64996 19.6962 5 17.6572\" stroke=\"currentColor\" stroke-width=\"2.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"/> <path d=\"M4.5 20.9998L4.5 16.9998L8.5 16.9998\" stroke=\"currentColor\" stroke-width=\"2.5\" stroke-linecap=\"round\" stroke-linejoin=\"round\"/> </svg>",
            )
        )
    }

    @Synchronized
    fun setLastActivityTypes() {
        val res = workoutController.dao().getLastWorkoutTypes()
        lastActivityTypes.clear()

        if (workoutController.dao().getUnsyncedWorkouts().isEmpty()) {
            lastActivityTypes.addAll(res)
        } else if (res.isNotEmpty()) {
            // Add a dummy "sync" SVG icon
            lastActivityTypes.add(TYPE_ID_SYNC)
            lastActivityTypes.addAll(res.subList(0, if (res.size >= 6) 5 else res.size))
        } else {
            lastActivityTypes.add(TYPE_ID_SYNC)
        }

    }

    override fun onWearMessageReceived(type: MessageType, data: String) {
        if (type == MessageType.SYNC_DATA_WORKOUT) {
            logger.log("d", "Updating workout types in MainActivity")
            activityTypes.clear()
            Thread{ setWorkoutTypes() }.start()
        }
    }

    private fun onActivityClicked(id: Long) {
        if (id == TYPE_ID_SYNC) {
            Thread {
                val success = workoutController.synchronizeWorkouts()

                // Vibrate device
                val vibrator = baseContext.getSystemService(Vibrator::class.java)
                val pattern = longArrayOf(90,  55, 75, 40, 90, 55, 75, 40, 90,  55, 75, 40, 90, 55, 75, 40, 90, 55, 75, 50)
                val amplitude = intArrayOf(255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0, 255, 0)
                val vibrationEffect = VibrationEffect.createWaveform(pattern, amplitude,-1)
                vibrator.vibrate(vibrationEffect)

                if (success) {
                    // Remove sync button from last types
                    setLastActivityTypes()

                    // Synchronize PAI
                    if (metricController.synchronizePai()) androidx.wear.tiles.TileService.getUpdater(this).requestUpdate(PaiTile::class.java)
                }
            }.start()
        } else {
            // Start start screen
            val intent = Intent(this,  WorkoutStartActivity::class.java).apply {
                putExtra(WorkoutStartActivity.KEY_TYPE_ID, id)
            }
            startActivity(intent)
        }
    }

}


/** Note: this UI does lag extremely in debug mode! */
@Composable
fun ActivityList(activityTypes: List<WorkoutType>, lastActivityTypes: List<Long>, onClick: (id: Long) -> Unit) {
    val listState = remember { ScalingLazyListState() }
    val coroutineScope = rememberCoroutineScope()

    // Sort activity types by name
    val sortedActivityTypes = remember { derivedStateOf { activityTypes.filter { it.id != MainActivity.TYPE_ID_SYNC }.sortedBy { it.getName(Tr.getUsedLanguage()) } } }
    val resolvedLastActivityTypes =  remember { derivedStateOf {
        lastActivityTypes.mapNotNull { id ->
            // Find activity type with provided ID
            activityTypes.find { id == it.id }
        }
    }}

    Scaffold(
        positionIndicator = { PositionIndicator(scalingLazyListState = listState) },
        modifier = Modifier.onPreRotaryScrollEvent { event ->
            // That doesn't look smooth
            var blockEvent = false
            // if (listState.centerItemIndex <= 1) {
            //    val viewportHeight = listState.layoutInfo.viewportSize.height.toFloat()
            //    coroutineScope.launch {
            //        listState.animateScrollToItem(ceil(resolvedLastActivityTypes.size / 3.0).toInt() + 2, )
            //    }
            //    blockEvent = true
            // }

            blockEvent
        }
    ) {
        ScalingLazyColumn(
            modifier = Modifier.fillMaxWidth(),
            state = listState,
            flingBehavior = ScalingLazyColumnDefaults.snapFlingBehavior(state = listState),
            contentPadding = PaddingValues(
                top = 8.dp,
                start = 12.dp,
                end = 12.dp,
                bottom = 35.dp
            ),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(6.dp),
            anchorType = ScalingLazyListAnchorType.ItemCenter,
            // Do not center the first elements index => use contentPadding or AutoCenteringParams(itemIndex = 3)
            autoCentering = null
        ) {
            item(key = "static-last") {
                Column {
                    Box(
                        modifier = Modifier
                            .background(defaultBackground, RoundedCornerShape(16.dp))
                            .border(1.dp, accentBlueBorder, shape = RoundedCornerShape(16.dp))
                            .padding(7.dp)
                    ) {
                        Text(
                            text = stringResource(R.string.main_lastTypes),
                            textAlign = TextAlign.Center,
                            modifier = Modifier.padding(start = 8.dp, end = 8.dp),
                            color = textBlue,
                            fontWeight = FontWeight.Bold
                        )
                    }
                    Spacer(modifier = Modifier.height(6.dp))
                }
            }

            items(ceil(resolvedLastActivityTypes.value.size / 3.0).toInt()) { index ->
                Row(
                    horizontalArrangement = Arrangement.Center,
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(top = 0.dp, bottom = 5.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    for (i in index * 3 until resolvedLastActivityTypes.value.size step 1) {
                        if (i > (index * 3) + 2) break

                        val item = resolvedLastActivityTypes.value[i]
                        Button(
                            onClick = { onClick(item.id) },
                            colors = ButtonDefaults.primaryButtonColors(
                                backgroundColor = backgroundLighter
                            ),
                            modifier = Modifier
                                .padding(
                                    start = if (i == index * 3) 0.dp else 3.dp,
                                    end = if (i == (index * 3) + 2) 0.dp else 3.dp
                                )
                                .width(49.dp)
                                .height(49.dp)
                        ) {
                            SvgIcon(
                                svgString = item.icon,
                                size = 29.dp,
                                hexTint = item.tagDark,
                            )
                        }
                    }
                }
            }

            item(key = "static-all") {
                Column {
                    // To show on extra screen: Default: 6, Pixel Watch 1 & 2: 24
                    Spacer(modifier = Modifier.height(if (ceil(resolvedLastActivityTypes.value.size / 3.0).toInt() >= 2) 24.dp else 6.dp))
                    Box(
                        modifier = Modifier
                            .background(defaultBackground, RoundedCornerShape(16.dp))
                            .border(1.dp, accentBlueBorder, shape = RoundedCornerShape(16.dp))
                            .padding(7.dp)
                    ) {
                        Text(
                            text = stringResource(R.string.main_allTypes),
                            textAlign = TextAlign.Center,
                            modifier = Modifier.padding(start = 8.dp, end = 8.dp),
                            color = textBlue,
                            fontWeight = FontWeight.Bold
                        )
                    }
                    Spacer(modifier = Modifier.height(6.dp))
                }
            }

            itemsIndexed(sortedActivityTypes.value, key = { _, item -> "overview-${item.id}" }) { _, item ->
                Chip(
                    modifier = Modifier
                        .fillMaxWidth(),
                    colors = ChipDefaults.primaryChipColors(
                        backgroundColor = backgroundLighter
                    ),
                    icon = {
                        SvgIcon(
                            svgString = item.icon,
                            size = 28.dp,
                            hexTint = item.tagDark,
                        )
                    },
                    label = {
                        Text(
                            modifier = Modifier.fillMaxWidth(),
                            color = text,
                            text = item.getName(Tr.getUsedLanguage())
                        )
                    },
                    onClick = { onClick(item.id) }
                )
            }
        }
    }

}

@Composable
@SuppressLint("ModifierParameter")
fun SvgIcon(
    svgString: String, size: Dp,
    modifier: Modifier = Modifier.wrapContentSize(align = Alignment.Center),
    tint: Color = text, hexTint: String? = null
) {

    // Apply tint from hex color
    var iTint = tint
    if (hexTint != null) {
        iTint = Color(android.graphics.Color.parseColor(hexTint))
    }

    // Initialize SVG
    val svg = SVG.getFromString(svgString)
    svg.documentWidth = with(LocalDensity.current) { size.toPx() }
    svg.documentHeight = svg.documentWidth

    // Convert it into a drawable
    val drawable = svg.renderToPicture()
    val bitmap = Bitmap.createBitmap(svg.documentWidth.toInt(), svg.documentHeight.toInt(), Bitmap.Config.ARGB_8888)
    val canvas = Canvas(bitmap)
    drawable.draw(canvas)

    Icon(
        painter = BitmapPainter(bitmap.asImageBitmap()),
        tint = iTint,
        contentDescription = "Star",
        modifier = modifier
            .size(size),
    )
}

val sampleActivityTypes = mutableListOf(
    WorkoutType(
        id = 0, nameEn = "Hiking", nameDe = "Gehen", tagDark = "#ffffff", tagWhite = "",
        icon = "<svg class=\"icon\" viewBox=\"0 0 16 21\" fill=\"none\" xmlns=\"http://www.w3.org/2000/svg\"> <path transform=\"translate(-4,-2)\" fill-rule=\"evenodd\" clip-rule=\"evenodd\" d=\"M13 6C14.1046 6 15 5.10457 15 4C15 2.89543 14.1046 2 13 2C11.8955 2 11 2.89543 11 4C11 5.10457 11.8955 6 13 6ZM11.0528 6.60557C11.3841 6.43992 11.7799 6.47097 12.0813 6.68627L13.0813 7.40056C13.3994 7.6278 13.5559 8.01959 13.482 8.40348L12.4332 13.847L16.8321 20.4453C17.1384 20.9048 17.0143 21.5257 16.5547 21.8321C16.0952 22.1384 15.4743 22.0142 15.168 21.5547L10.5416 14.6152L9.72611 13.3919C9.58336 13.1778 9.52866 12.9169 9.57338 12.6634L10.1699 9.28309L8.38464 10.1757L7.81282 13.0334C7.70445 13.575 7.17759 13.9261 6.63604 13.8178C6.09449 13.7094 5.74333 13.1825 5.85169 12.641L6.51947 9.30379C6.58001 9.00123 6.77684 8.74356 7.05282 8.60557L11.0528 6.60557ZM16.6838 12.9487L13.8093 11.9905L14.1909 10.0096L17.3163 11.0513C17.8402 11.226 18.1234 11.7923 17.9487 12.3162C17.7741 12.8402 17.2078 13.1234 16.6838 12.9487ZM6.12844 20.5097L9.39637 14.7001L9.70958 15.1699L10.641 16.5669L7.87159 21.4903C7.60083 21.9716 6.99111 22.1423 6.50976 21.8716C6.0284 21.6008 5.85768 20.9911 6.12844 20.5097Z\" fill=\"currentColor\"/> </svg>",
    ),
    WorkoutType(
        id = 1, nameEn = "Running", nameDe = "Laufen", tagDark = "#ffffff", tagWhite = "",
        icon = "<svg fill=\"currentColor\" class=\"icon\" version=\"1.2\" baseProfile=\"tiny\" xmlns=\"http://www.w3.org/2000/svg\" xmlns:xlink=\"http://www.w3.org/1999/xlink\" viewBox=\"-191 65 256 256\"> <path d=\"M0.4,95.8c-5.1,10.8-17.9,15.5-28.8,10.5s-15.5-17.9-10.5-28.8S-20.9,62-10,67S5.5,85,0.4,95.8z M47.1,167.2 c0,5.1-4.3,9.5-9.5,9.5H-6.8c-4.1,0-7.6-2.7-8.9-6.2l-8.1-21.9l-27.1,57.9l35.2,96.6c2.4,7-1.1,14.9-8.1,17.3 c-7,2.4-14.9-1.1-17.3-8.1l-33.8-92.8l-17.3,36.8c-2.2,4.6-6.8,7.8-12.2,7.8h-54.1c-7.6,0-13.5-6-13.5-13.5s6-13.5,13.5-13.5h45.4 l49.2-105.5l-21.1,7.6l-17.3,36.8c-2.4,4.9-7.8,6.8-12.7,4.6c-4.9-2.4-6.8-7.8-4.6-12.7l18.9-40.3c1.1-2.4,3.2-4.3,5.7-5.1l36-13 c7.8-3.2,17.3-3.5,25.7,0.5l3.8,1.9c9.2,3.2,16,10.6,19.2,18.9l9.7,26.8h37.9C42.5,157.5,46.8,161.8,47.1,167.2z\"/> </svg>"
    )
)

@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun DefaultPreview() {
    RPoutTheme {
        Box(modifier = Modifier
            .background(defaultBackground)
            .fillMaxSize()) {
            ActivityList(
                activityTypes = sampleActivityTypes,
                lastActivityTypes = listOf(0, 1),
                onClick = {}
            )
        }
    }
}
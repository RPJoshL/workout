package de.rpjosh.rpout.android.activities.services

import android.media.MediaPlayer
import android.os.Bundle
import android.os.Handler
import android.os.Looper
import android.os.VibrationEffect
import android.os.Vibrator
import android.view.WindowManager
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableDoubleStateOf
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
import androidx.wear.compose.material.MaterialTheme
import androidx.wear.compose.material.Text
import androidx.wear.tooling.preview.devices.WearDevices
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.shared.controller.MetricController
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext

class NotActiveActivity: ComponentActivity() {

    private var progress = mutableDoubleStateOf(0.0)
    private var steps = mutableIntStateOf(0)

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        installSplashScreen()

        // Get data. We need to do this on the main thread to check if we have to show the activity
        // at all (the timeframe between starting this activity from the foreground service and
        // really showing it can be huge!)
        runBlocking {
            withContext(Dispatchers.IO) {
                setData()
            }
        }
        if (steps.intValue >= 150) {
            Singleton.getApp()?.sharedLogger?.log("d", "Tried to start activity check screen but threshold was already reached")
            finish(); return
        }
        // Thread{ setData() }.start()

        setTheme(android.R.style.Theme_DeviceDefault)

        // Do not turn display off
        window.addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON)
        // Turn on display automatically
        setShowWhenLocked(true)
        setTurnScreenOn(true)

        // Automatically close after 30 seconds
        Handler(Looper.getMainLooper()).postDelayed({
            finish()
        }, 30 * 1000)

        // Play sound
        val mediaPlayer = MediaPlayer.create(baseContext, R.raw.do_activity)
        mediaPlayer?.start()
        mediaPlayer?.setOnCompletionListener {
            mediaPlayer.release()
        }

        // Vibrate
        val vibrator = getSystemService(Vibrator::class.java)
        val pattern = longArrayOf(0,  150, 20, 100, 150)
        val amplitude = intArrayOf(0, 170, 0,  210, 255)
        val vibrationEffect = VibrationEffect.createWaveform(pattern, amplitude,-1)
        vibrator.vibrate(vibrationEffect)

        setContent {
            NotActiveScreen(
                { finish() },
                steps = steps.intValue,
                progress = progress.doubleValue
            )
        }
    }

    /**
     * Sets and fetches required data for this activity
     */
    private fun setData() {
        val metricController = Singleton.appController.injection.inject(MetricController::class.java, null, false)
        // Activity is triggered when a step count of 150 was not reached within the last 60 minutes => user 58 to not confuse user
        val stepsHour = metricController.dao().getStepsSince(58 * 60)
        steps.intValue = stepsHour
        progress.doubleValue = stepsHour / 150.0
    }
}


@Composable
fun NotActiveScreen(onTab: () -> Unit, progress: Double, steps: Int) {

    RPoutTheme {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .background(MaterialTheme.colors.background)
                .padding(start = 5.dp, end = 5.dp)
                .pointerInput(Unit) {
                    detectTapGestures(
                        onTap = {
                            onTab()
                        }
                    )
                },
            verticalArrangement = Arrangement.spacedBy(8.dp),
        ) {
            Box(
                contentAlignment = Alignment.Center,
                modifier = Modifier
                    .size(80.dp)
                    .padding(10.dp)
                    .align(Alignment.CenterHorizontally)
            ) {
                // Progress bar around icon
                Canvas(modifier = Modifier.size(89.dp)) {
                    drawArc(
                        color = Color(0xFF4A4052),
                        startAngle = -90f,
                        sweepAngle = 360f,
                        useCenter = false,
                        style = Stroke(width = 8.dp.toPx())
                    )
                }
                Canvas(modifier = Modifier.size(90.dp)) {
                    val sweepAngle = 360f * progress
                    drawArc(
                        color = Color(0xFF11AD22),
                        startAngle = -90f,
                        sweepAngle = sweepAngle.toFloat(),
                        useCenter = false,
                        style = Stroke(width = 8.dp.toPx())
                    )
                }

                // Background of image
                Box(modifier = Modifier
                    .fillMaxSize()
                    .background(color = Color(0xFF021C1A), shape = CircleShape)
                ) {  }

                // Image itself
                Image(
                    painter = painterResource(R.drawable.walking),
                    contentDescription = "Logo",
                    modifier = Modifier
                        .width(65.dp)
                        .height(65.dp)
                        .padding(13.dp),
                    contentScale = ContentScale.FillWidth
                )
            }

            Text(
                text = stringResource(R.string.notActive_info),
                textAlign = TextAlign.Center,
                fontSize = 16.sp
            )

            Text(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 8.dp),
                text = "$steps / 150 ${stringResource(R.string.notActive_steps)}",
                textAlign = TextAlign.Center,
                color = Color(0xFFB8BEB9),
                fontSize = 15.sp
            )
        }
    }
}

@Preview(device = WearDevices.SMALL_ROUND, showSystemUi = true)
@Composable
fun NotActiveScreenPreview() {
    NotActiveScreen({}, steps = 125, progress = 0.75)
}
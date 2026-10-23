package de.rpjosh.rpout.android.activities.components

import androidx.compose.animation.Animatable
import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.layout.size
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableFloatStateOf
import androidx.compose.runtime.remember
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.wear.compose.foundation.ArcPaddingValues
import androidx.wear.compose.foundation.CurvedLayout
import androidx.wear.compose.foundation.CurvedModifier
import androidx.wear.compose.foundation.curvedComposable
import androidx.wear.compose.foundation.curvedRow
import androidx.wear.compose.foundation.padding
import androidx.wear.compose.material3.Icon
import androidx.wear.compose.material3.TimeText
import androidx.wear.compose.material3.curvedText
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textHint
import kotlinx.coroutines.delay
import kotlin.time.Duration.Companion.milliseconds

/**
 * Component that displays GPS status (Connecting / Connected) with animated icon and text
 * at the top of the watch screen (or TimeText if GPS is not supported on device)
 */
@Composable
fun GpsStatusIndicator(
    available: Boolean,
    modifier: Modifier = Modifier,
    supported: Boolean = true
) {
    if (!supported) {
        TimeText(modifier = modifier)
        return
    }

    /** (Animated) colors of the GPS signal icon and the text */
    val gpsColor = remember { Animatable(text) }
    /** Rotation value of the GPS icon */
    val gpsRotate = remember { mutableFloatStateOf(0f) }
    /** (Animated) rotation value of the GPS icon */
    val gpsRotating by animateFloatAsState(
        targetValue = gpsRotate.floatValue,
        animationSpec = tween(3000),
        label = "GPS rotating icon"
    )

    // Texts cannot be obtained inside curvedText
    val txtGpsConnected = stringResource(R.string.main_gpsConnected)
    val txtGpsConnecting = stringResource(R.string.main_gpsConnecting) + "..."

    LaunchedEffect(available) {
        var isDark = false
        while (!available) {
            gpsColor.animateTo(if (isDark) text else textHint, animationSpec = tween(1000))
            isDark = !isDark
            delay(1000.milliseconds)

            if (isDark) gpsRotate.floatValue += 360f
        }

        // Reset color if gps is connected
        gpsColor.animateTo(text, animationSpec = tween(500))
    }

    CurvedLayout(modifier = modifier) {
        curvedRow {
            curvedComposable(modifier = CurvedModifier.padding(ArcPaddingValues(after = 6.dp))) {
                Icon(
                    painter = painterResource(if (available) R.drawable.gps_connected else R.drawable.gps_connecting),
                    contentDescription = "GPS status",
                    modifier = Modifier
                        .size(20.dp)
                        .rotate(gpsRotating),
                    tint = gpsColor.value
                )
            }
            curvedText(
                text = if (available) txtGpsConnected else txtGpsConnecting,
                fontSize = 13.sp,
                color = gpsColor.value,
                maxSweepAngle = 360f
            )
        }
    }
}

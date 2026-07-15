package de.rpjosh.rpout.android.tiles

import android.annotation.SuppressLint
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.offset
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.remote.tooling.preview.RemoteDocumentPreview
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.glance.wear.GlanceWearWidget
import androidx.glance.wear.core.ContainerInfo
import androidx.glance.wear.core.WearWidgetParams
import androidx.wear.compose.material3.Text
import de.rpjosh.rpout.android.R
import kotlinx.coroutines.runBlocking

/**
 * Previews a [GlanceWearWidget] within a widget container in the Android Studio Preview.
 *
 * Note: This is taken from
 * https://android-review.googlesource.com/c/platform/frameworks/support/+/4045856 Once that change
 * lands in the library:
 * 1. Remove this file.
 * 2. Update imports and usages in the project to use
 *    `androidx.glance.wear.tooling.preview.WearWidgetPreview`.
 */
@SuppressLint("RestrictedApi")
@Composable
fun WearWidgetPreviewSnapshot(
    widget: GlanceWearWidget,
    params: WearWidgetParams,
    modifier: Modifier = Modifier,
    title: String = widget.javaClass.simpleName.replace(Regex("(?<=.)(?=\\p{Lu})"), " "),
) {
    val isTail = params.containerType == ContainerInfo.CONTAINER_TYPE_TILE_COMPAT

    val context = LocalContext.current
    val document =
        remember(widget, params, context) {
            runBlocking {
                val widgetData = widget.provideWidgetData(context, params)
                widgetData.captureRawContent(context, params).rcDocument
            }
        }

    Box(
        modifier = Modifier.size(227.dp).clip(CircleShape).background(Color.Black),
        contentAlignment = Alignment.Center,
    ) {
        RemoteDocumentPreview(
            document,
            modifier =
                modifier
                    .offset(y = if(isTail) 0.dp else 16.dp)
                    .width((params.widthDp + 2f * params.horizontalPaddingDp).dp)
                    .height((params.heightDp + 2f * params.verticalPaddingDp).dp),
        )

        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            modifier = Modifier.fillMaxSize().padding(top = if(isTail) 0.dp else 10.dp),
        ) {
            Box(
                modifier = Modifier.size(28.dp).clip(CircleShape).background(Color(0xFFE0E0E0)),
                contentAlignment = Alignment.Center,
            ) {
                Image(
                    painter = painterResource(id = R.drawable.ic_launcher_foreground),
                    contentDescription = "Android Logo",
                    modifier = Modifier.size(38.dp),
                    colorFilter = androidx.compose.ui.graphics.ColorFilter.tint(Color(0xFF424242)),
                )
            }
            Text(
                text = title,
                color = Color.White,
                fontSize = 14.sp,
                fontWeight = FontWeight.Bold,
                modifier = Modifier.padding(top = 2.dp),
            )
        }
    }
}
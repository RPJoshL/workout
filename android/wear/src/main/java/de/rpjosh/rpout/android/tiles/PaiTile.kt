package de.rpjosh.rpout.android.tiles

import android.annotation.SuppressLint
import android.content.Context
import android.graphics.Canvas
import android.util.TypedValue
import androidx.compose.remote.creation.compose.layout.RemoteAlignment
import androidx.compose.remote.creation.compose.layout.RemoteArrangement
import androidx.compose.remote.creation.compose.layout.RemoteBox
import androidx.compose.remote.creation.compose.layout.RemoteColumn
import androidx.compose.remote.creation.compose.layout.RemoteComposable
import androidx.compose.remote.creation.compose.layout.RemoteImage
import androidx.compose.remote.creation.compose.layout.RemoteRow
import androidx.compose.remote.creation.compose.layout.RemoteText
import androidx.compose.remote.creation.compose.modifier.RemoteModifier
import androidx.compose.remote.creation.compose.modifier.background
import androidx.compose.remote.creation.compose.modifier.clip
import androidx.compose.remote.creation.compose.modifier.fillMaxSize
import androidx.compose.remote.creation.compose.modifier.fillMaxWidth
import androidx.compose.remote.creation.compose.modifier.height
import androidx.compose.remote.creation.compose.modifier.padding
import androidx.compose.remote.creation.compose.modifier.size
import androidx.compose.remote.creation.compose.modifier.width
import androidx.compose.remote.creation.compose.shapes.RemoteRoundedCornerShape
import androidx.compose.remote.creation.compose.state.RemoteImageBitmap
import androidx.compose.remote.creation.compose.state.RemoteString
import androidx.compose.remote.creation.compose.state.asRemoteDp
import androidx.compose.remote.creation.compose.state.rb
import androidx.compose.remote.creation.compose.state.rc
import androidx.compose.remote.creation.compose.state.rdp
import androidx.compose.remote.creation.compose.state.rsp
import androidx.compose.remote.creation.compose.text.RemoteTextStyle
import androidx.compose.runtime.Composable
import androidx.compose.runtime.remember
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.glance.wear.AssociateWithGlanceWearWidget
import androidx.glance.wear.GlanceWearWidget
import androidx.glance.wear.GlanceWearWidgetService
import androidx.glance.wear.WearWidgetBrush
import androidx.glance.wear.WearWidgetData
import androidx.glance.wear.WearWidgetDocument
import androidx.glance.wear.core.WearWidgetParams
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.models.PaiDay
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.tooling.preview.PreviewParameter
import androidx.glance.wear.GlanceWearWidgetManager
import androidx.glance.wear.color
import de.rpjosh.rpout.android.activities.theme.paiFilled
import androidx.core.graphics.createBitmap
import androidx.glance.wear.core.WearWidgetRawContent
import androidx.wear.compose.remote.material3.RemoteMaterialTheme
import androidx.wear.tooling.preview.devices.WearDevices
import de.rpjosh.rpout.android.activities.theme.paiNone
import de.rpjosh.rpout.android.activities.theme.text

@AssociateWithGlanceWearWidget(PaiTileWidget::class)
class PaiTileService : GlanceWearWidgetService() {
    override val widget: GlanceWearWidget = PaiTileWidget()
}

class PaiTileWidget: GlanceWearWidget() {

    companion object {
        suspend fun triggerUpdate(context: Context) {
            val manager = GlanceWearWidgetManager(context)
            val widget = PaiTileWidget()
            manager.fetchActiveWidgets(widget::class).forEach {
                widget.triggerUpdate(context, it.instanceId)
            }
        }
    }

    override suspend fun provideWidgetData(
        context: Context,
        params: WearWidgetParams,
    ): WearWidgetData {
        val app = Singleton.getAppSec()
        val metricController = app.injection.inject(MetricController::class.java, null, false)

        // Fetch PAI progression from database
        val progression = withContext(Dispatchers.IO) {
            metricController.getPaiProgression()
        }

        return WearWidgetDocument(background = WearWidgetBrush.color(Color.Red.rc)) {
            if (progression.isEmpty()) NoPaisSynced()
            else PaiTileScreen(progression)
        }
    }

}

private class DummyPaiTileWidget(val progression: List<PaiDay>): GlanceWearWidget() {
    override suspend fun provideWidgetData(
        context: Context,
        params: WearWidgetParams,
    ): WearWidgetData {
        return WearWidgetDocument(background = WearWidgetBrush.color(Color(0xFF1E1D1D).rc)) {
            PaiTileScreen(progression)
            // NoPaisSynced()
        }
    }
}

@SuppressLint("RestrictedApi")
@RemoteComposable
@Composable
fun PaiTileScreen(progression: List<PaiDay>) {
    val context = LocalContext.current

    val paiImage = remember { renderBitmap(context, R.drawable.pai) }
    val paiNoneImage = remember { renderBitmap(context, R.drawable.pai_none) }

    // Maximum PAI score in all entries
    var maxVal = progression[0].value
    progression.forEach { if (it.value > maxVal) maxVal = it.value }
    val maxValue = maxVal.coerceAtLeast(1)

    // Default top padding for rows
    val imageHeight = 58.rdp

    RemoteBox(modifier = RemoteModifier.fillMaxSize()) {

        // Current score
        RemoteBox(contentAlignment = RemoteAlignment.Center, modifier = RemoteModifier.fillMaxWidth()) {
            RemoteImage(
                remoteBitmap = paiImage,
                contentDescription = RemoteString("PAI image"),
                modifier = RemoteModifier.size(imageHeight)
            )
            RemoteText(
                text = progression.last().value.toString(),
                style = RemoteTextStyle(
                    fontSize = 23.rsp, textAlign = TextAlign.Center,
                    fontWeight = FontWeight.Bold
                ),
                color = Color.White.rc
            )
        }

        RemoteRow(
            modifier = RemoteModifier.padding(top = 10.rdp, start = 4.rdp, end = 4.rdp)
                .fillMaxSize(),
            horizontalArrangement = RemoteArrangement.Center
        ) {
            progression.forEachIndexed { index, it ->
                val paddingTopOffset = when(index) {
                    0, 6 -> (-21).rdp
                    1, 5 -> (-14).rdp
                    2, 4 -> (-7).rdp
                    else -> 0.rdp
                }

                RemoteColumn(
                    horizontalAlignment = RemoteAlignment.CenterHorizontally,
                    modifier = RemoteModifier.padding(
                        start = 1.rdp,
                        end = 1.rdp,
                        top = (imageHeight.value + paddingTopOffset.value).asRemoteDp()
                    )
                ) {
                    // Score indicator
                    RemoteBox(contentAlignment = RemoteAlignment.Center) {
                        RemoteImage(
                            remoteBitmap = if(it.earned > 0) paiImage else paiNoneImage,
                            contentDescription = RemoteString("PAI image"),
                            modifier = RemoteModifier.size(22.rdp)
                        )
                        RemoteText(
                            text = it.earned.toString(),
                            style = RemoteTextStyle(fontSize = 10.rsp, textAlign = TextAlign.Center, fontWeight = FontWeight.Bold),
                            color = Color.White.rc
                        )
                    }

                    // Progress bar
                    val fullHeightDip = 65f
                    val fullHeightPx = TypedValue.applyDimension(TypedValue.COMPLEX_UNIT_DIP, fullHeightDip, context.resources.displayMetrics)
                    val density = context.resources.displayMetrics.density
                    val thisHeightDip = (fullHeightPx * (it.value.toDouble() / maxValue) / density).toFloat()

                    RemoteBox(
                        modifier = RemoteModifier.padding(top = 4.rdp).height(65.rdp)
                    ) {
                        // Partial borders are not supported and doesn't look good
                        if (thisHeightDip > 61) {
                            RemoteBox(
                                modifier = RemoteModifier
                                    .height(65.rdp)
                                    .width(9.rdp)
                                    .clip(RemoteRoundedCornerShape(4.rdp))
                                    .background(paiFilled),
                            )
                        } else {
                            // Overlay filled status
                            RemoteBox(
                                modifier = RemoteModifier
                                    .height(65.rdp)
                                    .width(9.rdp)
                                    .clip(RemoteRoundedCornerShape(4.rdp))
                                    .background(paiNone),
                                contentAlignment = RemoteAlignment.BottomCenter,
                            ){
                                RemoteBox(modifier = RemoteModifier
                                    .height(thisHeightDip.rdp).width(9.rdp)
                                    .background(paiFilled))
                            }
                        }
                    }

                    RemoteText(
                        text = it.weekdayAbbrevation,
                        style = RemoteTextStyle(fontSize = 10.rsp),
                        modifier = RemoteModifier.padding(top = 3.rdp),
                        color = Color.White.rc
                    )
                }
            }
        }
    }

}

/**
 * Prerender an image into a bitmap.
 *
 * This is required as "ImageBitmap.imageResource(R.drawable.pai).rb" results
 * into a timeout
 */
fun renderBitmap(context: Context, resID: Int): RemoteImageBitmap {
    val drawable = context.resources.getDrawable(resID, context.theme)

    val intrinsicWidth = drawable.intrinsicWidth.coerceAtLeast(1)
    val intrinsicHeight = drawable.intrinsicHeight.coerceAtLeast(1)

    val maxDimension = 128f
    val scale = (maxDimension / maxOf(intrinsicWidth, intrinsicHeight)).coerceAtMost(1f)
    val width = (intrinsicWidth * scale).toInt()
    val height = (intrinsicHeight * scale).toInt()

    val bitmap = createBitmap(width, height)
    val canvas = Canvas(bitmap)
    drawable.setBounds(0, 0, width, height)
    drawable.draw(canvas)

    return bitmap.asImageBitmap().rb
}

@RemoteComposable
@Composable
fun NoPaisSynced() {
    RemoteBox(
        modifier = RemoteModifier.fillMaxSize(),
        contentAlignment = RemoteAlignment.Center
    ) {
        RemoteText(
            text = stringResource(R.string.tilte_pai_notSynced),
            color = text.rc
        )
    }
}

@Preview(device = WearDevices.SMALL_ROUND)
@Composable
fun PaiTilePreview(
    @PreviewParameter(WearWidgetParamsProviderSnapshot::class) params: WearWidgetParams
) {
    val dummyProgression = listOf(
        PaiDay(0, 10, 2, "Mo", 0),
        PaiDay(1, 15, 5, "Di", 1),
        PaiDay(2, 48, 5, "Mi", 2),
        PaiDay(3, 25, 5, "Do", 3),
        PaiDay(4, 30, 5, "Fr", 4),
        PaiDay(5, 40, 10, "Sa", 5),
        PaiDay(6, 50, 10, "So", 6)
    )

    WearWidgetPreviewSnapshot(DummyPaiTileWidget(dummyProgression), params, title = "")
}

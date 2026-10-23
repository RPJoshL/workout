package de.rpjosh.rpout.android.tiles

import android.annotation.SuppressLint
import android.util.TypedValue
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.glance.GlanceComposable
import androidx.glance.GlanceModifier
import androidx.glance.Image
import androidx.glance.ImageProvider
import androidx.glance.LocalContext
import androidx.glance.background
import androidx.glance.layout.Alignment
import androidx.glance.layout.Box
import androidx.glance.layout.Column
import androidx.glance.layout.Row
import androidx.glance.layout.fillMaxSize
import androidx.glance.layout.fillMaxWidth
import androidx.glance.layout.height
import androidx.glance.layout.padding
import androidx.glance.layout.size
import androidx.glance.layout.width
import androidx.glance.wear.tiles.GlanceTileService
import androidx.glance.text.Text
import androidx.glance.text.TextAlign
import androidx.glance.text.TextStyle
import androidx.glance.unit.ColorProvider
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.models.PaiDay
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.runBlocking
import kotlinx.coroutines.withContext

class PaiTile : GlanceTileService() {

    @GlanceComposable
    @Composable
    override fun Content() {
        val app = Singleton.getAppSec(this)
        val metricController = app.injection.inject(MetricController::class.java, null, false)

        app.sharedLogger.log("d", "Rendering PAI progression tile")

        // Hack: database request on main thread :)
        val progression = runBlocking {
            withContext(Dispatchers.IO) {
                metricController.getPaiProgression()
            }
        }

        if (progression.isEmpty()) NoPaisSynced()
        else PaiTileScreen(progression)
    }

}

@SuppressLint("RestrictedApi")
@Composable
@GlanceComposable
fun PaiTileScreen(progression: List<PaiDay>) {
    val context = LocalContext.current

    // Scaling for large devices
    val displayMetrics = context.resources.displayMetrics
    val screenWidthDp = displayMetrics.widthPixels / displayMetrics.density
    val scale = (screenWidthDp / 192f).coerceAtLeast(1.0f)

    // Maximum PAI score in all entries
    var max = progression[0]
    progression.forEach { if (it.value > max.value) max = it }

    // Scaled element dimensions
    val imageHeight = 58.dp * scale
    val barWidth = 9.dp * scale
    val barFullHeightDpVal = 65f * scale
    val barFullHeightDp = barFullHeightDpVal.dp

    Box(modifier = GlanceModifier.fillMaxSize().padding(top = 2.dp * scale)) {

        // Current score
        Box(contentAlignment = Alignment.Center, modifier = GlanceModifier.fillMaxWidth()) {
            Image(
                provider = ImageProvider(R.drawable.pai),
                contentDescription = "PAI image",
                modifier = GlanceModifier.size(imageHeight)
            )
            Text(
                text = progression.last().value.toString(),
                style = TextStyle(
                    fontSize = (23 * scale).sp, textAlign = TextAlign.Center, fontWeight = androidx.glance.text.FontWeight.Bold,
                    color = ColorProvider(Color.White)
                ),
                modifier = GlanceModifier.fillMaxSize()
            )
        }

        Row(
            modifier = GlanceModifier.padding(top = 10.dp * scale, start = 4.dp * scale, end = 4.dp * scale).fillMaxSize().height(380.dp * scale),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            progression.forEachIndexed { index, it ->
                val paddingTopOffset = when(index) {
                    0, 6 -> (-21).dp * scale
                    1, 5 -> (-14).dp * scale
                    2, 4 -> (-7).dp * scale
                    else -> 0.dp
                }

                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    modifier = GlanceModifier.padding(horizontal = 1.dp * scale, vertical = imageHeight + paddingTopOffset)
                ) {
                    // Score indicator
                    Box {
                        Image(
                            provider = ImageProvider(if(it.earned > 0) R.drawable.pai else R.drawable.pai_none),
                            contentDescription = "PAI image",
                            modifier = GlanceModifier.size(22.dp * scale)
                        )
                        Text(
                            text = it.earned.toString(),
                            style = TextStyle(
                                fontSize = (10 * scale).sp, textAlign = TextAlign.Center, fontWeight = androidx.glance.text.FontWeight.Bold,
                                color = ColorProvider(Color.White)
                            ),
                            modifier = GlanceModifier.fillMaxSize(),
                        )
                    }

                    // Progress bar
                    val fullHeight = TypedValue.applyDimension(TypedValue.COMPLEX_UNIT_DIP, barFullHeightDpVal, displayMetrics)
                    val density = displayMetrics.density
                    val thisHeight = fullHeight * (it.value.toDouble() / max.value) / density

                    Box(modifier = GlanceModifier.padding(top = 4.dp * scale).height(barFullHeightDp)) {
                        // If height > 61.dp, we have to use an image (because we cannot use rounded corners in glance)
                        if (thisHeight > 61 * scale) {
                            Image(
                                provider = ImageProvider(R.drawable.pai_bar_filled), contentDescription = "PAI bar",
                                modifier = GlanceModifier.width(barWidth).height(barFullHeightDp)
                            )
                        } else {
                            // Default background image
                            Image(
                                provider = ImageProvider(R.drawable.pai_bar), contentDescription = "PAI bar",
                                modifier = GlanceModifier.width(barWidth).height(barFullHeightDp)
                            )

                            // Overlay filled status
                            Box(modifier = GlanceModifier.height(barFullHeightDp), contentAlignment = Alignment.BottomCenter){
                                Box(modifier = GlanceModifier.height(thisHeight.dp).width(barWidth).background(R.color.paiFilled)){}
                            }
                        }
                    }

                    Text(
                        text = it.weekdayAbbrevation,
                        style = TextStyle(fontSize = (10 * scale).sp, color = ColorProvider(Color.White)),
                        modifier = GlanceModifier.padding(top = 3.dp * scale)
                    )
                }
            }
        }
    }

}

@Composable
@GlanceComposable
fun NoPaisSynced() {
    Box {
        Text(
            text = LocalContext.current.getString(R.string.tilte_pai_notSynced),
            style = TextStyle()
        )
    }
}

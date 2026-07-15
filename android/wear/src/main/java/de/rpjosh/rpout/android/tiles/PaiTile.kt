package de.rpjosh.rpout.android.tiles

import android.annotation.SuppressLint
import android.content.Context
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Paint
import android.text.TextPaint
import android.util.TypedValue
import androidx.annotation.FontRes
import androidx.compose.runtime.Composable
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.unit.TextUnit
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.res.ResourcesCompat
import androidx.glance.GlanceComposable
import androidx.glance.GlanceModifier
import androidx.glance.Image
import androidx.glance.ImageProvider
import androidx.glance.LocalContext
import androidx.glance.background
import androidx.glance.color.ColorProviders
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
        val app = Singleton.getAppSec()
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

    // Maximum PAI score in all entries
    var max = progression[0]
    progression.forEach { if (it.value > max.value) max = it }

    // Default top padding for rows
    val imageHeight = 58.dp

    Box(modifier = GlanceModifier.fillMaxSize().padding(top = 2.dp)) {

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
                    fontSize = 23.sp, textAlign = TextAlign.Center, fontWeight = androidx.glance.text.FontWeight.Bold,
                    color = ColorProvider(Color.White)
                ),
                modifier = GlanceModifier.fillMaxSize()
            )
        }

        Row(modifier = GlanceModifier.padding(top = 10.dp, start = 4.dp, end = 4.dp).fillMaxSize().height(380.dp), horizontalAlignment = Alignment.CenterHorizontally) {
            progression.forEachIndexed { index, it ->
                val paddingTopOffset = when(index) {
                    0, 6 -> (-21).dp
                    1, 5 -> (-14).dp
                    2, 4 -> (-7).dp
                    else -> 0.dp
                }

                Column(
                    horizontalAlignment = Alignment.CenterHorizontally,
                    modifier = GlanceModifier.padding(horizontal = 1.dp, vertical = imageHeight + paddingTopOffset)
                ) {
                    // Score indicator
                    Box {
                        Image(
                            provider = ImageProvider(if(it.earned > 0) R.drawable.pai else R.drawable.pai_none),
                            contentDescription = "PAI image",
                            modifier = GlanceModifier.size(22.dp)
                        )
                        Text(
                            text = it.earned.toString(),
                            style = TextStyle(
                                fontSize = 10.sp, textAlign = TextAlign.Center, fontWeight = androidx.glance.text.FontWeight.Bold,
                                color = ColorProvider(Color.White)
                            ),
                            modifier = GlanceModifier.fillMaxSize(),
                        )
                    }

                    // Progress bar
                    val fullHeight = TypedValue.applyDimension(TypedValue.COMPLEX_UNIT_DIP, 65f, context.resources.displayMetrics)
                    val density = context.resources.displayMetrics.density
                    val thisHeight = fullHeight * (it.value.toDouble() / max.value) / density

                    Box(modifier = GlanceModifier.padding(top = 4.dp).height(65.dp)) {
                        // If height > 61.dp, we have to use an image (because we cannot use rounded corners in glance)
                        if (thisHeight > 61) {
                            Image(
                                provider = ImageProvider(R.drawable.pai_bar_filled), contentDescription = "PAI bar",
                                modifier = GlanceModifier.width(9.dp).height(65.dp)
                            )
                        } else {
                            // Default background image
                            Image(
                                provider = ImageProvider(R.drawable.pai_bar), contentDescription = "PAI bar",
                                modifier = GlanceModifier.width(9.dp).height(65.dp)
                            )

                            // Overlay filled status
                            Box(modifier = GlanceModifier.height(65.dp), contentAlignment = Alignment.BottomCenter){
                                Box(modifier = GlanceModifier.height(thisHeight.dp).width(9.dp).background(R.color.paiFilled)){}
                            }
                        }
                    }

                    Text(
                        text = it.weekdayAbbrevation,
                        style = TextStyle(fontSize = 10.sp, color = ColorProvider(Color.White)),
                        modifier = GlanceModifier.padding(top = 3.dp)
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

@Composable
fun GlanceText(
    text: String,
    @FontRes font: Int,
    fontSize: TextUnit,
    modifier: GlanceModifier = GlanceModifier,
    color: Color = Color.Black,
    letterSpacing: TextUnit = 0.1.sp
) {
    Image(
        modifier = modifier,
        provider = ImageProvider(
            LocalContext.current.textAsBitmap(
                text = text,
                fontSize = fontSize,
                color = color,
                font = font,
                letterSpacing = letterSpacing.value
            )
        ),
        contentDescription = null,
    )
}
fun Context.textAsBitmap(
    text: String,
    fontSize: TextUnit,
    color: Color = Color.Black,
    letterSpacing: Float = 0.1f,
    font: Int
): Bitmap {
    val paint = TextPaint(Paint.ANTI_ALIAS_FLAG)
    paint.textSize = TypedValue.applyDimension(TypedValue.COMPLEX_UNIT_SP, fontSize.value, this.resources.displayMetrics)
    paint.color = color.toArgb()
    paint.letterSpacing = letterSpacing
    paint.typeface = ResourcesCompat.getFont(this, font)

    val baseline = -paint.ascent()
    val width = (paint.measureText(text)).toInt()
    val height = (baseline + paint.descent()).toInt()
    val image = Bitmap.createBitmap(width, height, Bitmap.Config.ARGB_8888)
    val canvas = Canvas(image)
    canvas.drawText(text, 0f, baseline, paint)
    return image
}
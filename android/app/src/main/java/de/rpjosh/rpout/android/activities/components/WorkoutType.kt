package de.rpjosh.rpout.android.activities.components

import android.annotation.SuppressLint
import android.graphics.Bitmap
import android.graphics.Canvas
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.wrapContentSize
import androidx.compose.material3.Icon
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.asImageBitmap
import androidx.compose.ui.graphics.painter.BitmapPainter
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.core.graphics.createBitmap
import androidx.core.graphics.toColorInt
import com.caverock.androidsvg.SVG
import de.rpjosh.rpout.android.activities.theme.RPoutTheme

const val dummyWorkoutIcon = "<svg fill=\"currentColor\" viewBox=\"0 0 552.855 552.855\" version=\"1.1\" xml:space=\"preserve\" xmlns=\"http://www.w3.org/2000/svg\">\n" +
        "  <path d=\"m511.9 157.42c-3.408-25.845-17.057-53.513-40-76.463-22.943-22.944-50.605-36.585-76.445-39.994-11.695-1.542-27.307-8.005-36.664-15.184-20.691-15.869-49.902-25.784-82.357-25.784s-61.665 9.915-82.351 25.784c-9.357 7.179-24.97 13.642-36.665 15.184-25.845 3.409-53.501 17.05-76.445 39.994-22.944 22.95-36.592 50.619-40 76.463-1.536 11.695-8.005 27.295-15.178 36.653-15.875 20.686-25.79 49.896-25.79 82.35 0 32.455 9.915 61.666 25.784 82.352 7.179 9.357 13.642 24.963 15.178 36.652 3.409 25.844 17.056 53.514 40 76.463 22.944 22.943 50.606 36.586 76.445 39.994 11.695 1.543 27.308 8.006 36.665 15.184 20.686 15.869 49.896 25.783 82.351 25.783s61.666-9.914 82.352-25.783c9.357-7.178 24.969-13.641 36.664-15.184 25.846-3.408 53.502-17.051 76.445-39.994 22.943-22.949 36.592-50.619 40-76.463 1.537-11.695 8.006-27.295 15.178-36.652 15.869-20.686 25.783-49.896 25.783-82.352 0-32.454-9.914-61.665-25.783-82.35-7.167-9.358-13.629-24.958-15.167-36.653zm-202.38 275.77c0 6.764-5.484 12.24-12.24 12.24h-39.652c-6.756 0-12.24-5.477-12.24-12.24v-39.65c0-6.764 5.483-12.24 12.24-12.24h39.652c6.756 0 12.24 5.477 12.24 12.24v39.65zm74.977-189.52c-7.994 12.632-25.068 29.823-51.238 51.58-13.543 11.26-21.951 20.312-25.221 27.16-2.447 5.135-3.904 13.305-4.344 24.51-0.264 6.758-5.588 12.234-12.352 12.234h-33.72c-6.757 0-12.301-3.428-12.369-7.65-0.061-3.916-0.098-6.463-0.098-7.643 0-18.869 3.122-34.389 9.357-46.562 6.243-12.172 18.715-25.862 37.429-41.083 18.717-15.214 29.896-25.184 33.545-29.896 5.631-7.454 8.445-15.673 8.445-24.651 0-12.479-4.975-23.164-14.945-32.069-9.969-8.904-23.391-13.354-40.281-13.354-11.604 0-21.842 2.356-30.741 7.068-5.973 3.164-14.406 10.355-18.476 15.753-4.541 6.022-8.219 13.268-11.034 21.738-2.136 6.414-8.636 11.138-15.349 10.306l-34.584-4.29c-6.714-0.833-11.897-6.995-10.698-13.648 3.978-22.032 15.098-41.114 33.354-57.24 21.53-19.021 49.792-28.531 84.786-28.531 36.824 0 66.107 9.626 87.865 28.874 21.756 19.248 32.639 41.653 32.639 67.216 0.012 14.162-3.984 27.552-11.97 40.178z\"/>\n" +
        "</svg>"

@Composable
@SuppressLint("ModifierParameter")
fun SvgIcon(
    svgString: String, size: Dp,
    modifier: Modifier = Modifier.wrapContentSize(align = Alignment.Center),
    tint: Color? = null, hexTint: String? = null
) {

    // Apply tint from hex color
    var iTint = tint ?: RPoutTheme.colors.text
    if (hexTint != null) {
        iTint = Color(hexTint.toColorInt())
    }

    // Initialize SVG
    val svg = SVG.getFromString(svgString)
    svg.documentWidth = with(LocalDensity.current) { size.toPx() }
    svg.documentHeight = svg.documentWidth

    // Convert it into a drawable
    val drawable = svg.renderToPicture()
    val bitmap = createBitmap(svg.documentWidth.toInt(), svg.documentHeight.toInt())
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
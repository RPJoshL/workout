package de.rpjosh.rpout.android.activities.theme

import android.app.Activity
import android.os.Build
import androidx.compose.foundation.isSystemInDarkTheme
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.material3.contentColorFor
import androidx.compose.material3.darkColorScheme
import androidx.compose.material3.dynamicDarkColorScheme
import androidx.compose.material3.dynamicLightColorScheme
import androidx.compose.material3.lightColorScheme
import androidx.compose.runtime.Composable
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.ReadOnlyComposable
import androidx.compose.runtime.SideEffect
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalView
import androidx.core.view.WindowCompat

/** Helper to access theme variables based on system theme */
object RPoutTheme {
    val colors: RPoutCustomColors
        @Composable
        @ReadOnlyComposable
        get() = LocalRPoutColors.current
}

private val DarkColorScheme = darkColorScheme(
    primary = DarkColorPalette.secondary,
    secondary = DarkColorPalette.secondary,
    tertiary = DarkColorPalette.backgroundWork1,
    surface = DarkColorPalette.backgroundDarker,
    error = DarkColorPalette.error,
    background = DarkColorPalette.defaultBackground
)

private val LightColorScheme = lightColorScheme(
    primary = LightColorPalette.secondary,
    secondary = LightColorPalette.secondary,
    tertiary = LightColorPalette.backgroundWork1,
    surface = LightColorPalette.backgroundDarker,
    error = LightColorPalette.error,
    background = LightColorPalette.defaultBackground
)

@Composable
fun RPoutTheme(
    darkTheme: Boolean = isSystemInDarkTheme(),
    dynamicColor: Boolean = true,
    content: @Composable () -> Unit
) {
    val customColors = if (darkTheme) DarkColorPalette else LightColorPalette

    val colorScheme = when {
        dynamicColor && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S -> {
            val context = LocalContext.current
            if (darkTheme) dynamicDarkColorScheme(context) else dynamicLightColorScheme(context)
        }
        darkTheme -> DarkColorScheme
        else -> LightColorScheme
    }

    val view = LocalView.current
    if (!view.isInEditMode) {
        SideEffect {
            val window = (view.context as Activity).window
            WindowCompat.getInsetsController(window, view).isAppearanceLightStatusBars = !darkTheme
        }
    }

    CompositionLocalProvider(
        LocalRPoutColors provides customColors
    ) {
        MaterialTheme(
            colorScheme = colorScheme,
            typography = Typography,
            content = {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background,
                    contentColor = contentColorFor(MaterialTheme.colorScheme.background)
                ) {
                    content()
                }
            }
        )
    }
}

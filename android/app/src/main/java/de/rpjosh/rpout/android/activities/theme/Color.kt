package de.rpjosh.rpout.android.activities.theme

import androidx.compose.runtime.Immutable
import androidx.compose.runtime.staticCompositionLocalOf
import androidx.compose.ui.graphics.Color

@Immutable
data class RPoutCustomColors(
    val defaultBackground: Color,
    val backgroundDarker: Color,
    val backgroundLightDarker: Color,
    val backgroundSelection: Color,
    val backgroundLighter: Color,
    val background1: Color,
    val background2: Color,
    val backgroundDisabled: Color,
    val backgroundDisabledDarker: Color,
    val backgroundError: Color,
    val backgroundSuccess: Color,
    val backgroundWork1: Color,
    val backgroundWork2: Color,
    val buttonHover: Color,
    val text: Color,
    val textDarker: Color,
    val textHint: Color,
    val textHintDarker: Color,
    val textBlue: Color,
    val textGreen: Color,
    val secondary: Color,
    val accentBlueBorder: Color,
    val error: Color,
    val errorStatic: Color,
    val success: Color,
    val webviewHeaderColor: Color,
)

val DarkColorPalette = RPoutCustomColors(
    defaultBackground = Color(0xFF232629),
    backgroundDarker = Color(0xFF16191C),
    backgroundLightDarker = Color(0xFF1B1E21),
    backgroundSelection = Color(0xFF514B4B),
    backgroundLighter = Color(0xFF31363C),
    background1 = Color(0xFF32214B),
    background2 = Color(0xFF304B33),
    backgroundDisabled = Color(0xff31363f),
    backgroundDisabledDarker = Color(0xFF292D35),
    backgroundError = Color(0xFF680101),
    backgroundSuccess = Color(0xFF035C07),
    backgroundWork1 = Color(0xFFDD8C00),
    backgroundWork2 = Color(0xFFF48236),
    buttonHover = Color(0xFF716D6D),
    text = Color(0xFFFFFFFF),
    textDarker = Color(0xCDFFFFFF),
    textHint = Color(0xFD888181),
    textHintDarker = Color(0x727D7878),
    textBlue = Color(0xFF41C4FF),
    textGreen = Color(0xFF85FF8A),
    secondary = Color(0xFF0A80DD),
    accentBlueBorder = Color(0xFF085B9D),
    error = Color(0xFFFF0000),
    errorStatic = Color(0xFF8D0505),
    success = Color(0xFF2AA836),
    webviewHeaderColor = Color(0xFF141414)
)

val LightColorPalette = RPoutCustomColors(
    defaultBackground = Color(0xFFFFFFFF),
    backgroundDarker = Color(0xFFF1F1F1),
    backgroundLightDarker = Color(0xFFEEEEEE),
    backgroundSelection = Color(0xFFE0E0E0),
    backgroundLighter = Color(0xFFF9F9F9),
    background1 = Color(0xFFEDE7F6),
    background2 = Color(0xFFE8F5E9),
    backgroundDisabled = Color(0xFFF1F1F1),
    backgroundDisabledDarker = Color(0xFFE0E0E0),
    backgroundError = Color(0xFFFFEBEE),
    backgroundSuccess = Color(0xFFE8F5E9),
    backgroundWork1 = Color(0xFFFFF3E0),
    backgroundWork2 = Color(0xFFFBE9E7),
    buttonHover = Color(0xFFE0E0E0),
    text = Color(0xFF000000),
    textDarker = Color(0xFF212121),
    textHint = Color(0xFF757575),
    textHintDarker = Color(0xFF9E9E9E),
    textBlue = Color(0xFF01579B),
    textGreen = Color(0xFF1B5E20),
    secondary = Color(0xFF0A80DD),
    accentBlueBorder = Color(0xFF085B9D),
    error = Color(0xFFD32F2F),
    errorStatic = Color(0xFFB71C1C),
    success = Color(0xFF388E3C),
    webviewHeaderColor = Color(0xFFFFFFFF)
)

val LocalRPoutColors = staticCompositionLocalOf { DarkColorPalette }

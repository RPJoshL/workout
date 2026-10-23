package de.rpjosh.rpout.android.activities.components

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.wear.compose.foundation.CurvedLayout
import androidx.wear.compose.foundation.CurvedModifier
import androidx.wear.compose.foundation.curvedRow
import androidx.wear.compose.foundation.padding
import androidx.wear.compose.material3.curvedText

@Composable
fun CustomTimeText(text: String, modifier: Modifier = Modifier, color: Color = Color.White, alpha: Float = 1.0f, clickable: (() -> Unit)? = null, clickableHeight: Dp = 22.dp) {
    Box(modifier = modifier) {
        CurvedLayout {
                curvedText(
                    text = text,
                    fontSize = 13.sp,
                    color = color.copy(alpha = alpha)
                )
        }

        if (clickable != null && alpha != 0f) {
            Box(
                modifier = Modifier
                    .fillMaxWidth().height(clickableHeight)
                    .clickable { clickable() }
            )
        }
    }
}

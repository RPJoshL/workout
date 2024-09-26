package de.rpjosh.rpout.android.activities.theme

import androidx.compose.runtime.Composable
import androidx.wear.compose.material.MaterialTheme

@Composable
fun RPoutTheme(
    content: @Composable () -> Unit
) {
    MaterialTheme(
        content = content,
    )
}
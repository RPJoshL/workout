package de.rpjosh.rpout.android.activities.theme

import androidx.compose.runtime.Composable
import androidx.compose.ui.text.font.Font
import androidx.wear.compose.material.MaterialTheme

val FontSegoeUi = Font(de.rpjosh.rpout.android.shared.R.font.segoe_ui)
val FontSourceCodeProSemibold = Font(de.rpjosh.rpout.android.shared.R.font.source_code_pro_semibold)
val FontSourceSanseProSemibold = Font(de.rpjosh.rpout.android.shared.R.font.source_sanse_pro_semibold)
val FontSourceSansePro = Font(de.rpjosh.rpout.android.shared.R.font.source_sanse_pro_regular)


@Composable
fun RPoutTheme(
    content: @Composable () -> Unit
) {
    MaterialTheme(
        content = content,
    )
}
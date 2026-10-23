package de.rpjosh.rpout.android.activities

import androidx.compose.ui.tooling.preview.Preview
import androidx.wear.tooling.preview.devices.WearDevices


@Target(AnnotationTarget.FUNCTION, AnnotationTarget.ANNOTATION_CLASS)
@Retention(AnnotationRetention.BINARY)
@Preview(
    name = "Small Round",
    device = WearDevices.SMALL_ROUND,
    showSystemUi = true
)
@Preview(
    name = "Large Round",
    device = WearDevices.LARGE_ROUND,
    showSystemUi = true
)
annotation class WearPreview
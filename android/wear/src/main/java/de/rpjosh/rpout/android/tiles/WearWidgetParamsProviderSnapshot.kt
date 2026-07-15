package de.rpjosh.rpout.android.tiles

import android.annotation.SuppressLint
import androidx.compose.ui.tooling.preview.PreviewParameterProvider
import androidx.glance.wear.core.ContainerInfo
import androidx.glance.wear.core.WearWidgetParams
import androidx.glance.wear.core.WidgetInstanceId

/**
 * A [PreviewParameterProvider] that provides a variety of [WearWidgetParams] for Wear previews.
 *
 * Note: This is taken from
 * https://android-review.googlesource.com/c/platform/frameworks/support/+/4045856 Once that change
 * lands in the library:
 * 1. Remove this file.
 * 2. Update imports and usages in the project to use
 *    `androidx.glance.wear.tooling.preview.WearWidgetParamsProvider`.
 */
// Suppressed RestrictedApi because WearWidgetParams is currently restricted to LIBRARY_GROUP.
class WearWidgetParamsProviderSnapshot : PreviewParameterProvider<WearWidgetParams> {
    @SuppressLint("RestrictedApi")
    override val values: Sequence<WearWidgetParams> =
        sequenceOf(
            // Small Widget Preview
            WearWidgetParams(
                instanceId = WidgetInstanceId("widgets", 2),
                containerType = ContainerInfo.CONTAINER_TYPE_SMALL,
                widthDp = 180f,
                heightDp = 60f,
                verticalPaddingDp = 6f,
                horizontalPaddingDp = 8f,
                cornerRadiusDp = 26f,
            ),
            // Large Widget Preview
            WearWidgetParams(
                instanceId = WidgetInstanceId("widgets", 1),
                containerType = ContainerInfo.CONTAINER_TYPE_LARGE,
                widthDp = 180f,
                heightDp = 82f,
                verticalPaddingDp = 6f,
                horizontalPaddingDp = 8f,
                cornerRadiusDp = 32f,
            ),
            // Tile Preview
            WearWidgetParams(
                instanceId = WidgetInstanceId("widgets", 3),
                containerType = ContainerInfo.CONTAINER_TYPE_TILE_COMPAT,
                widthDp = 200f,
                heightDp = 200f,
                verticalPaddingDp = 0f,
                horizontalPaddingDp = 0f,
                cornerRadiusDp = 100f,
            ),
        )
}
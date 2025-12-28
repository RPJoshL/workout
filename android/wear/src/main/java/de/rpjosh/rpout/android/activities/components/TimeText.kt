package de.rpjosh.rpout.android.activities.components

import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.mutableIntStateOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.snapshotFlow
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.wear.compose.foundation.CurvedLayout
import androidx.wear.compose.foundation.CurvedModifier
import androidx.wear.compose.foundation.curvedRow
import androidx.wear.compose.foundation.lazy.ScalingLazyListState
import androidx.wear.compose.foundation.padding
import androidx.wear.compose.material.curvedText
import androidx.wear.compose.material.scrollAway
import kotlinx.coroutines.flow.distinctUntilChanged
import androidx.compose.foundation.clickable

@Composable
fun CustomTimeText(text: String, modifier: Modifier = Modifier, color: Color = Color.White, alpha: Float = 1.0f, clickable: (() -> Unit)? = null, clickableHeight: Dp = 22.dp) {
    Box(modifier = modifier) {
        CurvedLayout {
            curvedRow(modifier = CurvedModifier.padding(1.dp)) {
                curvedText(
                    text = text,
                    fontSize = 13.sp,
                    color = color.copy(alpha = alpha)
                )
            }
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


/** Provides a modifier for elements that should automatically scrolls away with it's root scrollable lazy list */
@Composable
fun AutoScrollAwayTimeText(state: ScalingLazyListState, scrollOffset: Dp = 0.dp, component: @Composable (modifier: Modifier) -> Unit) {
    val showAlways = remember { mutableStateOf(true) }
    val scrollAwayOffset = remember { mutableIntStateOf(0) }

    LaunchedEffect(state) {
        snapshotFlow { /* state.centerItemIndex.toString() + "-" + state.centerItemScrollOffset*/ state.layoutInfo }
            .distinctUntilChanged()
            .collect { _ ->
                val canScroll = state.canScrollForward || state.canScrollBackward

                // Always show time text if view cannot be scrolled
                if (!canScroll) {
                    showAlways.value = true
                    return@collect
                } else {
                    showAlways.value = false
                }

                scrollAwayOffset.intValue = ((state.layoutInfo.visibleItemsInfo.getOrNull(0)?.size ?: 0) / -2) + (state.layoutInfo.beforeContentPadding * 2)
            }
    }

    component(if(showAlways.value) Modifier else Modifier.scrollAway(state, offset = with(LocalDensity.current){ scrollAwayOffset.intValue.toDp() + scrollOffset}, itemIndex = 0))
}
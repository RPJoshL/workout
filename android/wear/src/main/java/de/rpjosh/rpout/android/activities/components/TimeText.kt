package de.rpjosh.rpout.android.activities.components

import androidx.compose.animation.core.animateFloatAsState
import androidx.compose.animation.core.tween
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.BoxScope
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.unit.Dp
import androidx.compose.ui.unit.LayoutDirection
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.wear.compose.foundation.CurvedDirection
import androidx.wear.compose.foundation.CurvedLayout
import androidx.wear.compose.foundation.CurvedModifier
import androidx.wear.compose.foundation.ScrollInfoProvider
import androidx.wear.compose.foundation.curvedRow
import androidx.wear.compose.foundation.lazy.ScalingLazyColumn
import androidx.wear.compose.foundation.lazy.ScalingLazyListItemScope
import androidx.wear.compose.foundation.lazy.ScalingLazyListScope
import androidx.wear.compose.foundation.lazy.ScalingLazyListState
import androidx.wear.compose.foundation.lazy.items
import androidx.wear.compose.foundation.padding
import androidx.wear.compose.material3.AppScaffold
import androidx.wear.compose.material3.ScreenScaffold
import androidx.wear.compose.material3.ScreenScaffoldDefaults
import androidx.wear.compose.material3.TimeText
import androidx.wear.compose.material3.curvedText
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.distinctUntilChanged
import kotlinx.coroutines.launch
import kotlin.time.Duration.Companion.milliseconds

@Composable
fun CustomTimeText(text: String, modifier: Modifier = Modifier, color: Color = Color.White, alpha: Float = 1.0f, clickable: (() -> Unit)? = null, clickableHeight: Dp = 22.dp, bottom: Boolean = false) {
    Box(modifier = modifier) {
        CurvedLayout(
            anchor = if(bottom) 90f else 270f,
            angularDirection = if(bottom) CurvedDirection.Angular.Reversed else CurvedDirection.Angular.Normal
        ) {
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
fun AutoScrollAwayTimeText(
    state: ScalingLazyListState,
    scrollOffset: Dp = 0.dp,
    component: @Composable (modifier: Modifier) -> Unit
) {
    val height = LocalConfiguration.current.screenHeightDp
    // The TimeText itself needs space on the screen
    val additionalOffset = (-15).dp

    val density = LocalDensity.current
    val thresholdPx = with(density) { maxOf(0.dp, scrollOffset + additionalOffset).toPx() }

    val fadeDistancePx = with(density) { 25.dp.toPx() }
    var timeTextJob by remember { mutableStateOf<Job?>(null) }

    val scrollInfoProvider = remember(state) { ScrollInfoProvider(state) }

    val scrollProgress by remember(state, thresholdPx, fadeDistancePx) {
        derivedStateOf {
            val offsetPx = scrollInfoProvider.anchorItemOffset
            if (offsetPx.isNaN()) {
                1f
            } else if (offsetPx <= thresholdPx) {
                0f
            } else {
                ((offsetPx - thresholdPx) / fadeDistancePx).coerceIn(0f, 1f)
            }
        }
    }

    var showTimeText by remember { mutableStateOf(true) }

    LaunchedEffect(state.isScrollInProgress) {
        if (state.isScrollInProgress) {
            timeTextJob?.cancel()
            showTimeText = false
        } else {
            timeTextJob = launch {
                delay(2000.milliseconds)
                showTimeText = true
            }
        }
    }

    val timeTextAlpha by animateFloatAsState(
        targetValue = if (showTimeText) 1f else 0f,
        animationSpec = tween(durationMillis = 200),
        label = "TimeTextAlpha"
    )

    var modifier = Modifier.graphicsLayer {
        alpha = 1f - scrollProgress
        translationY = -scrollProgress * 40f
    }

    if(showTimeText && scrollProgress == 1f) {
        modifier = Modifier.alpha(timeTextAlpha)
    }

    //CustomTimeText(text = scrollInfoProvider.anchorItemOffset .toString())
    component(modifier)
}


@Composable
fun AppListWithTimeText(
    paddingTop: Dp = 26.dp,
    paddingBottom: Dp = 45.dp,
    paddingInline: Dp = 0.dp,
    verticalArrangement: Dp = 4.dp,
    timeText: (@Composable (modifier: Modifier) -> Unit)? = null,
    edgeButton: (@Composable BoxScope.() -> Unit)? = null,
    content: ListItemScope.(ScalingLazyListState) -> Unit
) {
    val displayHeight = LocalConfiguration.current.screenHeightDp

    val listState = remember { ScalingLazyListState(
        initialCenterItemIndex = 1,
    )}

    val defaultInline = ScreenScaffoldDefaults.contentPadding.calculateLeftPadding(LayoutDirection.Ltr)
    val defaultTop = ScreenScaffoldDefaults.contentPadding.calculateTopPadding()

    AppScaffold(
        timeText = {}
    ) {
        ScreenScaffold(
            scrollState = listState,
            edgeButton = edgeButton ?: {}
        ) {
            ScalingLazyColumn(
                modifier = Modifier.fillMaxSize(),
                state = listState,
                contentPadding = PaddingValues(
                    start = defaultInline + paddingInline, end = defaultInline + paddingInline,
                    top = paddingTop - verticalArrangement, bottom = defaultTop + paddingBottom
                ),
                horizontalAlignment =  Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(verticalArrangement),
                autoCentering = null,
            ) {
                val appScope = ListItemScope(this)

                item(key = "spacer-top") {
                    Spacer(modifier = Modifier.height(if(appScope.itemCount == 1) (displayHeight / 5).dp else 0.dp))
                }

                appScope.content(listState)
            }
        }
    }

    AutoScrollAwayTimeText(
        state = listState,
        scrollOffset = paddingTop - 10.dp,
    ) { mod ->
        if(timeText != null) timeText(mod)
        else TimeText(mod)
    }
}

class ListItemScope(
    private val scope: ScalingLazyListScope
) {
    var itemCount = 0
        private set

    fun item(
        key: Any? = null,
        content: @Composable ScalingLazyListItemScope.() -> Unit
    ) {
        itemCount++

        scope.item(
            key = key,
            content = content
        )
    }

    fun <T> items(
        items: List<T>,
        key: ((item: T) -> Any)? = null,
        itemContent: @Composable ScalingLazyListItemScope.(item: T) -> Unit,
    ) {
        itemCount += items.size

        scope.items(
            items = items,
            key = key,
            itemContent = itemContent
        )
    }
    fun items(
        count: Int,
        key: ((index: Int) -> Any)? = null,
        itemContent: @Composable ScalingLazyListItemScope.(index: Int) -> Unit,
    ) {
        itemCount += count

        scope.items(
            count = count,
            key = key,
            itemContent = itemContent
        )
    }

}
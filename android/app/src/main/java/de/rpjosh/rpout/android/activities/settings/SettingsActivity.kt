package de.rpjosh.rpout.android.activities.settings

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.TabRowDefaults
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.shared.controller.UserController
import kotlinx.coroutines.launch

class SettingsActivity : ComponentActivity() {

    private lateinit var userController: UserController
    private lateinit var tabs: Array<Tab>

    companion object {
        const val INTENT_INITIAL_TAB = "INTENT_KEY_INITIAL_TAB"
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Singleton is already loaded (Settings activity is only called through MainActivity after app startup)
        Singleton.appController.activityStarted(applicationContext, this)
        userController = Singleton.getAppSec(true).injection.inject(UserController::class.java, null ,false)

        // Get tabs
        val inj = Singleton.appController.injection
        tabs = arrayOf<Tab>(
            inj.inject(GeneralTab::class.java, null, false),
            inj.inject(UserTab::class.java, null, false),
        )
        tabs.forEach { it.activity = this; it.loadData() }

        val initialTab = when(intent.getStringExtra(INTENT_INITIAL_TAB)) {
            "user" -> 1
            else -> 0
        }

        // Configure elements
        enableEdgeToEdge()
        setContent {
            RPoutTheme {
                Settings(
                    onGoBack = { finish() },
                    tabs = tabs,
                    initialTab = initialTab
                )
            }
        }
    }

    override fun onPause() {
        Singleton.getApp()?.activityPaused(this)
        super.onPause()
        tabs.forEach { it.onPause() }
    }
    override fun onDestroy() {
        Singleton.appController.activityDestroyed(this)
        super.onDestroy()
        tabs.forEach { it.onDestroy() }
    }

    override fun onStart() {
        Singleton.appController.activityStarted(applicationContext, this)
        super.onStart()
    }
    override fun onResume() {
        Singleton.appController.activityStarted(applicationContext, this)
        super.onResume()
    }

}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun Settings(onGoBack: () -> Unit, tabs: Array<Tab>, initialTab: Int = 0) {
    val colors = RPoutTheme.colors

    // Current tab positions
    val scope = rememberCoroutineScope()
    val pagerState = rememberPagerState(initialPage = initialTab, pageCount = { tabs.size })
    val selectedTabIndex = remember { derivedStateOf { pagerState.currentPage } }

    Column(modifier = Modifier
        .fillMaxSize()
        .background(colors.backgroundDarker),
    ) {
        // Don't draw onto status bar
        with(LocalDensity.current) {
            val paddingTop = WindowInsets.statusBars.getTop(LocalDensity.current) - 12.dp.toPx()
            Spacer(Modifier.height(if (paddingTop < 0) 0.dp else paddingTop.toDp()))
        }

        Scaffold(
            topBar = {
                TopAppBar(
                    colors = TopAppBarDefaults.topAppBarColors(
                        containerColor = colors.backgroundDarker,
                        titleContentColor = colors.text,
                    ),
                    windowInsets = WindowInsets(
                        top = 0.dp,
                        bottom = 0.dp,
                    ),
                    title = { Text("Einstellungen") },
                    navigationIcon = {
                        IconButton(onClick = onGoBack) {
                            Icon(
                                painter = painterResource(R.drawable.arrow_back),
                                contentDescription = "Go back to main view"
                            )
                        }
                    },
                )
            }
        ) { innerPadding ->
            Column(modifier = Modifier.padding(top = innerPadding.calculateTopPadding() - 12.dp)) {
                TabRow(
                    selectedTabIndex = selectedTabIndex.value,
                    containerColor = colors.backgroundDarker,
                    modifier = Modifier.fillMaxWidth(),
                    contentColor = colors.secondary,
                    divider = {},
                    indicator = { tabPositions ->
                        if (selectedTabIndex.value < tabPositions.size) {
                            TabRowDefaults.SecondaryIndicator(
                                modifier = Modifier.tabIndicatorOffset(tabPositions[selectedTabIndex.value]),
                                color = colors.secondary,
                            )
                        }
                    },
                ) {
                    tabs.forEachIndexed { index, tab ->
                        Tab(
                            selected = selectedTabIndex.value == index,
                            onClick = {
                                scope.launch {
                                    pagerState.animateScrollToPage(index)
                                }
                            },
                            text = { Text(
                                text = tab.getLabel(),
                                fontSize = 15.sp,
                                color = if (selectedTabIndex.value == index) colors.secondary else colors.text
                            ) },
                            modifier = Modifier.fillMaxWidth(),
                        )
                    }
                }
                HorizontalPager(
                    modifier = Modifier.fillMaxSize().weight(1f),
                    state = pagerState
                ) { page ->
                    Box(modifier = Modifier.fillMaxSize().weight(1f).background(colors.defaultBackground).padding(8.dp)) {
                        tabs[page].Content()
                    }
                }
            }
        }
    }
}

@Composable
fun SectionText(txt: String, noExtraPaddingTop: Boolean = false) {
    val colors = RPoutTheme.colors

    Box(
        modifier = Modifier.fillMaxWidth()
            .padding(4.dp, top = if (noExtraPaddingTop) 4.dp else 10.dp)
            .border(1.dp, colors.backgroundLighter, shape = RoundedCornerShape(5.dp))
            .background(colors.backgroundDarker, shape = RoundedCornerShape(5.dp))
    ) {
        Text(
            text = txt,
            textAlign = TextAlign.Center,
            modifier = Modifier.fillMaxWidth().padding(3.dp),
            color = colors.text,
            fontWeight = FontWeight.Bold
        )
    }

}

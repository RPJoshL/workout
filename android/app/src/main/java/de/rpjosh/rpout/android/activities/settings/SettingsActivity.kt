package de.rpjosh.rpout.android.activities.settings

import android.content.Intent
import android.os.Build
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.Image
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.WindowInsets
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBars
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.layout.windowInsetsPadding
import androidx.compose.foundation.pager.HorizontalPager
import androidx.compose.foundation.pager.rememberPagerState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.automirrored.filled.ArrowBack
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.ScrollableTabRow
import androidx.compose.material3.Tab
import androidx.compose.material3.TabRow
import androidx.compose.material3.TabRowDefaults
import androidx.compose.material3.TabRowDefaults.tabIndicatorOffset
import androidx.compose.material3.Text
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.TopAppBarDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.derivedStateOf
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.ExperimentalComposeUiApi
import androidx.compose.ui.Modifier
import androidx.compose.ui.autofill.AutofillType
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalConfiguration
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.res.dimensionResource
import androidx.compose.ui.res.painterResource
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.tooling.preview.Preview
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import de.rpjosh.rpout.android.R
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.main.MainActivity
import de.rpjosh.rpout.android.activities.theme.OutlinedTextField
import de.rpjosh.rpout.android.activities.theme.OutlinedTextFieldWithLabel
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.activities.theme.autofill
import de.rpjosh.rpout.android.activities.theme.backgroundDarker
import de.rpjosh.rpout.android.activities.theme.backgroundLighter
import de.rpjosh.rpout.android.activities.theme.defaultBackground
import de.rpjosh.rpout.android.activities.theme.secondary
import de.rpjosh.rpout.android.activities.theme.text
import de.rpjosh.rpout.android.activities.theme.textBlue
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.models.ApiKey
import kotlinx.coroutines.launch
import java.time.LocalDateTime

class SettingsActivity : ComponentActivity() {

    private lateinit var userController: UserController
    private lateinit var tabs: Array<Tab>

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

        // Configure elements
        enableEdgeToEdge()
        setContent {
            RPoutTheme {
                Settings(
                    onGoBack = { finish() },
                    tabs = tabs
                )
            }
        }
    }

    private fun doLogin(username: String, password: String, serverURL: String) {
        val key = ApiKey(
            alias = Build.MODEL,
            validUntil = LocalDateTime.now().plusYears(2),
        )

        Thread {
            if (userController.getApiKey(serverURL, username, password, key)) {
                startActivity(Intent(this, MainActivity::class.java))
            }
        }.start()
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
fun Settings(onGoBack: () -> Unit, tabs: Array<Tab>) {

    // Current tab positions
    val scope = rememberCoroutineScope()
    val pagerState = rememberPagerState(pageCount = { tabs.size })
    val selectedTabIndex = remember { derivedStateOf { pagerState.currentPage } }

    Column(modifier = Modifier
        .fillMaxSize()
        .background(backgroundDarker),
    ) {
        // Don't draw onto status bar
        with(LocalDensity.current) {
            val paddingTop = WindowInsets.statusBars.getTop(LocalDensity.current) - 12.dp.toPx()
            Spacer(Modifier.height(if (paddingTop < 0) 0.dp else paddingTop.toDp()))
        }

        Scaffold(
            topBar = {
                TopAppBar(
                    colors = TopAppBarDefaults.centerAlignedTopAppBarColors(
                        containerColor = backgroundDarker,
                        titleContentColor = text,
                    ),
                    windowInsets = WindowInsets(
                        top = 0.dp,
                        bottom = 0.dp,
                    ),
                    title = { Text("Einstellungen") },
                    navigationIcon = {
                        IconButton(onClick = onGoBack) {
                            Icon(
                                imageVector = Icons.AutoMirrored.Filled.ArrowBack,
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
                    containerColor = backgroundDarker,
                    modifier = Modifier.fillMaxWidth(),
                    contentColor = secondary,
                    divider = {},
                    indicator = { tabPositions ->
                        if (selectedTabIndex.value < tabPositions.size) {
                            TabRowDefaults.SecondaryIndicator(
                                modifier = Modifier.tabIndicatorOffset(tabPositions[selectedTabIndex.value]),
                                color = secondary,
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
                                color = if (selectedTabIndex.value == index) secondary else text
                            ) },
                            modifier = Modifier.fillMaxWidth(),
                        )
                    }
                }
                HorizontalPager(
                    modifier = Modifier.fillMaxSize().weight(1f),
                    state = pagerState
                ) { page ->
                    Box(modifier = Modifier.fillMaxSize().weight(1f).background(defaultBackground).padding(8.dp)) {
                        tabs[page].Content()
                    }
                }
            }
        }
    }
}

@Composable
fun SectionText(txt: String, noExtraPaddingTop: Boolean = false) {
    Box(
        modifier = Modifier.fillMaxWidth()
            .padding(4.dp, top = if (noExtraPaddingTop) 4.dp else 10.dp)
            .border(1.dp, backgroundLighter, shape = RoundedCornerShape(5.dp))
            .background(backgroundDarker, shape = RoundedCornerShape(5.dp))
    ) {
        Text(
            text = txt,
            textAlign = TextAlign.Center,
            modifier = Modifier.fillMaxWidth().padding(3.dp),
            color = text,
            fontWeight = FontWeight.Bold
        )
    }

}

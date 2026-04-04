package de.rpjosh.rpout.android.activities.main

import android.annotation.SuppressLint
import android.content.Context
import android.content.Intent
import android.content.res.Configuration
import android.graphics.Bitmap
import android.graphics.Color
import android.util.Log
import android.view.ViewGroup
import android.webkit.CookieManager
import android.webkit.JavascriptInterface
import android.webkit.WebResourceError
import android.webkit.WebResourceRequest
import android.webkit.WebResourceResponse
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.compose.BackHandler
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.runtime.Composable
import androidx.compose.runtime.mutableStateOf
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.toArgb
import androidx.compose.ui.viewinterop.AndroidView
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.activities.login.LoginActivity
import de.rpjosh.rpout.android.activities.settings.SettingsActivity
import de.rpjosh.rpout.android.activities.theme.RPoutTheme
import de.rpjosh.rpout.android.helper.VersionHelper
import de.rpjosh.rpout.android.shared.controller.UserController
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.models.User

@SuppressLint("SetJavaScriptEnabled")
class Webview(
    private val context: Context,
    private val user: User,
    private val onFinish: () -> Unit = {},
    private val onConnectionError: (isError: Boolean) -> Unit = {},
    private val onPageStarted: () -> Unit = {},
    private val onPageFinished: () -> Unit = {}
): WebViewClient() {

    val webView = WebView(context)
    val cookieManager: CookieManager = CookieManager.getInstance()
    val canGoBack = mutableStateOf(false)
    val isLoading = mutableStateOf(true)

    @Inject private lateinit var userController: UserController

    init {
        Singleton.appController.injection.inject(Webview::class.java, null, false, this)

        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            useWideViewPort = true
        }

        webView.setBackgroundColor(Color.TRANSPARENT)

        webView.addJavascriptInterface(this, "Android")
        webView.webViewClient = this

        val isDarkMode = (context.resources.configuration.uiMode and Configuration.UI_MODE_NIGHT_MASK) == Configuration.UI_MODE_NIGHT_YES

        cookieManager.setAcceptCookie(true)
        setCookie("AndroidClientVersion", VersionHelper.getVersionName())
        setCookie("AndroidClientTheme", if (isDarkMode) "dark" else "light")
        setCookie("WorkoutCookie", user.apikey)
        cookieManager.flush()

        loadUrlWithHeaders(user.serverUrl)
    }

    private fun setCookie(key: String, value: String) {
        cookieManager.setCookie(
            user.serverUrl,
            "$key=$value; path=/; HttpOnly; SameSite=Lax"
        )
    }

    private fun loadUrlWithHeaders(url: String) {
        val headers = mutableMapOf<String, String>()
        headers["Client-Version"] = VersionHelper.getVersionName()
        headers["Client-Webview"] = "true"
        headers["Time-Zone"] = java.time.ZoneId.systemDefault().id

        webView.loadUrl(url, headers)
    }

    override fun onReceivedError(view: WebView?, request: WebResourceRequest, error: WebResourceError) {
        onConnectionError(true)
    }

    override fun onPageStarted(view: WebView?, url: String?, favicon: Bitmap?) {
        super.onPageStarted(view, url, favicon)
        isLoading.value = true
        onPageStarted()
    }

    override fun onPageFinished(view: WebView?, url: String?) {
        super.onPageFinished(view, url)
        isLoading.value = false
        onPageFinished()
    }

    fun reload() {
        onConnectionError(false)
        webView.reload()
    }

    override fun doUpdateVisitedHistory(view: WebView?, url: String?, isReload: Boolean) {
        super.doUpdateVisitedHistory(view, url, isReload)
        canGoBack.value = view?.canGoBack() ?: false
    }

    override fun shouldOverrideUrlLoading(view: WebView?, request: WebResourceRequest?): Boolean {
        val url = request?.url?.toString() ?: return false

        val isInternal = url.startsWith(user.serverUrl)
        Log.d("RPout-Logger", "Navigation to URL '$url'. Internal = $isInternal")
        if (isInternal) {
            loadUrlWithHeaders(url)
            return true
        }

        // All external links should not be opened in the webview
        return true
    }

    override fun onReceivedHttpError(
        view: WebView?,
        request: WebResourceRequest?,
        errorResponse: WebResourceResponse?
    ) {
        // Resources to ignore errors
        val ignoredPaths = mutableListOf("favicon.ico")
        val isIgnored = ignoredPaths.find{ request?.url?.path?.endsWith(it) ?: false }
        if (isIgnored != null) {
            return
        }

        Log.d("RPout-Logger", "Got status code ${errorResponse?.statusCode} for URL ${request?.url}")

        if (errorResponse?.statusCode == 401) {
            onConnectionError(true)
            context.startActivity(Intent(context, SettingsActivity::class.java).apply {
                putExtra(SettingsActivity.INTENT_INITIAL_TAB, "user")
            })
        }
    }

    @JavascriptInterface
    fun openNativeSettings() {
        context.startActivity(Intent(context, SettingsActivity::class.java))
    }

    @JavascriptInterface
    fun logout() {
        Thread {
            userController.logout {
                context.startActivity(Intent(context, LoginActivity::class.java))
                onFinish()
            }
        }.start()
    }

}

@Composable
fun WebViewScreen(webView: Webview) {
    val backgroundColor = RPoutTheme.colors.defaultBackground

    BackHandler(enabled = webView.canGoBack.value) {
        webView.webView.goBack()
    }

    Box(modifier = Modifier.fillMaxSize().background(backgroundColor)) {
        AndroidView(
            modifier = Modifier.fillMaxSize(),
            factory = {
                webView.webView.apply {
                    setBackgroundColor(backgroundColor.toArgb())

                    layoutParams = ViewGroup.LayoutParams(
                        ViewGroup.LayoutParams.MATCH_PARENT,
                        ViewGroup.LayoutParams.MATCH_PARENT
                    )
                }
            },
            update = {
                it.setBackgroundColor(backgroundColor.toArgb())
            }
        )

        AnimatedVisibility(
            visible = webView.isLoading.value,
            enter = fadeIn(),
            exit = fadeOut(animationSpec = tween(400)),
            modifier = Modifier.align(Alignment.Center)
        ) {
            CircularProgressIndicator(
                color = RPoutTheme.colors.secondary
            )
        }
    }
}

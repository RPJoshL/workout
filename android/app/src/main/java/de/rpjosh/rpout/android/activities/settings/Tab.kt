package de.rpjosh.rpout.android.activities.settings

import android.app.Activity
import android.content.Context
import androidx.compose.runtime.Composable

/**  Represents a single tab in the settings pager */
interface Tab {

    /** Label to display on the tab pager */
    fun getLabel(): String

    /** Load data needed for the UI */
    fun loadData()

    fun onPause()
    fun onDestroy()

    /** Set context of main activity */
    var activity: Activity

    /** Content to render in tab */
    @Composable
    fun Content()
}
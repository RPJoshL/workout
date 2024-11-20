package de.rpjosh.rpout.android.shared.customizations

import android.content.Context
import android.content.Intent
import android.net.Uri

class RPdbBroadcaster(val context: Context) {

    companion object {
        const val KEY_NAME = "requestor_name"
        const val KEY_TIMEOUT = "requestor_timeout"
        const val KEY_LOCAL_CONNECTIVITY = "requestor_local_connectivity"

        const val ACTION_PREFIX = "de.rpjosh.rpdb.testandroid.customizations.connectivity.RemoteManager."
    }


    fun requestConnectivity(name: String, timeoutSeconds: Int) {
        sendIntent("REQUEST") {
            it.putExtra(KEY_NAME, name)
            it.putExtra(KEY_TIMEOUT, timeoutSeconds)
        }
    }
    fun dropConnectivity(name: String) {
        sendIntent("DROP") {
            it.putExtra(KEY_NAME, name)
        }
    }

    fun startWorkout() {
        sendIntent("WORKOUT_START") {
            it.putExtra(KEY_NAME, "WORKOUT")
        }
    }
    fun stopWorkout() {
        sendIntent("WORKOUT_END") {
            it.putExtra(KEY_NAME, "WORKOUT")
        }
    }

    private fun sendIntent(action: String, setIntent: (intent: Intent) -> Unit) {
        val intent = Intent(ACTION_PREFIX + action).apply{
            setClassName(
                "de.rpjosh.rpdb.testandroid",
                "de.rpjosh.rpdb.testandroid.customizations.connectivity.RemoteManager"
            )
            setIntent(this)
        }

        context.sendBroadcast(intent)
    }


}
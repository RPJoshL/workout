package de.rpjosh.rpout.android.shared.customizations

import android.content.Context
import android.content.Intent
import android.net.Uri

class RPdbBroadcaster(val context: Context) {

    companion object {
        const val KEY_NAME = "requestor_name"
        const val KEY_TIMEOUT = "requestor_timeout"
        const val KEY_LOCAL_CONNECTIVITY = "requestor_local_connectivity"
        const val KEY_FORWARD_REQUEST = "requestor_forward_request"

        const val ACTION_PREFIX_TEST = "de.rpjosh.rpdb.testandroid.customizations.connectivity.RemoteManager."
        const val ACTION_PREFIX_PROD = "de.rpjosh.rpdb.android.customizations.connectivity.RemoteManager."
    }


    fun requestConnectivity(name: String, timeoutSeconds: Int, forward: Boolean = false) {
        sendIntent("REQUEST") {
            it.putExtra(KEY_NAME, name)
            it.putExtra(KEY_TIMEOUT, timeoutSeconds)
            it.putExtra(KEY_FORWARD_REQUEST, forward)
        }
    }
    fun dropConnectivity(name: String, forward: Boolean = false) {
        sendIntent("DROP") {
            it.putExtra(KEY_NAME, name)
            it.putExtra(KEY_FORWARD_REQUEST, forward)
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
        // Send message both to test and prod version of RPdb
        sendIntentWithPrefix(ACTION_PREFIX_PROD, action, setIntent = setIntent)
        sendIntentWithPrefix(ACTION_PREFIX_TEST, action, setIntent = setIntent)
    }

    private fun sendIntentWithPrefix(prefix: String, action: String, setIntent: (intent: Intent) -> Unit) {
        val intent = Intent(prefix + action).apply{
            val packageName = if(prefix.contains("testandroid")) "de.rpjosh.rpdb.testandroid" else "de.rpjosh.rpdb.android"
            setClassName(
                packageName,
                "${packageName}.customizations.connectivity.RemoteManager"
            )
            setIntent(this)
        }

        context.sendBroadcast(intent)
    }


}
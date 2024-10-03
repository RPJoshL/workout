package de.rpjosh.rpout.android.services

import android.content.Context
import android.content.Intent
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import android.net.NetworkRequest
import androidx.core.content.ContextCompat
import androidx.work.Worker
import androidx.work.WorkerParameters
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.controller.MetricController
import de.rpjosh.rpout.android.shared.controller.WorkoutController

/**
 * Uploader uploads and synchronizes metrics like steps and workout data
 */
public class Uploader(appContext: Context, workerParams: WorkerParameters): Worker(appContext, workerParams) {

    companion object {
        const val TAG_UPLOADER = "UPLOADER"
        // One time work requests that are synchronized immediately when there is a network connection available
        const val TAG_UPLOADER_PRIO = "UPLOADER_PRIO"
    }

    override fun doWork(): Result {
        // We should have a app reference
        val app = Singleton.getApp() ?: return Result.failure()

        // Inject controller
        val metricController = app.injection.inject(MetricController::class.java, null,  false)
        val workoutController = app.injection.inject(WorkoutController::class.java, null, false)

        // Synchronize everything
        var success = true
        if (!metricController.synchronizeSteps()) success = false
        if (!workoutController.synchronizeWorkouts()) success = false

        return if(success) Result.success() else if (TAG_UPLOADER_PRIO in tags) Result.retry() else Result.failure()
    }

    /**
     * Registers an callback whenever the connection state of the internet connection changed
     */
    private fun registerNetworkCallback() {
        val connectivityManager = RPout.getAppContext().getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

        val networkRequest = NetworkRequest.Builder()
            .addCapability(NetworkCapabilities.NET_CAPABILITY_INTERNET)
            .addTransportType(NetworkCapabilities.TRANSPORT_WIFI)
            .addTransportType(NetworkCapabilities.TRANSPORT_CELLULAR)
            .addTransportType(NetworkCapabilities.TRANSPORT_BLUETOOTH)
            .build()

        val networkCallback = object : ConnectivityManager.NetworkCallback() {

            override fun onAvailable(network: Network) {
                super.onAvailable(network)

                if (doWork() == Result.success()) {
                    // Work successfully -> unregister
                    connectivityManager.unregisterNetworkCallback(this)
                }
            }

            override fun onCapabilitiesChanged(network: Network, networkCapabilities: NetworkCapabilities) {
                super.onCapabilitiesChanged(network, networkCapabilities)
            }

            override fun onLost(network: Network) {
                super.onLost(network)
            }
        }

        connectivityManager.registerNetworkCallback(networkRequest, networkCallback as ConnectivityManager.NetworkCallback)
    }

}
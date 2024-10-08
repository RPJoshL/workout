package de.rpjosh.rpout.android.services

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.SystemUtilsInterface
import de.rpjosh.rpout.android.RPout

class WearUtils: SystemUtilsInterface {

    @Inject( parameters = ["AndroidUtils"])
    private var logger: Logger? = null;

    override fun checkInternetConnection(localConnectivity: Boolean, url: String): Boolean {
        return if (localConnectivity) isLocaleNetworkAvailable()
        else isNetworkAvailable()
    }

    private fun isNetworkAvailable(): Boolean {
        val cm = RPout.getAppContext().getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val network = cm.activeNetwork ?: return false
        return cm.getNetworkCapabilities(network)?.run {
            hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR)
                    || hasTransport(NetworkCapabilities.TRANSPORT_WIFI)
                    || hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET)
                    || hasTransport(NetworkCapabilities.TRANSPORT_BLUETOOTH)
        } ?: return false
    }

    private fun isLocaleNetworkAvailable(): Boolean {
        val cm = RPout.getAppContext().getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val network = cm.activeNetwork ?: return false
        return cm.getNetworkCapabilities(network)?.run {
            hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR)
                    || hasTransport(NetworkCapabilities.TRANSPORT_WIFI)
        } ?: return false
    }

}
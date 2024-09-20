package de.rpjosh.rpout.android.services

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities
import android.os.Build
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.shared.inject.Inject
import de.rpjosh.rpout.android.shared.services.Logger
import de.rpjosh.rpout.android.shared.services.SystemUtilsInterface

class AndroidUtils: SystemUtilsInterface {

    @Inject( parameters = ["AndroidUtils"])
    private var logger: Logger? = null;

    override fun checkInternetConnection(localConnectivity: Boolean, url: String): Boolean {
        return isNetworkAvailable()
    }

    private fun isNetworkAvailable(): Boolean {
        val cm = RPout.getAppContext().getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
        val network = cm.activeNetwork ?: return false
        return cm.getNetworkCapabilities(network)?.run {
            hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR)
                    || hasTransport(NetworkCapabilities.TRANSPORT_WIFI)
                    || hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET)
        } ?: return false
    }
}
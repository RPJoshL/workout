package de.rpjosh.rpout.android.shared.helper

import android.content.Context
import android.net.ConnectivityManager
import android.net.NetworkCapabilities

class Helper {

    companion object {

        fun isNetworkAvailable(context: Context): Boolean {
            val cm = context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
            val network = cm.activeNetwork ?: return false
            return cm.getNetworkCapabilities(network)?.run {
                hasTransport(NetworkCapabilities.TRANSPORT_CELLULAR)
                        || hasTransport(NetworkCapabilities.TRANSPORT_WIFI)
                        || hasTransport(NetworkCapabilities.TRANSPORT_ETHERNET)
                        || hasTransport(NetworkCapabilities.TRANSPORT_BLUETOOTH)
            } ?: return false
        }

    }

}
package de.rpjosh.rpout.android.shared.services

public interface SystemUtilsInterface {
    /**
     * Checks whether this device has a working internet connectivity
     *
     * @param localeConnectivity        Whether the connectivity has to be local (no bluetooth
     *                                  tethering to the phone for smartwatch)
     * @param  url                      Server URL to test the connectivity against
     */
    fun checkInternetConnection(localeConnectivity: Boolean, url: String): Boolean
}
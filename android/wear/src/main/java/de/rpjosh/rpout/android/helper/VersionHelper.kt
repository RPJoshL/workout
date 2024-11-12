package de.rpjosh.rpout.android.helper

import android.content.pm.PackageInfo
import android.content.pm.PackageManager
import de.rpjosh.rpout.android.RPout
import de.rpjosh.rpout.android.Singleton

class VersionHelper {

    companion object {

        /**
         * Returns the current version name like '1.0.3' of the app
         */
        fun getVersionName(): String {
            var versionName = "0.0.0"
            try {
                val pInfo: PackageInfo = RPout.getAppContext().packageManager.getPackageInfo(RPout.getAppContext().packageName, 0)
                versionName = pInfo.versionName ?: ""
            } catch (e: PackageManager.NameNotFoundException) {
                Singleton.appController.sharedLogger.log("e", "Failed to get the version name of the app", e)
            }

            return versionName
        }
    }

}
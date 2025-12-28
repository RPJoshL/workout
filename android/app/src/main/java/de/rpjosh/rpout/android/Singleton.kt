package de.rpjosh.rpout.android

import android.util.Log
import de.rpjosh.rpout.android.shared.services.MessageType
import java.util.concurrent.atomic.AtomicInteger

class Singleton {

    companion object {

        const val TAG = "RPout-Logger"

        lateinit var appController: AndroidAppController
            private set
        val notificationId = AtomicInteger(1)

        private val onAppLoaded = mutableListOf<AppLoadable>()
        private val onWearMessageReceived = mutableListOf<WearMessageReceiver>()

        /**
         * Initializes the application Controller
         *
         * @return          if the application controller was already initialized
         */
        fun app(): Boolean {
            if (!::appController.isInitialized) {
                appController = AndroidAppController()
                return false
            }
            return true
        }

        fun getApp(): AndroidAppController? {
            return if (!::appController.isInitialized) null
            else appController
        }

        fun setApp(appController: AndroidAppController) {
            this.appController = appController
            synchronized(onAppLoaded) {
                onAppLoaded.forEach{ Thread { it.appLoaded() }.start() }
                onAppLoaded.clear()
            }
        }

        fun setAppSilent(appController: AndroidAppController) {
            this.appController = appController
        }

        /**
         * Returns the application controller securely after it has been initialized.
         * In some rare cases the controller is not initialized in time. So this method does block
         * until the controller has been initialized (at a max rate of 800ms)
         */
        fun getAppSec(retry: Boolean = true): AndroidAppController {
            // Application controller already initialized
            if (::appController.isInitialized) return appController

            var i = 0;
            while (i++ < 20) {
                Thread.sleep(50)
                if (::appController.isInitialized) return appController
            }

            Log.e(TAG, "Application controller got not initialized. This should not happen.")
            if (retry) {
                Log.e(TAG, "The controller will be initialized because the app was started from an unknown context. This could lead to problems....")
                app()
                // We do log the attempt. May this be a security risk when we receive an intent where we do not expect an entrypoint?
                appController.sharedLogger.log("e", "Because the app was started from an unknown context the application controller was initialized in fallback mode. There may be some problems until the app will be restarted...")

                // Because we don't known the invocation context we don't do anything
                return appController
            }

            Thread.sleep(50)
            // This throws an exception and the app does crash
            return appController
        }


        /**
         * Adds to the given number a leading zero (6 -> 06)
         *
         * @param number        number to add the leading zero
         *
         * @return              number as string with the leading zero
         */
        fun addLeadingZero(number: Int): String {
            return if (number in 0..9)  ("0$number")
            else if (number in -9..0)   ("-0${number*-1}")
            else                        number.toString()
        }

        fun registerOnAppLoaded(loadable: AppLoadable): Boolean {
            if (::appController.isInitialized) {
                loadable.appLoaded()
                return true
            } else {
                synchronized(onAppLoaded) {
                    onAppLoaded.add(loadable)
                }
                return false
            }
        }

        fun deRegisterOnAppLoaded(loadable: AppLoadable) {
            if (::appController.isInitialized) return

            synchronized(onAppLoaded) {
                onAppLoaded.remove(loadable)
            }
        }

        fun registerOnWearMessageReceived(receiver: WearMessageReceiver) {
            synchronized(onWearMessageReceived) {
                onWearMessageReceived.add(receiver)
            }
        }
        fun deRegisterOnWearMessageReceived(receiver: WearMessageReceiver) {
            synchronized(onWearMessageReceived) {
                onWearMessageReceived.remove(receiver)
            }
        }

        fun sendMessageToWearMessageReceiver(type: MessageType, data: String) {
            synchronized(onWearMessageReceived) {
                onWearMessageReceived.forEach {
                    it.onWearMessageReceived(type, data)
                }
            }
        }

    }
}
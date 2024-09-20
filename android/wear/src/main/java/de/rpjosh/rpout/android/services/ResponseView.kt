package de.rpjosh.rpout.android.services

import android.app.Activity
import android.content.Context
import android.os.VibrationEffect
import android.os.VibratorManager
import android.util.Log
import android.widget.Toast
import de.rpjosh.rpout.android.shared.services.ResponseViewInterface
import de.rpjosh.rpout.android.RPout
import java.time.LocalDateTime
import java.time.temporal.ChronoUnit
import java.util.concurrent.atomic.AtomicInteger

class ResponseView: ResponseViewInterface {

    companion object {
        val messageId = AtomicInteger(0)
        var lastMessage = ""
        var lastMessageTime: LocalDateTime = LocalDateTime.now()
        var lastMessageStatic = ""
    }
    // time before the background of the action bar is been reset
    val TIME_TO_WAIT = 2400L

    // Activity context that is currently displayed
    @Volatile var activityContext: Context? = null
    @Volatile var activity: Activity? = null
    private val activityLock = Object()

    /** Sets the activity for displaying the errors and success in the activity */
    fun setActivity(context: Context?, activity: Activity?) {
        synchronized(activityLock) {
            this.activityContext = context
            this.activity = activity
        }
    }

    /**
     * Removes the currently used toolbar
     */
    fun removeActivity(activity: Activity) {
        synchronized(activityLock) {
            if (this.activity == activity) {
                this.activity = null
                this.activityContext = null
            }
        }
    }

    @Synchronized
    override fun displayError(message: String?) {
        // When the message to print out equals the static message, don't do anything
        if (message == lastStaticMessage) return

        Log.d("RPdb", message ?: "")
        displayOnToolbar(message ?: "")
    }

    @Synchronized
    override fun displaySuccess(message: String?) {
        if (!printOut(message ?: "")) return

        Log.d("RPdb", message ?: "")
        displayOnToolbar(message ?: "")
    }

    @Synchronized
    override fun displayStatic(message: String?) {
        val msg: String = message ?: ""
        if (lastMessageStatic == message) return

        Log.d("RPdb", "Displaying static message $msg")
        lastMessageStatic = msg

        displayOnToolbar(message ?: "")
    }

    override fun getLastStaticMessage(): String {
        return lastMessageStatic
    }

    override fun resetStatic() {
        // Nothing to do here
        if (lastMessageStatic.isBlank()) return

        lastMessageStatic = ""
    }

    /**
     * Checks if the given message should be printed out
     */
    @Synchronized
    private fun printOut(message: String): Boolean {
        synchronized(lastMessage) {
            var rtc = true
            val timeDiff = ChronoUnit.MILLIS.between(lastMessageTime, LocalDateTime.now());

            if (lastMessage == message && timeDiff <= TIME_TO_WAIT - 600) {
                rtc = false
            } else if (activity != null) {
                lastMessageTime = LocalDateTime.now()
                lastMessage = message
            }

            return rtc
        }
    }

    /**
     * Displays the given message to the user
     */
    @Synchronized
    private fun displayOnToolbar(message: String) {
        synchronized(activityLock) {
            if (activity == null || activityContext == null) {
                return
            }

            // Display a simple toast to provide the message to the user
            if (message.isNotEmpty())
                activity?.runOnUiThread {
                    Toast.makeText(activityContext, message, Toast.LENGTH_SHORT).show()
                }
        }
    }

    /**
     * Vibrates the device for the given amount of milliseconds.
     *
     * @param milliseconds         How long the vibration effect should last
     * @param amplitude            How "heavy" the vibration should be (0 = low, 255 = high)
     */
    public fun vibrate(milliseconds: Long, amplitude: Int = -1) {

        // Get vibrator
        val vib = RPout.getAppContext().getSystemService(Context.VIBRATOR_MANAGER_SERVICE) as VibratorManager

        // Vibrate for the given amount
        vib.defaultVibrator.vibrate(VibrationEffect.createOneShot(milliseconds, if (amplitude > 255) 255 else amplitude))
    }

}
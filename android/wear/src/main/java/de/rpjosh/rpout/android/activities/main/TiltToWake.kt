package de.rpjosh.rpout.android.activities.main

import android.content.Context
import android.hardware.Sensor
import android.hardware.SensorEvent
import android.hardware.SensorEventListener
import android.hardware.SensorManager
import android.provider.Settings
import de.rpjosh.rpout.android.Singleton
import de.rpjosh.rpout.android.shared.models.WorkoutType
import de.rpjosh.rpout.android.shared.services.Logger

/** Helper class to enable a tilt to wake like feature for an activity */
class TiltToWake(
    val context: Context,
    val type: WorkoutType,
    val onTilted: () -> Unit,
    val logger: Logger
): SensorEventListener {

    private val sensorManager = context.getSystemService(Context.SENSOR_SERVICE) as SensorManager
    private val tiltSensor: Sensor?
    private val lock = Object()
    private val tiltToWakeEnabled = isTiltToWakeEnabled()
    @Volatile private var isListenerRegistered = false

    init {
        synchronized(lock) {
            sensorManager.getDefaultSensor(26).let {
                if (it == null) {
                    logger.log("e", "Device does not have a tilt to wake sensor")
                    tiltSensor = null
                } else if (!it.isWakeUpSensor) {
                    logger.log("e", "Tilt to wake sensor is not a wakeup device!")
                    tiltSensor = null
                } else {
                    tiltSensor = it
                }
            }
        }

        logger.log("d", "Tilt to wake enabled = $tiltToWakeEnabled")
    }


    /** Registers the tilt to wake sensor as a wakeup sensor if tilt to wake is not enabled globally  */
    fun register() {
        synchronized(lock) {
            if (tiltSensor != null && !tiltToWakeEnabled) {
                val success = sensorManager.registerListener(this, tiltSensor, 3)
                isListenerRegistered = success
                logger.log("d", "Registered tilt to wake sensor: $success")
            }
        }
    }

    /** Removes the event listener from the tilt to wake sensor */
    fun deRegister() {
        synchronized(lock) {
            if (tiltSensor != null && isListenerRegistered) {
                sensorManager.unregisterListener(this, tiltSensor)
                isListenerRegistered = false
                logger.log("d", "Unregistered tilt to wake sensor")
            }
        }
    }

    private fun isTiltToWakeEnabled(): Boolean {
        return Settings.Global.getInt(context.contentResolver, "ambient_tilt_to_wake", 0) == 1
    }

    override fun onSensorChanged(event: SensorEvent?) {
        if (event?.sensor?.type == 26 && event.values[0] == 1f) {
            logger.log("d", "Detected a tilt to wake gesture")
            onTilted()
        }
    }

    override fun onAccuracyChanged(sensor: Sensor?, accuracy: Int) {}

}
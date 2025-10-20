package de.rpjosh.rpout.android.shared.workout

import android.annotation.SuppressLint
import android.content.Context
import android.location.Location
import com.google.android.gms.location.CurrentLocationRequest
import com.google.android.gms.location.Granularity
import com.google.android.gms.location.LocationServices
import com.google.android.gms.location.Priority
import com.google.android.gms.tasks.CancellationTokenSource
import de.rpjosh.rpout.android.shared.services.Logger

class WorkoutLocation(
    val context: Context,
    val logger: Logger
) {

    private val lock = Object()
    private var cancellationToken: CancellationTokenSource? = null
    private var fusedLocationClient = LocationServices.getFusedLocationProviderClient(context)

    @SuppressLint("MissingPermission")
    fun getCurrentLocation(onSuccess: (location: Location) -> Unit, onFailure: () -> Unit) {
        synchronized(lock) {
            if (cancellationToken != null) {
                logger.log("d", "Another location request is still in process. Not requesting location again")
                return
            }

            // Build request constraints
            val request = CurrentLocationRequest.Builder()
                .setPriority(Priority.PRIORITY_HIGH_ACCURACY)
                .setMaxUpdateAgeMillis(10 * 1000)
                .setGranularity(Granularity.GRANULARITY_FINE)
                .build()

            // Get new cancellation token
            cancellationToken = CancellationTokenSource()

            // Do request
            val task = fusedLocationClient.getCurrentLocation(request, cancellationToken!!.token)
            task.addOnSuccessListener {
                synchronized(lock) { cancellationToken = null }

                if (it == null) {
                    logger.log("d", "Could not obtain current location (got null as a result)")
                    onFailure()
                    return@addOnSuccessListener
                }

                logger.log("d", "Got result from location request: accuracy = ${if(it.hasAccuracy()) it.accuracy else "?"} | lat = ${it.latitude} | lon = ${it.longitude} | speed = ${if(it.hasSpeed()) (it.speed.toString() + "m/s") else "?"}")
                onSuccess(it)
            }
            task.addOnFailureListener {
                synchronized(lock) { cancellationToken = null }

                logger.log("d", it, "Could not obtain current location")
                onFailure()
            }
        }
    }

    /** Aborts any running location requests. You cannot use this instance anymore after calling this function */
    fun abort() {
        synchronized(lock) {
            if (cancellationToken != null) cancellationToken?.cancel()

            // We don't reset the cancellationToken variable
        }
    }

}
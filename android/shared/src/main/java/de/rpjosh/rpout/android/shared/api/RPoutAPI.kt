package de.rpjosh.rpout.android.shared.api

import de.rpjosh.rpout.android.shared.models.ApiKey
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.Pai
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.models.StepPostResult
import de.rpjosh.rpout.android.shared.models.WorkoutSummary
import de.rpjosh.rpout.android.shared.models.WorkoutType
import retrofit2.Call
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
import retrofit2.http.PUT
import retrofit2.http.Path

/** API endpoints that are accessible and used from this application */
interface RPoutAPI {

    // API key
    @POST("api-key")
    fun createApiKey(@Body body: ApiKey): Call<ApiKey>
    @GET("api-key/-1")
    fun getApiKey(): Call<ApiKey>
    @DELETE("api-key/{id}")
    fun deleteApiKey(@Path("id") id: Long): Call<String>

    // Step metric
    @POST("metric/steps")
    fun postSteps(@Body body: List<Step>): Call<StepPostResult>

    // PAI metrics
    @GET("metric/pai")
    fun getPaiValues(): Call<Pai>

    // Workout types
    @GET("workout/types")
    fun getWorkoutTypes(): Call<List<WorkoutType>>

    // Workouts
    @POST("workout")
    fun postWorkout(@Body workout: GpsWorkout): Call<WorkoutSummary>
    @PUT("workout/{id1}/merge/{id2}")
    fun mergeWorkouts(@Path("id1") id1: Long, @Path("id2") id2: Long): Call<String>

}
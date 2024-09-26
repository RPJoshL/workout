package de.rpjosh.rpout.android.shared.api

import de.rpjosh.rpout.android.shared.models.ApiKey
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.models.StepPostResult
import de.rpjosh.rpout.android.shared.models.WorkoutType
import retrofit2.Call
import retrofit2.http.Body
import retrofit2.http.DELETE
import retrofit2.http.GET
import retrofit2.http.POST
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

    // Workout types
    @GET("workout/types")
    fun getWorkoutTypes(): Call<List<WorkoutType>>
}
package de.rpjosh.rpout.android.shared.api;

import android.util.Log;

import androidx.annotation.NonNull;

import java.io.IOException;
import java.util.concurrent.TimeUnit;

import de.rpjosh.rpout.android.shared.config.GlobalConfiguration;
import de.rpjosh.rpout.android.shared.inject.Inject;
import de.rpjosh.rpout.android.shared.services.Tr;
import okhttp3.OkHttpClient;
import okhttp3.OkHttpClient.Builder;
import okhttp3.FormBody;
import okhttp3.HttpUrl;
import okhttp3.Interceptor;
import okhttp3.Request;
import okhttp3.Response;
import retrofit2.Retrofit;
import retrofit2.converter.gson.GsonConverterFactory;

public class APIClient {

    @Inject GlobalConfiguration globalConfig;

    public Retrofit.Builder getRetrofit(String url, String username, String password, boolean addApiKey, Integer timeout, String apiKey, boolean jsonResponse) {

        boolean addPasswordAuth = username != null && password != null;

        Builder okHttpBuilder = new OkHttpClient.Builder()
                .connectTimeout(4, TimeUnit.SECONDS)
                .readTimeout(timeout == null ? 20 : timeout + 3, TimeUnit.SECONDS)
                .writeTimeout(timeout == null ? 20 : 65, TimeUnit.SECONDS)
                .addInterceptor(new HeaderInterceptor("Android-Webview", "true"))
                .addInterceptor(new HeaderInterceptor("Accept-Language", Tr.getUsedLanguage().locale.getLanguage()))
                .addInterceptor(new HeaderInterceptor("Time-Zone", "UTC"));

        if (addApiKey && apiKey != null)    okHttpBuilder.addInterceptor(new HeaderInterceptor("X-Api-Key", apiKey));

        if (addPasswordAuth) {
            okHttpBuilder
                .addInterceptor(new HeaderInterceptor("Username", username))
                .addInterceptor(new HeaderInterceptor("Password", password));
        }

        OkHttpClient okHttpClient = okHttpBuilder.build();
        Retrofit.Builder retrofitBuilder = new Retrofit.Builder()
                .client(okHttpClient)
                .baseUrl(url);
        if (jsonResponse) retrofitBuilder.addConverterFactory(GsonConverterFactory.create());
        return retrofitBuilder;
    }

    /**
     * Gets a new service object with the default API-Key authentication
     *
     * @param <T>	Type of the response Object
     * @param ca	Class of the response Object
     *
     * @return		The RPout API service
     */
    public <T> RPoutAPI getRetrofitService(Class<T> ca) {
        return (RPoutAPI) getRetrofit(getUrl(), null, null, true, null, globalConfig.getUser().getApikey(), true).build().create(ca);
    }

    /**
     * Gets a new service object with the default API-Key authentication
     *
     * @param <T>	Type of the response Object
     * @param ca	Class of the response Object
     *
     * @return		The RPout API service
     */
    public <T> RPoutAPI getRetrofitService(Class<T> ca, boolean isJson) {
        return (RPoutAPI) getRetrofit(getUrl(), null, null, true, null, globalConfig.getUser().getApikey(), isJson).build().create(ca);
    }

    /**
     * Creates a new service object with authentication by username and password
     *
     * @param <T>       Type of the response Object
     * @param ca        Class of the response object
     * @param baseURL   Base URL of the API
     * @param username  Username to authenticate against the API
     * @param password  Password for the user
     *
     * @return          The RPout API service
     */
    public <T> RPoutAPI getRetrofitService(Class<T> ca, String baseURL, String username, String password) {
        if (!baseURL.endsWith("/api/v1")) {
            if (!baseURL.endsWith("/")) baseURL += "/api/v1";
            else                        baseURL += "api/v1";
        }
        if (!baseURL.endsWith("/")) baseURL += "/";

        return (RPoutAPI) getRetrofit(baseURL, username, password, false, null, null, true).build().create(ca);
    }

    private String getUrl() {
        String baseURL = globalConfig.getUser().getServerUrl();
        if (!baseURL.endsWith("/api/v1")) {
            if (!baseURL.endsWith("/")) baseURL += "/api/v1";
            else                        baseURL += "api/v1";
        }
        if (!baseURL.endsWith("/")) baseURL += "/";

        return baseURL;
    }

    /**
     * Add a header to the request
     */
    public static class HeaderInterceptor implements Interceptor {

        private final String value;
        private final String headerName;

        public HeaderInterceptor(String headerName, String value) {
            this.headerName = headerName;
            this.value = value;
        }

        @NonNull
        @Override
        public Response intercept(Chain chain) throws IOException {
            Request request = chain.request();
            Request authenticatedRequest = request.newBuilder()
                    .header(headerName, value)
                    .build();
            return chain.proceed(authenticatedRequest);
        }

    }

    /**
     * Add a field to the request
     */
    public static class FieldInterceptor implements Interceptor {

        private final String key;
        private final String value;

        public FieldInterceptor(String key, String value) {
            this.key = key;
            this.value = value;
        }

        @NonNull
        @Override
        public Response intercept(Chain chain) throws IOException {
            Request request = chain.request();

            request = modifyRequestBody(request);
            return chain.proceed(request);
        }

        /**
         * add new post fields
         */
        private Request modifyRequestBody(Request request) {

            if ("GET".equals(request.method())) {
                HttpUrl url = request.url().newBuilder().addQueryParameter(key,value).build();
                request = request.newBuilder().url(url).build();
            } else if ("POST".equals(request.method())) {

                if (request.body() instanceof FormBody formBody) {

                    FormBody.Builder bodyBuilder = new FormBody.Builder();

                    // Copy the original parameters first
                    for (int i = 0; i < formBody.size(); i++) {
                        bodyBuilder.addEncoded(formBody.encodedName(i), formBody.encodedValue(i));
                    }
                    formBody = bodyBuilder
                            .addEncoded(key, value)
                            .build();
                    request = request.newBuilder().post(formBody).build();
                }
            }

            return request;
        }

    }

}

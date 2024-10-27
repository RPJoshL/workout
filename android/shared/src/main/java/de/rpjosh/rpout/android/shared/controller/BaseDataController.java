package de.rpjosh.rpout.android.shared.controller;

import java.io.IOException;
import java.net.SocketException;
import java.net.UnknownHostException;
import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;

import de.rpjosh.rpout.android.shared.api.APIClient;
import de.rpjosh.rpout.android.shared.config.GlobalConfiguration;
import de.rpjosh.rpout.android.shared.exceptions.AuthenticationException;
import de.rpjosh.rpout.android.shared.exceptions.OfflineException;
import de.rpjosh.rpout.android.shared.exceptions.ServerException;
import de.rpjosh.rpout.android.shared.exceptions.UnknownServerException;
import de.rpjosh.rpout.android.shared.inject.Inject;
import de.rpjosh.rpout.android.shared.models.ErrorResponse;
import de.rpjosh.rpout.android.shared.services.Logger;
import de.rpjosh.rpout.android.shared.services.ResponseViewInterface;
import de.rpjosh.rpout.android.shared.services.SystemUtilsInterface;
import retrofit2.Call;
import retrofit2.Response;

public class BaseDataController {

    @Inject protected APIClient apiClient;
    @Inject protected ResponseViewInterface responseView;
    @Inject protected SystemUtilsInterface systemUtils;
    @Inject GlobalConfiguration globalConfig;
    @Inject(parameters = {"BaseDataController"}) protected Logger baseLogger;

    // 1 = Yes | 0 = No | 99 = Not known
    private static int isUserLoggedIn = 99;
    private static AuthenticationException lastException;
    private static LocalDateTime lastAuthCheck = LocalDateTime.now();

    /**
     * Checks if an Internet Connection and an user account is established
     *
     * @param  localeInternetConnection     Weather the internet connection must be established on the
     *                                      watch and cannot be used over smartphone (bluetooth tethering)
     */
    protected void ensureConnection(boolean localeInternetConnection) throws AuthenticationException, OfflineException {

        // Was a URL provided?
        if (globalConfig.getUser() == null) {
            throw new AuthenticationException(AuthenticationException.TYPE.NO_CREDENTIALS);
        }

        // check Internet connection
        boolean hasAConnection = systemUtils.checkInternetConnection(false, globalConfig.getUser().getServerUrl());
        if (!hasAConnection) throw new OfflineException();

        // Basic Auth | API | Username + Password
        // Reduce spam of failed Login attempts
        if (isUserLoggedIn == -1 || isUserLoggedIn == 0) {
            if (ChronoUnit.SECONDS.between(lastAuthCheck, LocalDateTime.now()) > 360) {		// 6 Minutes
                baseLogger.log("i", "Resetting lock for failed authentications: last check {0} minutes before",ChronoUnit.MINUTES.between(lastAuthCheck, LocalDateTime.now()) );
                isUserLoggedIn = 99;
                return;
            }

            if (lastException != null) {
                throw lastException;
            } else {
                if (isUserLoggedIn == 0)
                    throw new AuthenticationException(AuthenticationException.TYPE.NO_CREDENTIALS);
            }

        }
    }
    public static AuthenticationException getLastAuthenticationException() { return lastException; }

    public static void resetUserStatus() { isUserLoggedIn = 99; }

    /**
     * Handles the processing of special error messages like Authentication exception or
     * offline exception.
     *
     * @param ex       Exception to process
     */
    protected void handleSpecialExceptions(Exception ex) {
        // No exception is handled different
    }


    /**
     * Determines the error cause of the failed request and sets the message in the view (error from server side)
     *
     * @param response						Response object
     * @param location						Path of the request (e.g. /attribute)
     * @param responseClass					Class of the response type (for bulk error handling)
     * @throws Exception
     */
    protected <T> void handleErrorResponse(Response<?> response, String location, Class<T> responseClass) throws Exception {

        if (response.code() >= 500) {
            String body = "";
            try { body = response.errorBody().string(); } catch (Exception ex) { }
            baseLogger.log(
                    "e", "An unknown error ({0}) occurred while queuing the server. URL: '{1}'\nBody: '{2}'",
                    response.code(),
                    location,
                    body
            );
            throw new UnknownServerException();
        }

        // API does return errors in text format
        String errorText = response.errorBody().string();
        ErrorResponse errorResponse = new ErrorResponse(errorText, response.code(), location, response.headers());

        // Incorrect username, password or API key given
        if (errorResponse.getCode() == 401) {
            isUserLoggedIn = 0;
            lastAuthCheck = LocalDateTime.now();

            lastException = new AuthenticationException(errorResponse.getText(), AuthenticationException.TYPE.API_KEY);
            throw lastException;
        }

        baseLogger.log("d", "Request failed:\n" + errorResponse);
        throw new ServerException(errorResponse);
    }

    /**
     * Checks the error cause of the occurred exception on calling the API (error from client side)
     *
     * @param ex		thrown exception from retrofit
     *
     * @throws Exception
     */
    protected void handleFailedResponse(Exception ex) throws Exception {
        if (ex instanceof IOException) {
            // The client didn't receive a "correct" answer from the server. This can be because of no Internet connection, or an invalid domain (certificate, no WebServer, ...)

            // Try to determine if the client is offline
            if (globalConfig.getUser() != null && !systemUtils.checkInternetConnection(false, globalConfig.getUser().getServerUrl())) {
                throw new OfflineException();
            } else if (ex instanceof SocketException) {
                baseLogger.log("d", ex, "No internet connection - Received a socket exception");
                throw new OfflineException();
            } else if (ex instanceof UnknownHostException) {
                baseLogger.log("d", "Unable to resolve the IP for the domain - Probably because of no internet connection");
                throw new OfflineException();
            } else {
                baseLogger.log("w", ex, "Couldn't establish an connection to the server");
            }
        }

        throw ex;
    }

    /**
     * Gets the response to the given call object. Exceptions and "bad" error codes will be catched
     *
     * @param <T>			Object-Type of the response result
     * @param call			Call to execute
     * @param responseType	Generic class of the response type (for bulk error handling). When bulk isn't used this should be null
     *
     * @return			A response from the call
     *
     * @throws			Exception
     */
    protected <T> Response<T> getResponse(Call<T> call, Class<?> responseType) throws Exception {

        try {
            Response<T> response;

            try {
                response = call.execute();
            } catch (Exception ex) {
                try {
                    handleFailedResponse(ex);
                } catch (Exception e) {
                    handleSpecialExceptions(e);

                    // Will be left -> an exception is thrown
                    throw e;
                }

                return null;
            }

            if (! (response.code() >= 200 && response.code() < 300) ) {
                // Not a successful response
                try {
                    String baseURL = globalConfig.getUser() != null ? globalConfig.getUser().getServerUrl() :null;
                    String requestURL = call.request().url().toString();
                    if (baseURL != null) requestURL = requestURL.replace(baseURL, "/").replace("/api/v1/", "");
                    handleErrorResponse(response, call.request().method() + " " + requestURL, responseType);
                } catch (Exception ex) {
                    handleSpecialExceptions(ex);

                    // Will be left -> an exception is thrown
                    throw ex;
                }
                return null;
            }

            // Reset any static error messages on a successful response
            String path = call.request().url().uri().getPath();
            if (path.contains("/entry") || path.contains("/attribute") || path.contains("/update")) {
                responseView.resetStatic();
            }

            return response;
        } catch (Exception ex) {
            //responseView.displayErr@or(ex.getMessage());
            throw ex;
        }

    }
    /**
     * {@link #getResponse(Call, Class)}
     */
    protected <T> Response<T> getResponse(Call<T> call) throws Exception {
        return getResponse(call, null);
    }

}

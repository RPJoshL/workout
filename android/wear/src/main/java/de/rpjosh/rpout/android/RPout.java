package de.rpjosh.rpout.android;

import android.annotation.SuppressLint;
import android.app.Application;
import android.content.Context;

public class RPout extends Application {

    @SuppressLint("StaticFieldLeak")
    private static Context context;

    public void onCreate() {
        super.onCreate();
        RPout.context = getApplicationContext();
    }

    /**
     * Returns the application context of this app.
     * Please don't use it for your activities because of the missing
     * garbage collection
     *
     * @return  App context
     */
    public static Context getAppContext() {
        return RPout.context;
    }
}

package de.rpjosh.rpout.android.shared.services;

public interface ResponseViewInterface {

    void displayError(String message);
    void displaySuccess(String message);

    /**
     * This function should display a static "error" message that is been shown until
     * the "resetStatic()" function is called.
     *
     * @param message      Message to display
     */
    void displayStatic(String message);
    String getLastStaticMessage();
    void resetStatic();
}
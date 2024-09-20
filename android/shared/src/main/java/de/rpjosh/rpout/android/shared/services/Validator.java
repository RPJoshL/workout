package de.rpjosh.rpout.android.shared.services;

import java.net.URL;
import java.util.ArrayList;

/**
 * A class for very simple validation rules
 */
public class Validator {

    private ArrayList<String> replaceParameters = new ArrayList<String>();
    private String errorMessage;

    public String validate() {
        if (errorMessage == null) return null;
        return Tr.get(errorMessage, replaceParameters.toArray(new String[replaceParameters.size()]));
    }

    public Validator notNull(Object value, String message) {
        if (value == null) errorMessage = message;
        return this;
    }
    public Validator isNull(Object value, String message) {
        if (value != null) errorMessage = message;
        return this;
    }
    public Validator notBlank(String value, String message) {
        if (value == null || value.isBlank()) errorMessage = message;
        return this;
    }
    public Validator lengthOf(String value, int length, String message) {
        if (value == null || value.length() != length) {
            errorMessage = message;
            replaceParameters.add(value == null ? "-1" : String.valueOf(value.length()));
        }
        return this;
    }
    public Validator greaterThan(Long value, Long greaterThan, String message) {
        if (value == null || value >  greaterThan) errorMessage = message;
        return this;
    }
    public Validator oneNotNull(String message, Object[] values) {
        for (Object value : values) {
            if (value != null) return this;
        }
        errorMessage = message;
        return this;
    }
    public Validator onlyOneNullOrAllNull(String message, Object[] values) {
        int nullCount = 0;
        for (Object value : values) {
            if (value == null) nullCount++;
        }

        // Only one value is null
        if (nullCount == 1) return this;
        // All values are null
        if (nullCount == values.length) return this;

        errorMessage = message;
        return this;
    }

    public Validator url(String value, String message) {
        try {
            URL url = new URL(value);

            // URL accepts all protocols like file, ftp, jar, ...
            if (!url.getProtocol().equals("http") && !url.getProtocol().equals("https")) {
                throw new Exception("Invalid URL");
            }
        } catch (Exception ex) {
            errorMessage = message;
        }

        return this;
    }

}


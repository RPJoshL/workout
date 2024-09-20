package de.rpjosh.rpout.android.shared.exceptions;

import java.io.Serial;

import de.rpjosh.rpout.android.shared.services.Tr;

public class AuthenticationException extends Exception {

    public enum TYPE {
        API_KEY,
        NO_CREDENTIALS,
        UNKNOWN
    }

    @Serial
    private static final long serialVersionUID = -8319318060206597978L;

    private final TYPE type;

    public AuthenticationException(TYPE type) {
        // Translations
        super (
                type == TYPE.API_KEY ? Tr.get("auth_failed_api") :
                type == TYPE.NO_CREDENTIALS ? Tr.get("auth_noCredentials") : "Unknown"
        );
        this.type = type;
    }

    public AuthenticationException(String message) {
        super(message);
        this.type = TYPE.UNKNOWN;
    }

    public AuthenticationException(String message, TYPE type) {
        super (
                type == TYPE.API_KEY ? Tr.get("auth_failed_api") :
                type == TYPE.NO_CREDENTIALS ? Tr.get("auth_no_credentials") : message
        );
        this.type = type;
    }

    public TYPE getType() { return type; }

}


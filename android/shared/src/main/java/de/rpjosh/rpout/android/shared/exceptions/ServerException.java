package de.rpjosh.rpout.android.shared.exceptions;

import java.io.Serial;

import de.rpjosh.rpout.android.shared.models.ErrorResponse;

public class ServerException extends Exception {

    @Serial
    private static final long serialVersionUID = -288684084823659556L;

    private final ErrorResponse response;

    public ServerException(ErrorResponse response) {
        super(response.getText());

        this.response = response;
    }

    public ErrorResponse getResponse() { return this.response; }
}


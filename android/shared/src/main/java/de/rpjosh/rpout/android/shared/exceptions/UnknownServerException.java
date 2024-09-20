package de.rpjosh.rpout.android.shared.exceptions;

import java.io.Serial;

import de.rpjosh.rpout.android.shared.services.Tr;

public class UnknownServerException extends Exception {

    @Serial
    private static final long serialVersionUID = -7871327859232753776L;

    public UnknownServerException() {
        super(Tr.get("unknown_serverError"));
    }
}
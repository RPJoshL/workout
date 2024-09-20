package de.rpjosh.rpout.android.shared.exceptions;

import java.io.Serial;

import de.rpjosh.rpout.android.shared.services.Tr;

public class OfflineException  extends IllegalStateException {

    @Serial
    private static final long serialVersionUID = -7296975904398911566L;

    public OfflineException() {
        super(Tr.get("noInternetConnection"));
    }
}

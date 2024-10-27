package de.rpjosh.rpout.android.shared.models

import okhttp3.Headers

/** ErrorResponse is used as a generic API responses for error codes between 300 and 499 */
data class ErrorResponse(
    val text: String,
    val code: Int,
    val path: String,
    val headers: Headers
) {
    override fun toString(): String {
        return "   Path: $path\n" +
               "   Code: $code\n" +
               "   Text: $text\n";
    }
}
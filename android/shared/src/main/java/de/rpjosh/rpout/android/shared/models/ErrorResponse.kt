package de.rpjosh.rpout.android.shared.models

/** ErrorResponse is used as a generic API responses for error codes between 300 and 499 */
data class ErrorResponse(
    val text: String,
    val code: Int,
    val path: String
) {
    override fun toString(): String {
        return "   Path: $path\n" +
               "   Code: $code\n" +
               "   Text: $text\n";
    }
}
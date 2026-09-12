package de.rpjosh.rpout.android.shared.models

data class ExternalApi(
    val key: String,
    val name: String,
)

enum class ExternalApiType(val key: String) {
    Strava("strava"),
    PumpfoilOrg("pumpfoilorg");

    companion object {
        fun fromKey(key: String): ExternalApiType? = entries.find { it.key == key }
    }
}
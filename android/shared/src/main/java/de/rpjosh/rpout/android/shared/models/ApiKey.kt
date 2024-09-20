package de.rpjosh.rpout.android.shared.models

import de.rpjosh.rpout.android.shared.helper.TimeHelper
import java.time.LocalDateTime

data class ApiKey(
    var id: Long,
    var key: String,
    var userId: Long,
    var obfuscated: String,
    var creationTime: String,
    var validUntil: String,
    var alias: String,
    var darkTheme: Int
) {
    constructor(alias: String, validUntil: LocalDateTime) : this(
        0, "", 0, "", TimeHelper.fromClientToServer(LocalDateTime.now()),
        TimeHelper.fromClientToServer(validUntil), alias, 1
    )
}
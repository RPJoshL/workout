package de.rpjosh.rpout.android.shared.helper

import de.rpjosh.rpout.android.shared.services.Logger
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZonedDateTime
import java.time.format.DateTimeFormatter

class TimeHelper {

    companion object {

        val formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm:ssX")
        lateinit var logger: Logger

        /**
         * Converts the provided time zone string from the server to the client time zone
         */
        fun fromServerToClient(dateString: String?): LocalDateTime {
            if (dateString === null) {
                logger.log("w", "Received a null date string in time converter (server -> client)")
                return LocalDateTime.now()
            }

            try {
                val parsed = ZonedDateTime.parse(dateString, formatter)
                return parsed.toLocalDateTime()
            } catch (ex: Exception) {
                logger.log("w", ex, "Failed to parse time received from server: {0}", dateString)
                return LocalDateTime.now()
            }
        }

        /**
         * Converts the provided date time object into a string that can be parsed by the server
         */
        fun fromClientToServer(dateTime: LocalDateTime): String {
            val zonedDateTime = dateTime.atZone(ZoneId.systemDefault())
            val rtc = zonedDateTime.format(formatter)
            return rtc + ":00"
        }

    }
}
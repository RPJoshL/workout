package de.rpjosh.rpout.android.shared.helper

import android.os.SystemClock
import de.rpjosh.rpout.android.shared.services.Logger
import java.time.Duration
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

        /** Returns the unix time (in seconds) of the provided duration since the device has booted */
        fun getUnixTimeFromBootTime(durationSinceBoot: Duration): Long {
            val timeSinceBoot = durationSinceBoot.toMillis()
            val elapsedTime = SystemClock.elapsedRealtime()
            val unixTime = System.currentTimeMillis()
            val bootTime = unixTime - elapsedTime
            val unixTimestampMillis = bootTime + timeSinceBoot

            return unixTimestampMillis / 1000
        }

        /** Returns the boot duration (in millis) of the provided unix time stamp in seconds */
        fun getBootTimeFromUnixTime(unixTimeSeconds: Long): Long {
            val unixTimeMillis = unixTimeSeconds * 1000
            val currentTimeMillis = System.currentTimeMillis()
            val elapsedRealtimeMillis = SystemClock.elapsedRealtime()
            val bootTimeMillis = currentTimeMillis - elapsedRealtimeMillis

            return unixTimeMillis - bootTimeMillis
        }

    }
}
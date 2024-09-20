package de.rpjosh.rpout.android.shared.models

import androidx.room.Entity
import androidx.room.PrimaryKey
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import java.time.LocalDateTime

@Entity(tableName = "steps")
data class Step(

    /** Internal ID of this step entry */
    @PrimaryKey(autoGenerate = true) val id: Long = 0,

    /** Start date of the step count. This time will be truncated to full minutes */
    val start: String,

    /** End date of the step count. This time will be truncated to full minutes */
    var end: String,

    /** Number of steps that were made between start and end */
    var count: Int,

    /** Internal unix timestamp of start time */
    val startUnix: Long,
    /** Internal unix timestamp of end time */
    var endUnix: Long,

    /** Internal number of steps since last reboot (of start time) */
    var stepsSinceLastReboot: Long,

    /** Whether this step entry was already synchronized */
    var wasSynchronized: Boolean = false
) {

    companion object {
        fun Empty(): Step {
            val nowString = TimeHelper.fromClientToServer(LocalDateTime.now())
            val nowInt = System.currentTimeMillis() / 1000

            return Step(
                0, nowString, nowString, 0,
                nowInt, nowInt, 0, false
            )
        }
    }

    /** Ends this data point at this time */
    fun endNow(count: Float) {
        val nowString = TimeHelper.fromClientToServer(LocalDateTime.now())
        val nowInt = System.currentTimeMillis() / 1000

        end = nowString
        endUnix = nowInt
        this.count = count.toInt()
    }
}

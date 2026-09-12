package de.rpjosh.rpout.android.shared.models

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Ignore
import androidx.room.PrimaryKey
import androidx.room.TypeConverter
import androidx.room.util.TableInfo
import de.rpjosh.rpout.android.shared.helper.TimeHelper
import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.time.Instant
import java.time.LocalDateTime
import java.util.Collections
import java.util.TimeZone

@Entity(tableName = "gpsWorkout")
data class GpsWorkout(

    /** Internal ID of this workout */
    @PrimaryKey(autoGenerate = true) var id: Long = 0,

    /** ID from RPout */
    @ColumnInfo(defaultValue = "0")
    var serverId: Long = 0,

    val type: Long,

    /** Whether this workout entry was already synchronized */
    var wasSynchronized: Boolean = false,

    /** Whether this workout entry is already finished */
    @Volatile var isFinished: Boolean = false,

    /** Unix time stamp (seconds) this workout was started */
    @ColumnInfo(defaultValue = "0")
    var startTime: Long = System.currentTimeMillis() / 1000,
    /** Unix time stamp (seconds) this workout was ended */
    @ColumnInfo(defaultValue = "0")
    var endTime: Long = System.currentTimeMillis() / 1000,

    /** Average traveling speed in sec/km */
    @ColumnInfo(defaultValue = "0")
    var speedAvg: Int = 0,
    /** Total distance in meters */
    @ColumnInfo(defaultValue = "0")
    var distanceTotal: Int = 0,


    @ColumnInfo(defaultValue = "0")
    var useDeviceData: Boolean = false,
    @ColumnInfo(defaultValue = "0")
    var useHighSamplingInterval: Boolean = false,
) {
    @Ignore
    var points: MutableList<GpsWorkoutPoint> = arrayListOf()
}

@Entity(
    tableName = "gpsWorkoutStep",
    foreignKeys = [ForeignKey(
        entity = GpsWorkout::class,
        parentColumns = [ "id" ],
        childColumns = [ "workoutId" ],
        onDelete = ForeignKey.CASCADE
    )]
)

data class GpsWorkoutPoint(

    /** Internal ID of this point */
    @PrimaryKey(autoGenerate = true) val id: Long = 0,

    /** Reference to the root workout */
    @ColumnInfo(index = true)
    val workoutId: Long,

    /** Unix time stamp (in seconds) of this data point */
    val unixTime: Long,

    val time: String,
    var elevation: Int,
    var latitude: Float,
    var longitude: Float,
    var horizontalAccuracy: Float? = null,
    var heartRate: Int,
    var steps: Int,
    @ColumnInfo(defaultValue = "0")
    var totalDistance: Int = 0,
    @ColumnInfo(defaultValue = "0")
    var speed: Int = 0,

    /** Unix timestamp (in milliseconds) from which root point was mainly filled. It's only used internally  */
    @ColumnInfo(defaultValue = "0")
    var refUnixTime: Long = 0,

    /** Stored as "relative timestamp to point,x,y,z,..."
     * Values are a signed Int16, scaled in g: 2048 = 1 g (9.80665 m/s²)
     */
    @Suppress("ArrayInDataClass")
    var acceleration: ShortArray? = null,
) {
    companion object {

        /** Returns a new empty workout point with the provided time set  */
        fun emptyPoint(unixTimeSec: Long, rootId: Long): GpsWorkoutPoint {
            val dateTime = LocalDateTime.ofInstant(Instant.ofEpochSecond(unixTimeSec), TimeZone.getDefault().toZoneId())
            val serverDateTime = TimeHelper.fromClientToServer(dateTime)

            return GpsWorkoutPoint(
                workoutId = rootId,
                steps = 0,
                unixTime = unixTimeSec,
                time = serverDateTime,
                elevation = 0,
                heartRate = 0,
                longitude = 0f,
                latitude = 0f,
                speed = 0,
                totalDistance = 0,
            )
        }
    }

    /**
     * Returns whether this GPS point contains only empty / dummy values that were initialized with "emptyPoint"
     * and aren't a fixed value for RPout
     */
    fun isEmpty(): Boolean {
        return elevation == 0 && heartRate == 0 && longitude == 0f && latitude == 0f && acceleration == null
    }

    override fun toString(): String {
        return "Heartrate = $heartRate | steps = $steps | Elevation = $elevation | Lat = $latitude | Lon = $longitude | Distance = $totalDistance | Speed = $speed"
    }
}

class ShortArrayConverter {

    @TypeConverter
    fun fromShortArray(value: ShortArray): ByteArray =
        ByteBuffer
            .allocate(value.size * 2)
            .order(ByteOrder.LITTLE_ENDIAN)
            .apply {
                value.forEach { putShort(it) }
            }
            .array()

    @TypeConverter
    fun toShortArray(value: ByteArray): ShortArray =
        ByteBuffer
            .wrap(value)
            .order(ByteOrder.LITTLE_ENDIAN)
            .let { buffer ->
                ShortArray(value.size / 2) {
                    buffer.getShort()
                }
            }
}
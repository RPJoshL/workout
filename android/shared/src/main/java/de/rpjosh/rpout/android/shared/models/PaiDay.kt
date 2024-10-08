package de.rpjosh.rpout.android.shared.models

import androidx.room.Entity
import androidx.room.PrimaryKey

data class Pai(
    /** Current PAI score */
    val score: Int,
    /** Progression over last seven days */
    val progression: List<PaiDay>
)

@Entity(tableName = "paiDay")
data class PaiDay(
    /** Unique and incrementing ID of this PAI day (days since unix epoch with client timezone offset applied) */
    @PrimaryKey val dayIndex: Int,

    /** The current PAI score at that specific day */
    val value: Int,
    /** How many PAIs were earned at this day */
    val earned: Int,
    /** Short abbreviation name of the weekday */
    val weekdayAbbrevation: String,
    /** Indexing of the weekday (0 = MONDAY, 1 = TUESDAY) */
    val weekdayIndex: Int
)

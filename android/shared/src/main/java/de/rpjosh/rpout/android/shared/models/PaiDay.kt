package de.rpjosh.rpout.android.shared.models

data class PaiDay(
    /** The current PAI score at that specific day */
    val value: Int,
    /** Short abbreviation name of the weekday */
    val weekdayAbbrevation: String,
    /** Indexing of the weekday (0 = MONDAY, 1 = TUESDAY) */
    val weekdayIndex: Int,
    /** How many PAIs were earned at this day */
    val earned: Int
)

package de.rpjosh.rpout.android.shared.persistence

import androidx.room.AutoMigration
import androidx.room.Database
import androidx.room.RoomDatabase
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.models.User

@Database(
    entities = [
        User::class, Step::class
    ],
    version = 2,
    autoMigrations = [
        AutoMigration (from = 1, to = 2)
    ],
    exportSchema = true
)
abstract class Database: RoomDatabase() {
    abstract fun userDao(): UserDao
    abstract fun metricDao(): MetricDao
}
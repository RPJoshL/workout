package de.rpjosh.rpout.android.shared.persistence

import androidx.room.AutoMigration
import androidx.room.Database
import androidx.room.RoomDatabase
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.models.User
import de.rpjosh.rpout.android.shared.models.Version
import de.rpjosh.rpout.android.shared.models.WorkoutType

@Database(
    entities = [
        User::class, Step::class, WorkoutType::class, Version::class
    ],
    version = 3,
    autoMigrations = [
        AutoMigration (from = 1, to = 2),
        AutoMigration (from = 2, to = 3)
    ],
    exportSchema = true
)
abstract class Database: RoomDatabase() {
    abstract fun userDao(): UserDao
    abstract fun metricDao(): MetricDao
    abstract fun WorkoutDao(): WorkoutDao
}
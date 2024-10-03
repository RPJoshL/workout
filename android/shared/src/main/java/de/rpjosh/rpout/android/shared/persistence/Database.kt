package de.rpjosh.rpout.android.shared.persistence

import androidx.room.AutoMigration
import androidx.room.Database
import androidx.room.RoomDatabase
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.models.User
import de.rpjosh.rpout.android.shared.models.Version
import de.rpjosh.rpout.android.shared.models.WorkoutType

@Database(
    entities = [
        User::class, Step::class, WorkoutType::class, Version::class,
        GpsWorkout::class, GpsWorkoutPoint::class
    ],
    version = 4,
    autoMigrations = [
        AutoMigration (from = 1, to = 2),
        AutoMigration (from = 2, to = 3),
        AutoMigration(from = 3, to = 4)
    ],
    exportSchema = true
)
abstract class Database: RoomDatabase() {
    abstract fun userDao(): UserDao
    abstract fun metricDao(): MetricDao
    abstract fun WorkoutDao(): WorkoutDao
}
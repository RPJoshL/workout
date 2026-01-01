package de.rpjosh.rpout.android.shared.persistence

import androidx.room.AutoMigration
import androidx.room.Database
import androidx.room.DeleteColumn
import androidx.room.RoomDatabase
import androidx.room.migration.AutoMigrationSpec
import de.rpjosh.rpout.android.shared.models.GpsWorkout
import de.rpjosh.rpout.android.shared.models.GpsWorkoutPoint
import de.rpjosh.rpout.android.shared.models.PaiDay
import de.rpjosh.rpout.android.shared.models.Step
import de.rpjosh.rpout.android.shared.models.User
import de.rpjosh.rpout.android.shared.models.Version
import de.rpjosh.rpout.android.shared.models.WorkoutType

@Database(
    entities = [
        User::class, Step::class, WorkoutType::class, Version::class,
        GpsWorkout::class, GpsWorkoutPoint::class, PaiDay::class
    ],
    version = 10,
    autoMigrations = [
        AutoMigration (from = 1, to = 2),
        AutoMigration (from = 2, to = 3),
        AutoMigration(from = 3, to = 4),
        AutoMigration(from = 4, to = 5),
        AutoMigration(from = 5, to = 6),
        AutoMigration(from = 6, to = 7),
        AutoMigration(from = 7, to = 8, spec = DeletePrefGpsColumnMigration::class),
        AutoMigration(from = 8, to = 9),
        AutoMigration(from = 9, to = 10),
    ],
    exportSchema = true
)
abstract class Database: RoomDatabase() {
    abstract fun userDao(): UserDao
    abstract fun metricDao(): MetricDao
    abstract fun WorkoutDao(): WorkoutDao
}

@DeleteColumn.Entries(
    DeleteColumn(
        tableName = "workout_type",
        columnName = "preferSmartphoneGps"
    )
)
class DeletePrefGpsColumnMigration: AutoMigrationSpec
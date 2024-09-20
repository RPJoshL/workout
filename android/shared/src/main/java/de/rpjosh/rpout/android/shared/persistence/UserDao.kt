package de.rpjosh.rpout.android.shared.persistence

import androidx.room.Dao
import androidx.room.Delete
import androidx.room.Insert
import androidx.room.Query
import androidx.room.Update
import de.rpjosh.rpout.android.shared.models.User

@Dao
interface UserDao {

    /**
     * Returns a list of all users in the database.
     * This query should either return no user or exactly
     * a single user
     */
    @Query("SELECT * FROM user")
    fun getAll(): List<User>

    /**
     * Creates a new user in the database
     */
    @Insert
    fun login(user: User)

    @Update
    fun update(user: User)

    @Query("DELETE FROM user")
    fun logout()
}
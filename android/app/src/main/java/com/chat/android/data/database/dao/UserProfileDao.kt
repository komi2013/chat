package com.chat.android.data.database.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import kotlinx.coroutines.flow.Flow
import com.chat.android.data.database.entities.UserProfileEntity

@Dao
interface UserProfileDao {
    @Query("SELECT * FROM user_profile WHERE profileId = 'current' LIMIT 1")
    fun observeCurrentProfile(): Flow<UserProfileEntity?>

    @Query("SELECT * FROM user_profile WHERE profileId = 'current' LIMIT 1")
    suspend fun getCurrentProfileSnapshot(): UserProfileEntity?

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun save(profile: UserProfileEntity)
}
package com.chat.android.data.database.dao

import androidx.room.Dao
import androidx.room.Insert
import androidx.room.OnConflictStrategy
import androidx.room.Query
import androidx.room.Transaction
import kotlinx.coroutines.flow.Flow
import com.chat.android.data.database.entities.UserNicknameEntity

@Dao
interface UserNicknameDao {
    @Query("SELECT * FROM user_nickname ORDER BY createdAt, nickname")
    fun observeNicknames(): Flow<List<UserNicknameEntity>>

    @Query("SELECT * FROM user_nickname WHERE nickname = :nickname LIMIT 1")
    suspend fun getByNickname(nickname: String): UserNicknameEntity?

    @Query("DELETE FROM user_nickname")
    suspend fun clear()

    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAll(nicknames: List<UserNicknameEntity>)

    @Transaction
    suspend fun replaceAll(nicknames: List<UserNicknameEntity>) {
        clear()
        if (nicknames.isNotEmpty()) insertAll(nicknames)
    }
}
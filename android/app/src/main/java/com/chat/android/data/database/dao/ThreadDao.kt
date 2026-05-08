package com.chat.android.data.database.dao

import androidx.room.*
import kotlinx.coroutines.flow.Flow
import com.chat.android.data.database.entities.ThreadEntity

@Dao
interface ThreadDao {
    
    @Query("SELECT * FROM thread")
    fun getAllThreads(): Flow<List<ThreadEntity>>
    
    @Query("SELECT * FROM thread WHERE messageID = :id")
    suspend fun getThreadById(id: String): ThreadEntity?
    
    @Query("SELECT * FROM thread WHERE parentID = :parentId ORDER BY timestamp DESC LIMIT :limit")
    suspend fun getThreadsByParent(parentId: String, limit: Int = 5): List<ThreadEntity>
    
    @Query("SELECT * FROM thread WHERE channelID = :channelId ORDER BY timestamp DESC LIMIT :limit")
    suspend fun getThreadsByChannel(channelId: String, limit: Int = 10): List<ThreadEntity>
    
    @Query("SELECT * FROM thread WHERE channelID = :channelId AND parentID = :parentId ORDER BY timestamp DESC LIMIT :limit")
    suspend fun getThreadsByChannelAndParent(channelId: String, parentId: String, limit: Int = 5): List<ThreadEntity>
    
    @Query("SELECT * FROM thread WHERE aliasName = :aliasName ORDER BY timestamp DESC LIMIT :limit")
    suspend fun getThreadsByAlias(aliasName: String, limit: Int = 10): List<ThreadEntity>
    
    @Query("SELECT * FROM thread WHERE bookmark = 1 ORDER BY timestamp DESC")
    suspend fun getBookmarkedThreads(): List<ThreadEntity>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertThread(thread: ThreadEntity)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertThreads(threads: List<ThreadEntity>)
    
    @Update
    suspend fun updateThread(thread: ThreadEntity)
    
    @Delete
    suspend fun deleteThread(thread: ThreadEntity)
    
    @Query("DELETE FROM thread WHERE messageID = :id")
    suspend fun deleteThreadById(id: String)
    
    @Query("DELETE FROM thread WHERE channelID = :channelId")
    suspend fun deleteThreadsByChannel(channelId: String)
    
    @Query("DELETE FROM thread")
    suspend fun deleteAllThreads()
}

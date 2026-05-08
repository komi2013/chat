package com.chat.android.data.database.dao

import androidx.room.*
import kotlinx.coroutines.flow.Flow
import com.chat.android.data.database.entities.ChannelEntity

@Dao
interface ChannelDao {
    
    @Query("SELECT * FROM channel")
    fun getAllChannels(): Flow<List<ChannelEntity>>
    
    @Query("SELECT * FROM channel WHERE channelID = :id")
    suspend fun getChannelById(id: String): ChannelEntity?
    
    @Query("SELECT * FROM channel WHERE displayStatus = :status")
    suspend fun getChannelsByStatus(status: String): List<ChannelEntity>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertChannel(channel: ChannelEntity)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertChannels(channels: List<ChannelEntity>)
    
    @Update
    suspend fun updateChannel(channel: ChannelEntity)
    
    @Delete
    suspend fun deleteChannel(channel: ChannelEntity)
    
    @Query("DELETE FROM channel WHERE channelID = :id")
    suspend fun deleteChannelById(id: String)
    
    @Query("DELETE FROM channel")
    suspend fun deleteAllChannels()
}

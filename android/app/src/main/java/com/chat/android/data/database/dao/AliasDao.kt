package com.chat.android.data.database.dao

import androidx.room.*
import kotlinx.coroutines.flow.Flow
import com.chat.android.data.database.entities.AliasEntity

@Dao
interface AliasDao {
    
    @Query("SELECT * FROM alias")
    fun getAllAliases(): Flow<List<AliasEntity>>
    
    @Query("SELECT * FROM alias WHERE aliasID = :id")
    suspend fun getAliasById(id: String): AliasEntity?
    
    @Query("SELECT * FROM alias WHERE channelID = :channelId")
    suspend fun getAliasesByChannel(channelId: String): List<AliasEntity>
    
    @Query("SELECT * FROM alias WHERE channelID = :channelId LIMIT :limit OFFSET :offset")
    suspend fun getAliasesByChannelPaginated(channelId: String, limit: Int, offset: Int): List<AliasEntity>
    
    @Query("SELECT * FROM alias WHERE aliasName = :name")
    suspend fun getAliasByName(name: String): AliasEntity?
    
    @Query("SELECT * FROM alias WHERE accessRight = 'guest' AND aliasName = :name")
    suspend fun getGuestAliasByName(name: String): AliasEntity?
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAlias(alias: AliasEntity)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAliases(aliases: List<AliasEntity>)
    
    @Update
    suspend fun updateAlias(alias: AliasEntity)
    
    @Delete
    suspend fun deleteAlias(alias: AliasEntity)
    
    @Query("DELETE FROM alias WHERE aliasID = :id")
    suspend fun deleteAliasById(id: String)
    
    @Query("DELETE FROM alias WHERE channelID = :channelId")
    suspend fun deleteAliasesByChannel(channelId: String)
    
    @Query("DELETE FROM alias")
    suspend fun deleteAllAliases()
}

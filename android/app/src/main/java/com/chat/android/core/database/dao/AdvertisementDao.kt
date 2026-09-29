package com.chat.android.core.database.dao

import androidx.room.*
import kotlinx.coroutines.flow.Flow
import com.chat.android.core.database.entities.AdvertisementEntity

@Dao
interface AdvertisementDao {
    
    @Query("SELECT * FROM advertisement")
    fun getAllAdvertisements(): Flow<List<AdvertisementEntity>>
    
    @Query("SELECT * FROM advertisement WHERE advertisementID = :id")
    suspend fun getAdvertisementById(id: String): AdvertisementEntity?
    
    @Query("SELECT * FROM advertisement WHERE channelID = :channelId")
    suspend fun getAdvertisementsByChannel(channelId: String): List<AdvertisementEntity>
    
    @Query("SELECT * FROM advertisement WHERE adPublic = 1")
    suspend fun getPublicAdvertisements(): List<AdvertisementEntity>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAdvertisement(advertisement: AdvertisementEntity)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertAdvertisements(advertisements: List<AdvertisementEntity>)
    
    @Update
    suspend fun updateAdvertisement(advertisement: AdvertisementEntity)
    
    @Delete
    suspend fun deleteAdvertisement(advertisement: AdvertisementEntity)
    
    @Query("DELETE FROM advertisement WHERE advertisementID = :id")
    suspend fun deleteAdvertisementById(id: String)
    
    @Query("DELETE FROM advertisement")
    suspend fun deleteAllAdvertisements()
    
    @Query("DELETE FROM advertisement WHERE adEnd < :currentTime")
    suspend fun deleteExpiredAdvertisements(currentTime: String)
}

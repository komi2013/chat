package com.chat.android.core.database.dao

import androidx.room.*
import kotlinx.coroutines.flow.Flow
import com.chat.android.core.database.entities.TicketEntity

@Dao
interface TicketDao {
    
    @Query("SELECT * FROM ticket")
    fun getAllTickets(): Flow<List<TicketEntity>>
    
    @Query("SELECT * FROM ticket WHERE ticketID = :id")
    suspend fun getTicketById(id: String): TicketEntity?
    
    @Query("SELECT * FROM ticket WHERE channelID = :channelId")
    suspend fun getTicketsByChannel(channelId: String): List<TicketEntity>
    
    @Query("SELECT * FROM ticket WHERE status = :status")
    suspend fun getTicketsByStatus(status: String): List<TicketEntity>
    
    @Query("SELECT * FROM ticket WHERE aliasName = :aliasName")
    suspend fun getTicketsByAlias(aliasName: String): List<TicketEntity>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertTicket(ticket: TicketEntity)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertTickets(tickets: List<TicketEntity>)
    
    @Update
    suspend fun updateTicket(ticket: TicketEntity)
    
    @Delete
    suspend fun deleteTicket(ticket: TicketEntity)
    
    @Query("DELETE FROM ticket WHERE ticketID = :id")
    suspend fun deleteTicketById(id: String)
    
    @Query("DELETE FROM ticket WHERE channelID = :channelId")
    suspend fun deleteTicketsByChannel(channelId: String)
    
    @Query("DELETE FROM ticket")
    suspend fun deleteAllTickets()
}

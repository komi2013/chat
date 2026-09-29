package com.chat.android.core.database.dao

import androidx.room.*
import kotlinx.coroutines.flow.Flow
import com.chat.android.core.database.entities.CalendarEntity

@Dao
interface CalendarDao {
    
    @Query("SELECT * FROM calendar")
    fun getAllCalendars(): Flow<List<CalendarEntity>>
    
    @Query("SELECT * FROM calendar WHERE calendarID = :id")
    suspend fun getCalendarById(id: String): CalendarEntity?
    
    @Query("SELECT * FROM calendar WHERE channelID = :channelId")
    suspend fun getCalendarsByChannel(channelId: String): List<CalendarEntity>
    
    @Query("SELECT * FROM calendar WHERE date = :date")
    suspend fun getCalendarsByDate(date: String): List<CalendarEntity>
    
    @Query("SELECT * FROM calendar WHERE channelID = :channelId AND date = :date")
    suspend fun getCalendarsByChannelAndDate(channelId: String, date: String): List<CalendarEntity>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertCalendar(calendar: CalendarEntity)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertCalendars(calendars: List<CalendarEntity>)
    
    @Update
    suspend fun updateCalendar(calendar: CalendarEntity)
    
    @Delete
    suspend fun deleteCalendar(calendar: CalendarEntity)
    
    @Query("DELETE FROM calendar WHERE calendarID = :id")
    suspend fun deleteCalendarById(id: String)
    
    @Query("DELETE FROM calendar WHERE channelID = :channelId")
    suspend fun deleteCalendarsByChannel(channelId: String)
    
    @Query("DELETE FROM calendar")
    suspend fun deleteAllCalendars()
}

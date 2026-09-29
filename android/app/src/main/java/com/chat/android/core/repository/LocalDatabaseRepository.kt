package com.chat.android.core.repository

import android.content.Context
import com.chat.android.core.database.ChatDatabase
import com.chat.android.core.database.entities.*
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.first
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class LocalDatabaseRepository @Inject constructor(
    context: Context
) {
    private val database = ChatDatabase.getDatabase(context)
    
    // Advertisement operations (Vue IndexedDB: advertisement)
    suspend fun getAllAdvertisements(): List<AdvertisementEntity> = 
        database.advertisementDao().getAllAdvertisements().first()
    
    suspend fun getAdvertisementsByChannel(channelId: String): List<AdvertisementEntity> = 
        database.advertisementDao().getAdvertisementsByChannel(channelId)
    
    suspend fun getPublicAdvertisements(): List<AdvertisementEntity> = 
        database.advertisementDao().getPublicAdvertisements()
    
    suspend fun insertAdvertisements(advertisements: List<AdvertisementEntity>) = 
        database.advertisementDao().insertAdvertisements(advertisements)
    
    suspend fun deleteExpiredAdvertisements(currentTime: String) = 
        database.advertisementDao().deleteExpiredAdvertisements(currentTime)
    
    // Alias operations (Vue IndexedDB: alias)
    suspend fun getAllAliases(): List<AliasEntity> = 
        database.aliasDao().getAllAliases().first()
    
    suspend fun getAliasesByChannel(channelId: String, limit: Int = 10000, offset: Int = 0): List<AliasEntity> = 
        database.aliasDao().getAliasesByChannelPaginated(channelId, limit, offset)
    
    suspend fun getAliasByName(name: String): AliasEntity? = 
        database.aliasDao().getAliasByName(name)
    
    suspend fun getGuestAliasByName(name: String): AliasEntity? = 
        database.aliasDao().getGuestAliasByName(name)
    
    suspend fun insertAliases(aliases: List<AliasEntity>) = 
        database.aliasDao().insertAliases(aliases)
    
    // Bookmark operations (Vue IndexedDB: bookmark)
    suspend fun getAllBookmarks(): List<BookmarkEntity> = 
        database.bookmarkDao().getAllBookmarks().first()
    
    suspend fun getBookmarksByChannel(channelId: String): List<BookmarkEntity> = 
        database.bookmarkDao().getBookmarksByChannel(channelId)
    
    suspend fun getBookmarksByAlias(aliasName: String): List<BookmarkEntity> = 
        database.bookmarkDao().getBookmarksByAlias(aliasName)
    
    suspend fun getActiveBookmarks(): List<BookmarkEntity> = 
        database.bookmarkDao().getActiveBookmarks()
    
    suspend fun insertBookmarks(bookmarks: List<BookmarkEntity>) = 
        database.bookmarkDao().insertBookmarks(bookmarks)
    
    // Channel operations (Vue IndexedDB: channel)
    suspend fun getAllChannels(): List<ChannelEntity> = 
        database.channelDao().getAllChannels().first()
    
    suspend fun getChannelById(id: String): ChannelEntity? = 
        database.channelDao().getChannelById(id)
    
    suspend fun getChannelsByStatus(status: String): List<ChannelEntity> = 
        database.channelDao().getChannelsByStatus(status)
    
    suspend fun insertChannels(channels: List<ChannelEntity>) = 
        database.channelDao().insertChannels(channels)
    
    // Thread operations (Vue IndexedDB: thread)
    suspend fun getAllThreads(): List<ThreadEntity> = 
        database.threadDao().getAllThreads().first()
    
    suspend fun getThreadsByParent(parentId: String, limit: Int = 5): List<ThreadEntity> = 
        database.threadDao().getThreadsByParent(parentId, limit)
    
    suspend fun getThreadsByChannel(channelId: String, limit: Int = 10): List<ThreadEntity> = 
        database.threadDao().getThreadsByChannel(channelId, limit)
    
    suspend fun getThreadsByChannelAndParent(channelId: String, parentId: String, limit: Int = 5): List<ThreadEntity> = 
        database.threadDao().getThreadsByChannelAndParent(channelId, parentId, limit)
    
    suspend fun getThreadsByAlias(aliasName: String, limit: Int = 10): List<ThreadEntity> = 
        database.threadDao().getThreadsByAlias(aliasName, limit)
    
    suspend fun getBookmarkedThreads(): List<ThreadEntity> = 
        database.threadDao().getBookmarkedThreads()
    
    suspend fun insertThreads(threads: List<ThreadEntity>) = 
        database.threadDao().insertThreads(threads)
    
    // Calendar operations (Vue IndexedDB: calendar)
    suspend fun getAllCalendars(): List<CalendarEntity> = 
        database.calendarDao().getAllCalendars().first()
    
    suspend fun getCalendarsByChannel(channelId: String): List<CalendarEntity> = 
        database.calendarDao().getCalendarsByChannel(channelId)
    
    suspend fun getCalendarsByDate(date: String): List<CalendarEntity> = 
        database.calendarDao().getCalendarsByDate(date)
    
    suspend fun getCalendarsByChannelAndDate(channelId: String, date: String): List<CalendarEntity> = 
        database.calendarDao().getCalendarsByChannelAndDate(channelId, date)
    
    suspend fun insertCalendars(calendars: List<CalendarEntity>) = 
        database.calendarDao().insertCalendars(calendars)
    
    // Ticket operations (Vue IndexedDB: ticket)
    suspend fun getAllTickets(): List<TicketEntity> = 
        database.ticketDao().getAllTickets().first()
    
    suspend fun getTicketsByChannel(channelId: String): List<TicketEntity> = 
        database.ticketDao().getTicketsByChannel(channelId)
    
    suspend fun getTicketsByStatus(status: String): List<TicketEntity> = 
        database.ticketDao().getTicketsByStatus(status)
    
    suspend fun getTicketsByAlias(aliasName: String): List<TicketEntity> = 
        database.ticketDao().getTicketsByAlias(aliasName)
    
    suspend fun insertTickets(tickets: List<TicketEntity>) = 
        database.ticketDao().insertTickets(tickets)
    
    // Utility operations
    suspend fun clearAllData() {
        database.advertisementDao().deleteAllAdvertisements()
        database.aliasDao().deleteAllAliases()
        database.bookmarkDao().deleteAllBookmarks()
        database.channelDao().deleteAllChannels()
        database.threadDao().deleteAllThreads()
        database.calendarDao().deleteAllCalendars()
        database.ticketDao().deleteAllTickets()
    }
    
    suspend fun clearChannelData(channelId: String) {
        database.aliasDao().deleteAliasesByChannel(channelId)
        database.channelDao().deleteChannelById(channelId)
        database.threadDao().deleteThreadsByChannel(channelId)
        database.calendarDao().deleteCalendarsByChannel(channelId)
        database.ticketDao().deleteTicketsByChannel(channelId)
    }
}

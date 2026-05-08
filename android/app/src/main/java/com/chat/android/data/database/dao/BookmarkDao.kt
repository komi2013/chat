package com.chat.android.data.database.dao

import androidx.room.*
import kotlinx.coroutines.flow.Flow
import com.chat.android.data.database.entities.BookmarkEntity

@Dao
interface BookmarkDao {
    
    @Query("SELECT * FROM bookmark")
    fun getAllBookmarks(): Flow<List<BookmarkEntity>>
    
    @Query("SELECT * FROM bookmark WHERE messageID = :id")
    suspend fun getBookmarkById(id: String): BookmarkEntity?
    
    @Query("SELECT * FROM bookmark WHERE channelID = :channelId")
    suspend fun getBookmarksByChannel(channelId: String): List<BookmarkEntity>
    
    @Query("SELECT * FROM bookmark WHERE aliasName = :aliasName")
    suspend fun getBookmarksByAlias(aliasName: String): List<BookmarkEntity>
    
    @Query("SELECT * FROM bookmark WHERE bookmarked = 1")
    suspend fun getActiveBookmarks(): List<BookmarkEntity>
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertBookmark(bookmark: BookmarkEntity)
    
    @Insert(onConflict = OnConflictStrategy.REPLACE)
    suspend fun insertBookmarks(bookmarks: List<BookmarkEntity>)
    
    @Update
    suspend fun updateBookmark(bookmark: BookmarkEntity)
    
    @Delete
    suspend fun deleteBookmark(bookmark: BookmarkEntity)
    
    @Query("DELETE FROM bookmark WHERE messageID = :id")
    suspend fun deleteBookmarkById(id: String)
    
    @Query("DELETE FROM bookmark")
    suspend fun deleteAllBookmarks()
}

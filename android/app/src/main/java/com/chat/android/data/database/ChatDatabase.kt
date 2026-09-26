package com.chat.android.data.database

import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import androidx.room.migration.Migration
import androidx.sqlite.db.SupportSQLiteDatabase
import android.content.Context
import com.chat.android.data.database.entities.*
import com.chat.android.data.database.dao.*

@Database(
    entities = [
        AdvertisementEntity::class,
        AliasEntity::class,
        BookmarkEntity::class,
        CalendarEntity::class,
        ChannelEntity::class,
        ThreadEntity::class,
        TicketEntity::class,
        UserProfileEntity::class,
        UserNicknameEntity::class
        // Add more entities as needed
    ],
    version = 2,
    exportSchema = false
)
abstract class ChatDatabase : RoomDatabase() {
    
    abstract fun advertisementDao(): AdvertisementDao
    abstract fun aliasDao(): AliasDao
    abstract fun bookmarkDao(): BookmarkDao
    abstract fun calendarDao(): CalendarDao
    abstract fun channelDao(): ChannelDao
    abstract fun threadDao(): ThreadDao
    abstract fun ticketDao(): TicketDao
    abstract fun userProfileDao(): UserProfileDao
    abstract fun userNicknameDao(): UserNicknameDao
    
    companion object {
        @Volatile
        private var INSTANCE: ChatDatabase? = null
        
        fun getDatabase(context: Context): ChatDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    ChatDatabase::class.java,
                    "chat_database"
                )
                .addMigrations(MIGRATION_1_2)
                .fallbackToDestructiveMigration()
                .build()
                INSTANCE = instance
                instance
            }
        }
        
        // Migration for future schema updates
        private val MIGRATION_1_2 = object : Migration(1, 2) {
            override fun migrate(database: SupportSQLiteDatabase) {
                database.execSQL(
                    "CREATE TABLE IF NOT EXISTS user_profile (" +
                        "profileId TEXT NOT NULL PRIMARY KEY, " +
                        "name TEXT, mail TEXT, nickname TEXT, channelID TEXT, " +
                        "accessRight TEXT, admin INTEGER, telephone TEXT, walletAddress TEXT, " +
                        "latitude REAL, longitude REAL, nickImg TEXT, nickBio TEXT)"
                )
                database.execSQL(
                    "CREATE TABLE IF NOT EXISTS user_nickname (" +
                        "nickname TEXT NOT NULL PRIMARY KEY, nickImg TEXT, nickBio TEXT, " +
                        "good INTEGER NOT NULL, bad INTEGER NOT NULL, createdAt TEXT, accessRight TEXT)"
                )
            }
        }
    }
}

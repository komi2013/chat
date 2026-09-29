package com.chat.android.di

import android.content.Context
import androidx.room.Room
import com.chat.android.core.database.ChatDatabase
import com.chat.android.core.database.dao.AdvertisementDao
import com.chat.android.core.database.dao.AliasDao
import com.chat.android.core.database.dao.BookmarkDao
import com.chat.android.core.database.dao.CalendarDao
import com.chat.android.core.database.dao.ChannelDao
import com.chat.android.core.database.dao.ThreadDao
import com.chat.android.core.database.dao.TicketDao
import com.chat.android.core.database.dao.UserNicknameDao
import com.chat.android.core.database.dao.UserProfileDao
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object DatabaseModule {

    @Provides
    @Singleton
    fun provideChatDatabase(@ApplicationContext context: Context): ChatDatabase {
        return ChatDatabase.getDatabase(context)
    }

    @Provides
    fun provideAdvertisementDao(database: ChatDatabase): AdvertisementDao {
        return database.advertisementDao()
    }

    @Provides
    fun provideAliasDao(database: ChatDatabase): AliasDao {
        return database.aliasDao()
    }

    @Provides
    fun provideBookmarkDao(database: ChatDatabase): BookmarkDao {
        return database.bookmarkDao()
    }

    @Provides
    fun provideCalendarDao(database: ChatDatabase): CalendarDao {
        return database.calendarDao()
    }

    @Provides
    fun provideChannelDao(database: ChatDatabase): ChannelDao {
        return database.channelDao()
    }

    @Provides
    fun provideThreadDao(database: ChatDatabase): ThreadDao {
        return database.threadDao()
    }

    @Provides
    fun provideTicketDao(database: ChatDatabase): TicketDao {
        return database.ticketDao()
    }

    @Provides
    fun provideUserProfileDao(database: ChatDatabase): UserProfileDao {
        return database.userProfileDao()
    }

    @Provides
    fun provideUserNicknameDao(database: ChatDatabase): UserNicknameDao {
        return database.userNicknameDao()
    }
}

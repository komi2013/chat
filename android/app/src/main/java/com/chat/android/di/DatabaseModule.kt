package com.chat.android.di

import android.content.Context
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.reception.ReceptionDbHelper
import com.chat.android.feature.ticket.TicketDbHelper
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
    fun provideChannelDbHelper(@ApplicationContext context: Context): ChannelDbHelper {
        return ChannelDbHelper(context)
    }

    @Provides
    @Singleton
    fun provideReceptionDbHelper(@ApplicationContext context: Context): ReceptionDbHelper {
        return ReceptionDbHelper(context)
    }

    @Provides
    @Singleton
    fun provideTicketDbHelper(@ApplicationContext context: Context): TicketDbHelper {
        return TicketDbHelper(context)
    }

    @Provides
    @Singleton
    fun provideChannelProfileStore(sessionManager: com.chat.android.core.network.SessionManager): com.chat.android.feature.channel.ChannelProfileStore {
        return com.chat.android.feature.channel.ChannelProfileStore(sessionManager)
    }
}

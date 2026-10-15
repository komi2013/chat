package com.chat.android.di

import com.chat.android.core.network.ApiService
import com.chat.android.core.network.RetrofitClient
import com.chat.android.core.network.SessionManager
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object NetworkModule {

    @Provides
    @Singleton
    fun provideApiService(sessionManager: SessionManager): ApiService {
        RetrofitClient.configureSessionManager(sessionManager)
        return RetrofitClient.apiService
    }

    @Provides
    @Singleton
    fun provideEntryFormApiService(): com.chat.android.feature.entryform.EntryFormApiService {
        return RetrofitClient.createService(com.chat.android.feature.entryform.EntryFormApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideChannelApiService(): com.chat.android.feature.channel.ChannelApiService {
        return RetrofitClient.createService(com.chat.android.feature.channel.ChannelApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideCalendarApiService(): com.chat.android.feature.calendar.CalendarApiService {
        return RetrofitClient.createService(com.chat.android.feature.calendar.CalendarApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideReceptionApiService(): com.chat.android.feature.reception.ReceptionApiService {
        return RetrofitClient.createService(com.chat.android.feature.reception.ReceptionApiService::class.java)
    }

    @Provides
    @Singleton
    fun provideTicketApiService(): com.chat.android.feature.ticket.TicketApiService {
        return RetrofitClient.createService(com.chat.android.feature.ticket.TicketApiService::class.java)
    }
}

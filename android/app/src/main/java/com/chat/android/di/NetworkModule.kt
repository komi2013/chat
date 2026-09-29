package com.chat.android.di

import com.chat.android.core.data.SessionManager
import com.chat.android.core.network.ApiService
import com.chat.android.core.network.RetrofitClient
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
}

package com.chat.android.di

import android.content.Context
import android.database.sqlite.SQLiteOpenHelper
import com.chat.android.core.push.PushDatabaseHelper
import com.chat.android.core.push.PushReceiveDispatcher
import dagger.Module
import dagger.Provides
import dagger.hilt.InstallIn
import dagger.hilt.android.qualifiers.ApplicationContext
import dagger.hilt.components.SingletonComponent
import javax.inject.Singleton

@Module
@InstallIn(SingletonComponent::class)
object PushModule {

    @Provides
    @Singleton
    fun providePushDatabaseHelper(@ApplicationContext context: Context): SQLiteOpenHelper {
        return PushDatabaseHelper(context)
    }

    @Provides
    @Singleton
    fun providePushReceiveDispatcher(
        @ApplicationContext context: Context,
        dbHelper: SQLiteOpenHelper,
        sessionManager: com.chat.android.core.network.SessionManager,
        apiService: com.chat.android.core.network.ApiService,
        receptionApiService: com.chat.android.feature.reception.ReceptionApiService,
        userRepository: com.chat.android.core.repository.UserRepository,
        profileStore: com.chat.android.feature.channel.ChannelProfileStore
    ): PushReceiveDispatcher {
        return PushReceiveDispatcher(
            context,
            dbHelper,
            sessionManager,
            apiService,
            receptionApiService,
            userRepository,
            profileStore
        )
    }
}

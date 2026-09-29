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
        dbHelper: SQLiteOpenHelper
    ): PushReceiveDispatcher {
        return PushReceiveDispatcher(context, dbHelper)
    }
}

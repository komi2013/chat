package com.chat.android.core.push

import android.content.Context
import android.util.Log
import androidx.core.content.edit
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PushTokenManager @Inject constructor(
    @ApplicationContext private val context: Context
) {
    fun storeFCMTokenLocally(token: String) {
        val prefs = context.getSharedPreferences("chat_prefs", Context.MODE_PRIVATE)
        prefs.edit {
            putString("fcm_token", token)
        }
        Log.d("PushTokenManager", "FCM Token stored locally")
    }

    fun getStoredFCMToken(): String? {
        val prefs = context.getSharedPreferences("chat_prefs", Context.MODE_PRIVATE)
        return prefs.getString("fcm_token", null)
    }
}

package com.chat.android.data

import android.content.Context
import android.content.SharedPreferences
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SessionManager @Inject constructor(
    @ApplicationContext context: Context
) {
    private val prefs: SharedPreferences = context.getSharedPreferences("chat_prefs", Context.MODE_PRIVATE)

    companion object {
        private const val KEY_CSRF = "csrf"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_NICKNAME = "nickname"
    }

    fun saveSession(csrf: String, userId: String?, nickname: String?) {
        prefs.edit().apply {
            putString(KEY_CSRF, csrf)
            putString(KEY_USER_ID, userId)
            putString(KEY_NICKNAME, nickname)
            apply()
        }
    }

    fun getCsrf(): String? = prefs.getString(KEY_CSRF, null)
    
    fun getUserId(): String? = prefs.getString(KEY_USER_ID, null)

    fun getNickname(): String? = prefs.getString(KEY_NICKNAME, null)

    fun clearSession() {
        prefs.edit().clear().apply()
    }
}

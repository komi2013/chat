package com.chat.android.data.repository

import android.content.Context
import android.content.SharedPreferences
import com.chat.android.data.model.TopLink
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UserRepository @Inject constructor(
    @ApplicationContext private val context: Context
) {
    private val sharedPreferences: SharedPreferences = 
        context.getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
    
    private val gson = Gson()

    companion object {
        private const val CSRF_TOKEN_KEY = "csrf"
        private const val NICKNAME_KEY = "nickname"
        private const val CHANNEL_ID_KEY = "channelID"
        private const val TO_KEY = "TO"
        private const val TOP_LINKS_PREFIX = "topLinks"
    }

    fun getCsrfToken(): String? {
        return sharedPreferences.getString(CSRF_TOKEN_KEY, null)
    }

    fun setCsrfToken(token: String) {
        sharedPreferences.edit()
            .putString(CSRF_TOKEN_KEY, token)
            .apply()
    }

    fun getNickname(): String? {
        return sharedPreferences.getString(NICKNAME_KEY, null)
    }

    fun setNickname(nickname: String) {
        sharedPreferences.edit()
            .putString(NICKNAME_KEY, nickname)
            .apply()
    }

    fun getCurrentChannelId(): String? {
        return sharedPreferences.getString(CHANNEL_ID_KEY, null)
    }

    fun setCurrentChannelId(channelId: String) {
        sharedPreferences.edit()
            .putString(CHANNEL_ID_KEY, channelId)
            .apply()
    }

    fun getTO(): String? {
        return sharedPreferences.getString(TO_KEY, null)
    }

    fun setTO(to: String) {
        sharedPreferences.edit()
            .putString(TO_KEY, to)
            .apply()
    }

    fun clearTO() {
        sharedPreferences.edit()
            .remove(TO_KEY)
            .apply()
    }

    fun getStoredTopLinks(channelId: String): List<TopLink>? {
        val key = "$TOP_LINKS_PREFIX$channelId"
        val json = sharedPreferences.getString(key, null)
        return if (json != null) {
            val type = object : TypeToken<List<TopLink>>() {}.type
            gson.fromJson(json, type)
        } else null
    }

    fun setStoredTopLinks(channelId: String, topLinks: List<TopLink>) {
        val key = "$TOP_LINKS_PREFIX$channelId"
        val json = gson.toJson(topLinks)
        sharedPreferences.edit()
            .putString(key, json)
            .apply()
    }

    fun clearUserData() {
        sharedPreferences.edit()
            .remove(CSRF_TOKEN_KEY)
            .remove(NICKNAME_KEY)
            .remove(CHANNEL_ID_KEY)
            .remove(TO_KEY)
            .apply()
    }

    fun isSignedIn(): Boolean {
        return !getCsrfToken().isNullOrEmpty()
    }
}

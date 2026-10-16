package com.chat.android.feature.channel

import com.chat.android.core.network.SessionManager
import kotlinx.coroutines.runBlocking

/** 
 * Device-local current alias and icon, keyed by channel. 
 * Delegated to SessionManager (DataStore).
 */
class ChannelProfileStore(private val sessionManager: SessionManager) {

    suspend fun profile(channelID: String): Pair<String, String> {
        return sessionManager.getChannelProfile(channelID)
    }

    suspend fun save(channelID: String, name: String? = null, image: String? = null) {
        sessionManager.saveChannelProfile(channelID, name, image)
    }

    suspend fun saveLastProfile(userId: String, channelID: String, name: String, image: String) {
        sessionManager.saveLastProfile(userId, channelID, name, image)
    }

    suspend fun getLastProfile(userId: String): Triple<String, String, String> {
        return sessionManager.getLastProfile(userId)
    }
}

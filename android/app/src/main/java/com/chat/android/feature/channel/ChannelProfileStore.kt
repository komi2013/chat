package com.chat.android.feature.channel

import android.content.Context
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import kotlinx.coroutines.flow.first

private val Context.channelProfileDataStore by preferencesDataStore(name = "channel_profiles")

/** Device-local current alias and icon, keyed by channel. */
class ChannelProfileStore(private val context: Context) {
    private fun nameKey(channelID: String) = stringPreferencesKey("${channelID}_name")
    private fun imageKey(channelID: String) = stringPreferencesKey("${channelID}_image")
    
    // Global keys for "the latest/top profile" per user
    private fun lastChannelKey(userId: String) = stringPreferencesKey("${userId}_last_channel_id")
    private fun lastNameKey(userId: String) = stringPreferencesKey("${userId}_last_name")
    private fun lastImageKey(userId: String) = stringPreferencesKey("${userId}_last_image")

    suspend fun profile(channelID: String): Pair<String, String> {
        val prefs = context.channelProfileDataStore.data.first()
        var name = prefs[nameKey(channelID)]
        var image = prefs[imageKey(channelID)]
        if (name == null || image == null) {
            val legacy = context.getSharedPreferences("channel_profiles_migration", Context.MODE_PRIVATE)
            val oldName = legacy.getString("${channelID}_name", null)
            val oldImage = legacy.getString("${channelID}_image", null)
            if (name == null && oldName != null) name = oldName
            if (image == null && oldImage != null) image = oldImage
            if (oldName != null || oldImage != null) {
                save(channelID, name, image)
                legacy.edit().remove("${channelID}_name").remove("${channelID}_image").apply()
            }
        }
        return name.orEmpty() to image.orEmpty()
    }

    suspend fun save(channelID: String, name: String? = null, image: String? = null) {
        context.channelProfileDataStore.edit { prefs ->
            if (name != null) prefs[nameKey(channelID)] = name
            if (image != null) prefs[imageKey(channelID)] = image
        }
    }

    suspend fun saveLastProfile(userId: String, channelID: String, name: String, image: String) {
        context.channelProfileDataStore.edit { prefs ->
            prefs[lastChannelKey(userId)] = channelID
            prefs[lastNameKey(userId)] = name
            prefs[lastImageKey(userId)] = image
            
            // Also update the per-channel one
            prefs[nameKey(channelID)] = name
            prefs[imageKey(channelID)] = image
        }
    }

    suspend fun getLastProfile(userId: String): Triple<String, String, String> {
        val prefs = context.channelProfileDataStore.data.first()
        return Triple(
            prefs[lastChannelKey(userId)].orEmpty(),
            prefs[lastNameKey(userId)].orEmpty(),
            prefs[lastImageKey(userId)].orEmpty()
        )
    }
}

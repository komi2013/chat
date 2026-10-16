package com.chat.android.core.push

import org.json.JSONArray
import org.json.JSONObject

/**
 * Wraps the raw JSON array (pd) wire format.
 * Indices: [0]=pushID, [1]=pushTitle, [2]=channelID, [3]=updatedBy, [4]=contents, [5]=extra
 */
data class PushData(
    val rawJson: JSONArray,
    val directPush: Boolean = false
) {
    val pushID: String
        get() = rawJson.optString(0, "")
        
    val title: String
        get() = rawJson.optString(1, "")
        
    val channelID: String
        get() = rawJson.optString(2, "")
        
    val updatedBy: String
        get() = rawJson.optString(3, "")

    // contents can be a String, JSONObject, or JSONArray. Handlers cast it as needed.
    val contents: Any?
        get() {
            val raw = rawJson.opt(4)
            if (raw is String && (raw.trim().startsWith("{") || raw.trim().startsWith("["))) {
                return try {
                    val trimmed = raw.trim()
                    if (trimmed.startsWith("{")) JSONObject(trimmed) else JSONArray(trimmed)
                } catch (e: Exception) {
                    raw
                }
            }
            return raw
        }
        
    val extra: Any?
        get() {
            if (rawJson.length() > 5) {
                val raw = rawJson.opt(rawJson.length() - 1)
                if (raw is String && (raw.trim().startsWith("{") || raw.trim().startsWith("["))) {
                    return try {
                        val trimmed = raw.trim()
                        if (trimmed.startsWith("{")) JSONObject(trimmed) else JSONArray(trimmed)
                    } catch (e: Exception) {
                        raw
                    }
                }
                return raw
            }
            return null
        }

    // Helper utilities for handlers
    fun getContentsAsObject(): JSONObject? = contents as? JSONObject
    fun getContentsAsArray(): JSONArray? = contents as? JSONArray
    fun getContentsAsString(): String? = contents as? String

    companion object {
        fun parse(notificationData: String, fromPush: Boolean): PushData {
            return PushData(JSONArray(notificationData), fromPush)
        }
    }
}

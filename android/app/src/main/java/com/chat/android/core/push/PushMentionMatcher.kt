package com.chat.android.core.push

import com.chat.android.feature.channel.ChannelDbHelper
import org.json.JSONArray

/** Matches a message mention against the current user's alias and group membership. */
object PushMentionMatcher {
    fun mentionsCurrentUser(message: String, channelID: String, dbHelper: ChannelDbHelper): Boolean {
        val myname = dbHelper.getChannel(channelID)?.myname.orEmpty()
        if (myname.isBlank()) return false
        if (message.contains("＠＠$myname・＠＠")) return true
        if (message.contains("＠＠channel・＠＠")) return true

        return dbHelper.getGroupsForChannel(channelID).any { group ->
            val mentioned = message.contains("＠＠${group.groupName}・＠＠")
            if (!mentioned) return@any false
            val members = runCatching { JSONArray(group.aliasNamesJson) }.getOrNull() ?: return@any false
            (0 until members.length()).any { index -> members.optString(index) == myname }
        }
    }
}

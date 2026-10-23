package com.chat.android.feature.thread

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.core.util.Base62
import com.chat.android.core.util.formatThreadTime
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbThread
import org.json.JSONArray

/**
 * thread / threadEdit push を処理する。
 */
class ThreadHandler(
    private val context: Context,
    private val profileStore: com.chat.android.feature.channel.ChannelProfileStore
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val channelID = pd.channelID
        val updatedBy = pd.updatedBy
        val array = pd.getContentsAsArray() ?: return

        val parentID = array.optString(0, "")
        val messageID = array.optString(1, "")
        if (messageID.isEmpty()) return

        val dbHelper = ChannelDbHelper(context)

        if (pd.title == "thread") {
            if (dbHelper.getThread(messageID) != null) return

            val messageTxt = array.optString(2, "") + imageAwareFileLinks(pd.rawJson.optJSONArray(5))
            val incomingUpdatedUnixAt = array.optLong(8, 0L)

            dbHelper.saveThread(
                DbThread(
                    messageID = messageID,
                    parentID = parentID,
                    channelID = channelID,
                    messageTxt = messageTxt,
                    aliasName = array.optString(7, "").ifBlank { updatedBy },
                    aliasImg = array.optString(3, ""),
                    aliasNamesJson = array.optJSONArray(4)?.toString() ?: "[]",
                    backID = array.optString(5, ""),
                    emojisJson = array.optJSONArray(6)?.toString() ?: "[]",
                    createdAt = createdAtOf(messageID),
                    isBookmarked = false,
                    updatedUnixAt = incomingUpdatedUnixAt
                )
            )
            Log.i("ThreadHandler", "Upserted thread $messageID")

            val myname = profileStore.profile(channelID).first
            val isDm = parentID.contains('@')
            val isDmTarget = isDm && parentID.split('@').contains(myname) && updatedBy != myname

            val mention = ThreadMentionMatcher.mentionsCurrentUser(messageTxt, channelID, myname, dbHelper) || isDmTarget
            val displayStatus = if (mention) 2 else 1
            dbHelper.getThreadHead(parentID)?.let { head ->
                dbHelper.saveThreadHead(head.copy(displayStatus = displayStatus))
            }

            if (mention) {
                val emojis = array.optJSONArray(6)
                if (emojis != null && emojis.length() > 0) {
                    dbHelper.getThread(messageID)?.let {
                        dbHelper.saveThread(it.copy(isBookmarked = true))
                    }
                }
            }
        } else if (pd.title == "threadEdit") {
            val existing = dbHelper.getThread(messageID)
            val messageTxt = array.optString(2, "")
            val incomingUpdatedUnixAt = array.optLong(8, 0L)

            if (messageTxt.isEmpty()) {
                dbHelper.deleteThread(messageID)
                return
            }

            if (existing == null) return
            if (existing.updatedUnixAt > incomingUpdatedUnixAt && incomingUpdatedUnixAt != 0L) return

            val newTxt = messageTxt + imageAwareFileLinks(pd.rawJson.optJSONArray(5))
            dbHelper.saveThread(
                existing.copy(
                    messageTxt = newTxt,
                    aliasImg = array.optString(3, "").ifBlank { existing.aliasImg },
                    aliasNamesJson = array.optJSONArray(4)?.toString() ?: existing.aliasNamesJson,
                    emojisJson = mergeEmojis(existing.emojisJson, array.optJSONArray(6)),
                    isBookmarked = array.optBoolean(8, existing.isBookmarked),
                    updatedUnixAt = incomingUpdatedUnixAt
                )
            )
            Log.i("ThreadHandler", "Edited thread $messageID")

            val myname = profileStore.profile(channelID).first
            val isDm = parentID.contains('@')
            val isDmTarget = isDm && parentID.split('@').contains(myname) && updatedBy != myname
            val isMention = ThreadMentionMatcher.mentionsCurrentUser(newTxt, channelID, myname, dbHelper) || isDmTarget
            if (isMention) {
                dbHelper.getThreadHead(parentID)?.let { head ->
                    dbHelper.saveThreadHead(head.copy(displayStatus = 2))
                }
                
                val incomingEmojis = array.optJSONArray(6)
                if (incomingEmojis != null && incomingEmojis.length() > 0) {
                    dbHelper.getThread(messageID)?.let {
                        dbHelper.saveThread(it.copy(isBookmarked = true))
                    }
                }
            }
        }
    }

    private fun imageAwareFileLinks(files: JSONArray?): String = buildLinks(files)

    private fun isImageFile(link: String): Boolean {
        val path = link.substringBefore('?').lowercase()
        return listOf(".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp")
            .any { path.endsWith(it) }
    }

    private fun buildLinks(files: JSONArray?): String {
        if (files == null) return ""
        var out = ""
        for (i in 0 until files.length()) {
            val link = files.optString(i, "")
            if (link.isEmpty()) continue
            val m = if (isImageFile(link)) "img" else "f"
            out += "＊${m}＊${link}・＊${m}＊ "
        }
        return out
    }

    private fun createdAtOf(messageID: String): String {
        val body = if (messageID.length > 1) messageID.dropLast(1) else messageID
        return formatThreadTime(Base62.decode(body))
    }

    private fun mergeEmojis(existingJson: String, incoming: JSONArray?): String {
        val result = JSONArray()
        val seen = mutableSetOf<String>()
        fun add(array: JSONArray?) {
            if (array == null) return
            for (i in 0 until array.length()) {
                val obj = array.optJSONObject(i) ?: continue
                val name = obj.optString("aliasName", "")
                if (seen.add(name)) result.put(obj)
            }
        }
        add(runCatching { JSONArray(existingJson) }.getOrNull())
        add(incoming)
        return result.toString()
    }
}

package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper
import org.json.JSONArray

/**
 * emoji push を処理する（vue/src/pushReceive/emoji.js に対応）。
 *
 * contents = [messageID, emoji, parentID, del]
 * messageID == parentID の場合は threadHead、それ以外は thread に反映する。
 */
class EmojiHandler(private val context: Context) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val aliasName = pd.updatedBy
        val array = pd.getContentsAsArray() ?: return

        val messageID = array.optString(0, "")
        val emoji = array.optString(1, "")
        val parentID = array.optString(2, "")
        val del = array.optBoolean(3, false)
        if (messageID.isEmpty()) return

        val dbHelper = ChannelDbHelper(context)

        // threadHead 自身へのリアクションかどうか
        if (messageID == parentID) {
            val head = dbHelper.getThreadHead(messageID) ?: return
            val updated = applyEmoji(head.emojisJson, aliasName, emoji, del)
            dbHelper.saveThreadHead(head.copy(emojisJson = updated))
            Log.i("EmojiHandler", "Emoji updated on threadHead $messageID")
            return
        }

        val thread = dbHelper.getThread(messageID) ?: return
        val updated = applyEmoji(thread.emojisJson, aliasName, emoji, del)
        dbHelper.saveThread(thread.copy(emojisJson = updated))
        Log.i("EmojiHandler", "Emoji updated on thread $messageID")
    }

    /**
     * リアクションの追加・削除。同じエイリアスが同じ絵文字を重複して持たないようにする
     * （vue の addEmoji と同じ）。
     */
    private fun applyEmoji(existingJson: String, aliasName: String, emoji: String, del: Boolean): String {
        val result = JSONArray()
        val current = runCatching { JSONArray(existingJson) }.getOrNull()
        // 同じエイリアス・同じ絵文字が既にあるか（追加時の重複防止用）
        var exists = false
        for (i in 0 until (current?.length() ?: 0)) {
            val obj = current!!.optJSONObject(i) ?: continue
            val sameAlias = obj.optString("aliasName", "") == aliasName
            val sameEmoji = obj.optString("emoji", "") == emoji
            if (sameAlias && sameEmoji) {
                if (del) continue
                exists = true
            }
            result.put(obj)
        }
        if (!del && !exists && emoji.isNotEmpty()) {
            val obj = org.json.JSONObject()
            obj.put("aliasName", aliasName)
            obj.put("emoji", emoji)
            result.put(obj)
        }
        return result.toString()
    }
}
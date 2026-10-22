package com.chat.android.core.push.handlers

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
 * thread / threadEdit push を処理する（vue/src/pushReceive/thread.js, threadEdit.js に対応）。
 *
 * contents は位置配列:
 *   [0]=parentID [1]=messageID [2]=messageTxt [3]=aliasImg
 *   [4]=aliasNames [5]=backID [6]=emojis [7]=aliasName （thread のみ）
 * 添付ファイルは pd[5] に配列で来る。
 */
class ThreadHandler(private val context: Context) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val channelID = pd.channelID
        val updatedBy = pd.updatedBy
        val array = pd.getContentsAsArray() ?: return

        val parentID = array.optString(0, "")
        val messageID = array.optString(1, "")
        if (messageID.isEmpty()) return

        val dbHelper = ChannelDbHelper(context)

        if (pd.title == "thread") {
            // 既存なら何もしない（vue と同じ重複ガード）
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

            // メンション判定と displayStatus 更新 (thread.js と同様)
            // DM の場合は自分以外からの投稿ならメンション扱いで displayStatus=2 とする。
            val myname = dbHelper.getChannel(channelID)?.myname.orEmpty()
            val isDm = parentID.contains('@')
            val isDmTarget = isDm && parentID.split('@').contains(myname) && updatedBy != myname

            val mention = ThreadMentionMatcher.mentionsCurrentUser(messageTxt, channelID, dbHelper) || isDmTarget
            val displayStatus = if (mention) 2 else 1
            dbHelper.getThreadHead(parentID)?.let { head ->
                dbHelper.saveThreadHead(head.copy(displayStatus = displayStatus))
            }

            // 自動ブックマーク (メンションかつ絵文字がある場合)
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

            // messageTxt が空なら削除（vue と同じ）
            if (messageTxt.isEmpty()) {
                dbHelper.deleteThread(messageID)
                return
            }

            if (existing == null) return

            // 順序制御: 既存データの updatedUnixAt の方が新しい場合は更新しない（vue と同じ）
            // 同時刻または incoming が 0 の場合は更新を許可する
            if (existing.updatedUnixAt > incomingUpdatedUnixAt && incomingUpdatedUnixAt != 0L) {
                return
            }

            val newTxt = messageTxt + imageAwareFileLinks(pd.rawJson.optJSONArray(5))
            dbHelper.saveThread(
                existing.copy(
                    // Keep image attachments renderable when an edited message is received.
                    messageTxt = newTxt,
                    aliasImg = array.optString(3, "").ifBlank { existing.aliasImg },
                    aliasNamesJson = array.optJSONArray(4)?.toString() ?: existing.aliasNamesJson,
                    emojisJson = mergeEmojis(existing.emojisJson, array.optJSONArray(6)),
                    isBookmarked = array.optBoolean(8, existing.isBookmarked), // Note: array index might vary between thread and threadEdit, verifying...
                    updatedUnixAt = incomingUpdatedUnixAt
                )
            )
            Log.i("ThreadHandler", "Edited thread $messageID")

            // メンションチェックとブックマーク自動登録 (vue/src/pushReceive/threadEdit.js の後半部分)
            val myname = dbHelper.getChannel(channelID)?.myname.orEmpty()
            val isDm = parentID.contains('@')
            val isDmTarget = isDm && parentID.split('@').contains(myname) && updatedBy != myname
            val isMention = ThreadMentionMatcher.mentionsCurrentUser(newTxt, channelID, dbHelper) || isDmTarget
            if (isMention) {
                // displayStatus を 2 (メンション) に更新
                dbHelper.getThreadHead(parentID)?.let { head ->
                    dbHelper.saveThreadHead(head.copy(displayStatus = 2))
                }
                
                // ブックマーク自動登録 (emojis がある場合)
                val incomingEmojis = array.optJSONArray(6)
                if (incomingEmojis != null && incomingEmojis.length() > 0) {
                    dbHelper.getThread(messageID)?.let {
                        dbHelper.saveThread(it.copy(isBookmarked = true))
                    }
                }
            }
        }
    }

    /**
     * 添付ファイルリンク。vue の thread.js と同じで、画像は ＊img＊、それ以外は ＊f＊。
     */
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

    /** 既存と更新の絵文字を aliasName で重複なく統合する。 */
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

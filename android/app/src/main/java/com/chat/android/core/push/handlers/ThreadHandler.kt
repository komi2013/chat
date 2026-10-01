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
                    createdAt = createdAtOf(messageID)
                )
            )
            Log.i("ThreadHandler", "Upserted thread $messageID")

            // 新着なので親スレッドを「未読」にする（vue の thread.js と同じ）。
            // メンションされた場合は 2、それ以外は 1。
            // なお displayStatus を 0 にする処理はここではなく、
            // ユーザーがスレッドを開いたときの readStatus()（Thread.vue）が行う。
            val myname = dbHelper.getChannel(channelID)?.myname.orEmpty()
            val mention = myname.isNotEmpty() && messageTxt.contains("＠＠$myname・＠＠")
            val displayStatus = if (mention) 2 else 1
            dbHelper.getThreadHead(parentID)?.let { head ->
                dbHelper.saveThreadHead(head.copy(displayStatus = displayStatus))
            }
        } else if (pd.title == "threadEdit") {
            val existing = dbHelper.getThread(messageID)
            val messageTxt = array.optString(2, "")
            // messageTxt が空なら削除（vue と同じ）
            if (messageTxt.isEmpty()) {
                dbHelper.deleteThread(messageID)
                return
            }
            if (existing == null) return

            dbHelper.saveThread(
                existing.copy(
                    // threadEdit は添付を一律 ＊f＊ で書く（vue の threadEdit.js と同じ）
                    messageTxt = messageTxt + plainFileLinks(pd.rawJson.optJSONArray(5)),
                    aliasImg = array.optString(3, "").ifBlank { existing.aliasImg },
                    aliasNamesJson = array.optJSONArray(4)?.toString() ?: existing.aliasNamesJson,
                    emojisJson = mergeEmojis(existing.emojisJson, array.optJSONArray(6))
                )
            )
            Log.i("ThreadHandler", "Edited thread $messageID")
        }
    }

    /**
     * 添付ファイルリンク。vue の thread.js と同じで、画像は ＊img＊、それ以外は ＊f＊。
     */
    private fun imageAwareFileLinks(files: JSONArray?): String = buildLinks(files) { link ->
        val lower = link.lowercase()
        if (lower.endsWith(".jpg") || lower.endsWith(".jpeg") || lower.endsWith(".png")) "img" else "f"
    }

    /**
     * 添付ファイルリンク。vue の threadEdit.js と同じで、
     * 拡張子に関係なく一律 ＊f＊ で書く。
     */
    private fun plainFileLinks(files: JSONArray?): String = buildLinks(files) { "f" }

    private fun buildLinks(files: JSONArray?, marker: (String) -> String): String {
        if (files == null) return ""
        var out = ""
        for (i in 0 until files.length()) {
            val link = files.optString(i, "")
            if (link.isEmpty()) continue
            val m = marker(link)
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
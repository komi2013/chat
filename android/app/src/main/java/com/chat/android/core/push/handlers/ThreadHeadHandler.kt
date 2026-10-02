package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbThreadHead

/**
 * threadHead push を処理する（vue/src/pushReceive/threadHead.js に対応）。
 *
 * contents は threadHead オブジェクト本体（配列ではない）。
 */
class ThreadHeadHandler(private val context: Context) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val channelID = pd.channelID
        val updatedBy = pd.updatedBy
        val obj = pd.getContentsAsObject() ?: return

        val parentID = obj.optString("parentID", "")
        if (parentID.isEmpty()) return

        val dbHelper = ChannelDbHelper(context)
        val myname = dbHelper.getChannel(channelID)?.myname.orEmpty()

        // タイトルを DM の相手名に書き換える（vue と同じ）。
        //   parentID が "@グループ名"    → 自分が更新者でなければグループ名をタイトルに
        //   parentID が "自分@相手" 形式 → 相手側の名前をタイトルに
        var title = obj.optString("title", "")
        if (parentID.startsWith('@')) {
            val nameTail = parentID.substring(1)
            if (myname.isNotEmpty() && myname != updatedBy) {
                title = nameTail
            }
        } else if (parentID.contains('@')) {
            val namePre = parentID.substringBefore('@')
            val nameTail = parentID.substringAfter('@')
            if (myname == namePre) {
                title = nameTail
            } else if (myname == nameTail) {
                title = namePre
            }
        }

        var displayStatus = obj.optInt("displayStatus", 0)
        if (myname.isNotEmpty() && myname == updatedBy) {
            displayStatus = 1
        }

        val messageTxt = obj.optString("messageTxt", "") + imageAwareFileLinks(pd.rawJson.optJSONArray(5))

        dbHelper.saveThreadHead(
            DbThreadHead(
                parentID = parentID,
                channelID = obj.optString("channelID", channelID),
                title = title,
                messageTxt = messageTxt,
                description = obj.optString("description", ""),
                aliasName = obj.optString("aliasName", updatedBy),
                aliasNamesJson = obj.optJSONArray("aliasNames")?.toString() ?: "[]",
                adminNamesJson = obj.optJSONArray("adminNames")?.toString() ?: "[]",
                displayStatus = displayStatus,
                broadcastFlag = obj.optInt("broadcastFlag", 0),
                updatedAt = obj.optString("updatedAt", ""),
                emojisJson = obj.optJSONArray("emojis")?.toString() ?: "[]",
                backID = obj.optString("backID", ""),
                isBookmarked = obj.optBoolean("isBookmarked", false)
            )
        )
        Log.i("ThreadHeadHandler", "Upserted threadHead $parentID")
    }

    private fun imageAwareFileLinks(files: org.json.JSONArray?): String {
        if (files == null) return ""
        var out = ""
        for (i in 0 until files.length()) {
            val link = files.optString(i, "")
            if (link.isEmpty()) continue
            val path = link.substringBefore('?').lowercase()
            val m = if (listOf(".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp").any { path.endsWith(it) }) "img" else "f"
            out += "＊${m}＊${link}・＊${m}＊ "
        }
        return out
    }
}

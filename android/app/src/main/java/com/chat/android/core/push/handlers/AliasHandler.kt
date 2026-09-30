package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbAlias

/**
 * alias イベントのハンドラ。
 *
 * サーバーが送るワイヤ形式は **位置配列**（vue/src/pushReceive/alias.js と同一）:
 *   pd = [pushID, "alias", channelID, updatedBy,
 *         [userID, aliasName, aliasBio, accessRight], aliasImg]
 *
 * 旧実装は contents を JSONObject として読んでいたため、配列が常に null 扱いになり
 * alias がローカルDBへ一切保存されていなかった（= PUSHでDB同期する仕様が動かない原因）。
 */
class AliasHandler(
    private val context: Context
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val dbHelper = ChannelDbHelper(context)
        val channelID = pd.channelID

        val contents = pd.getContentsAsArray()
        if (contents == null) {
            Log.e("AliasHandler", "contents is not a positional array: ${pd.contents}")
            return
        }

        val userID = contents.optString(0, "")
        val aliasName = contents.optString(1, "")
        val aliasBio = contents.optString(2, "")
        val accessRight = contents.optString(3, "")
        val aliasImg = pd.rawJson.optString(5, "")

        // aliasID = channelID + (accessRight=='inquirer'?'@':'') + aliasName
        val atMark = if (accessRight == "inquirer") "@" else ""
        val aliasID = "$channelID$atMark$aliasName"

        if (accessRight == "delete") {
            dbHelper.deleteAlias(aliasID)
            Log.i("AliasHandler", "Deleted alias $aliasID")
            return
        }

        dbHelper.saveAlias(
            DbAlias(
                aliasID = aliasID,
                channelID = channelID,
                aliasName = "$atMark$aliasName",
                aliasImg = aliasImg,
                userID = userID,
                aliasBio = aliasBio,
                accessRight = accessRight
            )
        )
        Log.i("AliasHandler", "Upserted alias $aliasID")
    }
}

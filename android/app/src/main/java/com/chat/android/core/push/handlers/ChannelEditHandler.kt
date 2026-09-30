package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbChannel

/**
 * channelEdit イベントのハンドラ。
 *
 * ワイヤ形式（vue/src/pushReceive/channelEdit.js と同一）:
 *   pd = [pushID, "channelEdit", channelID, updatedBy, [channelName, channelDescription]]
 *   pd[4] == "delete" のときはチャネル削除。
 *
 * 旧実装は contents を JSONObject として読んでいたため、配列が常に null 扱いになり
 * 新規チャネルがローカルDBへ保存されなかった。
 */
class ChannelEditHandler(
    private val context: Context
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val dbHelper = ChannelDbHelper(context)
        val channelID = pd.channelID
        val contents = pd.contents

        if (contents is String && contents == "delete") {
            dbHelper.deleteChannel(channelID)
            Log.i("ChannelEditHandler", "Deleted channel $channelID")
            return
        }

        val array = pd.getContentsAsArray()
        if (array == null) {
            Log.e("ChannelEditHandler", "contents is not a positional array: $contents")
            return
        }

        val channelName = array.optString(0, "")
        val channelDescription = array.optString(1, "")

        // Vue と同じく既存レコードは差分更新する（invitationCode 等のローカル値を消さない）。
        val existing = dbHelper.getChannel(channelID)
        val channel = if (existing != null) {
            existing.copy(
                channelName = channelName,
                channelDescription = channelDescription
            )
        } else {
            DbChannel(
                channelID = channelID,
                channelName = channelName,
                channelDescription = channelDescription,
                myname = pd.updatedBy,
                myimg = "",
                displayStatus = 0,
                invitationCode = "",
                invitationGuestCode = ""
            )
        }

        dbHelper.saveChannel(channel)
        Log.i("ChannelEditHandler", "Upserted channel $channelID")
    }
}

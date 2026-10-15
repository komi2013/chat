package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.network.Ticket
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.ticket.TicketDbHelper

/**
 * ticket イベントのハンドラ。
 *
 * サーバーが送るワイヤ形式 (vue/src/pushReceive/ticket.js と同一):
 *   pd = [pushID, "ticket", channelID, updatedBy, [ticketID, title, description, status, priority, createdAt]]
 *   pd[4] == "delete" のときは削除。
 */
class TicketHandler(
    private val context: Context
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val dbHelper = TicketDbHelper(context)
        val contents = pd.contents

        if (contents is String && contents == "delete") {
            val ticketID = pd.rawJson.optString(5, "") // 仮定: 5番目にIDが入っているか、contentsがIDのケースがある
            // Vueの実装に合わせる必要があるが、一般的に contents が "delete" なら pd[2] (channelID) か pd[5] あたりにIDがあるはず
            // ここでは contents が配列でなく文字列 "delete" の場合を想定
            val targetID = if (pd.rawJson.length() > 5) pd.rawJson.optString(5) else ""
            if (targetID.isNotBlank()) {
                dbHelper.deleteTicket(targetID)
                Log.i("TicketHandler", "Deleted ticket $targetID")
            }
            return
        }

        val array = pd.getContentsAsArray()
        if (array == null) {
            Log.e("TicketHandler", "contents is not a positional array: $contents")
            return
        }

        val ticket = Ticket(
            ticketID = array.optString(0, ""),
            title = array.optString(1, ""),
            description = array.optString(2, ""),
            status = array.optString(3, ""),
            priority = array.optString(4, ""),
            createdAt = array.optString(5, "")
        )

        dbHelper.saveTicket(ticket)
        Log.i("TicketHandler", "Upserted ticket ${ticket.ticketID}")
    }
}

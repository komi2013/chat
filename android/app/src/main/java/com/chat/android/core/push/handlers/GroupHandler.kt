package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbGroup
import org.json.JSONArray
import org.json.JSONObject

class GroupHandler(
    private val context: Context
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val dbHelper = ChannelDbHelper(context)
        val channelID = pd.channelID
        
        // contents=[groupName, aliasNames, groupID, groupImg, groupBio]
        // Wire format from Go/Vue often uses positional arrays for group push
        val contents = pd.contents
        
        if (contents is String && contents == "delete") {
             // If extra or some other field has groupID, we'd delete it.
             // Usually groupID is in the contents array or extra.
             val groupID = pd.extra as? String ?: return
             dbHelper.deleteGroup(groupID)
             return
        }

        val array = pd.getContentsAsArray() ?: return

        // サーバーが送る group push（controller/ChannelEdit.go:387）は
        //   pd = [pushID, "group", channelID, updatedBy, [groupName, aliasNames], groupImg]
        // つまり contents には groupName と aliasNames の **2要素しかない**。
        // groupID を contents から読むと必ず空になり、処理全体が
        // 旧来の if (groupID.isEmpty()) return で打ち切られていた（＝グループが
        // ローカルDBに保存されない原因）。
        // vue/src/pushReceive/group.js と同じく channelID + groupName で合成する。
        val groupName = array.optString(0, "")
        val aliasNames = array.optJSONArray(1)?.toString() ?: "[]"
        val groupImg = pd.rawJson.optString(5, "")
        // groupBio は group push に含まれない（サーバーも送信していない）

        if (groupName.isEmpty()) return

        val groupID = channelID + groupName

        val group = DbGroup(
            groupID = groupID,
            channelID = channelID,
            groupName = groupName,
            groupImg = groupImg,
            aliasNamesJson = aliasNames,
            groupBio = ""
        )

        dbHelper.saveGroup(group)
        Log.i("GroupHandler", "Upserted group $groupID")
    }
}

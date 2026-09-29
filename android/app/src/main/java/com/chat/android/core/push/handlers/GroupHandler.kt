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
        
        // Example mapping based on Vue group.js:
        // contents: [groupName, aliasNames, groupID, groupImg, groupBio]
        val groupName = array.optString(0, "")
        val aliasNames = array.optJSONArray(1)?.toString() ?: "[]"
        val groupID = array.optString(2, "")
        val groupImg = pd.rawJson.optString(5, "") // imgPath is often at index 5 in pd
        val groupBio = array.optString(4, "")

        if (groupID.isEmpty()) return

        val group = DbGroup(
            groupID = groupID,
            channelID = channelID,
            groupName = groupName,
            groupImg = groupImg,
            aliasNamesJson = aliasNames,
            groupBio = groupBio
        )
        
        dbHelper.saveGroup(group)
        Log.i("GroupHandler", "Upserted group $groupID")
    }
}

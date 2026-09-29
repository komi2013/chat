package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbAlias
import org.json.JSONObject

class AliasHandler(
    private val context: Context
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val dbHelper = ChannelDbHelper(context)
        val channelID = pd.channelID
        
        val json = pd.getContentsAsObject() ?: return
        val aliasName = json.optString("aliasName", "")
        val accessRight = json.optString("accessRight", "")
        
        // aliasID = channelID + (accessRight=='inquirer'?'@':'') + aliasName
        val prefix = if (accessRight == "inquirer") "@" else ""
        val aliasID = "$channelID$prefix$aliasName"

        if (accessRight == "delete") {
            dbHelper.deleteAlias(aliasID)
            Log.i("AliasHandler", "Deleted alias $aliasID")
            return
        }

        val alias = DbAlias(
            aliasID = aliasID,
            channelID = channelID,
            aliasName = aliasName,
            aliasImg = json.optString("aliasImg", ""),
            userID = json.optString("userID", ""),
            aliasBio = json.optString("aliasBio", ""),
            accessRight = accessRight
        )
        
        dbHelper.saveAlias(alias)
        Log.i("AliasHandler", "Upserted alias $aliasID")
    }
}

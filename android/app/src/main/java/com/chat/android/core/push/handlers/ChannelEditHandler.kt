package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbChannel
import org.json.JSONObject

class ChannelEditHandler(
    private val context: Context
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val dbHelper = ChannelDbHelper(context)
        val channelID = pd.channelID
        val contents = pd.contents

        if (contents == "delete") {
            dbHelper.deleteChannel(channelID)
            Log.i("ChannelEditHandler", "Deleted channel $channelID")
            return
        }

        val json = pd.getContentsAsObject() ?: return
        
        // The backend often returns the whole channel object in contents for channelEdit
        // or a partial update. Vue's channelEdit.js upserts based on these fields:
        val channel = DbChannel(
            channelID = channelID,
            channelName = json.optString("channelName", ""),
            channelDescription = json.optString("channelDescription", ""),
            myname = pd.updatedBy, // In Vue, myname is often set to updatedBy in this context
            myimg = json.optString("myimg", ""),
            displayStatus = json.optInt("displayStatus", 0),
            invitationCode = json.optString("invitationCode", ""),
            invitationGuestCode = json.optString("invitationGuestCode", "")
        )
        
        dbHelper.saveChannel(channel)
        Log.i("ChannelEditHandler", "Upserted channel $channelID")
    }
}

package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.channel.ChannelDbHelper

/**
 * bookmark push を処理する（vue/src/pushReceive/bookmark.js に対応）。
 *
 * contents = [messageID, parentID, isBookmarked]
 * messageID == parentID の場合は threadHead、それ以外は thread に反映する。
 */
class BookmarkHandler(private val context: Context) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val array = pd.getContentsAsArray() ?: return

        val messageID = array.optString(0, "")
        val parentID = array.optString(1, "")
        val isBookmarked = array.optBoolean(2, false)
        
        if (messageID.isEmpty()) return

        val dbHelper = ChannelDbHelper(context)

        // threadHead 自身へのブックマークかどうか
        if (messageID == parentID) {
            val head = dbHelper.getThreadHead(messageID) ?: return
            dbHelper.saveThreadHead(head.copy(isBookmarked = isBookmarked))
            Log.i("BookmarkHandler", "Bookmark updated on threadHead $messageID to $isBookmarked")
            return
        }

        val thread = dbHelper.getThread(messageID) ?: return
        dbHelper.saveThread(thread.copy(isBookmarked = isBookmarked))
        Log.i("BookmarkHandler", "Bookmark updated on thread $messageID to $isBookmarked")
    }
}

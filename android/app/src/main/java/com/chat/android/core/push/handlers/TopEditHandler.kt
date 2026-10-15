package com.chat.android.core.push.handlers

import android.util.Log
import com.chat.android.core.network.model.TopLink
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.core.repository.UserRepository

/**
 * topEdit イベントのハンドラ。
 *
 * サーバーが送るワイヤ形式 (vue/src/pushReceive/topEdit.js と同一):
 *   pd = [pushID, "topEdit", channelID, updatedBy, [[topText, topLink], ...]]
 */
class TopEditHandler(
    private val userRepository: UserRepository
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val channelID = pd.channelID
        val contents = pd.getContentsAsArray()

        if (contents == null) {
            Log.e("TopEditHandler", "contents is not a JSONArray: ${pd.contents}")
            return
        }

        val topLinks = mutableListOf<TopLink>()
        for (i in 0 until contents.length()) {
            val item = contents.optJSONArray(i)
            if (item != null) {
                topLinks.add(
                    TopLink(
                        topText = item.optString(0, ""),
                        topLink = item.optString(1, "")
                    )
                )
            }
        }

        userRepository.setStoredTopLinks(channelID, topLinks)
        Log.i("TopEditHandler", "Updated top links for channel $channelID: ${topLinks.size} items")
    }
}

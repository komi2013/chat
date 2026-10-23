package com.chat.android.feature.thread

import android.content.Context
import android.net.Uri
import androidx.core.net.toUri
import com.chat.android.core.push.PushData
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.firebase.PushNotificationManager
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ThreadNotificationProcessor @Inject constructor(
    @ApplicationContext private val context: Context,
    private val notificationManager: PushNotificationManager,
    private val profileStore: com.chat.android.feature.channel.ChannelProfileStore
) {
    suspend fun processNotification(pd: PushData) {
        // This processor is specialized for "thread" events
        if (pd.title != "thread") return

        val array = pd.getContentsAsArray() ?: return
        val extra = pd.extra as? org.json.JSONObject
        
        // 1. Handle Incoming Call (threadCall)
        val callUrl = extra?.optString("url")?.takeIf { it.isNotBlank() }
        if (callUrl != null) {
            val segments = callUrl.toUri().pathSegments
            val callPathIndex = segments.indexOfLast { it == "threadCall" }
            val callCode = segments.getOrNull(callPathIndex + 1).orEmpty()
            if (callCode.isNotBlank()) {
                notificationManager.showIncomingCall(pd.updatedBy, callCode)
                return
            }
        }

        // 2. Handle Mention / DM Notification
        val channelID = pd.channelID
        val parentID = array.optString(0)
        val messageID = array.optString(1)
        val message = array.optString(2)
        val updatedBy = pd.updatedBy
        if (channelID.isBlank() || messageID.isBlank()) return

        val db = ChannelDbHelper(context)
        val myname = profileStore.profile(channelID).first

        // DM Detection: parentID contains '@', I am one of the parties, and I am not the sender.
        val isDm = parentID.contains('@')
        val isDmTarget = isDm && parentID.split('@').contains(myname) && updatedBy != myname

        if (!isDmTarget && !ThreadMentionMatcher.mentionsCurrentUser(message, channelID, myname, db)) return

        val author = array.optString(7).ifBlank { updatedBy }
        val title = if (isDmTarget) "DM from $author" else "$author mentioned you"
        
        // Strip mention tags and clean up whitespace for the notification body
        val body = message
            .replace(Regex("＠＠[^・＠]+・＠＠"), "")
            .replace(Regex("\\s+"), " ")
            .trim()
            .take(120)
            .ifBlank { if (isDmTarget) "New direct message" else "You were mentioned in a group" }

        notificationManager.showNotification(
            title = title,
            body = body,
            channelId = channelID,
            collapseKey = "mention|$messageID"
        )
    }
}

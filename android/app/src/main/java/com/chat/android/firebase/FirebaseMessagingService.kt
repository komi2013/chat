package com.chat.android.firebase

import android.util.Log
import android.net.Uri
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import com.chat.android.core.push.PushReceiveDispatcher
import com.chat.android.core.push.handlers.ThreadMentionMatcher
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.entryform.EntryFormCodec
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import dagger.hilt.android.AndroidEntryPoint
import org.json.JSONArray
import javax.inject.Inject

@AndroidEntryPoint
class FirebaseMessagingService : com.google.firebase.messaging.FirebaseMessagingService() {
    
    @Inject
    lateinit var pushNotificationManager: PushNotificationManager

    @Inject
    lateinit var pushDispatcher: PushReceiveDispatcher
    
    override fun onNewToken(token: String) {
        super.onNewToken(token)
        Log.d(TAG, "FCM Token: $token")
        
        // Note: CSRF token should be retrieved from user session
        // This is typically done after user logs in
        // For now, we'll store the token locally and send it after login
        storeFCMTokenLocally(token)
    }
    
    private fun storeFCMTokenLocally(token: String) {
        // Store token in SharedPreferences or local storage
        // This token will be sent to backend after user authentication
        val prefs = getSharedPreferences("chat_prefs", MODE_PRIVATE)
        prefs.edit().putString("fcm_token", token).apply()
        Log.d(TAG, "FCM Token stored locally")
    }
    
    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        super.onMessageReceived(remoteMessage)

        Log.d(TAG, "From: ${remoteMessage.from}")
        Log.d(TAG, "Message data payload: ${remoteMessage.data}")

        // Data-only pass-through: "payload" is the exact JSON array the web
        // client receives via VAPID. Dispatch on pd[1] like vue pushReceive.js —
        // the client decides importance/display, the server never maps events.
        // There is intentionally no handling of remoteMessage.notification:
        // the server no longer sends notification blocks.
        val payload = remoteMessage.data["payload"]
        if (payload.isNullOrEmpty()) {
            Log.w(TAG, "Message has no data payload; ignoring")
            return
        }

        // 1) 状態同期: vue/src/pushReceive/pushReceive.js と同じく、受信した payload を
        //    そのまま PushReceiveDispatcher に渡して channel / alias / group /
        //    entryForm を SQLite へ反映する。従来は API レスポンスの pushContents
        //    経由でしか同期されず、PUSH 受信だけではローカルDBが更新されなかった。
        pushDispatcher.receive(payload, fromPush = true)

        // Show a high priority alert for a direct mention or a group mention
        // when the signed-in alias belongs to that group.
        notifyForPayload(payload)
    }

    // Mirrors vue/src/pushReceive/pushReceive.js: parse the array
    // [pushID, event, channelID, updatedBy, ...] and dispatch on pd[1].
    //
    // alias / group / channelEdit は他端末の状態同期専用イベントなので通知しない
    // （Web の service-worker.js も sw 要素を持たない payload では通知しない。
    //   DB への反映は pushDispatcher 側で完了している）。
    // 通知するイベントは (event, channelID) 単位で1件に集約し、6セッション分の
    // 重複配信などで同じ通知が積み上がらないようにする。
    private fun notifyForPayload(payload: String) {
        val push = runCatching { JSONArray(payload) }.getOrNull() ?: return
        if (push.optString(1) != "thread") return
        val extra = push.opt(push.length() - 1) as? org.json.JSONObject
        val callUrl = extra?.optString("url")?.takeIf { !it.isNullOrBlank() } ?: run {
            notifyForMention(push)
            return
        }
        val segments = Uri.parse(callUrl).pathSegments
        val callPathIndex = segments.indexOfLast { it == "threadCall" }
        val callCode = segments.getOrNull(callPathIndex + 1).orEmpty()
        if (callCode.isNotBlank()) {
            // The push notification only names the caller; the web domain and URL
            // are used internally to find the shared SkyWay room.
            pushNotificationManager.showIncomingCall(push.optString(3), callCode)
        }
    }

    private fun notifyForMention(push: JSONArray) {
        val channelID = push.optString(2)
        val contents = push.optJSONArray(4) ?: return
        val parentID = contents.optString(0)
        val messageID = contents.optString(1)
        val message = contents.optString(2)
        val updatedBy = push.optString(3)
        if (channelID.isBlank() || messageID.isBlank()) return

        val db = ChannelDbHelper(this)
        val myname = db.getChannel(channelID)?.myname.orEmpty()

        // DM 判定: parentID に @ が含まれ、自分がその一人であり、かつ自分が送信者でない場合
        val isDm = parentID.contains('@')
        val isDmTarget = isDm && parentID.split('@').contains(myname) && updatedBy != myname

        if (!isDmTarget && !ThreadMentionMatcher.mentionsCurrentUser(message, channelID, db)) return

        val author = contents.optString(7).ifBlank { updatedBy }
        val title = if (isDmTarget) "DM from $author" else "$author mentioned you"
        val body = message
            .replace(Regex("＠＠[^・＠]+・＠＠"), "")
            .replace(Regex("\\s+"), " ")
            .trim()
            .take(120)
            .ifBlank { if (isDmTarget) "New direct message" else "You were mentioned in a group" }
        pushNotificationManager.showNotification(
            title = title,
            body = body,
            channelId = channelID,
            collapseKey = "mention|$messageID"
        )
    }

    /** 同じチャネルの同じイベントは同じ通知IDに上書きして1件に集約する。 */
    private fun notificationCollapseKey(event: String, channelID: String): String =
        "$event|$channelID"

    companion object {
        private const val TAG = "FCMService"

        // Events handled by the actions map in vue pushReceive.js.
        private val KNOWN_EVENTS = setOf(
            "advertisement", "alias", "answer", "bookmark", "calendar",
            "channelEdit", "chunk", "emoji", "entryForm", "group",
            "receptionOrder", "shiftStaffEdit", "storeSelect", "storeShare",
            "schedule", "thread", "threadEdit", "threadHead", "ticket",
            "timestamp", "timestampCode", "timestampReport", "timestampRevert",
            "topEdit", "tweetHead"
        )
    }
}

package com.chat.android.firebase

import android.util.Log
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import dagger.hilt.android.AndroidEntryPoint
import org.json.JSONArray
import org.json.JSONException
import javax.inject.Inject

@AndroidEntryPoint
class FirebaseMessagingService : FirebaseMessagingService() {
    
    @Inject
    lateinit var pushNotificationManager: PushNotificationManager
    
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
        dispatchPayload(payload)
    }

    // Mirrors vue/src/pushReceive/pushReceive.js: parse the array
    // [pushID, event, channelID, updatedBy, ...] and dispatch on pd[1].
    // Known events show a client-side notification; unknown events are
    // dropped, same as pushReceive.js logging "Unknown action.".
    private fun dispatchPayload(payload: String) {
        val pd = try {
            JSONArray(payload)
        } catch (e: JSONException) {
            Log.w(TAG, "Unparseable payload: $payload")
            return
        }

        val pushID = pd.optString(0, "")
        val event = pd.optString(1, "")
        val channelID = pd.optString(2, "")
        Log.d(TAG, "pushID=$pushID event=$event channelID=$channelID")

        // pushReceive.js special-cases pushCheck: it alerts pd[2] instead of
        // dispatching an action.
        if (event == "pushCheck") {
            pushNotificationManager.showNotification(
                "Push登録完了",
                pd.optString(2, "通知の登録が完了しました"),
                ""
            )
            return
        }

        val (title, body) = when (event) {
            // chunk hides the channel id — pd[2] is the chunk body there.
            "chunk", "thread", "threadHead" -> "新着メッセージ" to "新しい投稿があります"
            "channelEdit" -> "チャンネル更新" to "チャンネル情報が更新されました"
            "alias" -> "メンバー更新" to "メンバー情報が更新されました"
            "group" -> "グループ更新" to "グループ情報が更新されました"
            in KNOWN_EVENTS -> "新着通知" to event
            else -> {
                Log.d(TAG, "Unknown action: $event")
                return
            }
        }

        pushNotificationManager.showNotification(
            title,
            body,
            if (event == "chunk") "" else channelID
        )
    }

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
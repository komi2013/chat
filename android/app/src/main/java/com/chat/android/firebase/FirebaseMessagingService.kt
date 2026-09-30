package com.chat.android.firebase

import android.util.Log
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import com.chat.android.core.push.PushReceiveDispatcher
import com.chat.android.feature.entryform.EntryFormCodec
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

        // 2) 通知: ユーザーに見せる意味のあるイベントのみ表示する。
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

        when (event) {
            // pushReceive.js special-cases pushCheck: it alerts pd[2] instead of
            // dispatching an action. 登録確認は毎回表示する。
            "pushCheck" -> pushNotificationManager.showNotification(
                "Push登録完了",
                pd.optString(2, "通知の登録が完了しました"),
                ""
            )

            "entryForm" -> {
                val formTitle = pd.optJSONObject(4)
                    ?.toString()
                    ?.let { runCatching { EntryFormCodec.parse(it) }.getOrNull() }
                    ?.title
                    ?.takeIf(String::isNotBlank)
                    ?: "入力フォームが更新されました"
                pushNotificationManager.showNotification(
                    "フォーム更新",
                    formTitle,
                    channelID,
                    collapseKey = notificationCollapseKey(event, channelID)
                )
            }

            // chunk hides the channel id — pd[2] is the chunk body there.
            "chunk", "thread", "threadHead" -> pushNotificationManager.showNotification(
                "新着メッセージ",
                "新しい投稿があります",
                if (event == "chunk") "" else channelID,
                collapseKey = notificationCollapseKey(event, channelID)
            )

            // 状態同期イベント（ローカルDBは pushDispatcher が更新済み）
            "alias", "group", "channelEdit" ->
                Log.i(TAG, "State sync event; notification suppressed: $event")

            in KNOWN_EVENTS -> pushNotificationManager.showNotification(
                "新着通知",
                event,
                channelID,
                collapseKey = notificationCollapseKey(event, channelID)
            )

            else -> Log.d(TAG, "Unknown action: $event")
        }
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
package com.chat.android.firebase

import android.util.Log
import android.net.Uri
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import com.chat.android.core.push.PushReceiveDispatcher
import com.chat.android.core.push.PushTokenManager
import dagger.hilt.android.AndroidEntryPoint
import javax.inject.Inject

@AndroidEntryPoint
class FirebaseMessagingService : com.google.firebase.messaging.FirebaseMessagingService() {

    @Inject
    lateinit var pushDispatcher: PushReceiveDispatcher

    @Inject
    lateinit var pushTokenManager: PushTokenManager

    override fun onNewToken(token: String) {
        super.onNewToken(token)
        Log.d(TAG, "FCM Token: $token")
        pushTokenManager.storeFCMTokenLocally(token)
    }

    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        super.onMessageReceived(remoteMessage)

        Log.d(TAG, "From: ${remoteMessage.from}")
        Log.d(TAG, "Message data payload: ${remoteMessage.data}")

        val payload = remoteMessage.data["payload"]
        if (payload.isNullOrEmpty()) {
            Log.w(TAG, "Message has no data payload; ignoring")
            return
        }

        // Delegate both DB sync AND notification processing to the dispatcher layer.
        // The dispatcher will use ThreadNotificationProcessor internally if fromPush is true.
        pushDispatcher.receive(payload, fromPush = true)
    }

    companion object {
        private const val TAG = "FCMService"
    }
}

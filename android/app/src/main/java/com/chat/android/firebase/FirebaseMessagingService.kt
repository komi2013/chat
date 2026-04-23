package com.chat.android.firebase

import android.util.Log
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import dagger.hilt.android.AndroidEntryPoint
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
        
        // Check if message contains a data payload
        if (remoteMessage.data.isNotEmpty()) {
            Log.d(TAG, "Message data payload: ${remoteMessage.data}")
            
            val title = remoteMessage.data["title"] ?: "New Message"
            val body = remoteMessage.data["body"] ?: "You have a new message"
            val channelId = remoteMessage.data["channelId"] ?: ""
            
            pushNotificationManager.showNotification(title, body, channelId)
        }
        
        // Check if message contains notification payload
        remoteMessage.notification?.let {
            Log.d(TAG, "Message Notification Body: ${it.body}")
            pushNotificationManager.showNotification(
                it.title ?: "New Message",
                it.body ?: "You have a new message",
                ""
            )
        }
    }
    
    companion object {
        private const val TAG = "FCMService"
    }
}
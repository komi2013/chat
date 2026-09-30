package com.chat.android.firebase

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import com.chat.android.MainActivity
import com.chat.android.R
import com.chat.android.core.network.ApiService
import com.chat.android.core.network.SessionManager
import kotlinx.coroutines.withContext
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class PushNotificationManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ApiService,
    private val sessionManager: SessionManager
) {
    
    private val notificationManager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
    
    init {
        createNotificationChannel()
    }
    
    private fun createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                "Chat Notifications",
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "Notifications for new messages"
            }
            notificationManager.createNotificationChannel(channel)
        }
    }
    
    suspend fun sendTokenToBackend(token: String, csrf: String) {
        try {
            val response = apiService.subscribeMobilePush(token, csrf)
            if (response.isSuccessful) {
                // PushSubscribeMobile はセッション確認の一部として CSRF を回転させ、
                // 新しい値を応答へ返す（common/session.go の CSRFcheckMake）。
                // ここで保存しないと、以降の API がすべて
                // "SessionCheckTake token error" で失敗し続ける。
                sessionManager.applyResponseCsrf(csrf, response.body()?.csrf)
            } else {
                throw Exception("Server returned error: ${response.code()}")
            }
        } catch (e: Exception) {
            throw Exception("Failed to send FCM token: ${e.message}")
        }
    }
    
    /**
     * 通知を表示する。
     *
     * [collapseKey] を渡すと通知IDがそのキーの hashCode になるため、同じキーの通知は
     * 積み上がらず1件に集約される（サーバー側の FCM `tag` と同じ考え方）。
     * 省略した場合は従来どおり毎回新しいIDで表示する。
     */
    fun showNotification(title: String, body: String, channelId: String, collapseKey: String? = null) {
        val intent = Intent(context, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TASK
            putExtra("channelId", channelId)
        }
        
        val pendingIntent = PendingIntent.getActivity(
            context,
            0,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
        
        val notification = NotificationCompat.Builder(context, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_notification)
            .setContentTitle(title)
            .setContentText(body)
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .build()
        
        // collapseKey があるときは同じキーの通知を上書きして1件に集約する。
        val notificationId = collapseKey?.hashCode() ?: System.currentTimeMillis().toInt()
        notificationManager.notify(notificationId, notification)
    }
    
    companion object {
        private const val CHANNEL_ID = "chat_notifications"
    }
}
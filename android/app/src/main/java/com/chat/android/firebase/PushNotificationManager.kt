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
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.ChannelPayload
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbChannel
import com.chat.android.feature.channel.DbGroup
import kotlinx.coroutines.Dispatchers
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
    
    // ChannelDbHelper は Hilt の提供定義が無いので、他の組む場所と同じく
    // context から直接組み立てる（ChannelRepository / 各PushHandler と同じ方式）。
    private val dbHelper by lazy { ChannelDbHelper(context) }

    suspend fun sendTokenToBackend(token: String, csrf: String) {
        try {
            val response = apiService.subscribeMobilePush(token, csrf)
            if (response.isSuccessful) {
                // PushSubscribeMobile はセッション確認の一部として CSRF を回転させ、
                // 新しい値を応答へ返す（common/session.go の CSRFcheckMake）。
                // ここで保存しないと、以降の API がすべて
                // "SessionCheckTake token error" で失敗し続ける。
                val body = response.body()
                sessionManager.applyResponseCsrf(csrf, body?.csrf)

                // 所属チャネルをローカルSQLiteへUpsertする。
                // アプリはローカルDBを唯一のデータ源にしており、再インストールで
                // DBが消えるとドロワーのチャネルが戻らないため、サインイン時に
                // サーバーから復元しておく必要がある（Web版のPushSubscribeと同じ）。
                syncChannels(body?.channels)
            } else {
                throw Exception("Server returned error: ${response.code()}")
            }
        } catch (e: Exception) {
            throw Exception("Failed to send FCM token: ${e.message}")
        }
    }

    /**
     * 応答に含まれる channels をローカルSQLiteへ保存する。
     *
     * vue/view/pushSubscription.html と同じく、channel だけでなく
     * その channel に含まれる aliases と groups もまとめて upsert する。
     * これを省略すると、サインイン直後にエイリアスとグループが空のままになる。
     *
     * myimg は応答に含まれない（sanitizeMyChannel は UserID と一緒に落とす）が、
     * 既にローカルにmyimg がある場合は上書きしない。
     */
    private suspend fun syncChannels(channels: List<ChannelPayload>?) {
        val list = channels.orEmpty()
        if (list.isEmpty()) return

        withContext(Dispatchers.IO) {
            list.forEach channelLoop@{ c ->
                val id = c.channelID?.takeIf { it.isNotBlank() } ?: return@channelLoop
                val existing = dbHelper.getChannel(id)
                dbHelper.saveChannel(
                    DbChannel(
                        channelID = id,
                        channelName = c.channelName ?: existing?.channelName.orEmpty(),
                        channelDescription = c.channelDescription ?: existing?.channelDescription.orEmpty(),
                        myname = c.myname ?: existing?.myname.orEmpty(),
                        myimg = existing?.myimg.orEmpty()
                    )
                )

                // エイリアス（UserID はサーバー側の sanitizeMyChannel で除去されている）
                c.aliases?.forEach aliasLoop@{ alias ->
                    val aliasID = alias.aliasID?.takeIf { it.isNotBlank() } ?: return@aliasLoop
                    dbHelper.saveAlias(
                        DbAlias(
                            aliasID = aliasID,
                            channelID = id,
                            aliasName = alias.aliasName.orEmpty(),
                            aliasImg = alias.aliasImg.orEmpty(),
                            userID = alias.userID.orEmpty(),
                            aliasBio = alias.aliasBio.orEmpty(),
                            accessRight = alias.accessRight.orEmpty()
                        )
                    )
                }

                // グループ
                c.groups?.forEach groupLoop@{ group ->
                    val groupID = group.groupID?.takeIf { it.isNotBlank() } ?: return@groupLoop
                    dbHelper.saveGroup(
                        DbGroup(
                            groupID = groupID,
                            channelID = id,
                            groupName = group.groupName.orEmpty(),
                            groupImg = group.groupImg.orEmpty(),
                            aliasNamesJson = group.aliasNames?.toString() ?: "[]",
                            groupBio = group.groupBio.orEmpty()
                        )
                    )
                }
            }
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
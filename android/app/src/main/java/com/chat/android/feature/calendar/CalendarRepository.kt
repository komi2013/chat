package com.chat.android.feature.calendar

import android.content.Context
import com.chat.android.core.network.SessionManager
import com.chat.android.core.push.PushReceiveDispatcher
import com.chat.android.core.repository.UserRepository
import com.chat.android.feature.channel.ChannelRepository
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import org.json.JSONArray
import org.json.JSONObject
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class CalendarRepository @Inject constructor(
    @ApplicationContext context: Context,
    private val api: CalendarApiService,
    private val session: SessionManager,
    private val users: UserRepository,
    private val channels: ChannelRepository,
    private val dispatcher: PushReceiveDispatcher
) {
    private val db = CalendarDbHelper(context)
    suspend fun currentChannelId(): String? = withContext(Dispatchers.IO) {
        users.getCurrentChannelId() ?: channels.getAllChannels().firstOrNull()?.channelID
    }
    suspend fun currentAlias(channelID: String) = withContext(Dispatchers.IO) { channels.getChannel(channelID)?.myname.orEmpty() }
    suspend fun aliases(channelID: String): List<DbAlias> = withContext(Dispatchers.IO) { channels.getAliases(channelID) }
    suspend fun groups(channelID: String): List<DbGroup> = withContext(Dispatchers.IO) { channels.getGroups(channelID) }
    suspend fun events(channelID: String): List<CalendarItem> = withContext(Dispatchers.IO) { db.all().filter { it.channelID == channelID } }
    suspend fun event(id: String): CalendarItem? = withContext(Dispatchers.IO) { db.get(id) }

    /** storeSelect で対象エイリアスを POST し、返ってきたカレンダーを push 経由で取り込む。 */
    suspend fun requestUserCalendars(channelID: String, aliasNames: List<String>, date: String): String? {
        if (aliasNames.isEmpty()) return null
        val csrf = session.getCsrf()?.takeIf(String::isNotBlank) ?: return "サインインが必要です"
        val updatedBy = currentAlias(channelID).ifBlank { session.getNickname().orEmpty() }
        if (updatedBy.isBlank()) return "送信者名が設定されていません"
        return try {
            val contents = JSONArray().put("calendar").put(JSONObject().put("date", date)).toString()
            val response = api.requestCalendars(
                JSONArray(aliasNames).toString().part(), channelID.part(), updatedBy.part(), csrf.part(),
                contents.part(), "storeSelect".part()
            )
            if (!response.isSuccessful) return "データ取得に失敗しました (${response.code()})"
            val body = response.body() ?: return "サーバーから空の応答が返されました"
            session.applyResponseCsrf(csrf, body.csrf)
            if (body.csrf.isNullOrBlank()) return "データ取得に失敗しました。サインイン状態を確認してください"
            if (!body.error.isNullOrBlank()) return body.error
            body.pushContents.orEmpty().forEach { dispatcher.receive(it) }
            null
        } catch (e: Exception) { e.message ?: "データ取得に失敗しました" }
    }

    suspend fun removeUserCalendar(channelID: String, aliasName: String) = withContext(Dispatchers.IO) {
        db.deleteByAlias(channelID, aliasName)
    }

    suspend fun send(item: CalendarItem, attendees: List<String>, events: List<CalendarItem>, delete: Boolean = false): String? {
        val csrf = session.getCsrf()?.takeIf(String::isNotBlank) ?: return "サインインが必要です"
        val alias = currentAlias(item.channelID).ifBlank { session.getNickname().orEmpty() }
        if (alias.isBlank()) return "送信者名が設定されていません"
        for (event in events) {
            val id = event.calendarID.ifBlank { java.util.UUID.randomUUID().toString().replace("-", "").take(8) }
            val outgoing = event.copy(calendarID = id, channelID = item.channelID, aliasName = alias, aliasNamesJson = JSONArray(attendees).toString())
            val json = outgoing.toJson().apply { if (delete) put("delete", 1) }.toString()
            try {
                val response = api.pushCalendar(
                    JSONArray(attendees).toString().part(), item.channelID.part(), alias.part(), csrf.part(), json.part(), "calendar".part()
                )
                if (!response.isSuccessful) return "送信に失敗しました (${response.code()})"
                val body = response.body() ?: return "サーバーから空の応答が返されました"
                session.applyResponseCsrf(csrf, body.csrf)
                if (!body.error.isNullOrBlank()) return body.error
                if (body.csrf.isNullOrBlank()) return "CSRFトークンを更新できませんでした"
                body.pushContents.orEmpty().forEach { dispatcher.receive(it) }
                withContext(Dispatchers.IO) { if (delete) db.delete(id) else db.save(outgoing) }
            } catch (e: Exception) { return e.message ?: "送信に失敗しました" }
        }
        return null
    }
    private fun String.part() = toRequestBody("text/plain".toMediaType())
}

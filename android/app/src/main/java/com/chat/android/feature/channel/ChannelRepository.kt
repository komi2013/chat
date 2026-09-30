package com.chat.android.feature.channel

import android.content.Context
import com.chat.android.core.data.SessionManager
import com.chat.android.core.push.PushReceiveDispatcher
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody
import okhttp3.RequestBody.Companion.toRequestBody
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChannelRepository @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ChannelApiService,
    private val sessionManager: SessionManager,
    private val pushDispatcher: PushReceiveDispatcher
) {
    private val dbHelper = ChannelDbHelper(context)
    
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.Default)

    private val _dbUpdateFlow = MutableSharedFlow<Unit>(replay = 0)
    val dbUpdateFlow: SharedFlow<Unit> = _dbUpdateFlow.asSharedFlow()

    init {
        // FCM 受信で PushReceiveDispatcher がローカルDBを更新したとき、画面にも
        // 反映させる。dispatcher.onUpdated を購読している箇所が他に無いための橋渡し。
        scope.launch {
            pushDispatcher.onUpdated.collect { _dbUpdateFlow.emit(Unit) }
        }
    }

    suspend fun getChannel(channelID: String): DbChannel? = dbHelper.getChannel(channelID)
    suspend fun getAllChannels(): List<DbChannel> = dbHelper.getAllChannels()
    suspend fun getAliases(channelID: String): List<DbAlias> = dbHelper.getAliasesForChannel(channelID)
    suspend fun getGroups(channelID: String): List<DbGroup> = dbHelper.getGroupsForChannel(channelID)

    private fun toPart(value: String): RequestBody = value.toRequestBody("text/plain".toMediaTypeOrNull())

    suspend fun createChannel(name: String, description: String, myname: String, myimg: String): Result<String> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelAdd(
            toPart(name), toPart(description), toPart(myname), toPart(myimg), toPart(csrf)
        )

        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return Result.failure(Exception("サーバーから空の応答が返されました"))

        // CSRF は成功・エラーどちらの応答でも回転しているので必ず反映する。
        sessionManager.applyResponseCsrf(csrf, body.csrf)
        body.pushContents?.forEach { pushDispatcher.receive(it) }

        // 「すでにチャネル作成の上限です」等は HTTP 200 + error で返る。
        // csrf の有無で成功判定するとエラーを握りつぶしてしまう。
        if (!body.error.isNullOrBlank()) {
            return Result.failure(Exception(body.error))
        }

        val channelID = body.channelID?.takeIf(String::isNotBlank)
            ?: return Result.failure(Exception("サーバーからチャネルIDが返されませんでした"))

        // ChannelAdd の応答は channelID のみなので、作成直後に画面を開いても
        // "Channel not found" にならないよう送信値でローカルを先に埋めておく。
        // 後続の channelEdit / alias push がサーバー値で上書きする。
        dbHelper.saveChannel(
            DbChannel(
                channelID = channelID,
                channelName = name,
                channelDescription = description,
                myname = myname,
                myimg = myimg,
                displayStatus = 0,
                invitationCode = "",
                invitationGuestCode = ""
            )
        )

        _dbUpdateFlow.emit(Unit)
        return Result.success(channelID)
    }

    suspend fun editChannel(
        channelID: String,
        updatedBy: String,
        pushNamesJson: String,
        channelName: String? = null,
        groupsJson: String? = null,
        description: String? = null
    ): Result<Unit> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(updatedBy),
            toPart(pushNamesJson),
            channelName?.let { toPart(it) },
            groupsJson?.let { toPart(it) },
            description?.let { toPart(it) },
            csrf = toPart(csrf)
        )

        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        body.pushContents?.forEach { pushDispatcher.receive(it) }
        _dbUpdateFlow.emit(Unit)

        // 権限エラー等は HTTP 200 + error で返る
        if (!body.error.isNullOrBlank()) {
            return Result.failure(Exception(body.error))
        }
        return Result.success(Unit)
    }

    suspend fun deleteChannel(channelID: String, updatedBy: String): Result<Unit> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelDelete(
            toPart(channelID),
            toPart(updatedBy),
            toPart("1"),
            toPart(csrf)
        )
        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        body.pushContents?.forEach { pushDispatcher.receive(it) }
        _dbUpdateFlow.emit(Unit)

        if (!body.error.isNullOrBlank()) {
            return Result.failure(Exception(body.error))
        }
        return Result.success(Unit)
    }
}

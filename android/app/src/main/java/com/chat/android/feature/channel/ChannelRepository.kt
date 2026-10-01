package com.chat.android.feature.channel

import android.content.Context
import com.chat.android.core.network.SessionManager
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
import org.json.JSONObject
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

    /** スレッド一覧（トップレベルのみ）。vue の Channel.vue と同じ。 */
    suspend fun getThreadHeads(channelID: String): List<DbThreadHead> =
        dbHelper.getThreadHeadsForChannel(channelID)

    suspend fun getThreadHead(parentID: String): DbThreadHead? = dbHelper.getThreadHead(parentID)

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
                myimg = myimg
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

    /**
     * 招待コードを生成する。
     *
     * サーバーは generateInvitation を受けたときだけ invitationCode / invitationGuestCode を
     * 回転させ지만、channelEdit push には [channelName, channelDescription] しか載せない
     * （controller/ChannelEdit.go:337）。そのため push 経由ではローカルDBに新しいコードが
     * 反映されないので、応答の channel を使ってここで DB を更新する。
     */
    suspend fun generateInvitation(channelID: String, updatedBy: String, guest: Boolean): Result<String> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(updatedBy),
            toPart("[]"),
            null,
            null,
            null,
            generateInvitation = toPart("1"),
            guest = if (guest) toPart("1") else null,
            csrf = toPart(csrf)
        )

        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)

        if (!body.error.isNullOrBlank()) {
            return Result.failure(Exception(body.error))
        }

        body.pushContents?.forEach { pushDispatcher.receive(it) }

        // コードは保存せず、呼び出し元へ返すだけにする。
        val code = body.channel?.let { if (guest) it.invitationGuestCode else it.invitationCode }
        if (code.isNullOrBlank()) {
            return Result.failure(Exception("招待コードが返されませんでした"))
        }
        _dbUpdateFlow.emit(Unit)
        return Result.success(code)
    }

    /**
     * 招待コードでチャネルに参加する（vue/src/views/Profile.vue の join() に対応）。
     *
     * サーバーは参加後に pushContents で
     *  - channelEdit（チャネル名・説明）
     *  - alias（参加者全員）
     *  - group（既存グループ）
     * を返すので、これをそのまま PushReceiveDispatcher へ流して
     * ローカルSQLiteを同期する。
     */
    suspend fun joinChannel(
        channelID: String,
        code: String,
        myname: String,
        myimg: String
    ): Result<Unit> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelJoin(
            toPart(channelID),
            toPart(code),
            toPart(myname),
            toPart(myimg),
            toPart(csrf)
        )
        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)

        // エラー応答には pushContents が含まれないので先に判定する
        if (!body.error.isNullOrBlank()) {
            return Result.failure(Exception(body.error))
        }

        body.pushContents?.forEach { pushDispatcher.receive(it) }
        _dbUpdateFlow.emit(Unit)
        return Result.success(Unit)
    }

    /**
     * エイリアス（プロフィール）を編集する（Profile.vue の editAlias 相当）。
     *
     * 注意: Profile.vue は pushTitle/contents を送っているが、サーバーは
     * `admin` フィールド（collection.Alias のJSON）だけを見ている
     * （controller/ChannelEdit.go:306）。画像は admin.AliasImg 経由で
     * alias push の index 5 へ流れる（同ファイル:381）ので、
     * ここで aliasImg を含めて JSON として送る。
     */
    suspend fun editAlias(
        channelID: String,
        updatedBy: String,
        aliasName: String,
        aliasBio: String,
        aliasImg: String
    ): Result<Unit> {
        val csrf = sessionManager.getCsrf() ?: ""
        val aliases = dbHelper.getAliasesForChannel(channelID)
        val existing = aliases.find { it.aliasName == updatedBy }

        val adminJson = JSONObject().apply {
            put("aliasID", existing?.aliasID ?: (channelID + aliasName))
            put("channelID", channelID)
            put("aliasName", aliasName)
            put("aliasImg", aliasImg)
            put("userID", existing?.userID.orEmpty())
            put("aliasBio", aliasBio)
            put("accessRight", existing?.accessRight.orEmpty())
        }.toString()

        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(updatedBy),
            toPart("[]"),
            null,
            null,
            null,
            admin = toPart(adminJson),
            csrf = toPart(csrf)
        )
        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        if (!body.error.isNullOrBlank()) {
            return Result.failure(Exception(body.error))
        }
        body.pushContents?.forEach { pushDispatcher.receive(it) }
        _dbUpdateFlow.emit(Unit)
        return Result.success(Unit)
    }

    /**
     * 参加済みチャネルの myname（現在のエイリアス）を切り替える。
     *
     * サーバーへは送らずローカルSQLiteだけ更新する。
     * Profile.vue の switchAlias も IndexedDB への upsert のみで、
     * サーバーAPIは呼んでいないため、Web版と同じ挙動になる。
     */
    suspend fun switchChannelMyname(channelID: String, aliasName: String) {
        val existing = dbHelper.getChannel(channelID) ?: return
        dbHelper.saveChannel(existing.copy(myname = aliasName))
        _dbUpdateFlow.emit(Unit)
    }

    /**
     * グループの変更を保存する（vue/src/views/Group.vue の editGroup/removeGroup に対応）。
     *
     * [groupsJson] は差分（追加・更新・削除）だけを入れる。
     * 削除は aliasNames を null にした要素で表現する。
     * サーバーは group ごとに common.ImgSave() を呼ぶため、
     * 保存済み画像パスは呼び出し側で DataURI へ変換しておく必要がある。
     */
    suspend fun editGroups(
        channelID: String,
        updatedBy: String,
        groupsJson: String,
        pushNamesJson: String
    ): Result<Unit> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(updatedBy),
            toPart(pushNamesJson),
            null,
            toPart(groupsJson),
            null,
            csrf = toPart(csrf)
        )
        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        if (!body.error.isNullOrBlank()) {
            return Result.failure(Exception(body.error))
        }
        body.pushContents?.forEach { pushDispatcher.receive(it) }

        // サーバーの group push は FCM 経由の非同期配信で、HTTP 応答の
        // pushContents には含まれない（common/push.go の ChunkPush は
        // SendWebPushNotification を送るだけで session.PushContents に積まない）。
        // push が届く前に画面を再読込すると「保存成功だが一覧に出てこない」状態になる。
        // createChannel と同じ考え方（送信値で先にローカルへ埋める）で、保存内容を
        // ローカルSQLiteへ直接反映する。
        applyGroupsLocally(channelID, groupsJson)
        _dbUpdateFlow.emit(Unit)
        return Result.success(Unit)
    }

    /**
     * 保存した groups 差分（追加・更新・削除）をローカルSQLiteへ反映する。
     * aliasNames が null の要素は削除を表す（vue の Group.vue と同じ表現）。
     */
    private suspend fun applyGroupsLocally(channelID: String, groupsJson: String) {
        val array = runCatching { org.json.JSONArray(groupsJson) }.getOrNull() ?: return
        for (i in 0 until array.length()) {
            val obj = array.optJSONObject(i) ?: continue
            val groupName = obj.optString("groupName", "")
            if (groupName.isEmpty()) continue

            val aliasNames = obj.optJSONArray("aliasNames")
            if (aliasNames == null) {
                dbHelper.deleteGroup(channelID + groupName)
                continue
            }

            val groupID = obj.optString("groupID", "").takeIf { it.isNotBlank() }
                ?: (channelID + groupName)
            val existing = dbHelper.getGroupsForChannel(channelID)
                .firstOrNull { it.groupID == groupID }
            dbHelper.saveGroup(
                DbGroup(
                    groupID = groupID,
                    channelID = channelID,
                    groupName = groupName,
                    groupImg = obj.optString("groupImg", existing?.groupImg.orEmpty()),
                    aliasNamesJson = aliasNames.toString(),
                    groupBio = obj.optString("groupBio", existing?.groupBio.orEmpty())
                )
            )
        }
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

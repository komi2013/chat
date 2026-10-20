package com.chat.android.feature.channel

import android.content.Context
import com.chat.android.core.network.ApiService
import com.chat.android.core.network.SessionManager
import com.chat.android.core.push.PushReceiveDispatcher
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
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
    private val generalApiService: ApiService,
    private val sessionManager: SessionManager,
    private val pushDispatcher: PushReceiveDispatcher,
    private val profileStore: ChannelProfileStore
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

    suspend fun getChannel(channelID: String): DbChannel? = dbHelper.getChannel(channelID)?.withLocalProfile()
    suspend fun getAllChannels(): List<DbChannel> = dbHelper.getAllChannels().map { it.withLocalProfile() }

    private suspend fun resolveIdentity(channelID: String?): Pair<String, String> {
        if (channelID == null) return "" to ""

        // 1. DataStore (Channel-specific Cache)
        var (name, image) = profileStore.profile(channelID)
        if (name.isNotBlank()) return name to image

        // 2. Alias Table (Join by current userID)
        val myUserId = sessionManager.getUserId()
        if (myUserId != null) {
            val myAlias = dbHelper.getAliasesForChannel(channelID).find { it.userID == myUserId }
            if (myAlias != null) {
                name = myAlias.aliasName
                image = myAlias.aliasImg
                if (name.isNotBlank()) {
                    profileStore.save(channelID, name, image)
                }
                return name to image
            }

            // 3. Fallback to Last Used Profile (Cross-channel Cache)
            val last = profileStore.getLastProfile(myUserId)
            if (last.second.isNotBlank()) {
                return last.second to last.third
            }
        }

        // 4. Global Fallback (Session nickname)
        return sessionManager.getNickname().orEmpty() to ""
    }

    private suspend fun DbChannel.withLocalProfile(): DbChannel {
        val (name, image) = resolveIdentity(channelID)
        return copy(myname = name, myimg = image)
    }
    suspend fun getAliases(channelID: String): List<DbAlias> = dbHelper.getAliasesForChannel(channelID)
    suspend fun getGroups(channelID: String): List<DbGroup> = dbHelper.getGroupsForChannel(channelID)

    /** スレッド一覧（トップレベルのみ）。vue の Channel.vue と同じ。 */
    suspend fun getThreadHeads(channelID: String): List<DbThreadHead> =
        dbHelper.getThreadHeadsForChannel(channelID)

    suspend fun getThreadHead(parentID: String): DbThreadHead? = dbHelper.getThreadHead(parentID)

    suspend fun getLatestProfile(): Triple<String, String, String> {
        val userId = sessionManager.getUserId() ?: return Triple("", "", "")
        return profileStore.getLastProfile(userId)
    }

    suspend fun saveLastProfile(channelID: String, name: String, image: String) {
        val userId = sessionManager.getUserId() ?: return
        profileStore.saveLastProfile(userId, channelID, name, image)
    }

    suspend fun saveLastChannelID(channelID: String) {
        val userId = sessionManager.getUserId() ?: return
        val channel = getChannel(channelID) ?: return
        if (channel.myname.isNotBlank()) {
            profileStore.saveLastProfile(userId, channelID, channel.myname, channel.myimg)
        }
    }

    private suspend fun getEffectiveName(channelID: String?, providedName: String): String {
        if (providedName.isNotBlank()) return providedName

        if (channelID != null) {
            val (name, _) = resolveIdentity(channelID)
            if (name.isNotBlank()) return name
        }

        return sessionManager.getNickname().orEmpty()
    }

    private suspend fun getEffectiveImage(channelID: String?, providedImage: String): String {
        if (providedImage.isNotBlank()) return providedImage

        if (channelID != null) {
            val (_, image) = resolveIdentity(channelID)
            if (image.isNotBlank()) return image
        }

        return ""
    }

    private fun toPart(value: String): RequestBody = value.toRequestBody("text/plain".toMediaTypeOrNull())

    suspend fun createChannel(name: String, description: String, myname: String, myimg: String): Result<String> = sessionManager.serializeAuthenticatedRequest {
        val effectiveName = getEffectiveName(null, myname)
        val effectiveImg = getEffectiveImage(null, myimg)
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelAdd(
            toPart(name), toPart(description), toPart(effectiveName), toPart(effectiveImg), toPart(csrf)
        )

        if (!response.isSuccessful) {
            return@serializeAuthenticatedRequest Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーから空の応答が返されました"))

        // CSRF は成功・エラーどちらの応答でも回転しているので必ず反映する。
        sessionManager.applyResponseCsrf(csrf, body.csrf)
        body.pushContents?.forEach { pushDispatcher.receive(it) }

        // 「すでにチャネル作成の上限です」等は HTTP 200 + error で返る。
        // csrf の有無で成功判定するとエラーを握りつぶしてしまう。
        if (!body.error.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception(body.error))
        }

        val channelID = body.channelID?.takeIf(String::isNotBlank)
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーからチャネルIDが返されませんでした"))

        // ChannelAdd の応答は channelID のみなので、作成直後に画面を開いても
        // "Channel not found" にならないよう送信値でローカルを先に埋めておく。
        // 後続の channelEdit / alias push がサーバー値で上書きする。
        dbHelper.saveChannel(
            DbChannel(
                channelID = channelID,
                channelName = name,
                channelDescription = description,
                updatedBy = effectiveName,
                updatedAt = java.text.SimpleDateFormat("yyyy-MM-dd HH:mm:ss", java.util.Locale.getDefault()).format(java.util.Date()),
                myname = effectiveName,
                myimg = effectiveImg
            )
        )
        // profileStore is device-local, we save it here for auto-filling and authorized requests
        profileStore.save(channelID, name = effectiveName, image = effectiveImg)
        sessionManager.getUserId()?.let { userId ->
            profileStore.saveLastProfile(userId, channelID, effectiveName, effectiveImg)
        }

        _dbUpdateFlow.emit(Unit)
        return@serializeAuthenticatedRequest Result.success(channelID)
    }

    suspend fun editChannel(
        channelID: String,
        providedUpdatedBy: String,
        pushNamesJson: String,
        channelName: String? = null,
        groupsJson: String? = null,
        description: String? = null
    ): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val effectiveName = getEffectiveName(channelID, providedUpdatedBy)
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(effectiveName),
            toPart(pushNamesJson),
            channelName?.let { toPart(it) },
            groupsJson?.let { toPart(it) },
            description?.let { toPart(it) },
            csrf = toPart(csrf)
        )

        if (!response.isSuccessful) {
            return@serializeAuthenticatedRequest Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        body.pushContents?.forEach { pushDispatcher.receive(it) }
        _dbUpdateFlow.emit(Unit)

        // 権限エラー等は HTTP 200 + error で返る
        if (!body.error.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception(body.error))
        }
        return@serializeAuthenticatedRequest Result.success(Unit)
    }

    /**
     * 招待コードを生成する。
     *
     * サーバーは generateInvitation を受けたときだけ invitationCode / invitationGuestCode を
     * 回転させ지만、channelEdit push には [channelName, channelDescription] しか載せない
     * （controller/ChannelEdit.go:337）。そのため push 経由ではローカルDBに新しいコードが
     * 反映されないので、応答の channel を使ってここで DB を更新する。
     */
    suspend fun generateInvitation(channelID: String, providedUpdatedBy: String, guest: Boolean): Result<String> = sessionManager.serializeAuthenticatedRequest {
        val effectiveName = getEffectiveName(channelID, providedUpdatedBy)
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(effectiveName),
            toPart("[]"),
            null,
            null,
            null,
            generateInvitation = toPart("1"),
            guest = if (guest) toPart("1") else null,
            csrf = toPart(csrf)
        )

        if (!response.isSuccessful) {
            return@serializeAuthenticatedRequest Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)

        if (!body.error.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception(body.error))
        }

        body.pushContents?.forEach { pushDispatcher.receive(it) }

        // コードは保存せず、呼び出し元へ返すだけにする。
        val code = body.channel?.let { if (guest) it.invitationGuestCode else it.invitationCode }
        if (code.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception("招待コードが返されませんでした"))
        }
        _dbUpdateFlow.emit(Unit)
        return@serializeAuthenticatedRequest Result.success(code)
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
    ): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val effectiveName = getEffectiveName(null, myname)
        val effectiveImg = getEffectiveImage(null, myimg)
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelJoin(
            toPart(channelID),
            toPart(code),
            toPart(effectiveName),
            toPart(effectiveImg),
            toPart(csrf)
        )
        if (!response.isSuccessful) {
            return@serializeAuthenticatedRequest Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)

        // エラー応答には pushContents が含まれないので先に判定する
        if (!body.error.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception(body.error))
        }

        body.pushContents?.forEach { pushDispatcher.receive(it) }
        profileStore.save(channelID, name = effectiveName, image = effectiveImg)
        sessionManager.getUserId()?.let { userId ->
            profileStore.saveLastProfile(userId, channelID, effectiveName, effectiveImg)
        }
        _dbUpdateFlow.emit(Unit)
        return@serializeAuthenticatedRequest Result.success(Unit)
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
        providedUpdatedBy: String,
        aliasName: String,
        aliasBio: String,
        aliasImg: String
    ): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val effectiveUpdatedBy = getEffectiveName(channelID, providedUpdatedBy)
        val effectiveAliasName = if (aliasName.isNotBlank()) aliasName else effectiveUpdatedBy
        val effectiveAliasImg = getEffectiveImage(channelID, aliasImg)
        val csrf = sessionManager.getCsrf() ?: ""
        val aliases = dbHelper.getAliasesForChannel(channelID)
        val existing = aliases.find { it.aliasName == effectiveUpdatedBy }

        val myUserId = sessionManager.getUserId().orEmpty()
        val adminJson = JSONObject().apply {
            put("aliasID", existing?.aliasID ?: (channelID + effectiveAliasName))
            put("channelID", channelID)
            put("aliasName", effectiveAliasName)
            put("aliasImg", effectiveAliasImg)
            put("userID", existing?.userID?.takeIf { it.isNotBlank() } ?: myUserId)
            put("aliasBio", aliasBio)
            put("accessRight", existing?.accessRight.orEmpty())
        }.toString()

        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(effectiveUpdatedBy),
            toPart("[]"),
            null,
            null,
            null,
            admin = toPart(adminJson),
            csrf = toPart(csrf)
        )
        if (!response.isSuccessful) {
            return@serializeAuthenticatedRequest Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        if (!body.error.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception(body.error))
        }
        body.pushContents?.forEach { pushDispatcher.receive(it) }

        profileStore.save(channelID, name = effectiveAliasName, image = effectiveAliasImg)
        sessionManager.getUserId()?.let { userId ->
            profileStore.saveLastProfile(userId, channelID, effectiveAliasName, effectiveAliasImg)
        }
        _dbUpdateFlow.emit(Unit)
        return@serializeAuthenticatedRequest Result.success(Unit)
    }

    /**
     * 参加済みチャネルの myname（現在のエイリアス）を切り替える。
     *
     * サーバーへは送らずローカルSQLiteだけ更新する。
     * Profile.vue の switchAlias も IndexedDB への upsert のみで、
     * サーバーAPIは呼んでいないため、Web版と同じ挙動になる。
     */
    suspend fun switchChannelMyname(channelID: String, aliasName: String) {
        dbHelper.getChannel(channelID) ?: return
        profileStore.save(channelID, name = aliasName)
        sessionManager.getUserId()?.let { userId ->
            val profile = profileStore.profile(channelID)
            profileStore.saveLastProfile(userId, channelID, aliasName, profile.second)
        }
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
        providedUpdatedBy: String,
        groupsJson: String,
        pushNamesJson: String
    ): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val effectiveName = getEffectiveName(channelID, providedUpdatedBy)
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(effectiveName),
            toPart(pushNamesJson),
            null,
            toPart(groupsJson),
            null,
            csrf = toPart(csrf)
        )
        if (!response.isSuccessful) {
            return@serializeAuthenticatedRequest Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        if (!body.error.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception(body.error))
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
        return@serializeAuthenticatedRequest Result.success(Unit)
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

    suspend fun deleteChannel(channelID: String, providedUpdatedBy: String): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val effectiveName = getEffectiveName(channelID, providedUpdatedBy)
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelDelete(
            toPart(channelID),
            toPart(effectiveName),
            toPart("1"),
            toPart(csrf)
        )
        if (!response.isSuccessful) {
            return@serializeAuthenticatedRequest Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body()
            ?: return@serializeAuthenticatedRequest Result.failure(Exception("サーバーから空の応答が返されました"))

        sessionManager.applyResponseCsrf(csrf, body.csrf)
        body.pushContents?.forEach { pushDispatcher.receive(it) }
        _dbUpdateFlow.emit(Unit)

        if (!body.error.isNullOrBlank()) {
            return@serializeAuthenticatedRequest Result.failure(Exception(body.error))
        }
        return@serializeAuthenticatedRequest Result.success(Unit)
    }

    /**
     * FCM トークンをサーバーへ登録し、所属チャネル一覧を同期する。
     * サインイン直後またはトークン更新時に呼ばれる。
     */
    suspend fun registerPushToken(token: String, csrf: String): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        withContext(Dispatchers.IO) {
            try {
                val response = generalApiService.subscribeMobilePush(token, csrf)
                if (response.isSuccessful) {
                    val body = response.body()
                    sessionManager.applyResponseCsrf(csrf, body?.csrf)
                    syncChannels(body?.channels)
                    Result.success(Unit)
                } else {
                    Result.failure(Exception("Server returned error: ${response.code()}"))
                }
            } catch (e: Exception) {
                Result.failure(Exception("Failed to send FCM token: ${e.message}"))
            }
        }
    }

    private suspend fun syncChannels(channels: List<ChannelPayload>?) {
        val list = channels.orEmpty()
        if (list.isEmpty()) return

        list.forEach { c ->
            val id = c.channelID?.takeIf { it.isNotBlank() } ?: return@forEach
            val existing = dbHelper.getChannel(id)
            dbHelper.saveChannel(
                DbChannel(
                    channelID = id,
                    channelName = c.channelName ?: existing?.channelName.orEmpty(),
                    channelDescription = c.channelDescription ?: existing?.channelDescription.orEmpty(),
                    updatedBy = existing?.updatedBy ?: "system",
                    updatedAt = existing?.updatedAt ?: java.text.SimpleDateFormat("yyyy-MM-dd HH:mm:ss", java.util.Locale.getDefault()).format(java.util.Date())
                )
            )

            val currentUserId = sessionManager.getUserId()
            c.aliases?.forEach { alias ->
                val aliasID = alias.aliasID?.takeIf { it.isNotBlank() } ?: return@forEach
                val resolvedUserId = alias.userID?.takeIf { it.isNotBlank() }
                    ?: if (alias.aliasName == c.myname && !c.myname.isNullOrBlank()) currentUserId.orEmpty() else ""

                dbHelper.saveAlias(
                    DbAlias(
                        aliasID = aliasID,
                        channelID = id,
                        aliasName = alias.aliasName.orEmpty(),
                        aliasImg = alias.aliasImg.orEmpty(),
                        userID = resolvedUserId,
                        aliasBio = alias.aliasBio.orEmpty(),
                        accessRight = alias.accessRight.orEmpty()
                    )
                )
            }

            c.groups?.forEach { group ->
                val groupID = group.groupID?.takeIf { it.isNotBlank() } ?: return@forEach
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
        _dbUpdateFlow.emit(Unit)
    }
}

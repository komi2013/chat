package com.chat.android.feature.thread

import com.chat.android.core.network.ApiService
import com.chat.android.core.network.SessionManager
import com.chat.android.core.push.PushReceiveDispatcher
import com.chat.android.core.util.Base62
import com.chat.android.feature.channel.DbThreadHead
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import javax.inject.Inject
import javax.inject.Singleton

/**
 * スレッドの送信側（vue/src/components/EditBox.vue と EmojiModal.vue に対応）。
 *
 * 送信は ContentsPush/ 経由の push であり、サーバーの pushContents を
 * ローカルDBへ反映することで自分の投稿も画面に出る（Vue と同じ方式）。
 */
@Singleton
class ThreadRepository @Inject constructor(
    private val apiService: ApiService,
    private val sessionManager: SessionManager,
    private val pushDispatcher: PushReceiveDispatcher
) {

    private suspend fun push(
        channelID: String,
        updatedBy: String,
        pushNames: List<String>,
        pushTitle: String,
        contentsJson: String
    ): Result<Unit> = withContext(Dispatchers.IO) {
        val csrf = sessionManager.getCsrf()?.takeIf { it.isNotBlank() }
            ?: return@withContext Result.failure(Exception("サインインが必要です"))

        runCatching {
            val response = apiService.contentsPush(
                csrf = csrf.toPart(),
                channelID = channelID.toPart(),
                updatedBy = updatedBy.toPart(),
                pushNames = JSONArray(pushNames).toString().toPart(),
                contents = contentsJson.toPart(),
                pushTitle = pushTitle.toPart()
            )
            val body = response.body()
            if (!response.isSuccessful) {
                throw Exception("送信に失敗しました (${response.code()})")
            }
            // CSRF は必ず回転するので保存する
            body?.csrf?.takeIf { it.isNotBlank() }?.let { sessionManager.setCsrf(it) }
            body?.error?.takeIf { it.isNotBlank() }?.let { throw Exception(it) }
            // 自分の投稿も push 経由でローカルへ反映する
            body?.pushContents?.forEach { pushDispatcher.receive(it) }
            Unit
        }
    }

    /**
     * メッセージを投稿する（新規 or 編集、削除）。
     *
     * contents は vue と同じ位置配列:
     * [parentID, messageID, messageTxt, aliasImg, aliasNames, backID, yets, groupName, (updatedUnixAt)]
     */
    suspend fun postMessage(
        channelID: String,
        myname: String,
        parentID: String,
        messageTxt: String,
        aliasImg: String,
        totalNames: List<String>,
        backID: String,
        yets: List<Pair<String, String>>,
        groupName: String,
        editingMessageID: String? = null
    ): Result<Unit> {
        val pushTitle = if (editingMessageID != null) "threadEdit" else "thread"
        val messageID = editingMessageID ?: newMessageID()

        val contents = JSONArray().apply {
            put(parentID)
            put(messageID)
            put(messageTxt)
            put(aliasImg)
            put(JSONArray(totalNames))
            put(backID)
            put(JSONArray(yets.map { JSONArray().apply { put(it.first); put(it.second) } }))
            put(groupName)
            // 編集時だけ更新時刻を付ける（vue と同じ）
            if (editingMessageID != null) put(System.currentTimeMillis() / 1000)
        }

        val names = (totalNames + myname).distinct()
        return push(channelID, myname, names, pushTitle, contents.toString())
    }

    /** 絵文字リアクションを付ける／外す（vue/src/components/EmojiModal.vue と同じ）。 */
    suspend fun postEmoji(
        channelID: String,
        myname: String,
        pushNames: List<String>,
        messageID: String,
        emoji: String,
        parentID: String,
        delete: Boolean
    ): Result<Unit> {
        val contents = JSONArray().apply {
            put(messageID)
            put(emoji)
            put(parentID)
            put(delete)
        }
        return push(channelID, myname, pushNames, "emoji", contents.toString())
    }

    /**
     * スレッド見出しを更新する（vue/src/components/ThreadHead.vue と同じ）。
     * threadHead push はオブジェクト本体を送る。
     */
    suspend fun postThreadHead(
        channelID: String,
        myname: String,
        head: DbThreadHead
    ): Result<Unit> {
        val obj = org.json.JSONObject().apply {
            put("parentID", head.parentID)
            put("channelID", channelID)
            put("title", head.title)
            put("messageTxt", head.messageTxt)
            put("description", head.description)
            put("aliasName", head.aliasName)
            put("aliasNames", JSONArray(head.aliasNamesJson))
            put("adminNames", JSONArray(head.adminNamesJson))
            put("displayStatus", head.displayStatus)
            put("broadcastFlag", head.broadcastFlag)
            put("emojis", JSONArray(head.emojisJson))
        }
        val names = head.aliasNamesJson.toNameList() + myname
        return push(channelID, myname, names.distinct(), "threadHead", obj.toString())
    }

    /** vue の base62Encode(Date.now()) + ランダム1文字と同じ形式。 */
    private fun newMessageID(): String {
        val random = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
        val suffix = random[java.util.Random().nextInt(random.length)]
        return Base62.encode(System.currentTimeMillis() / 1000) + suffix
    }

    private fun String.toPart(): RequestBody =
        toRequestBody("text/plain".toMediaTypeOrNull())
}
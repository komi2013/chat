package com.chat.android.feature.thread

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.push.PushReceiveDispatcher
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbThread
import com.chat.android.feature.channel.DbThreadHead
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import javax.inject.Inject

data class ThreadUiState(
    val isLoading: Boolean = true,
    val channelID: String = "",
    val parentID: String = "",
    val myname: String = "",
    val title: String = "",
    val head: DbThreadHead? = null,
    val messages: List<DbThread> = emptyList(),
    /** 新規スレッド（threadHead が未作成）なら true。vue の newThread 相当。 */
    val isNewThread: Boolean = false,
    /** broadcastFlag により管理者以外は投稿できない。 */
    val canPost: Boolean = true,
    val isPosting: Boolean = false,
    val error: String? = null
)

/** aliasNames / emojis の JSON 文字列をリストへ。 */
internal fun String.toNameList(): List<String> = runCatching {
    val array = org.json.JSONArray(this)
    (0 until array.length()).mapNotNull { array.optString(it, "").ifBlank { null } }
}.getOrDefault(emptyList())

/** 本文から ＠＠名前・＠＠ 形式のメンションを抜き出す。 */
internal fun extractMentions(text: String): List<String> =
    Regex("＠＠([^・＠]+)・＠＠").findAll(text).map { it.groupValues[1] }.toList()

/** markdown の記号だけを取り除いた文字列（vue の removeMark の簡易版）。 */
internal fun removeMark(text: String): String =
    text.replace("＊", "").replace("・", "")

/** スレッド画面の状態（vue/src/views/Thread.vue に対応）。 */
@HiltViewModel
class ThreadViewModel @Inject constructor(
    @ApplicationContext private val context: Context,
    private val repository: ThreadRepository,
    private val pushDispatcher: PushReceiveDispatcher
) : ViewModel() {

    private val _uiState = MutableStateFlow(ThreadUiState())
    val uiState: StateFlow<ThreadUiState> = _uiState.asStateFlow()

    fun init(channelID: String?, parentID: String?) {
        val id = channelID ?: return
        val parent = parentID ?: return
        viewModelScope.launch {
            _uiState.value = withContext(Dispatchers.IO) { load(id, parent) }
            markRead()
        }
    }

    private fun load(channelID: String, parentID: String): ThreadUiState {
        val db = ChannelDbHelper(context)
        val channel = db.getChannel(channelID)
        val myname = channel?.myname.orEmpty()
        val groups = db.getGroupsForChannel(channelID)
        val existing = db.getThreadHead(parentID)

        if (existing != null) {
            // vue の EditBox: broadcastFlag が立っていれば管理者のみ投稿可
            val canPost = existing.broadcastFlag == 0 ||
                existing.adminNamesJson.toNameList().contains(myname)
            return ThreadUiState(
                isLoading = false,
                channelID = channelID,
                parentID = parentID,
                myname = myname,
                title = existing.title,
                head = existing,
                messages = db.getThreads(parentID),
                canPost = canPost,
                error = if (channel == null) "チャネルが見つかりません" else null
            )
        }

        // 新規スレッド（vue の makeThreadHead と同じ作り方で組み立てる）
        var title = "新規スレッド"
        var aliasNames = listOf(myname)
        if (parentID.contains('@')) {
            val parts = parentID.split('@')
            val toWhom = if (parts.firstOrNull() == myname) parts.getOrNull(1).orEmpty()
            else parts.firstOrNull().orEmpty()
            title = toWhom.take(12)
            aliasNames = parts.toMutableList().apply {
                // グループ名が含まれていれば、そのメンバーを全員に含める（vue と同じ）
                parts.forEach { part ->
                    groups.find { it.groupName == part }
                        ?.let { addAll(it.aliasNamesJson.toNameList()) }
                }
            }.distinct()
        }

        val draft = DbThreadHead(
            parentID = parentID,
            channelID = channelID,
            title = title,
            messageTxt = if (parentID.contains('@')) title else "",
            description = "",
            aliasName = myname,
            aliasNamesJson = org.json.JSONArray(aliasNames).toString(),
            adminNamesJson = "[]",
            displayStatus = 0,
            broadcastFlag = 0,
            updatedAt = ""
        )

        return ThreadUiState(
            isLoading = false,
            channelID = channelID,
            parentID = parentID,
            myname = myname,
            title = title,
            head = draft,
            messages = emptyList(),
            isNewThread = true,
            error = if (channel == null) "チャネルが見つかりません" else null
        )
    }

    /** 開いたら既読にする（vue の readStatus と同じ）。 */
    private fun markRead() {
        val head = _uiState.value.head ?: return
        if (head.displayStatus != 1 && head.displayStatus != 2) return
        viewModelScope.launch(Dispatchers.IO) {
            ChannelDbHelper(context).saveThreadHead(head.copy(displayStatus = 0))
        }
    }
    /**
     * メッセージを投稿する（新規 / 編集 / 削除）。
     *
     * 新規スレッドのときは先に threadHead を作り、その後でメッセージを投げる
     * （vue の EditBox と同じ順序）。
     */
    fun postMessage(text: String, editingMessageID: String? = null, asDelete: Boolean = false) {
        val state = _uiState.value
        if (state.isPosting || !state.canPost) return
        if (!asDelete && editingMessageID == null && text.isBlank()) return

        viewModelScope.launch {
            _uiState.value = state.copy(isPosting = true, error = null)

            val body = if (asDelete) "" else text
            val head = state.head

            // 新規スレッドなら先に threadHead を送る
            if (head != null && state.isNewThread) {
                val savedHead = head.copy(
                    messageTxt = body,
                    title = removeMark(body).take(30).ifBlank { state.title },
                    adminNamesJson = org.json.JSONArray(listOf(state.myname)).toString()
                )
                val headResult = repository.postThreadHead(state.channelID, state.myname, savedHead)
                if (headResult.isFailure) {
                    _uiState.value = _uiState.value.copy(
                        isPosting = false,
                        error = headResult.exceptionOrNull()?.message
                    )
                    return@launch
                }
            }

            val aliasNames = head?.aliasNamesJson?.toNameList() ?: listOf(state.myname)
            val totalNames = (aliasNames + extractMentions(body) + state.myname).distinct()

            val result = repository.postMessage(
                channelID = state.channelID,
                myname = state.myname,
                parentID = state.parentID,
                messageTxt = body,
                aliasImg = "",
                totalNames = totalNames,
                backID = head?.backID.orEmpty(),
                yets = emptyList(),
                groupName = "",
                editingMessageID = editingMessageID
            )

            _uiState.value = _uiState.value.copy(
                isPosting = false,
                error = result.exceptionOrNull()?.message
            )
            if (result.isSuccess) reload()
        }
    }

    /** 絵文字リアクションを付ける／外す。 */
    fun postEmoji(messageID: String, emoji: String, parentID: String, delete: Boolean) {
        val state = _uiState.value
        viewModelScope.launch {
            val aliasNames = state.head?.aliasNamesJson?.toNameList() ?: listOf(state.myname)
            repository.postEmoji(
                channelID = state.channelID,
                myname = state.myname,
                pushNames = (aliasNames + state.myname).distinct(),
                messageID = messageID,
                emoji = emoji,
                parentID = parentID,
                delete = delete
            )
            reload()
        }
    }

    /** push 反映後に画面を最新化する。 */
    fun reload() {
        val state = _uiState.value
        viewModelScope.launch {
            val messages = withContext(Dispatchers.IO) {
                ChannelDbHelper(context).getThreads(state.parentID)
            }
            _uiState.value = _uiState.value.copy(
                messages = messages,
                // 初回の新規スレッドは保存済みになるのでフラグを落とす
                isNewThread = false
            )
        }
    }

    fun onErrorDismiss() {
        _uiState.value = _uiState.value.copy(error = null)
    }
}
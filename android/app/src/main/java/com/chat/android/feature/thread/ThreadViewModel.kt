package com.chat.android.feature.thread

import android.content.Context
import android.net.Uri
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.BuildConfig
import com.chat.android.core.push.PushReceiveDispatcher
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
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
    /** チャネルのグループ（👥 のグループ選択に使う）。 */
    val groups: List<DbGroup> = emptyList(),
    /** チャネルのエイリアス（自分のアイコン解決に使う）。 */
    val aliases: List<DbAlias> = emptyList(),
    /** DM（parentID に @ を含む）かどうか。通話ボタンの表示条件。 */
    val isDm: Boolean = false,
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

    private var updateJob: kotlinx.coroutines.Job? = null

    fun init(channelID: String?, parentID: String?) {
        val id = channelID ?: return
        val parent = parentID ?: return

        updateJob?.cancel()
        viewModelScope.launch {
            _uiState.value = withContext(Dispatchers.IO) { load(id, parent) }
            markRead()
        }

        updateJob = viewModelScope.launch {
            pushDispatcher.onUpdated.collect { updatedChannelID ->
                if (updatedChannelID == id) {
                    reload()
                }
            }
        }
    }

    private fun load(channelID: String, parentID: String): ThreadUiState {
        val db = ChannelDbHelper(context)
        val channel = db.getChannel(channelID)
        val myname = channel?.myname.orEmpty()
        val groups = db.getGroupsForChannel(channelID)
        val aliases = db.getAliasesForChannel(channelID)
        val isDm = parentID.contains('@')
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
                groups = groups,
                aliases = aliases,
                isDm = isDm,
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
            groups = groups,
            aliases = aliases,
            isDm = isDm,
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
     * （vue の EditBox と同じ順序）。メンションがあるときも threadHead を送る。
     * [task] は 🔖（タスク化）、[group] は 👥（グループ投稿）、
     * [fileUris] は 🌄 で選んだ添付ファイル。
     */
    fun postMessage(
        text: String,
        editingMessageID: String? = null,
        asDelete: Boolean = false,
        task: Boolean = false,
        group: DbGroup? = null,
        fileUris: List<Uri> = emptyList()
    ) {
        val state = _uiState.value
        if (state.isPosting || !state.canPost) return
        if (!asDelete && editingMessageID == null && text.isBlank() && fileUris.isEmpty()) return

        viewModelScope.launch {
            _uiState.value = state.copy(isPosting = true, error = null)

            val body = if (asDelete) "" else text
            val head = state.head

            // 自分のエイリアス（アイコン解決）。グループ選択時はグループのアイコンを使う。
            val myAlias = state.aliases.firstOrNull { it.aliasName == state.myname }
            val aliasImg = group?.groupImg?.takeIf { it.isNotBlank() }
                ?: myAlias?.aliasImg.orEmpty()
            val groupName =
                if (group == null || group.groupName == "グループなし") "" else group.groupName

            // メンションと、タスク（🔖）時の ☑️ リアクション（vue の yets と同じ）
            val mentions = extractMentions(body)
            val yets = if (task && !state.isDm) mentions.map { it to "☑️" } else emptyList()

            val aliasNames = head?.aliasNamesJson?.toNameList() ?: listOf(state.myname)
            val totalNames = (mentions + aliasNames + state.myname).distinct()

            // 新規スレッド、頭の編集、またはメンションがあるときは threadHead を先に送る
            if (head != null && (state.isNewThread || editingMessageID == state.parentID || mentions.isNotEmpty())) {
                val isHeadEdit = editingMessageID == state.parentID
                val files = withContext(Dispatchers.IO) { readFiles(fileUris) }
                val savedHead = head.copy(
                    messageTxt = if (state.isNewThread || isHeadEdit) body else head.messageTxt,
                    title = if (state.isNewThread || isHeadEdit) {
                        removeMark(body).take(30).ifBlank { state.title }
                    } else head.title,
                    aliasNamesJson = org.json.JSONArray(totalNames).toString(),
                    adminNamesJson = if (state.isNewThread) {
                        org.json.JSONArray(listOf(state.myname)).toString()
                    } else head.adminNamesJson
                )
                val headResult = repository.postThreadHead(state.channelID, state.myname, savedHead, files)
                if (headResult.isFailure) {
                    _uiState.value = _uiState.value.copy(
                        isPosting = false,
                        error = headResult.exceptionOrNull()?.message
                    )
                    return@launch
                }
                // 頭の編集のみの場合はここで終了（threadEdit を送る必要はない）
                if (isHeadEdit) {
                    _uiState.value = _uiState.value.copy(isPosting = false)
                    reload()
                    return@launch
                }
            }

            val files = withContext(Dispatchers.IO) { readFiles(fileUris) }
            val result = repository.postMessage(
                channelID = state.channelID,
                myname = state.myname,
                parentID = state.parentID,
                messageTxt = body,
                aliasImg = aliasImg,
                totalNames = totalNames,
                backID = head?.backID.orEmpty(),
                yets = yets,
                groupName = groupName,
                editingMessageID = editingMessageID,
                files = files
            )

            _uiState.value = _uiState.value.copy(
                isPosting = false,
                error = result.exceptionOrNull()?.message
            )
            if (result.isSuccess) reload()
        }
    }

    /**
     * DM の通話を開始する（vue/src/components/EditBox.vue の calling と同じ）。
     * 呼び出しメッセージを送り、通話URLを [onOpen] に渡す。
     */
    fun postCall(onOpen: (String) -> Unit) {
        val state = _uiState.value
        val head = state.head ?: return
        viewModelScope.launch {
            val names = state.parentID.split('@')
            val opponent = if (names.firstOrNull() == state.myname) {
                names.getOrNull(1).orEmpty()
            } else {
                names.firstOrNull().orEmpty()
            }
            val callUrl = BuildConfig.BASE_URL.trimEnd('/') +
                "/threadCall/" + randomCode(8) + "/"
            val aliasImg =
                state.aliases.firstOrNull { it.aliasName == state.myname }?.aliasImg.orEmpty()
            val result = repository.postCall(
                channelID = state.channelID,
                myname = state.myname,
                parentID = state.parentID,
                opponent = opponent,
                aliasImg = aliasImg,
                backID = head.backID,
                callUrl = callUrl
            )
            if (result.isFailure) {
                _uiState.value = _uiState.value.copy(error = result.exceptionOrNull()?.message)
                return@launch
            }
            reload()
            onOpen(callUrl)
        }
    }

    /** 選択されたファイル Uri を読み込んで送信できる形にする。 */
    private fun readFiles(uris: List<Uri>): List<ThreadFile> = uris.mapNotNull { uri ->
        runCatching {
            val resolver = context.contentResolver
            val bytes = resolver.openInputStream(uri)?.use { it.readBytes() }
                ?: return@runCatching null
            ThreadFile(
                fileName = queryDisplayName(uri) ?: uri.lastPathSegment.orEmpty().ifBlank { "file" },
                mimeType = resolver.getType(uri) ?: "application/octet-stream",
                bytes = bytes
            )
        }.getOrNull()
    }

    private fun queryDisplayName(uri: Uri): String? = runCatching {
        context.contentResolver.query(uri, null, null, null, null)?.use { cursor ->
            val index = cursor.getColumnIndex(android.provider.OpenableColumns.DISPLAY_NAME)
            if (index >= 0 && cursor.moveToFirst()) cursor.getString(index) else null
        }
    }.getOrNull()

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
            val db = ChannelDbHelper(context)
            val updatedState = withContext(Dispatchers.IO) {
                val head = db.getThreadHead(state.parentID)
                val messages = db.getThreads(state.parentID)
                if (head != null) {
                    val canPost = head.broadcastFlag == 0 ||
                            head.adminNamesJson.toNameList().contains(state.myname)
                    state.copy(
                        head = head,
                        messages = messages,
                        title = head.title,
                        canPost = canPost,
                        isNewThread = false
                    )
                } else {
                    state.copy(
                        messages = messages,
                        isNewThread = false
                    )
                }
            }
            _uiState.value = updatedState
        }
    }

    fun onErrorDismiss() {
        _uiState.value = _uiState.value.copy(error = null)
    }
}
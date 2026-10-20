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
    text
        .replace(Regex("＊(?:太|p)＊|・＊(?:太|p)＊|〜〜|・〜〜|＜quote＞|＜・quote＞|｀｀｀|・｀｀｀|色＊赤|赤＊色"), "")
        .replace(Regex("＠＠|・＠＠"), "")
        .replace(Regex("「([^」]*)」（[^）]*）"), "\$1")
        .replace(Regex("!?\\[([^\\]]*)\\]\\([^)]*\\)"), "\$1")
        .replace(Regex("(?m)^\\s{0,3}#{1,6}\\s+|^\\s*>\\s?"), "")
        .replace(Regex("[*_~`]"), "")
        .replace(Regex("\\s+"), " ")
        .trim()

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

    fun init(channelID: String?, parentID: String?, backID: String? = null) {
        val id = channelID ?: return
        val parent = parentID ?: return

        updateJob?.cancel()
        _uiState.value = ThreadUiState(
            isLoading = true,
            channelID = id,
            parentID = parent
        )
        viewModelScope.launch {
            _uiState.value = withContext(Dispatchers.IO) { load(id, parent, backID.orEmpty()) }
            markRead()
            
            // Save local dataStore when coming to thread screen
            val currentState = _uiState.value
            if (currentState.myname.isNotBlank()) {
                repository.saveLocalProfile(id, currentState.myname, currentState.aliases.find { it.aliasName == currentState.myname }?.aliasImg.orEmpty())
            }
        }

        updateJob = viewModelScope.launch {
            pushDispatcher.onUpdated.collect { updatedChannelID ->
                if (updatedChannelID == id) {
                    reload()
                }
            }
        }
    }

    private fun load(channelID: String, parentID: String, backID: String): ThreadUiState {
        val db = ChannelDbHelper(context)
        val channel = db.getChannel(channelID)
        val myname = channel?.myname.orEmpty()
        val groups = db.getGroupsForChannel(channelID)
        val aliases = db.getAliasesForChannel(channelID)
        val isDm = parentID.contains('@')
        val existing = db.getThreadHead(parentID)

        if (existing != null) {
            // DM の場合はタイトルを相手の名前に固定する（既存の保存タイトルより優先）
            val displayTitle = if (isDm) {
                val parts = parentID.split('@')
                if (parts.size == 2) {
                    if (parts[0] == myname) parts[1] else parts[0]
                } else existing.title
            } else existing.title

            // vue の EditBox: broadcastFlag が立っていれば管理者のみ投稿可
            val canPost = existing.broadcastFlag == 0 ||
                existing.adminNamesJson.toNameList().contains(myname)
            return ThreadUiState(
                isLoading = false,
                channelID = channelID,
                parentID = parentID,
                myname = myname,
                title = displayTitle,
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

        val replySource = if (backID.isNotBlank()) db.getThread(parentID) else null
        val sourceHead = replySource?.let { db.getThreadHead(it.parentID) }
        val replyDraft = replySource != null && sourceHead != null
        val initialAliases = if (replyDraft) {
            (sourceHead!!.aliasNamesJson.toNameList() + replySource!!.aliasName).distinct()
        } else aliasNames
        val draft = DbThreadHead(
            parentID = parentID,
            channelID = channelID,
            title = if (replyDraft) removeMark(replySource!!.messageTxt).take(12) else title,
            messageTxt = if (replyDraft) replySource!!.messageTxt else if (parentID.contains('@')) title else "",
            description = "",
            aliasName = if (replyDraft) replySource!!.aliasName else myname,
            aliasNamesJson = org.json.JSONArray(initialAliases).toString(),
            adminNamesJson = sourceHead?.adminNamesJson ?: "[]",
            displayStatus = 0,
            broadcastFlag = sourceHead?.broadcastFlag ?: 0,
            updatedAt = "",
            backID = if (replyDraft) backID else ""
        )

        return ThreadUiState(
            isLoading = false,
            channelID = channelID,
            parentID = parentID,
            myname = myname,
            title = draft.title,
            head = draft,
            messages = emptyList(),
            isNewThread = true,
            canPost = draft.broadcastFlag == 0 || draft.adminNamesJson.toNameList().contains(myname),
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

            val files = try {
                withContext(Dispatchers.IO) { readFiles(fileUris) }
            } catch (error: Exception) {
                _uiState.value = _uiState.value.copy(
                    isPosting = false,
                    error = error.message ?: "添付ファイルを読み込めませんでした"
                )
                return@launch
            }

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
                val isReplyDraft = state.isNewThread && head.backID.isNotBlank()
                val savedHead = head.copy(
                    messageTxt = if ((state.isNewThread && !isReplyDraft) || isHeadEdit) body else head.messageTxt,
                    title = if ((state.isNewThread && !isReplyDraft) || isHeadEdit) {
                        removeMark(body).take(30).ifBlank { state.title }
                    } else head.title,
                    aliasNamesJson = org.json.JSONArray(totalNames).toString(),
                    adminNamesJson = if (state.isNewThread && !isReplyDraft) {
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

            val result = repository.postMessage(
                channelID = state.channelID,
                providedMyname = state.myname,
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

    /** ブックマークを保存する（サーバーと同期）。 */
    fun toggleBookmark(messageID: String) {
        val state = _uiState.value
        viewModelScope.launch {
            val isBookmarked = if (messageID == state.parentID) {
                state.head?.isBookmarked ?: false
            } else {
                state.messages.find { it.messageID == messageID }?.isBookmarked ?: false
            }

            val aliasNames = state.head?.aliasNamesJson?.toNameList() ?: listOf(state.myname)
            val result = repository.postBookmark(
                channelID = state.channelID,
                providedMyname = state.myname,
                pushNames = (aliasNames + state.myname).distinct(),
                messageID = messageID,
                parentID = state.parentID,
                isBookmarked = !isBookmarked
            )

            if (result.isSuccess) {
                reload()
            } else {
                _uiState.value = _uiState.value.copy(error = result.exceptionOrNull()?.message)
            }
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
                providedMyname = state.myname,
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
            onOpen(randomCodeFromUrl(callUrl))
        }
    }

    private fun randomCodeFromUrl(url: String): String = url.trimEnd('/').substringAfterLast('/')

    /** 選択されたファイル Uri を読み込んで送信できる形にする。 */
    private fun readFiles(uris: List<Uri>): List<ThreadFile> = uris.map { uri ->
        try {
            val resolver = context.contentResolver
            val bytes = resolver.openInputStream(uri)?.use { it.readBytes() }
                ?: throw IllegalStateException("添付ファイルを開けませんでした")
            val fileName = queryDisplayName(uri)
                ?: uri.lastPathSegment.orEmpty().ifBlank { "file" }
            if (bytes.isEmpty()) {
                throw IllegalStateException("添付ファイルが空です: $fileName")
            }
            ThreadFile(
                fileName = fileName,
                mimeType = resolver.getType(uri) ?: "application/octet-stream",
                bytes = bytes
            )
        } catch (error: Exception) {
            throw IllegalStateException(
                "添付ファイルを読み込めませんでした: ${error.message ?: uri.lastPathSegment.orEmpty()}",
                error
            )
        }
    }

    private fun queryDisplayName(uri: Uri): String? = runCatching {
        context.contentResolver.query(uri, null, null, null, null)?.use { cursor ->
            val index = cursor.getColumnIndex(android.provider.OpenableColumns.DISPLAY_NAME)
            if (index >= 0 && cursor.moveToFirst()) cursor.getString(index) else null
        }
    }.getOrNull()

    /** 絵文字リアクションを付ける／外す。 */
    fun postEmoji(messageID: String, emoji: String, parentID: String, forceDelete: Boolean = false) {
        val state = _uiState.value
        viewModelScope.launch {
            // Vue の clickEmoji 相当のトグルロジック:
            // 既にその絵文字を付けていれば削除、そうでなければ追加。
            val targetJson = if (messageID == parentID) {
                state.head?.emojisJson
            } else {
                state.messages.find { it.messageID == messageID }?.emojisJson
            }

            val isAlreadyReacted = runCatching {
                val array = org.json.JSONArray(targetJson ?: "[]")
                (0 until array.length()).any { i ->
                    val obj = array.optJSONObject(i)
                    obj?.optString("aliasName") == state.myname && obj.optString("emoji") == emoji
                }
            }.getOrDefault(false)

            val delete = forceDelete || isAlreadyReacted
            val aliasNames = state.head?.aliasNamesJson?.toNameList() ?: listOf(state.myname)

            val result = repository.postEmoji(
                channelID = state.channelID,
                providedMyname = state.myname,
                pushNames = (aliasNames + state.myname).distinct(),
                messageID = messageID,
                emoji = emoji,
                parentID = parentID,
                delete = delete
            )
            if (result.isSuccess) {
                reload()
            } else {
                _uiState.value = _uiState.value.copy(error = result.exceptionOrNull()?.message)
            }
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

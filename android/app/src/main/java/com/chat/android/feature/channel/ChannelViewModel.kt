package com.chat.android.feature.channel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collectLatest
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ChannelUiState(
    val isLoading: Boolean = false,
    val isCreateMode: Boolean = true,
    val channelID: String? = null,
    val channelName: String = "",
    val channelDescription: String = "",
    val myName: String = "",
    val myImg: String = "",
    val invitationCode: String = "",
    val invitationGuestCode: String = "",
    val aliases: List<DbAlias> = emptyList(),
    val groups: List<DbGroup> = emptyList(),
    val error: String? = null,
    val successMessage: String? = null,
    val iamAdmin: Boolean = false,
    val iamGuest: Boolean = false,
    /** ローカルSQLite (channel テーブル) から読んでいます。 */
    val channels: List<DbChannel> = emptyList()
)

@HiltViewModel
class ChannelViewModel @Inject constructor(
    private val repository: ChannelRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(ChannelUiState())
    val uiState: StateFlow<ChannelUiState> = _uiState.asStateFlow()

    fun init(channelID: String?) {
        _uiState.value = _uiState.value.copy(
            channelID = channelID,
            isCreateMode = channelID == null
        )

        // 一覧はどのモードでも表示するので常に取得する。
        loadChannelList()
        observeDbUpdates(channelID)

        if (channelID != null) {
            loadChannelData(channelID)
        }
    }

    /** ローカルSQLite の channel テーブルからチャネル一覧を読み込む。 */
    fun loadChannelList() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(
                channels = runCatching { repository.getAllChannels() }
                    .getOrDefault(emptyList())
            )
        }
    }

    /**
     * 一覧で選んだチャネルの詳細を表示する。
     *
     * ナビゲーションのルート引数は初期選択としてだけ使う。同じ画面に一覧と詳細を
     * 置いており、遷移のたびにバックスタックを積み上げないため。
     */
    fun selectChannel(channelID: String) {
        if (_uiState.value.channelID == channelID) return
        _uiState.value = _uiState.value.copy(
            channelID = channelID,
            isCreateMode = false,
            error = null,
            successMessage = null,
            isLoading = true
        )
        observeDbUpdates(channelID)
        loadChannelData(channelID)
    }

    /** 一覧の「+ 新規」から登録モードへ戻る。 */
    fun startCreateMode() {
        observeJob?.cancel()
        _uiState.value = _uiState.value.copy(
            channelID = null,
            isCreateMode = true,
            channelName = "",
            channelDescription = "",
            myName = "",
            myImg = "",
            aliases = emptyList(),
            groups = emptyList(),
            invitationCode = "",
            invitationGuestCode = "",
            iamAdmin = false,
            iamGuest = false,
            error = null,
            successMessage = null,
            isLoading = false
        )
    }

    private fun loadChannelData(id: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            val channel = repository.getChannel(id)
            if (channel != null) {
                val aliases = repository.getAliases(id)
                val groups = repository.getGroups(id)
                
                val iamAdmin = aliases.any { it.aliasName == channel.myname && it.accessRight == "admin" }
                val iamGuest = aliases.any { it.aliasName == channel.myname && it.accessRight == "guest" }

                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    channelName = channel.channelName,
                    channelDescription = channel.channelDescription,
                    myName = channel.myname,
                    myImg = channel.myimg,
                    invitationCode = channel.invitationCode,
                    invitationGuestCode = channel.invitationGuestCode,
                    aliases = aliases,
                    groups = groups,
                    iamAdmin = iamAdmin,
                    iamGuest = iamGuest
                )
            } else {
                _uiState.value = _uiState.value.copy(isLoading = false, error = "Channel not found")
            }
        }
    }

    /**
     * FCM 受信やローカルDB更新のたびに詳細と一覧を再読込する。
     * init() が何度も呼ばれても購読が重複しないように、Job を持つ。
     */
    private var observeJob: kotlinx.coroutines.Job? = null

    private fun observeDbUpdates(id: String?) {
        observeJob?.cancel()
        observeJob = viewModelScope.launch {
            repository.dbUpdateFlow.collectLatest {
                loadChannelList()
                if (id != null) {
                    loadChannelData(id)
                }
            }
        }
    }

    fun onNameChange(name: String) { _uiState.value = _uiState.value.copy(channelName = name) }
    fun onDescriptionChange(desc: String) { _uiState.value = _uiState.value.copy(channelDescription = desc) }
    fun onMyNameChange(name: String) { _uiState.value = _uiState.value.copy(myName = name) }
    fun onMyImgChange(img: String) { _uiState.value = _uiState.value.copy(myImg = img) }
    fun onErrorDismiss() { _uiState.value = _uiState.value.copy(error = null, successMessage = null) }

    /** 通知対象にするエイリアス名（Vue と同じ除外条件）。 */
    private fun pushTargetNames(aliases: List<DbAlias>): List<String> =
        aliases
            .filter { it.accessRight != "guest" && it.accessRight != "inquirer" }
            .map { it.aliasName }

    fun save() {
        viewModelScope.launch {
            val state = _uiState.value
            _uiState.value = _uiState.value.copy(isLoading = true, error = null, successMessage = null)

            val result = if (state.isCreateMode) {
                repository.createChannel(
                    state.channelName,
                    MarkdownCodec.fromHtml(state.channelDescription),
                    state.myName,
                    state.myImg
                )
            } else {
                repository.editChannel(
                    channelID = state.channelID!!,
                    updatedBy = state.myName,
                    // サーバーは pushNames から通知先の userID を解決する。
                    pushNamesJson = jsonArrayOfStrings(pushTargetNames(state.aliases)),
                    channelName = state.channelName,
                    // groups は意図的に null のままにする。
                    // サーバーは送られた group ごとに common.ImgSave() を呼ぶが
                    // (controller/ChannelEdit.go:152)、ローカルDBには /img/... のような
                    // 保存済みパスが入っているため "Emoji invalid:" でエラーになり
                    // チャネル名などの保存自体が失敗する
                    // (common/file.go:96)。diff が空ならサーバーは既存の groups を保持する。
                    groupsJson = null,
                    description = MarkdownCodec.fromHtml(state.channelDescription)
                )
            }

            result.onSuccess {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    successMessage = "保存しました",
                    error = null
                )
                loadChannelList()
                if (state.isCreateMode && it is String && it.isNotEmpty()) {
                    init(it)
                } else if (state.channelID != null) {
                    loadChannelData(state.channelID)
                }
            }.onFailure {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = it.message,
                    successMessage = null
                )
            }
        }
    }

    /** 招待コードを生成する（Vue の invite() に対応）。 */
    fun generateInvitation(guest: Boolean) {
        val state = _uiState.value
        val id = state.channelID ?: return
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null, successMessage = null)
            repository.generateInvitation(id, state.myName, guest)
                .onSuccess {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        successMessage = "招待URLを生成しました"
                    )
                    loadChannelData(id)
                }
                .onFailure {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        error = it.message
                    )
                }
        }
    }

    fun deleteChannel() {
        val id = _uiState.value.channelID ?: return
        val myName = _uiState.value.myName
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true, error = null, successMessage = null)
            repository.deleteChannel(id, myName).onSuccess {
                 _uiState.value = _uiState.value.copy(isLoading = false, successMessage = "削除しました")
                 loadChannelList()
            }.onFailure {
                 _uiState.value = _uiState.value.copy(isLoading = false, error = it.message)
            }
        }
    }
}

/** 文字列配列をサーバーが期待する JSON 配列文字列にする。 */
private fun jsonArrayOfStrings(values: List<String>): String =
    values.joinToString(separator = ",", prefix = "[", postfix = "]") { value ->
        "\"" + value.replace("\\", "\\\\").replace("\"", "\\\"") + "\""
    }

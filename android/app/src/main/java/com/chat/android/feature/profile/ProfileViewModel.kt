package com.chat.android.feature.profile

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.util.RandomAvatar
import com.chat.android.feature.channel.ChannelRepository
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbChannel
import com.chat.android.feature.channel.DbGroup
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

/**
 * /profile/{id}/?code=... 用の状態（vue/src/views/Profile.vue に対応）。
 *
 * code があるとき = 招待から参加する画面、
 * code がないとき = 参加済みチャネルのプロフィール編集画面。
 */
data class ProfileUiState(
    val isLoading: Boolean = false,
    val channelID: String? = null,
    val code: String? = null,
    /** 参加画面（code あり）か、編集画面（code なし）か。 */
    val isJoinMode: Boolean = false,
    val channel: DbChannel? = null,
    val aliasName: String = "",
    val aliasImg: String = "",
    val aliasBio: String = "",
    val aliases: List<DbAlias> = emptyList(),
    val joinGroups: List<DbGroup> = emptyList(),
    val sameUserAliases: List<DbAlias> = emptyList(),
    val error: String? = null,
    val successMessage: String? = null,
    val joined: Boolean = false
)

@HiltViewModel
class ProfileViewModel @Inject constructor(
    private val repository: ChannelRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(ProfileUiState())
    val uiState: StateFlow<ProfileUiState> = _uiState.asStateFlow()

    private var observeJob: Job? = null

    fun init(channelID: String?, code: String?) {
        _uiState.value = _uiState.value.copy(
            channelID = channelID,
            code = code,
            isJoinMode = !code.isNullOrBlank(),
            aliasImg = RandomAvatar.random(),
            error = null,
            successMessage = null
        )
        observeJob?.cancel()
        observeJob = viewModelScope.launch {
            repository.dbUpdateFlow.collect { load() }
        }
        load()
    }

    private fun load() {
        val state = _uiState.value
        val id = state.channelID ?: return
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            val channel = repository.getChannel(id)
            val aliases = repository.getAliases(id)
            val groups = repository.getGroups(id)

            if (channel == null && !state.isJoinMode) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "このチャネルは未参加です。招待URLから参加してください。"
                )
                return@launch
            }

            // 自分のエイリアスは channel.myname に記録されている
            val myName = channel?.myname.orEmpty()
            val myAlias = aliases.find { it.aliasName == myName }
            // 参加グループ: aliasNames に自分のニックネームを含むもの
            val joinGroups = groups.filter { it.aliasNamesJson.contains(myName) }
            // マイニックネーム: 同じ userID の他のエイリアス
            val sameUserAliases = if (myAlias == null) emptyList() else aliases.filter {
                it.userID == myAlias.userID && it.aliasID != myAlias.aliasID
            }

            _uiState.value = _uiState.value.copy(
                isLoading = false,
                channel = channel,
                aliases = aliases,
                joinGroups = joinGroups,
                sameUserAliases = sameUserAliases,
                aliasName = if (state.isJoinMode) _uiState.value.aliasName else myName,
                aliasBio = if (state.isJoinMode) "" else myAlias?.aliasBio.orEmpty()
            )
        }
    }

    fun onNameChange(name: String) { _uiState.value = _uiState.value.copy(aliasName = name) }
    fun onBioChange(bio: String) { _uiState.value = _uiState.value.copy(aliasBio = bio) }
    fun onImgChange(img: String) { _uiState.value = _uiState.value.copy(aliasImg = img) }
    fun onErrorDismiss() { _uiState.value = _uiState.value.copy(error = null, successMessage = null) }

    /** 「参加」または「保存」を実行する。 */
    fun submit() {
        val state = _uiState.value
        val id = state.channelID ?: return
        if (state.aliasName.isBlank()) {
            _uiState.value = state.copy(error = "ニックネームを入力してください")
            return
        }
        viewModelScope.launch {
            _uiState.value = state.copy(isLoading = true, error = null, successMessage = null)
            val result = if (state.isJoinMode) {
                val code = state.code
                if (code.isNullOrBlank()) {
                    _uiState.value = state.copy(isLoading = false, error = "招待コードがありません")
                    return@launch
                }
                repository.joinChannel(id, code, state.aliasName, state.aliasImg)
            } else {
                repository.editAlias(
                    channelID = id,
                    updatedBy = state.channel?.myname.orEmpty(),
                    aliasName = state.aliasName,
                    aliasBio = state.aliasBio,
                    aliasImg = state.aliasImg
                )
            }

            result.onSuccess {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    joined = true,
                    successMessage = if (state.isJoinMode) "参加しました" else "保存しました",
                    error = null
                )
                load()
            }.onFailure {
                _uiState.value = _uiState.value.copy(isLoading = false, error = it.message)
            }
        }
    }

    /** 他のニックネームへ切り替える（Profile.vue の switchAlias）。 */
    fun switchAlias(aliasName: String) {
        val id = _uiState.value.channelID ?: return
        viewModelScope.launch {
            repository.switchChannelMyname(id, aliasName)
            _uiState.value = _uiState.value.copy(successMessage = "ニックネームを切り替えました")
            load()
        }
    }
}

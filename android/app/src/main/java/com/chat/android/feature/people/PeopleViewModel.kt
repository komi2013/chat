package com.chat.android.feature.people

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.feature.channel.ChannelRepository
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbChannel
import com.chat.android.feature.channel.DbGroup
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import org.json.JSONArray
import javax.inject.Inject

/**
 * People 画面の状態（vue/src/views/People.vue に対応）。
 *
 * name はエイリアス名とグループ名のどちらでも受け付ける。
 * Vue と同じ判定で、グループ名が一致した場合はグループとして表示する。
 */
data class PeopleUiState(
    val isLoading: Boolean = true,
    val channel: DbChannel? = null,
    val isGroup: Boolean = false,
    val personName: String = "",
    val personImage: String = "",
    val personBio: String = "",
    /** 「ニックネーム一覧」。グループ表示時はグループメンバー、個人は同じ人の他ニックネーム。 */
    val myAliases: List<DbAlias> = emptyList(),
    /** 「参加グループ一覧」。個人表示のときだけ使う。 */
    val joinGroups: List<DbGroup> = emptyList(),
    val error: String? = null
)

@HiltViewModel
class PeopleViewModel @Inject constructor(
    private val repository: ChannelRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(PeopleUiState())
    val uiState: StateFlow<PeopleUiState> = _uiState.asStateFlow()

    fun init(channelID: String?, name: String?) {
        val id = channelID ?: return
        val target = name ?: return

        viewModelScope.launch {
            _uiState.value = PeopleUiState(isLoading = true)
            val channel = repository.getChannel(id)
            val groups = repository.getGroups(id)
            val aliases = repository.getAliases(id)

            if (channel == null) {
                _uiState.value = PeopleUiState(isLoading = false, error = "チャネルが見つかりません")
                return@launch
            }

            val group = groups.find { it.groupName == target }
            if (group != null) {
                // Vue と同じ。グループ名で渡された場合はグループとして扱う。
                val names = group.aliasNamesJson.toNameList()
                _uiState.value = PeopleUiState(
                    isLoading = false,
                    channel = channel,
                    isGroup = true,
                    personName = group.groupName,
                    personImage = group.groupImg,
                    personBio = "",
                    myAliases = aliases.filter { it.aliasName in names }
                )
                return@launch
            }

            val alias = aliases.find { it.aliasName == target }
            if (alias == null) {
                _uiState.value = PeopleUiState(isLoading = false, error = "ユーザーが見つかりません")
                return@launch
            }

            _uiState.value = PeopleUiState(
                isLoading = false,
                channel = channel,
                isGroup = false,
                personName = alias.aliasName,
                personImage = alias.aliasImg,
                personBio = alias.aliasBio,
                // 同じ userID の他ニックネーム（Vue の sameUserAliases 相当）
                myAliases = aliases.filter {
                    it.userID == alias.userID && it.aliasID != alias.aliasID
                },
                // 参加グループ一覧は自分が参加しているグループ（Vue と同じ）
                joinGroups = groups.filter { channel.myname in it.aliasNamesJson.toNameList() }
            )
        }
    }
}

/** aliasNames の JSON 文字列をリストへ。壊れている場合は空リスト。 */
private fun String.toNameList(): List<String> = runCatching {
    val array = JSONArray(this)
    (0 until array.length()).mapNotNull { index -> array.optString(index, "").ifBlank { null } }
}.getOrDefault(emptyList())
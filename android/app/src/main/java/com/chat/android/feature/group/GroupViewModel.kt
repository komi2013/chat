package com.chat.android.feature.group

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
import org.json.JSONArray
import org.json.JSONObject
import javax.inject.Inject

/** 編集中のグループ（Group.vue の groups に対応）。 */
data class EditableGroup(
    val groupID: String,
    val groupName: String,
    val groupImg: String,
    val aliasNames: List<String>,
    val groupBio: String = "",
    /** 編集可能か（Vue と同じ判定）。 */
    val editable: Boolean = true,
    /** 今回新建したグループか（名前をテキスト入力で編集できる）。 */
    val isNew: Boolean = false,
    /** 削除対象としてチェック済みか。 */
    val removed: Boolean = false
)

data class GroupUiState(
    val isLoading: Boolean = false,
    val channelID: String? = null,
    val channel: DbChannel? = null,
    val groups: List<EditableGroup> = emptyList(),
    val aliases: List<DbAlias> = emptyList(),
    val error: String? = null,
    val successMessage: String? = null
)

@HiltViewModel
class GroupViewModel @Inject constructor(
    private val repository: ChannelRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(GroupUiState())
    val uiState: StateFlow<GroupUiState> = _uiState.asStateFlow()

    /** 読み込み時点のスナップショット（差分判定の基準）。 */
    private var original: List<EditableGroup> = emptyList()
    private var observeJob: Job? = null

    fun init(channelID: String?) {
        _uiState.value = _uiState.value.copy(channelID = channelID, error = null)
        observeJob?.cancel()
        observeJob = viewModelScope.launch {
            repository.dbUpdateFlow.collect { load() }
        }
        load()
    }

    private fun load() {
        val id = _uiState.value.channelID ?: return
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            val channel = repository.getChannel(id)
            val groups = repository.getGroups(id)
            val aliases = repository.getAliases(id)

            if (channel == null) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    error = "チャネルが見つかりません"
                )
                return@launch
            }

            val myName = channel.myname
            val allNames = aliases.map { it.aliasName }.toSet()
            val editableGroups = groups.map { g ->
                val names = g.aliasNamesJson.toAliasNameList()
                // Vue と同じ判定: 既存メンバーのいるグループは自分が含まれる場合のみ編集可
                val hasMember = names.any { it in allNames }
                EditableGroup(
                    groupID = g.groupID,
                    groupName = g.groupName,
                    groupImg = g.groupImg,
                    aliasNames = names,
                    groupBio = g.groupBio,
                    editable = !hasMember || names.contains(myName),
                    isNew = false
                )
            }

            original = editableGroups.map { it.copy(removed = false) }
            _uiState.value = _uiState.value.copy(
                isLoading = false,
                channel = channel,
                aliases = aliases,
                groups = editableGroups,
                error = null
            )
        }
    }

    fun addGroup() {
        val state = _uiState.value
        val myName = state.channel?.myname.orEmpty()
        _uiState.value = state.copy(
            groups = state.groups + EditableGroup(
                groupID = "",
                groupName = "",
                groupImg = "",
                aliasNames = listOf(myName),
                editable = true,
                isNew = true
            )
        )
    }

    fun updateName(groupID: String, name: String) = mutate(groupID) { it.copy(groupName = name) }
    fun updateBio(groupID: String, bio: String) = mutate(groupID) { it.copy(groupBio = bio) }
    fun updateImg(groupID: String, img: String) = mutate(groupID) { it.copy(groupImg = img) }
    fun randomImage(groupID: String) = updateImg(groupID, RandomAvatar.random())

    fun addMember(groupID: String, aliasName: String) = mutate(groupID) {
        if (aliasName in it.aliasNames) it else it.copy(aliasNames = it.aliasNames + aliasName)
    }

    fun removeMember(groupID: String, aliasName: String) = mutate(groupID) {
        it.copy(aliasNames = it.aliasNames - aliasName)
    }

    fun toggleRemoved(groupID: String) = mutate(groupID) { it.copy(removed = !it.removed) }

    private fun mutate(groupID: String, block: (EditableGroup) -> EditableGroup) {
        _uiState.value = _uiState.value.copy(
            groups = _uiState.value.groups.map { if (it.groupID == groupID) block(it) else it }
        )
    }

    /** 変更・削除をまとめて保存する（Group.vue の editGroup / removeGroup）。 */
    fun save() {
        val state = _uiState.value
        val id = state.channelID ?: return
        val myName = state.channel?.myname.orEmpty()
        if (myName.isBlank()) {
            _uiState.value = state.copy(error = "自分のニックネームが不明です")
            return
        }
        val current = state.groups.filter { !it.removed }
        if (current.any { it.groupName.isBlank() }) {
            _uiState.value = state.copy(error = "グループ名を入力してください")
            return
        }

        viewModelScope.launch {
            _uiState.value = state.copy(isLoading = true, error = null, successMessage = null)

            val diffs = JSONArray()
            for (g in current) {
                val pre = original.find { it.groupID == g.groupID }
                if (pre == null || isChanged(pre, g)) {
                    diffs.put(
                        JSONObject().apply {
                            put("groupID", id + g.groupName)
                            put("channelID", id)
                            put("groupName", g.groupName)
                            // 保存済みパス("/" 始まり) は common.ImgSave がそのまま通す
                            // ので、敢えてDataURI へ再エンコードしない。
                            put("groupImg", g.groupImg)
                            put("aliasNames", JSONArray(g.aliasNames))
                            put("groupBio", g.groupBio)
                        }
                    )
                }
            }
            // 削除は aliasNames=null で表現する
            for (pre in original) {
                if (current.none { it.groupID == pre.groupID }) {
                    diffs.put(
                        JSONObject().apply {
                            put("groupName", pre.groupName)
                            put("aliasNames", JSONObject.NULL)
                        }
                    )
                }
            }

            if (diffs.length() == 0) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    successMessage = "変更はありません"
                )
                return@launch
            }

            val pushNames = state.aliases
                .filter { it.accessRight != "guest" && it.accessRight != "inquirer" }
                .joinToString(",", "[", "]") { "\"${it.aliasName.replace("\"", "\\\"")}\"" }

            repository.editGroups(id, myName, diffs.toString(), pushNames)
                .onSuccess {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        successMessage = "保存しました"
                    )
                    load()
                }
                .onFailure {
                    _uiState.value = _uiState.value.copy(isLoading = false, error = it.message)
                }
        }
    }

    private fun isChanged(a: EditableGroup, b: EditableGroup): Boolean =
        a.groupName != b.groupName ||
            a.groupImg != b.groupImg ||
            a.aliasNames != b.aliasNames ||
            a.groupBio != b.groupBio

    fun onErrorDismiss() {
        _uiState.value = _uiState.value.copy(error = null, successMessage = null)
    }
}

/** aliasNames の JSON 文字列をリストへ。壊れている場合は空リスト。 */
private fun String.toAliasNameList(): List<String> = runCatching {
    val array = JSONArray(this)
    (0 until array.length()).mapNotNull { index -> array.optString(index, "").ifBlank { null } }
}.getOrDefault(emptyList())

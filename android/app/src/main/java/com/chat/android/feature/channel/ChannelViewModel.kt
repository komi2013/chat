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
    val iamGuest: Boolean = false
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

        if (channelID != null) {
            loadChannelData(channelID)
            observeDbUpdates(channelID)
        }
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

    private fun observeDbUpdates(id: String) {
        viewModelScope.launch {
            repository.dbUpdateFlow.collectLatest {
                loadChannelData(id)
            }
        }
    }

    fun onNameChange(name: String) { _uiState.value = _uiState.value.copy(channelName = name) }
    fun onDescriptionChange(desc: String) { _uiState.value = _uiState.value.copy(channelDescription = desc) }
    fun onMyNameChange(name: String) { _uiState.value = _uiState.value.copy(myName = name) }
    fun onMyImgChange(img: String) { _uiState.value = _uiState.value.copy(myImg = img) }

    fun save() {
        viewModelScope.launch {
            val state = _uiState.value
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            val result = if (state.isCreateMode) {
                repository.createChannel(
                    state.channelName, 
                    MarkdownCodec.fromHtml(state.channelDescription),
                    state.myName,
                    state.myImg
                )
            } else {
                repository.editChannel(
                    state.channelID!!,
                    state.myName,
                    "[]", // pushNamesJson placeholder
                    state.channelName,
                    null, // groupsJson
                    MarkdownCodec.fromHtml(state.channelDescription)
                )
            }

            result.onSuccess {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    successMessage = "Saved successfully"
                )
                if (state.isCreateMode && it is String) {
                    init(it)
                }
            }.onFailure {
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
            _uiState.value = _uiState.value.copy(isLoading = true)
            repository.deleteChannel(id, myName).onSuccess {
                 _uiState.value = _uiState.value.copy(isLoading = false, successMessage = "Deleted")
            }.onFailure {
                 _uiState.value = _uiState.value.copy(isLoading = false, error = it.message)
            }
        }
    }
}

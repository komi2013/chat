package com.chat.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.data.repository.UserRepository
import com.chat.android.network.ApiService
import com.chat.android.network.ChannelDetail
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ChannelUiState(
    val isLoading: Boolean = false,
    val channels: List<ChannelDetail> = emptyList(),
    val currentChannel: ChannelDetail? = null,
    val channelName: String = "",
    val channelDescription: String = "",
    val myname: String = "",
    val myimg: String = "",
    val errorMessage: String? = null,
    val successMessage: String? = null,
    val isEditing: Boolean = false
)

@HiltViewModel
class ChannelViewModel @Inject constructor(
    private val userRepository: UserRepository,
    private val apiService: ApiService
) : ViewModel() {

    private val _uiState = MutableStateFlow(ChannelUiState())
    val uiState: StateFlow<ChannelUiState> = _uiState.asStateFlow()

    fun loadChannels() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Not authenticated"
                    )
                    return@launch
                }

                val response = apiService.getChannel(csrf)
                if (response.isSuccessful) {
                    val channel = response.body()
                    if (channel != null) {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            currentChannel = channel,
                            channels = listOf(channel)
                        )
                    } else {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            channels = emptyList()
                        )
                    }
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Failed to load channels"
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = "Error: ${e.message}"
                )
            }
        }
    }

    fun loadChannel(channelId: String) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Not authenticated"
                    )
                    return@launch
                }

                val response = apiService.getChannel(csrf, channelId)
                if (response.isSuccessful) {
                    val channel = response.body()
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        currentChannel = channel,
                        channelName = channel?.name ?: "",
                        channelDescription = channel?.description ?: "",
                        myname = channel?.myname ?: "",
                        myimg = channel?.myimg ?: ""
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Failed to load channel"
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = "Error: ${e.message}"
                )
            }
        }
    }

    fun updateChannelName(name: String) {
        _uiState.value = _uiState.value.copy(channelName = name)
    }

    fun updateChannelDescription(description: String) {
        _uiState.value = _uiState.value.copy(channelDescription = description)
    }

    fun updateMyname(name: String) {
        _uiState.value = _uiState.value.copy(myname = name)
    }

    fun updateMyimg(img: String) {
        _uiState.value = _uiState.value.copy(myimg = img)
    }

    fun saveChannel() {
        viewModelScope.launch {
            val currentState = _uiState.value
            val currentChannel = currentState.currentChannel
            if (currentChannel == null) {
                _uiState.value = currentState.copy(
                    errorMessage = "No channel selected"
                )
                return@launch
            }

            _uiState.value = currentState.copy(isLoading = true)

            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) {
                    _uiState.value = currentState.copy(
                        isLoading = false,
                        errorMessage = "Not authenticated"
                    )
                    return@launch
                }

                val response = apiService.editChannel(
                    csrf = csrf,
                    id = currentChannel.id,
                    name = currentState.channelName.takeIf { it.isNotBlank() },
                    description = currentState.channelDescription.takeIf { it.isNotBlank() },
                    myname = currentState.myname.takeIf { it.isNotBlank() },
                    myimg = currentState.myimg.takeIf { it.isNotBlank() }
                )

                if (response.isSuccessful) {
                    val body = response.body()
                    if (body?.csrf != null) {
                        userRepository.setCsrfToken(body.csrf)
                    }
                    
                    _uiState.value = currentState.copy(
                        isLoading = false,
                        isEditing = false,
                        successMessage = "Channel updated successfully"
                    )
                    
                    // Reload channel data
                    loadChannel(currentChannel.id)
                } else {
                    _uiState.value = currentState.copy(
                        isLoading = false,
                        errorMessage = "Failed to update channel"
                    )
                }
            } catch (e: Exception) {
                _uiState.value = currentState.copy(
                    isLoading = false,
                    errorMessage = "Error: ${e.message}"
                )
            }
        }
    }

    fun joinChannel(channelId: String, myname: String, myimg: String? = null) {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Not authenticated"
                    )
                    return@launch
                }

                val response = apiService.joinChannel(
                    csrf = csrf,
                    id = channelId,
                    myname = myname,
                    myimg = myimg
                )

                if (response.isSuccessful) {
                    val body = response.body()
                    if (body?.csrf != null) {
                        userRepository.setCsrfToken(body.csrf)
                    }
                    
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        successMessage = "Joined channel successfully"
                    )
                    
                    userRepository.setCurrentChannelId(channelId)
                    loadChannel(channelId)
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Failed to join channel"
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = "Error: ${e.message}"
                )
            }
        }
    }

    fun startEditing() {
        val currentChannel = _uiState.value.currentChannel
        if (currentChannel != null) {
            _uiState.value = _uiState.value.copy(
                isEditing = true,
                channelName = currentChannel.name,
                channelDescription = currentChannel.description ?: "",
                myname = currentChannel.myname ?: "",
                myimg = currentChannel.myimg ?: ""
            )
        }
    }

    fun cancelEditing() {
        _uiState.value = _uiState.value.copy(isEditing = false)
    }

    fun clearMessages() {
        _uiState.value = _uiState.value.copy(
            errorMessage = null,
            successMessage = null
        )
    }
}

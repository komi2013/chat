package com.chat.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.data.repository.UserRepository
import com.chat.android.network.ChannelDetail
import com.chat.android.network.NicknameResponse
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class DrawerUiState(
    val isSignedIn: Boolean = false,
    val hasChannel: Boolean = false,
    val isGuest: Boolean = false,
    val myname: String? = null
)

@HiltViewModel
class DrawerViewModel @Inject constructor(
    private val userRepository: UserRepository
) : ViewModel() {

    private val _uiState = MutableStateFlow(DrawerUiState())
    val uiState: StateFlow<DrawerUiState> = _uiState.asStateFlow()

    init {
        checkAuthStatus()
    }

    private fun checkAuthStatus() {
        val isSignedIn = !userRepository.getCsrfToken().isNullOrEmpty()
        val myname = userRepository.getNickname()
        
        _uiState.value = _uiState.value.copy(
            isSignedIn = isSignedIn,
            myname = myname
        )
    }

    fun updateDrawerState(
        channel: ChannelDetail? = null,
        aliases: List<NicknameResponse> = emptyList()
    ) {
        viewModelScope.launch {
            val isSignedIn = !userRepository.getCsrfToken().isNullOrEmpty()
            val myname = userRepository.getNickname()
            val hasChannel = channel != null
            
            val isGuest = if (channel != null && aliases.isNotEmpty()) {
                aliases.any { alias ->
                    alias.accessRight == "guest" && alias.aliasName == myname
                }
            } else {
                // Check from stored data if no channel provided
                false // Simplified for now
            }
            
            _uiState.value = DrawerUiState(
                isSignedIn = isSignedIn,
                hasChannel = hasChannel,
                isGuest = isGuest,
                myname = myname
            )
        }
    }

    fun refreshAuthStatus() {
        checkAuthStatus()
    }
}

package com.chat.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.data.model.TopLink
import com.chat.android.data.repository.UserRepository
import com.chat.android.network.ApiService
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class TopUiState(
    val isLoading: Boolean = false,
    val isSignedIn: Boolean = false,
    val topLinks: List<TopLink> = emptyList(),
    val errorMessage: String? = null
)

@HiltViewModel
class TopViewModel @Inject constructor(
    private val userRepository: UserRepository,
    private val apiService: ApiService
) : ViewModel() {

    private val _uiState = MutableStateFlow(TopUiState())
    val uiState: StateFlow<TopUiState> = _uiState.asStateFlow()

    init {
        checkSignInStatus()
        loadTopLinks()
    }

    private fun checkSignInStatus() {
        val csrf = userRepository.getCsrfToken()
        _uiState.value = _uiState.value.copy(isSignedIn = !csrf.isNullOrEmpty())
    }

    private fun loadTopLinks() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            try {
                val csrf = userRepository.getCsrfToken()
                val channelId = userRepository.getCurrentChannelId()
                
                // Check for user-edited top links
                val storedLinks = if (channelId != null) {
                    userRepository.getStoredTopLinks(channelId)
                } else null
                
                val links = if (storedLinks != null && storedLinks.isNotEmpty()) {
                    storedLinks
                } else {
                    val isSignedIn = !csrf.isNullOrEmpty()
                    getDefaultLinks(isSignedIn)
                }
                
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    topLinks = links
                )
                
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = "Failed to load top links: ${e.message}"
                )
            }
        }
    }

    private fun getDefaultLinks(isSignedIn: Boolean): List<TopLink> {
        return if (isSignedIn) {
            listOf(
                TopLink("組織・チャネル設定", "/channel/"),
                TopLink("ユーザー設定", "/user/"),
                TopLink("ツイート一覧", "/tweets/"),
                TopLink("広告設定", "/adSetting/"),
                TopLink("設定", "/setting/")
            )
        } else {
            listOf(
                TopLink("サインイン", "/sign/"),
                TopLink("ツイート一覧", "/tweets/"),
                TopLink("設定", "/setting/"),
                TopLink("規則", "/html/rule/"),
                TopLink("個人情報遵守", "/html/privacy/")
            )
        }
    }

    fun refreshTopLinks() {
        loadTopLinks()
    }

    fun clearError() {
        _uiState.value = _uiState.value.copy(errorMessage = null)
    }
}

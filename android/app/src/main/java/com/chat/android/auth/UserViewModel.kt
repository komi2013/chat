package com.chat.android.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.data.SessionManager
import com.chat.android.network.ApiService
import com.chat.android.network.GoogleSignInResponse
import com.chat.android.network.NicknameResponse
import com.chat.android.network.UserResponse
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class UserViewModel @Inject constructor(
    private val apiService: ApiService,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _uiState = MutableStateFlow<UserUiState>(UserUiState.Loading)
    val uiState = _uiState.asStateFlow()

    init {
        fetchUser()
    }

    fun fetchUser() {
        val csrf = sessionManager.getCsrf() ?: return
        
        viewModelScope.launch {
            try {
                _uiState.value = UserUiState.Loading
                val response = apiService.getUser(csrf)
                if (response.isSuccessful) {
                    val body = response.body()
                    if (body != null && (body.success == true || body.user != null)) {
                        sessionManager.saveSession(body.csrf, sessionManager.getUserId(), body.nickname)
                        _uiState.value = UserUiState.Success(body)
                    } else {
                        _uiState.value = UserUiState.Error(body?.message ?: "Failed to fetch user data")
                    }
                } else {
                    _uiState.value = UserUiState.Error("Server error: ${response.code()}")
                }
            } catch (e: Exception) {
                _uiState.value = UserUiState.Error(e.message ?: "An unexpected error occurred")
            }
        }
    }

    fun updateUser(
        nickname: String,
        nickImg: String?,
        nickBio: String?,
        mail: String?,
        telephone: String?,
        walletAddress: String?,
        latitude: Double?,
        longitude: Double?
    ) {
        val csrf = sessionManager.getCsrf() ?: return
        viewModelScope.launch {
            try {
                _uiState.value = UserUiState.Loading
                val response = apiService.editUser(
                    csrf, nickname, nickImg, nickBio, mail, telephone, walletAddress, latitude, longitude
                )
                if (response.isSuccessful) {
                    val body = response.body()
                    if (body != null && (body.success == true || body.user != null)) {
                        sessionManager.saveSession(body.csrf, sessionManager.getUserId(), body.nickname)
                        _uiState.value = UserUiState.Success(body)
                    } else {
                        _uiState.value = UserUiState.Error(body?.message ?: "Failed to update user data")
                    }
                } else {
                    _uiState.value = UserUiState.Error("Server error: ${response.code()}")
                }
            } catch (e: Exception) {
                _uiState.value = UserUiState.Error(e.message ?: "An unexpected error occurred")
            }
        }
    }
}

sealed class UserUiState {
    object Loading : UserUiState()
    data class Success(val data: GoogleSignInResponse) : UserUiState()
    data class Error(val message: String) : UserUiState()
}

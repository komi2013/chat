package com.chat.android.auth

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.network.GoogleSignInResponse
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class LoginViewModel @Inject constructor(
    private val googleSignInManager: GoogleSignInManager
) : ViewModel() {

    private val _uiState = MutableStateFlow<LoginUiState>(LoginUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun getGoogleSignInClient() = googleSignInManager.getClient()

    fun handleGoogleSignInResult(idToken: String?, errorMessage: String? = null) {
        if (idToken == null) {
            _uiState.value = LoginUiState.Error(errorMessage ?: "Google Sign-In failed: No ID Token")
            return
        }

        viewModelScope.launch {
            _uiState.value = LoginUiState.Loading
            // The Go backend expects a g_csrf_token cookie that matches the g_csrf_token form field
            val csrfToken = java.util.UUID.randomUUID().toString()
            
            val result = googleSignInManager.signInWithGoogle(idToken, csrfToken)
            
            result.onSuccess { response ->
                _uiState.value = LoginUiState.Success(response)
            }.onFailure { error ->
                _uiState.value = LoginUiState.Error(error.message ?: "Authentication failed")
            }
        }
    }
}

sealed class LoginUiState {
    object Idle : LoginUiState()
    object Loading : LoginUiState()
    data class Success(val response: GoogleSignInResponse) : LoginUiState()
    data class Error(val message: String) : LoginUiState()
}

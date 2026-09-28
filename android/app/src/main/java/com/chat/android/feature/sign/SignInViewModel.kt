package com.chat.android.feature.sign

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.network.GoogleSignInResponse
import com.chat.android.data.SessionManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class SignInViewModel @Inject constructor(
    private val googleSignInManager: GoogleSignInManager,
    private val sessionManager: SessionManager
) : ViewModel() {

    private val _uiState = MutableStateFlow<SignInUiState>(SignInUiState.Idle)
    val uiState = _uiState.asStateFlow()

    fun getGoogleSignInClient() = googleSignInManager.getClient()

    fun handleGoogleSignInResult(idToken: String?, errorMessage: String? = null) {
        if (idToken == null) {
            _uiState.value = SignInUiState.Error(errorMessage ?: "Google Sign-In failed: No ID Token")
            return
        }

        viewModelScope.launch {
            try {
                _uiState.value = SignInUiState.Loading
                android.util.Log.d("SignInViewModel", "Starting Google Sign-In with backend")
                
                val csrfToken = java.util.UUID.randomUUID().toString()
                
                val result = googleSignInManager.signInWithGoogle(idToken, csrfToken)
                
                result.onSuccess { response ->
                    android.util.Log.d("SignInViewModel", "Sign-in success: ${response.userId}")
                    val sessionId = response.sessionId?.takeIf(String::isNotBlank)
                    if (sessionId == null) {
                        _uiState.value = SignInUiState.Error("サーバーからセッション情報を取得できませんでした")
                    } else {
                        sessionManager.saveSession(
                            csrf = response.csrf,
                            userId = response.userId,
                            nickname = response.nickname,
                            sessionId = sessionId
                        )
                        _uiState.value = SignInUiState.Success(response)
                    }
                }.onFailure { error ->
                    android.util.Log.e("SignInViewModel", "Sign-in API failure", error)
                    _uiState.value = SignInUiState.Error(error.message ?: "Authentication failed")
                }
            } catch (e: Exception) {
                android.util.Log.e("SignInViewModel", "Crash prevented in handleGoogleSignInResult", e)
                _uiState.value = SignInUiState.Error("An unexpected error occurred")
            }
        }
    }
}

sealed class SignInUiState {
    object Idle : SignInUiState()
    object Loading : SignInUiState()
    data class Success(val response: GoogleSignInResponse) : SignInUiState()
    data class Error(val message: String) : SignInUiState()
}

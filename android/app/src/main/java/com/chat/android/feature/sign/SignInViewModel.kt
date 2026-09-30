package com.chat.android.feature.sign

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.network.GoogleSignInResponse
import com.chat.android.core.network.SessionManager
import com.chat.android.firebase.PushNotificationManager
import com.google.firebase.messaging.FirebaseMessaging
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.suspendCancellableCoroutine
import javax.inject.Inject
import kotlin.coroutines.resume

@HiltViewModel
class SignInViewModel @Inject constructor(
    private val googleSignInManager: GoogleSignInManager,
    private val sessionManager: SessionManager,
    private val pushNotificationManager: PushNotificationManager
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
                    if (response.sessionId == null) {
                        _uiState.value = SignInUiState.Error("サーバーからセッション情報を取得できませんでした")
                    } else {
                        sessionManager.saveSession(
                            csrf = response.csrf,
                            userId = response.userId,
                            nickname = response.nickname,
                            sessionId = response.sessionId
                        )
                        registerPushToken(response.csrf)
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

    /**
     * Registers the FCM token with the backend after sign-in/up so the server can push to
     * this device (POST /PushSubscribeMobile/ -> session.pushToken + deviceType 2).
     * Notification registration must never break the sign-in flow.
     */
    private suspend fun registerPushToken(csrf: String?) {
        val token = fetchFcmToken() ?: return
        val effectiveCsrf = csrf?.takeIf { it.isNotBlank() } ?: sessionManager.getCsrf() ?: return
        runCatching { pushNotificationManager.sendTokenToBackend(token, effectiveCsrf) }
            .onFailure { android.util.Log.w("SignInViewModel", "FCM token registration failed", it) }
    }

    /** Firebase returns the token through a callback, so bridge it into a suspend call. */
    private suspend fun fetchFcmToken(): String? = suspendCancellableCoroutine { continuation ->
        FirebaseMessaging.getInstance().token.addOnCompleteListener { task ->
            if (task.isSuccessful) {
                continuation.resume(task.result)
            } else {
                android.util.Log.w("SignInViewModel", "Fetching FCM registration token failed", task.exception)
                continuation.resume(null)
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

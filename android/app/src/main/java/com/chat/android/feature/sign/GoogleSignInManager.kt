package com.chat.android.feature.sign

import android.content.Context
import com.google.android.gms.auth.api.signin.GoogleSignIn
import com.google.android.gms.auth.api.signin.GoogleSignInClient
import com.google.android.gms.auth.api.signin.GoogleSignInOptions
import com.chat.android.BuildConfig
import com.chat.android.core.network.ApiService
import com.chat.android.core.data.SessionManager
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class GoogleSignInManager @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ApiService,
    private val sessionManager: SessionManager
) {
    
    private val googleSignInClient: GoogleSignInClient by lazy {
        try {
            val clientId = BuildConfig.GOOGLE_WEB_CLIENT_ID
            android.util.Log.d("GoogleSignInManager", "Initializing with Client ID: $clientId")
            
            if (clientId.isEmpty()) {
                android.util.Log.e("GoogleSignInManager", "GOOGLE_WEB_CLIENT_ID is empty!")
            }

            val gso = GoogleSignInOptions.Builder(GoogleSignInOptions.DEFAULT_SIGN_IN)
                .requestIdToken(clientId)
                .requestEmail()
                .build()
            
            GoogleSignIn.getClient(context, gso)
        } catch (e: Exception) {
            android.util.Log.e("GoogleSignInManager", "Failed to initialize GoogleSignInClient", e)
            throw e
        }
    }
    
    fun getClient(): GoogleSignInClient = googleSignInClient
    
    suspend fun signInWithGoogle(idToken: String, csrf: String) = try {
        android.util.Log.d("GoogleSignInManager", "Attempting sign-in with backend...")
        android.util.Log.d("GoogleSignInManager", "CSRF: $csrf")
        android.util.Log.d("GoogleSignInManager", "ID Token prefix: ${idToken.take(20)}...")

        val response = apiService.signInWithGoogle(idToken, csrf, sessionManager.getDeviceId())
        if (response.isSuccessful) {
            val responseBody = response.body()?.string()
            android.util.Log.d("GoogleSignInManager", "Response Body: $responseBody")
            
            if (responseBody.isNullOrBlank()) {
                Result.failure(Exception("Empty response body"))
            } else {
                try {
                    val gson = com.google.gson.Gson()
                    // Try parsing as the expected object first
                    val signInResponse = gson.fromJson(responseBody, com.chat.android.core.network.GoogleSignInResponse::class.java)
                    Result.success(signInResponse)
                } catch (e: Exception) {
                    android.util.Log.w("GoogleSignInManager", "Standard JSON parsing failed, attempting fallback: $responseBody")
                    
                    // Fallback: If it's a quoted string containing JSON, or just a plain string
                    try {
                        // Check if it's just a literal string (like "success" or a token)
                        if (!responseBody.trim().startsWith("{")) {
                             // It's likely a plain string or a quoted string. 
                             // We'll create a dummy response using this as the message/token
                             Result.success(com.chat.android.core.network.GoogleSignInResponse(
                                 csrf = csrf, // Use the one we sent
                                 success = true,
                                 message = responseBody,
                                 userId = "imported_user"
                             ))
                        } else {
                            Result.failure(Exception("Invalid JSON format: $responseBody"))
                        }
                    } catch (e2: Exception) {
                        Result.failure(Exception("Parsing error: $responseBody"))
                    }
                }
            }
        } else {
            val errorBody = response.errorBody()?.string()
            android.util.Log.e("GoogleSignInManager", "Response Error (${response.code()}): $errorBody")
            Result.failure(Exception("Server returned error ${response.code()}: ${errorBody ?: "No details"}"))
        }
    } catch (e: Exception) {
        android.util.Log.e("GoogleSignInManager", "Network or Parsing error", e)
        Result.failure(Exception("Network or Parsing error: ${e.localizedMessage}"))
    }
}

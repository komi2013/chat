package com.chat.android.auth

import android.content.Context
import com.google.android.gms.auth.api.signin.GoogleSignIn
import com.google.android.gms.auth.api.signin.GoogleSignInClient
import com.google.android.gms.auth.api.signin.GoogleSignInOptions
import com.chat.android.BuildConfig
import com.chat.android.network.RetrofitClient
import dagger.hilt.android.qualifiers.ApplicationContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class GoogleSignInManager @Inject constructor(
    @ApplicationContext private val context: Context
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
        // The Go backend needs the g_csrf_token cookie to match the g_csrf_token form field.
        RetrofitClient.setCsrfCookie(BuildConfig.BASE_URL, "g_csrf_token", csrf)

        val response = RetrofitClient.apiService.signInWithGoogle(idToken, csrf)
        if (response.isSuccessful) {
            response.body()?.let { Result.success(it) }
                ?: Result.failure(Exception("Empty response"))
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
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
        val gso = GoogleSignInOptions.Builder(GoogleSignInOptions.DEFAULT_SIGN_IN)
            .requestIdToken(BuildConfig.GOOGLE_WEB_CLIENT_ID)
            .requestEmail()
            .build()
        
        GoogleSignIn.getClient(context, gso)
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
            Result.failure(Exception("Google sign-in failed: ${response.code()}"))
        }
    } catch (e: Exception) {
        Result.failure(e)
    }
}
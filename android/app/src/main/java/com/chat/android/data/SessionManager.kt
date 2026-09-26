package com.chat.android.data

import android.content.Context
import android.content.SharedPreferences
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import dagger.hilt.android.qualifiers.ApplicationContext
import java.nio.ByteBuffer
import java.nio.charset.StandardCharsets
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class SessionManager @Inject constructor(
    @ApplicationContext context: Context
) {
    private val prefs: SharedPreferences = context.getSharedPreferences("chat_prefs", Context.MODE_PRIVATE)

    companion object {
        private const val KEY_CSRF = "csrf"
        private const val KEY_USER_ID = "user_id"
        private const val KEY_NICKNAME = "nickname"
        private const val KEY_SESSION_ID = "session_id"
        private const val SESSION_KEY_ALIAS = "chat_session_id"
        private const val ENCRYPTED_PREFIX = "enc:"
        private const val GCM_IV_SIZE = 12
    }

    fun saveSession(csrf: String, userId: String?, nickname: String?, sessionId: String? = null) {
        val encryptedSessionId = sessionId?.let { ENCRYPTED_PREFIX + encryptSessionId(it) }
        prefs.edit().apply {
            putString(KEY_CSRF, csrf)
            putString(KEY_USER_ID, userId)
            putString(KEY_NICKNAME, nickname)
            if (encryptedSessionId != null) {
                putString(KEY_SESSION_ID, encryptedSessionId)
            }
            apply()
        }
    }

    fun getSessionId(): String? {
        val storedValue = prefs.getString(KEY_SESSION_ID, null) ?: return null
        if (!storedValue.startsWith(ENCRYPTED_PREFIX)) {
            setSessionId(storedValue)
            return storedValue
        }

        return try {
            decryptSessionId(storedValue.removePrefix(ENCRYPTED_PREFIX))
        } catch (_: Exception) {
            prefs.edit().remove(KEY_SESSION_ID).apply()
            null
        }
    }

    fun setSessionId(sessionId: String) {
        val encryptedValue = ENCRYPTED_PREFIX + encryptSessionId(sessionId)
        prefs.edit().putString(KEY_SESSION_ID, encryptedValue).apply()
    }

    fun getCsrf(): String? = prefs.getString(KEY_CSRF, null)

    fun setCsrf(csrf: String) {
        prefs.edit().putString(KEY_CSRF, csrf).apply()
    }
    
    fun getUserId(): String? = prefs.getString(KEY_USER_ID, null)

    fun getNickname(): String? = prefs.getString(KEY_NICKNAME, null)

    fun setNickname(nickname: String) {
        prefs.edit().putString(KEY_NICKNAME, nickname).apply()
    }

    fun clearSession() {
        prefs.edit().clear().apply()
    }

    private fun encryptSessionId(sessionId: String): String {
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.ENCRYPT_MODE, getOrCreateSessionKey())
        val encrypted = cipher.doFinal(sessionId.toByteArray(StandardCharsets.UTF_8))
        val payload = ByteBuffer.allocate(cipher.iv.size + encrypted.size)
            .put(cipher.iv)
            .put(encrypted)
            .array()
        return Base64.encodeToString(payload, Base64.NO_WRAP)
    }

    private fun decryptSessionId(value: String): String {
        val payload = Base64.decode(value, Base64.NO_WRAP)
        require(payload.size > GCM_IV_SIZE) { "Invalid encrypted session ID" }
        val iv = payload.copyOfRange(0, GCM_IV_SIZE)
        val encrypted = payload.copyOfRange(GCM_IV_SIZE, payload.size)
        val cipher = Cipher.getInstance("AES/GCM/NoPadding")
        cipher.init(Cipher.DECRYPT_MODE, getOrCreateSessionKey(), GCMParameterSpec(128, iv))
        return String(cipher.doFinal(encrypted), StandardCharsets.UTF_8)
    }

    private fun getOrCreateSessionKey(): SecretKey {
        val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }
        (keyStore.getKey(SESSION_KEY_ALIAS, null) as? SecretKey)?.let { return it }

        val keyGenerator = KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore")
        keyGenerator.init(
            KeyGenParameterSpec.Builder(
                SESSION_KEY_ALIAS,
                KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
            )
                .setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                .setRandomizedEncryptionRequired(true)
                .build()
        )
        return keyGenerator.generateKey()
    }
}

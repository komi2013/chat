package com.chat.android.core.network

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import android.util.Log
import androidx.datastore.preferences.core.edit
import androidx.datastore.preferences.core.stringPreferencesKey
import androidx.datastore.preferences.preferencesDataStore
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.first
import kotlinx.coroutines.launch
import kotlinx.coroutines.runBlocking
import java.nio.ByteBuffer
import java.nio.charset.StandardCharsets
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec
import javax.inject.Inject
import javax.inject.Singleton

private val Context.sessionDataStore by preferencesDataStore(name = "session_prefs")

@Singleton
class SessionManager @Inject constructor(
    @ApplicationContext val context: Context
) {
    private val dataStore = context.sessionDataStore
    private val scope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    companion object {
        private val KEY_CSRF = stringPreferencesKey("csrf")
        private val KEY_USER_ID = stringPreferencesKey("user_id")
        private val KEY_NICKNAME = stringPreferencesKey("nickname")
        private val KEY_SESSION_ID = stringPreferencesKey("session_id")
        private val KEY_DEVICE_ID = stringPreferencesKey("device_id")
        private val KEY_TO = stringPreferencesKey("to_token")
        private val KEY_LAST_CHANNEL_ID = stringPreferencesKey("last_channel_id")
        
        // Channel Profile keys
        private fun nameKey(channelID: String) = stringPreferencesKey("${channelID}_name")
        private fun imageKey(channelID: String) = stringPreferencesKey("${channelID}_image")
        private fun topLinksKey(channelID: String) = stringPreferencesKey("topLinks_$channelID")
        private fun lastChannelKey(userId: String) = stringPreferencesKey("${userId}_last_channel_id")
        private fun lastNameKey(userId: String) = stringPreferencesKey("${userId}_last_name")
        private fun lastImageKey(userId: String) = stringPreferencesKey("${userId}_last_image")
        
        private const val SESSION_KEY_ALIAS = "chat_session_id"
        private const val ENCRYPTED_PREFIX = "enc:"
        private const val GCM_IV_SIZE = 12
    }

    @Volatile private var cachedCsrf: String? = null
    @Volatile private var cachedUserId: String? = null
    @Volatile private var cachedNickname: String? = null
    @Volatile private var cachedSessionId: String? = null
    @Volatile private var cachedDeviceId: String? = null
    @Volatile private var cachedTo: String? = null
    @Volatile private var cachedLastChannelId: String? = null

    init {
        // We use runBlocking here because SessionManager is a Singleton initialized once at app start,
        // and we need these values available synchronously for Interceptors and initial UI state.
        runBlocking {
            val prefs = dataStore.data.first()
            cachedCsrf = prefs[KEY_CSRF]
            cachedUserId = prefs[KEY_USER_ID]
            cachedNickname = prefs[KEY_NICKNAME]
            cachedSessionId = prefs[KEY_SESSION_ID]?.let { decryptSafe(it) }
            cachedDeviceId = prefs[KEY_DEVICE_ID]
            cachedTo = prefs[KEY_TO]
            cachedLastChannelId = prefs[KEY_LAST_CHANNEL_ID]

            // Migration from legacy SharedPreferences if DataStore is empty
            if (cachedCsrf == null && cachedUserId == null && cachedSessionId == null) {
                migrateFromPrefs()
            }
            
            // Generate Device ID if not exists
            if (cachedDeviceId == null) {
                cachedDeviceId = java.util.UUID.randomUUID().toString()
                scope.launch { dataStore.edit { it[KEY_DEVICE_ID] = cachedDeviceId!! } }
            }
        }
    }

    private suspend fun migrateFromPrefs() {
        val chatPrefs = context.getSharedPreferences("chat_prefs", Context.MODE_PRIVATE)
        val userPrefs = context.getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
        
        val csrf = chatPrefs.getString("csrf", null)
        val userId = chatPrefs.getString("user_id", null)
        val nickname = chatPrefs.getString("nickname", null)
        val sessionId = chatPrefs.getString("session_id", null)
        val deviceId = chatPrefs.getString("device_id", null)
        val toToken = userPrefs.getString("TO", null)
        val channelId = userPrefs.getString("channelID", null)

        if (csrf != null || userId != null || sessionId != null) {
            cachedCsrf = csrf
            cachedUserId = userId
            cachedNickname = nickname
            cachedSessionId = sessionId?.let { decryptSafe(it) }
            cachedDeviceId = deviceId
            cachedTo = toToken
            cachedLastChannelId = channelId
            
            dataStore.edit { prefs ->
                if (csrf != null) prefs[KEY_CSRF] = csrf
                if (userId != null) prefs[KEY_USER_ID] = userId
                if (nickname != null) prefs[KEY_NICKNAME] = nickname
                if (sessionId != null) prefs[KEY_SESSION_ID] = sessionId
                if (deviceId != null) prefs[KEY_DEVICE_ID] = deviceId
                if (toToken != null) prefs[KEY_TO] = toToken
                if (channelId != null) prefs[KEY_LAST_CHANNEL_ID] = channelId
            }
        }
    }

    private fun decryptSafe(storedValue: String): String? {
        if (!storedValue.startsWith(ENCRYPTED_PREFIX)) return storedValue
        return try {
            decryptSessionId(storedValue.removePrefix(ENCRYPTED_PREFIX))
        } catch (_: Exception) {
            null
        }
    }

    fun saveSession(csrf: String, userId: String?, nickname: String?, sessionId: String? = null) {
        cachedCsrf = csrf
        cachedUserId = userId
        cachedNickname = nickname
        
        val storedSessionId = sessionId?.let {
            try {
                ENCRYPTED_PREFIX + encryptSessionId(it)
            } catch (e: Exception) {
                Log.e("SessionManager", "Failed to encrypt session ID, storing as plain text", e)
                it
            }
        }

        if (sessionId != null) {
            cachedSessionId = sessionId
        }

        scope.launch {
            dataStore.edit { prefs ->
                prefs[KEY_CSRF] = csrf
                if (userId != null) prefs[KEY_USER_ID] = userId
                if (nickname != null) prefs[KEY_NICKNAME] = nickname
                if (storedSessionId != null) {
                    prefs[KEY_SESSION_ID] = storedSessionId
                }
            }
        }
    }

    fun getSessionId(): String? = cachedSessionId

    fun setSessionId(sessionId: String) {
        cachedSessionId = sessionId
        scope.launch {
            val storedValue = try {
                ENCRYPTED_PREFIX + encryptSessionId(sessionId)
            } catch (e: Exception) {
                Log.e("SessionManager", "Failed to encrypt session ID, storing as plain text", e)
                sessionId
            }
            dataStore.edit { it[KEY_SESSION_ID] = storedValue }
        }
    }

    fun getCsrf(): String? = cachedCsrf

    fun setCsrf(csrf: String) {
        cachedCsrf = csrf
        scope.launch {
            dataStore.edit { it[KEY_CSRF] = csrf }
        }
    }

    fun applyResponseCsrf(sentCsrf: String?, receivedCsrf: String?): String? {
        val received = receivedCsrf?.takeIf(String::isNotBlank) ?: return null
        if (received == sentCsrf) return null
        setCsrf(received)
        return received
    }

    fun getUserId(): String? = cachedUserId
    fun getNickname(): String? = cachedNickname

    fun setNickname(nickname: String) {
        cachedNickname = nickname
        scope.launch {
            dataStore.edit { it[KEY_NICKNAME] = nickname }
        }
    }

    fun getDeviceId(): String = cachedDeviceId ?: run {
        val id = java.util.UUID.randomUUID().toString()
        cachedDeviceId = id
        scope.launch { dataStore.edit { it[KEY_DEVICE_ID] = id } }
        id
    }

    fun getTO(): String? = cachedTo
    fun setTO(to: String) {
        cachedTo = to
        scope.launch { dataStore.edit { it[KEY_TO] = to } }
    }

    fun getLastChannelId(): String? = cachedLastChannelId
    fun setCurrentChannelId(channelId: String) {
        cachedLastChannelId = channelId
        scope.launch { dataStore.edit { it[KEY_LAST_CHANNEL_ID] = channelId } }
    }

    fun getStoredTopLinks(channelID: String): String? {
        return runBlocking { dataStore.data.first()[topLinksKey(channelID)] }
    }

    fun setStoredTopLinks(channelID: String, json: String) {
        scope.launch { dataStore.edit { it[topLinksKey(channelID)] = json } }
    }

    suspend fun getChannelProfile(channelID: String): Pair<String, String> {
        val prefs = dataStore.data.first()
        return (prefs[nameKey(channelID)] ?: "") to (prefs[imageKey(channelID)] ?: "")
    }

    suspend fun saveChannelProfile(channelID: String, name: String? = null, image: String? = null) {
        dataStore.edit { prefs ->
            if (name != null) prefs[nameKey(channelID)] = name
            if (image != null) prefs[imageKey(channelID)] = image
        }
    }

    suspend fun saveLastProfile(userId: String, channelID: String, name: String, image: String) {
        dataStore.edit { prefs ->
            prefs[lastChannelKey(userId)] = channelID
            prefs[lastNameKey(userId)] = name
            prefs[lastImageKey(userId)] = image
            
            // Also update the per-channel one
            prefs[nameKey(channelID)] = name
            prefs[imageKey(channelID)] = image
        }
    }

    suspend fun getLastProfile(userId: String): Triple<String, String, String> {
        val prefs = dataStore.data.first()
        return Triple(
            prefs[lastChannelKey(userId)].orEmpty(),
            prefs[lastNameKey(userId)].orEmpty(),
            prefs[lastImageKey(userId)].orEmpty()
        )
    }

    fun clearSession() {
        cachedCsrf = null
        cachedUserId = null
        cachedNickname = null
        cachedSessionId = null
        cachedTo = null
        cachedLastChannelId = null
        scope.launch {
            dataStore.edit { prefs ->
                val deviceId = prefs[KEY_DEVICE_ID]
                prefs.clear()
                if (deviceId != null) prefs[KEY_DEVICE_ID] = deviceId
            }
        }
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
        
        try {
            (keyStore.getKey(SESSION_KEY_ALIAS, null) as? SecretKey)?.let { return it }
        } catch (e: Exception) {
            Log.w("SessionManager", "Failed to get session key from Keystore, attempting to recreate", e)
            try {
                keyStore.deleteEntry(SESSION_KEY_ALIAS)
            } catch (_: Exception) {}
        }

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

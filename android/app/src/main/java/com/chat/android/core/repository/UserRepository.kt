package com.chat.android.core.repository

import android.content.ContentValues
import android.content.Context
import android.content.SharedPreferences
import android.database.Cursor
import android.database.sqlite.SQLiteDatabase
import com.chat.android.core.network.ApiService
import com.chat.android.core.network.GoogleSignInResponse
import com.chat.android.core.network.NicknameResponse
import com.chat.android.core.network.SessionManager
import com.chat.android.core.network.UserResponse
import com.chat.android.core.network.model.TopLink
import com.chat.android.core.network.model.UserNicknameEntity
import com.chat.android.core.network.model.UserProfileEntity
import com.chat.android.feature.channel.ChannelDbHelper
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.withContext
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UserRepository @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ApiService,
    private val sessionManager: SessionManager
) {
    // 生SQLite: ChannelFeatureDB.db を user_profile / user_nickname の保存先として共有する。
    private val dbHelper: ChannelDbHelper = ChannelDbHelper(context)
    private val sharedPreferences: SharedPreferences = 
        context.getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
    
    private val gson = Gson()

    private val _userProfile = MutableStateFlow<UserProfileEntity?>(null)
    private val _nicknames = MutableStateFlow<List<UserNicknameEntity>>(emptyList())

    init {
        // Initial load
        refreshLocalState()
    }

    private fun refreshLocalState() {
        _userProfile.value = getCurrentProfileSnapshot()
        _nicknames.value = getNicknamesSnapshot()
    }

    fun observeUserProfile(): Flow<UserProfileEntity?> = _userProfile.asStateFlow()

    fun observeNicknames(): Flow<List<UserNicknameEntity>> = _nicknames.asStateFlow()

    private fun getCurrentProfileSnapshot(): UserProfileEntity? {
        val db = dbHelper.readableDatabase
        val cursor = db.query(
            "user_profile",
            null,
            "profileId = ?",
            arrayOf(UserProfileEntity.CURRENT_PROFILE_ID),
            null, null, null
        )
        return cursor.use {
            if (it.moveToFirst()) it.toUserProfile() else null
        }
    }

    private fun getNicknamesSnapshot(): List<UserNicknameEntity> {
        val db = dbHelper.readableDatabase
        val cursor = db.query("user_nickname", null, null, null, null, null, "nickname ASC")
        val list = mutableListOf<UserNicknameEntity>()
        cursor.use {
            while (it.moveToNext()) {
                list.add(it.toUserNickname())
            }
        }
        return list
    }

    suspend fun refreshUser(): String? = withContext(Dispatchers.IO) {
        val csrf = getCsrfToken()?.takeIf(String::isNotBlank) ?: return@withContext "サインインが必要です"
        val response = apiService.getUser(csrf)
        if (!response.isSuccessful) return@withContext "ユーザー情報を取得できませんでした (${response.code()})"
        val body = response.body() ?: return@withContext "サーバーから空の応答が返されました"
        return@withContext storeUserResponse(body, csrf)
    }

    suspend fun updateUser(
        nickname: String,
        nickImg: String?,
        removeNickImg: Boolean,
        nickBio: String?,
        mail: String,
        telephone: String,
        walletAddress: String,
        latitude: Double,
        longitude: Double
    ): UserEditResult = withContext(Dispatchers.IO) {
        val csrf = getCsrfToken()?.takeIf(String::isNotBlank)
            ?: return@withContext UserEditResult(error = "サインインが必要です")
        val response = apiService.editUser(
            csrf = csrf,
            nickname = nickname,
            nickImg = nickImg,
            removeNickImg = removeNickImg,
            nickBio = nickBio,
            mail = mail,
            telephone = telephone,
            walletAddress = walletAddress,
            latitude = latitude,
            longitude = longitude
        )
        if (!response.isSuccessful) {
            return@withContext UserEditResult(error = "ユーザー情報を更新できませんでした (${response.code()})")
        }
        val body = response.body() ?: return@withContext UserEditResult(error = "サーバーから空の応答が返されました")
        val error = storeCsrfAndReadError(body, csrf)
        if (error != null) return@withContext UserEditResult(error = error)

        val db = dbHelper.writableDatabase
        db.beginTransaction()
        try {
            val current = getCurrentProfileSnapshot() ?: UserProfileEntity()
            val updatedProfile = current.copy(
                mail = mail,
                telephone = telephone,
                walletAddress = walletAddress,
                latitude = latitude,
                longitude = longitude,
                nickname = nickname.ifBlank { current.nickname },
                nickImg = if (removeNickImg) "" else nickImg ?: current.nickImg,
                nickBio = nickBio
            )
            saveProfile(db, updatedProfile)

            if (nickname.isNotBlank()) {
                val existing = getNickname(db, nickname)
                val savedNickImg = if (removeNickImg) "" else nickImg ?: existing?.nickImg
                val updatedNick = existing?.copy(nickImg = savedNickImg, nickBio = nickBio)
                    ?: UserNicknameEntity(nickname = nickname, nickImg = savedNickImg, nickBio = nickBio)
                saveNickname(db, updatedNick)
            }
            db.setTransactionSuccessful()
        } finally {
            db.endTransaction()
        }
        
        refreshLocalState()

        return@withContext UserEditResult(message = body.message ?: "ユーザー情報を更新しました", pushContents = body.pushContents.orEmpty())
    }

    private fun saveProfile(db: SQLiteDatabase, profile: UserProfileEntity) {
        val values = ContentValues().apply {
            put("profileId", profile.profileId)
            put("name", profile.name)
            put("mail", profile.mail)
            put("nickname", profile.nickname)
            put("channelID", profile.channelID)
            put("accessRight", profile.accessRight)
            put("admin", if (profile.admin == true) 1 else 0)
            put("telephone", profile.telephone)
            put("walletAddress", profile.walletAddress)
            put("latitude", profile.latitude)
            put("longitude", profile.longitude)
            put("nickImg", profile.nickImg)
            put("nickBio", profile.nickBio)
        }
        db.insertWithOnConflict("user_profile", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    private fun getNickname(db: SQLiteDatabase, nickname: String): UserNicknameEntity? {
        val cursor = db.query("user_nickname", null, "nickname = ?", arrayOf(nickname), null, null, null)
        return cursor.use {
            if (it.moveToFirst()) it.toUserNickname() else null
        }
    }

    private fun saveNickname(db: SQLiteDatabase, nickname: UserNicknameEntity) {
        val values = ContentValues().apply {
            put("nickname", nickname.nickname)
            put("nickImg", nickname.nickImg)
            put("nickBio", nickname.nickBio)
            put("good", nickname.good)
            put("bad", nickname.bad)
            put("createdAt", nickname.createdAt)
            put("accessRight", nickname.accessRight)
        }
        db.insertWithOnConflict("user_nickname", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    private fun storeUserResponse(body: GoogleSignInResponse, sentCsrf: String): String? {
        val error = storeCsrfAndReadError(body, sentCsrf)
        if (error != null) return error
        val user = body.user ?: return "ユーザー情報が応答に含まれていません"
        
        val db = dbHelper.writableDatabase
        db.beginTransaction()
        try {
            saveProfile(db, user.toEntity())
            
            db.delete("user_nickname", null, null)
            body.nicknames.orEmpty().forEach { 
                saveNickname(db, it.toEntity())
            }
            db.setTransactionSuccessful()
        } finally {
            db.endTransaction()
        }
        
        refreshLocalState()
        return null
    }

    private fun storeCsrfAndReadError(body: GoogleSignInResponse, sentCsrf: String): String? {
        sessionManager.applyResponseCsrf(sentCsrf, body.csrf)?.let(::setCsrfToken)
        if (body.csrf.isNullOrBlank() && body.error == null) {
            return "CSRFトークンの更新に失敗しました"
        }
        return body.error
    }

    private fun UserResponse.toEntity() = UserProfileEntity(
        name = name,
        mail = mail,
        nickname = nickname,
        channelID = channelID,
        accessRight = accessRight,
        admin = admin,
        telephone = telephone,
        walletAddress = walletAddress,
        latitude = latitude,
        longitude = longitude,
        nickImg = nickImg,
        nickBio = nickBio
    )

    private fun NicknameResponse.toEntity() = UserNicknameEntity(
        nickname = nickname,
        nickImg = nickImg,
        nickBio = nickBio,
        good = good,
        bad = bad,
        createdAt = createdAt,
        accessRight = accessRight
    )

    private fun Cursor.toUserProfile() = UserProfileEntity(
        profileId = getString(getColumnIndexOrThrow("profileId")),
        name = getString(getColumnIndexOrThrow("name")),
        mail = getString(getColumnIndexOrThrow("mail")),
        nickname = getString(getColumnIndexOrThrow("nickname")),
        channelID = getString(getColumnIndexOrThrow("channelID")),
        accessRight = getString(getColumnIndexOrThrow("accessRight")),
        admin = getInt(getColumnIndexOrThrow("admin")) == 1,
        telephone = getString(getColumnIndexOrThrow("telephone")),
        walletAddress = getString(getColumnIndexOrThrow("walletAddress")),
        latitude = getDouble(getColumnIndexOrThrow("latitude")),
        longitude = getDouble(getColumnIndexOrThrow("longitude")),
        nickImg = getString(getColumnIndexOrThrow("nickImg")),
        nickBio = getString(getColumnIndexOrThrow("nickBio"))
    )

    private fun Cursor.toUserNickname() = UserNicknameEntity(
        nickname = getString(getColumnIndexOrThrow("nickname")),
        nickImg = getString(getColumnIndexOrThrow("nickImg")),
        nickBio = getString(getColumnIndexOrThrow("nickBio")),
        good = getInt(getColumnIndexOrThrow("good")),
        bad = getInt(getColumnIndexOrThrow("bad")),
        createdAt = getString(getColumnIndexOrThrow("createdAt")),
        accessRight = getString(getColumnIndexOrThrow("accessRight"))
    )

    companion object {
        private const val CSRF_TOKEN_KEY = "csrf"
        private const val NICKNAME_KEY = "nickname"
        private const val CHANNEL_ID_KEY = "channelID"
        private const val TO_KEY = "TO"
        private const val TOP_LINKS_PREFIX = "topLinks"
    }

    fun getCsrfToken(): String? {
        return sessionManager.getCsrf() ?: sharedPreferences.getString(CSRF_TOKEN_KEY, null)
    }

    fun setCsrfToken(token: String) {
        sessionManager.setCsrf(token)
        sharedPreferences.edit()
            .putString(CSRF_TOKEN_KEY, token)
            .apply()
    }

    fun getNickname(): String? {
        return sessionManager.getNickname() ?: sharedPreferences.getString(NICKNAME_KEY, null)
    }

    fun setNickname(nickname: String) {
        sessionManager.setNickname(nickname)
        sharedPreferences.edit()
            .putString(NICKNAME_KEY, nickname)
            .apply()
    }

    fun getCurrentChannelId(): String? {
        return sharedPreferences.getString(CHANNEL_ID_KEY, null)
    }

    fun setCurrentChannelId(channelId: String) {
        sharedPreferences.edit()
            .putString(CHANNEL_ID_KEY, channelId)
            .apply()
    }

    fun getTO(): String? {
        return sharedPreferences.getString(TO_KEY, null)
    }

    fun setTO(to: String) {
        sharedPreferences.edit()
            .putString(TO_KEY, to)
            .apply()
    }

    fun clearTO() {
        sharedPreferences.edit()
            .remove(TO_KEY)
            .apply()
    }

    fun getStoredTopLinks(channelId: String): List<TopLink>? {
        val key = "$TOP_LINKS_PREFIX$channelId"
        val json = sharedPreferences.getString(key, null)
        return if (json != null) {
            val type = object : TypeToken<List<TopLink>>() {}.type
            gson.fromJson(json, type)
        } else null
    }

    fun setStoredTopLinks(channelId: String, topLinks: List<TopLink>) {
        val key = "$TOP_LINKS_PREFIX$channelId"
        val json = gson.toJson(topLinks)
        sharedPreferences.edit()
            .putString(key, json)
            .apply()
    }

    fun clearUserData() {
        sharedPreferences.edit()
            .remove(CSRF_TOKEN_KEY)
            .remove(NICKNAME_KEY)
            .remove(CHANNEL_ID_KEY)
            .remove(TO_KEY)
            .apply()
    }

    fun isSignedIn(): Boolean {
        return !getCsrfToken().isNullOrEmpty()
    }
}

data class UserEditResult(
    val error: String? = null,
    val message: String? = null,
    val pushContents: List<String> = emptyList()
)

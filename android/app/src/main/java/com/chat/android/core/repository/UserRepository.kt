package com.chat.android.core.repository

import android.content.Context
import android.content.SharedPreferences
import androidx.room.withTransaction
import com.chat.android.core.database.ChatDatabase
import com.chat.android.core.database.entities.UserNicknameEntity
import com.chat.android.core.database.entities.UserProfileEntity
import com.chat.android.core.data.SessionManager
import com.chat.android.core.data.model.TopLink
import com.chat.android.core.network.ApiService
import com.chat.android.core.network.GoogleSignInResponse
import com.chat.android.core.network.NicknameResponse
import com.chat.android.core.network.UserResponse
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.Flow
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class UserRepository @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ApiService,
    private val sessionManager: SessionManager
) {
    private val sharedPreferences: SharedPreferences = 
        context.getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
    
    private val gson = Gson()
    private val database = ChatDatabase.getDatabase(context)

    fun observeUserProfile(): Flow<UserProfileEntity?> = database.userProfileDao().observeCurrentProfile()

    fun observeNicknames(): Flow<List<UserNicknameEntity>> = database.userNicknameDao().observeNicknames()

    suspend fun refreshUser(): String? {
        val csrf = getCsrfToken()?.takeIf(String::isNotBlank) ?: return "サインインが必要です"
        val response = apiService.getUser(csrf)
        if (!response.isSuccessful) return "ユーザー情報を取得できませんでした (${response.code()})"
        val body = response.body() ?: return "サーバーから空の応答が返されました"
        return storeUserResponse(body, csrf)
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
    ): UserEditResult {
        val csrf = getCsrfToken()?.takeIf(String::isNotBlank)
            ?: return UserEditResult(error = "サインインが必要です")
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
            return UserEditResult(error = "ユーザー情報を更新できませんでした (${response.code()})")
        }
        val body = response.body() ?: return UserEditResult(error = "サーバーから空の応答が返されました")
        val error = storeCsrfAndReadError(body, csrf)
        if (error != null) return UserEditResult(error = error)

        database.withTransaction {
            val profileDao = database.userProfileDao()
            val profile = profileDao.getCurrentProfileSnapshot() ?: UserProfileEntity()
            profileDao.save(
                profile.copy(
                    mail = mail,
                    telephone = telephone,
                    walletAddress = walletAddress,
                    latitude = latitude,
                    longitude = longitude,
                    nickname = nickname.ifBlank { profile.nickname },
                    nickImg = if (removeNickImg) "" else nickImg ?: profile.nickImg,
                    nickBio = nickBio
                )
            )

            if (nickname.isNotBlank()) {
                val nicknameDao = database.userNicknameDao()
                val existing = nicknameDao.getByNickname(nickname)
                val savedNickImg = if (removeNickImg) "" else nickImg ?: existing?.nickImg
                nicknameDao.insertAll(
                    listOf(
                        existing?.copy(nickImg = savedNickImg, nickBio = nickBio)
                            ?: UserNicknameEntity(nickname = nickname, nickImg = savedNickImg, nickBio = nickBio)
                    )
                )
            }
        }

        return UserEditResult(message = body.message ?: "ユーザー情報を更新しました", pushContents = body.pushContents.orEmpty())
    }

    private suspend fun storeUserResponse(body: GoogleSignInResponse, sentCsrf: String): String? {
        val error = storeCsrfAndReadError(body, sentCsrf)
        if (error != null) return error
        val user = body.user ?: return "ユーザー情報が応答に含まれていません"
        database.withTransaction {
            database.userProfileDao().save(user.toEntity())
            database.userNicknameDao().replaceAll(body.nicknames.orEmpty().map { it.toEntity() })
        }
        return null
    }

    /**
     * 応答に含まれる CSRF を保存し、エラーメッセージを返す。
     *
     * UserGet / UserEdit はセッション確認に成功すると回転後の CSRF を返すので必ず保存する。
     * 送信値と同じ値はサーバーがエコーしただけ（セッション確認に失敗したときの
     * common/response.go の WriteResponseWithoutSession）なので保存しない。
     */
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

package com.chat.android.feature.reception

import android.content.Context
import com.chat.android.core.network.SessionManager
import com.chat.android.feature.channel.ChannelDbHelper
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
import com.google.gson.Gson
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ReceptionRepository @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ReceptionApiService,
    private val sessionManager: SessionManager,
    private val dbHelper: ChannelDbHelper
) {
    private val preferences = context.getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
    private val gson = Gson()

    fun getChannelId(): String = preferences.getString("channelID", "").orEmpty()
    fun getAliasName(): String = sessionManager.getNickname().orEmpty()

    suspend fun getAliases(channelID: String): List<DbAlias> = withContext(Dispatchers.IO) {
        dbHelper.getAliasesForChannel(channelID)
    }

    suspend fun getGroups(channelID: String): List<DbGroup> = withContext(Dispatchers.IO) {
        dbHelper.getGroupsForChannel(channelID)
    }

    suspend fun getReception(id: String, code: String?): Result<ReceptionGetResponse> = withContext(Dispatchers.IO) {
        try {
            val csrf = sessionManager.getCsrf() ?: return@withContext Result.failure(Exception("サインインが必要です"))
            val channelID = getChannelId()
            val aliasName = getAliasName()
            
            val response = apiService.getReception(csrf, id, channelID, aliasName, code)
            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    sessionManager.applyResponseCsrf(csrf, body.csrf)
                    if (body.error != null) {
                        Result.failure(Exception(body.error))
                    } else {
                        Result.success(body)
                    }
                } else {
                    Result.failure(Exception("空の応答です"))
                }
            } else {
                Result.failure(Exception("エラー: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun editReception(reception: Reception, menus: List<Menu>, itemDetails: List<ItemDetail>): Result<ReceptionEditResponse> = withContext(Dispatchers.IO) {
        try {
            val csrf = sessionManager.getCsrf() ?: return@withContext Result.failure(Exception("サインインが必要です"))
            val channelID = getChannelId()
            val aliasName = getAliasName()

            val menuPayload = MenuPayload(menus, itemDetails)
            
            val response = apiService.editReception(
                csrf.part(),
                channelID.part(),
                aliasName.part(),
                gson.toJson(reception).part(),
                gson.toJson(menuPayload).part()
            )
            
            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    sessionManager.applyResponseCsrf(csrf, body.csrf)
                    if (body.error != null) {
                        Result.failure(Exception(body.error))
                    } else {
                        Result.success(body)
                    }
                } else {
                    Result.failure(Exception("空の応答です"))
                }
            } else {
                Result.failure(Exception("エラー: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    private fun String.part() = toRequestBody("text/plain".toMediaType())

    suspend fun bookReception(
        receptionID: String,
        aliasName: String,
        bookStart: String,
        bookEnd: String,
        menuID: Int,
        answers: List<String>
    ): Result<ReceptionBookResponse> = withContext(Dispatchers.IO) {
        try {
            val csrf = sessionManager.getCsrf() ?: return@withContext Result.failure(Exception("サインインが必要です"))
            val response = apiService.bookReception(
                csrf,
                aliasName,
                receptionID,
                bookStart,
                bookEnd,
                menuID,
                gson.toJson(answers)
            )
            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    sessionManager.applyResponseCsrf(csrf, body.csrf)
                    if (body.error != null) {
                        Result.failure(Exception(body.error))
                    } else {
                        Result.success(body)
                    }
                } else {
                    Result.failure(Exception("空の応答です"))
                }
            } else {
                Result.failure(Exception("エラー: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun orderReception(
        receptionID: String,
        orders: List<ReceptionOrder>,
        code: String?
    ): Result<ReceptionOrderResponse> = withContext(Dispatchers.IO) {
        try {
            val csrf = sessionManager.getCsrf() ?: return@withContext Result.failure(Exception("サインインが必要です"))
            val response = apiService.orderReception(
                csrf,
                receptionID,
                gson.toJson(orders),
                code
            )
            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    sessionManager.applyResponseCsrf(csrf, body.csrf)
                    if (body.error != null) {
                        Result.failure(Exception(body.error))
                    } else {
                        Result.success(body)
                    }
                } else {
                    Result.failure(Exception("空の応答です"))
                }
            } else {
                Result.failure(Exception("エラー: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun deleteOrder(
        receptionID: String,
        code: String
    ): Result<ReceptionOrderDeleteResponse> = withContext(Dispatchers.IO) {
        try {
            val csrf = sessionManager.getCsrf() ?: return@withContext Result.failure(Exception("サインインが必要です"))
            val channelID = getChannelId()
            val aliasName = getAliasName()
            
            val response = apiService.deleteOrder(
                csrf,
                receptionID,
                code,
                channelID,
                aliasName
            )
            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    sessionManager.applyResponseCsrf(csrf, body.csrf)
                    if (body.error != null) {
                        Result.failure(Exception(body.error))
                    } else {
                        Result.success(body)
                    }
                } else {
                    Result.failure(Exception("空の応答です"))
                }
            } else {
                Result.failure(Exception("エラー: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun queueEdit(
        receptionID: String,
        code: String?,
        guestCount: Int?,
        queueName: String?,
        editType: Int
    ): Result<ReceptionQueueEditResponse> = withContext(Dispatchers.IO) {
        try {
            val csrf = sessionManager.getCsrf() ?: return@withContext Result.failure(Exception("サインインが必要です"))
            val channelID = getChannelId()
            val aliasName = getAliasName()
            
            val response = apiService.queueEdit(
                csrf,
                channelID,
                aliasName,
                receptionID,
                code,
                "1", // codeType = '1' from Vue
                guestCount,
                queueName,
                editType
            )

            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    sessionManager.applyResponseCsrf(csrf, body.csrf)
                    if (body.error != null) {
                        Result.failure(Exception(body.error))
                    } else {
                        Result.success(body)
                    }
                } else {
                    Result.failure(Exception("空の応答です"))
                }
            } else {
                Result.failure(Exception("エラー: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    suspend fun submitShift(
        receptionID: String,
        updatedShifts: List<Shift>
    ): Result<ReceptionShiftResponse> = withContext(Dispatchers.IO) {
        try {
            val csrf = sessionManager.getCsrf() ?: return@withContext Result.failure(Exception("サインインが必要です"))
            val channelID = getChannelId()
            val aliasName = getAliasName()
            
            val response = apiService.submitShift(
                csrf,
                receptionID,
                channelID,
                aliasName,
                gson.toJson(updatedShifts)
            )

            if (response.isSuccessful) {
                val body = response.body()
                if (body != null) {
                    sessionManager.applyResponseCsrf(csrf, body.csrf)
                    if (body.error != null) {
                        Result.failure(Exception(body.error))
                    } else {
                        Result.success(body)
                    }
                } else {
                    Result.failure(Exception("空の応答です"))
                }
            } else {
                Result.failure(Exception("エラー: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

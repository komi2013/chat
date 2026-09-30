package com.chat.android.feature.channel

import android.content.Context
import com.chat.android.core.data.SessionManager
import com.chat.android.core.push.PushReceiveDispatcher
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody
import okhttp3.RequestBody.Companion.toRequestBody
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChannelRepository @Inject constructor(
    @ApplicationContext private val context: Context,
    private val apiService: ChannelApiService,
    private val sessionManager: SessionManager,
    private val pushDispatcher: PushReceiveDispatcher
) {
    private val dbHelper = ChannelDbHelper(context)
    
    private val _dbUpdateFlow = MutableSharedFlow<Unit>(replay = 0)
    val dbUpdateFlow: SharedFlow<Unit> = _dbUpdateFlow.asSharedFlow()

    suspend fun getChannel(channelID: String): DbChannel? = dbHelper.getChannel(channelID)
    suspend fun getAllChannels(): List<DbChannel> = dbHelper.getAllChannels()
    suspend fun getAliases(channelID: String): List<DbAlias> = dbHelper.getAliasesForChannel(channelID)
    suspend fun getGroups(channelID: String): List<DbGroup> = dbHelper.getGroupsForChannel(channelID)

    private fun toPart(value: String): RequestBody = value.toRequestBody("text/plain".toMediaTypeOrNull())

    suspend fun createChannel(name: String, description: String, myname: String, myimg: String): Result<String> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelAdd(
            toPart(name), toPart(description), toPart(myname), toPart(myimg), toPart(csrf)
        )
        
        if (response.isSuccessful) {
            val body = response.body()
            if (body?.csrf != null) {
                sessionManager.setCsrf(body.csrf)
                body.pushContents?.forEach { 
                    pushDispatcher.receive(it) 
                }
                _dbUpdateFlow.emit(Unit)
                return Result.success(body.channelID ?: "")
            }
            return Result.failure(Exception(body?.error ?: "Unknown error"))
        }
        return Result.failure(Exception("Network error: ${response.code()}"))
    }

    suspend fun editChannel(
        channelID: String,
        updatedBy: String,
        pushNamesJson: String,
        channelName: String? = null,
        groupsJson: String? = null,
        description: String? = null
    ): Result<Unit> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelEdit(
            toPart(channelID),
            toPart(updatedBy),
            toPart(pushNamesJson),
            channelName?.let { toPart(it) },
            groupsJson?.let { toPart(it) },
            description?.let { toPart(it) },
            csrf = toPart(csrf)
        )

        if (response.isSuccessful) {
            val body = response.body()
            if (body?.csrf != null) {
                sessionManager.setCsrf(body.csrf)
                body.pushContents?.forEach { pushDispatcher.receive(it) }
                _dbUpdateFlow.emit(Unit)
                return Result.success(Unit)
            }
            return Result.failure(Exception(response.body()?.error ?: "Unknown error"))
        }
        return Result.failure(Exception("Network error"))
    }

    suspend fun deleteChannel(channelID: String, updatedBy: String): Result<Unit> {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = apiService.channelDelete(
            toPart(channelID),
            toPart(updatedBy),
            toPart("1"),
            toPart(csrf)
        )
        if (response.isSuccessful) {
            val body = response.body()
            if (body?.csrf != null) {
                sessionManager.setCsrf(body.csrf)
                body.pushContents?.forEach { pushDispatcher.receive(it) }
                _dbUpdateFlow.emit(Unit)
                return Result.success(Unit)
            }
        }
        return Result.failure(Exception("Delete failed"))
    }
}

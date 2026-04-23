package com.chat.android.data

import com.chat.android.network.ApiService
import com.chat.android.network.Channel
import com.chat.android.network.Message
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class ChannelRepository @Inject constructor(
    private val apiService: ApiService
) {
    
    suspend fun getChannels(csrf: String): Result<List<Channel>> {
        return try {
            val response = apiService.getChannels(csrf)
            if (response.isSuccessful) {
                response.body()?.let { Result.success(it) }
                    ?: Result.failure(Exception("Empty response"))
            } else {
                Result.failure(Exception("Failed to get channels: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
    
    suspend fun createChannel(csrf: String, channel: Channel): Result<Channel> {
        return try {
            val response = apiService.createChannel(csrf, channel.name, channel.description)
            if (response.isSuccessful) {
                response.body()?.let { Result.success(it) }
                    ?: Result.failure(Exception("Empty response"))
            } else {
                Result.failure(Exception("Failed to create channel: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }
}

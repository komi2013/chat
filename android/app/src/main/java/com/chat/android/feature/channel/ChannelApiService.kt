package com.chat.android.feature.channel

import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part

data class ChannelResponse(
    val error: String? = null,
    val csrf: String? = null,
    val pushContents: List<String>? = null,
    val channelID: String? = null
)

interface ChannelApiService {
    @Multipart
    @POST("/ChannelAdd/")
    suspend fun channelAdd(
        @Part("channelName") channelName: RequestBody,
        @Part("channelDescription") channelDescription: RequestBody,
        @Part("myname") myname: RequestBody,
        @Part("myimg") myimg: RequestBody,
        @Part("csrf") csrf: RequestBody
    ): Response<ChannelResponse>

    @Multipart
    @POST("/ChannelEdit/")
    suspend fun channelEdit(
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("pushNames") pushNames: RequestBody,
        @Part("channelName") channelName: RequestBody?,
        @Part("groups") groups: RequestBody?,
        @Part("channelDescription") channelDescription: RequestBody?,
        @Part("admin") admin: RequestBody? = null,
        @Part("deleteAliases") deleteAliases: RequestBody? = null,
        @Part("generateInvitation") generateInvitation: RequestBody? = null,
        @Part("guest") guest: RequestBody? = null,
        @Part("csrf") csrf: RequestBody
    ): Response<ChannelResponse>

    @Multipart
    @POST("/ChannelDelete/")
    suspend fun channelDelete(
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("channelDelete") channelDelete: RequestBody,
        @Part("csrf") csrf: RequestBody
    ): Response<ChannelResponse>
}

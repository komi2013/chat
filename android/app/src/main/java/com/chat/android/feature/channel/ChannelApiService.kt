package com.chat.android.feature.channel

import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part

/**
 * ChannelEdit/ の応答に含まれるチャネル本体。
 *
 * controller/ChannelEdit.go の `channel` フィールドに対応する。
 * 招待コードを生成した応答には channelEdit push が載らないため
 * （controller/ChannelEdit.go は channelName が空だと push を送らない）、
 * クライアントはこの応答からローカルDBを更新する必要がある。
 */
data class ChannelPayload(
    val channelID: String? = null,
    val channelName: String? = null,
    val channelDescription: String? = null,
    val myname: String? = null,
    val invitationCode: String? = null,
    val invitationGuestCode: String? = null
)

data class ChannelResponse(
    val error: String? = null,
    val csrf: String? = null,
    val pushContents: List<String>? = null,
    val channelID: String? = null,
    val channel: ChannelPayload? = null
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

package com.chat.android.feature.ticket

import com.chat.android.core.network.ApiResponse
import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part

interface TicketApiService {
    @Multipart
    @POST("/TicketPost/")
    suspend fun ticketPost(
        @Part("csrf") csrf: RequestBody,
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("pushNames") pushNames: RequestBody,
        @Part("title") title: RequestBody,
        @Part("description") description: RequestBody,
        @Part("status") status: RequestBody,
        @Part("priority") priority: RequestBody
    ): Response<ApiResponse<Any>>

    @Multipart
    @POST("/TicketEdit/")
    suspend fun ticketEdit(
        @Part("csrf") csrf: RequestBody,
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("pushNames") pushNames: RequestBody,
        @Part("ticketID") ticketID: RequestBody,
        @Part("title") title: RequestBody?,
        @Part("description") description: RequestBody?,
        @Part("status") status: RequestBody?,
        @Part("priority") priority: RequestBody?
    ): Response<ApiResponse<Any>>

    @Multipart
    @POST("/TicketDelete/")
    suspend fun ticketDelete(
        @Part("csrf") csrf: RequestBody,
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("pushNames") pushNames: RequestBody,
        @Part("ticketID") ticketID: RequestBody
    ): Response<ApiResponse<Any>>
}

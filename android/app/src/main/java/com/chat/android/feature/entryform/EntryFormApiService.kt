package com.chat.android.feature.entryform

import retrofit2.Response
import retrofit2.http.Headers
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part

data class PushResponse(
    val error: String? = null,
    val csrf: String? = null,
    val pushContents: List<String>? = null
)

interface EntryFormApiService {
    @Headers("Accept: application/json")
    @Multipart
    @POST("ContentsPush/")
    suspend fun sendEntryFormPush(
        @Part("pushNames") pushNames: okhttp3.RequestBody,
        @Part("channelID") channelID: okhttp3.RequestBody,
        @Part("updatedBy") updatedBy: okhttp3.RequestBody,
        @Part("csrf") csrf: okhttp3.RequestBody,
        @Part("contents") contents: okhttp3.RequestBody,
        @Part("pushTitle") pushTitle: okhttp3.RequestBody
    ): Response<PushResponse>
}

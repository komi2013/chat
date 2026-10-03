package com.chat.android.feature.calendar

import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.Headers
import retrofit2.http.Multipart
import retrofit2.http.POST
import retrofit2.http.Part

data class CalendarPushResponse(val error: String? = null, val csrf: String? = null, val pushContents: List<String>? = null)

interface CalendarApiService {
    @Headers("Accept: application/json")
    @Multipart
    @POST("ContentsPush/")
    suspend fun pushCalendar(
        @Part("pushNames") pushNames: RequestBody,
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("csrf") csrf: RequestBody,
        @Part("contents") contents: RequestBody,
        @Part("pushTitle") pushTitle: RequestBody
    ): Response<CalendarPushResponse>

    @Headers("Accept: application/json")
    @Multipart
    @POST("ContentsJustPush/")
    suspend fun requestCalendars(
        @Part("pushNames") pushNames: RequestBody,
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("csrf") csrf: RequestBody,
        @Part("contents") contents: RequestBody,
        @Part("pushTitle") pushTitle: RequestBody
    ): Response<CalendarPushResponse>
}

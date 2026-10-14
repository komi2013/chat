package com.chat.android.feature.reception

import okhttp3.RequestBody
import retrofit2.Response
import retrofit2.http.*

interface ReceptionApiService {
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionGet/")
    suspend fun getReception(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String,
        @Field("channelID") channelID: String,
        @Field("aliasName") aliasName: String,
        @Field("code") code: String? = null
    ): Response<ReceptionGetResponse>

    @Headers("Accept: application/json")
    @Multipart
    @POST("ReceptionEdit/")
    suspend fun editReception(
        @Part("csrf") csrf: RequestBody,
        @Part("channelID") channelID: RequestBody,
        @Part("aliasName") aliasName: RequestBody,
        @Part("reception") reception: RequestBody,
        @Part("menu") menu: RequestBody
    ): Response<ReceptionEditResponse>

    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionBook/")
    suspend fun bookReception(
        @Field("csrf") csrf: String,
        @Field("aliasName") aliasName: String,
        @Field("receptionID") receptionID: String,
        @Field("bookStart") bookStart: String,
        @Field("bookEnd") bookEnd: String,
        @Field("menuID") menuID: Int,
        @Field("answers") answersJson: String
    ): Response<ReceptionBookResponse>

    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionOrder/")
    suspend fun orderReception(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String,
        @Field("receptionOrders") ordersJson: String,
        @Field("code") code: String? = null
    ): Response<ReceptionOrderResponse>

    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionOrderDelete/")
    suspend fun deleteOrder(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String,
        @Field("code") code: String,
        @Field("channelID") channelID: String,
        @Field("aliasName") aliasName: String
    ): Response<ReceptionOrderDeleteResponse>

    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionQueueEdit/")
    suspend fun queueEdit(
        @Field("csrf") csrf: String,
        @Field("channelID") channelID: String,
        @Field("aliasName") aliasName: String,
        @Field("receptionID") receptionID: String,
        @Field("code") code: String? = null,
        @Field("codeType") codeType: String? = null,
        @Field("guestCount") guestCount: Int? = null,
        @Field("queueName") queueName: String? = null,
        @Field("editType") editType: Int
    ): Response<ReceptionQueueEditResponse>

    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionShift/")
    suspend fun submitShift(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String,
        @Field("channelID") channelID: String,
        @Field("aliasName") aliasName: String,
        @Field("updatedShifts") updatedShiftsJson: String
    ): Response<ReceptionShiftResponse>

    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("PushResponse/")
    suspend fun pushResponse(
        @Field("csrf") csrf: String,
        @Field("pushID") pushID: String
    ): Response<PushResponse>
}

data class ReceptionShiftResponse(
    val csrf: String? = null,
    val error: String? = null,
    val pushContents: List<String>? = null
)

data class ReceptionQueueEditResponse(
    val csrf: String? = null,
    val error: String? = null,
    val pushContents: List<String>? = null
)

data class ReceptionOrderResponse(
    val csrf: String? = null,
    val error: String? = null,
    val pushContents: List<String>? = null
)

data class ReceptionOrderDeleteResponse(
    val csrf: String? = null,
    val error: String? = null,
    val pushContents: List<String>? = null
)

data class ReceptionBookResponse(
    val csrf: String? = null,
    val error: String? = null,
    val pushContents: List<String>? = null
)

data class ReceptionGetResponse(
    val csrf: String? = null,
    val error: String? = null,
    val reception: Reception? = null,
    val nickname: String? = null,
    val facilityName: String? = null,
    val menu: MenuPayload? = null,
    val pushContents: List<String>? = null
)

data class MenuPayload(
    val menus: List<Menu>? = null,
    val itemDetails: List<ItemDetail>? = null
)

data class ReceptionEditResponse(
    val csrf: String? = null,
    val error: String? = null,
    val pushContents: List<String>? = null
)

data class PushResponse(
    val error: String? = null,
    val csrf: String? = null,
    val pushContents: List<String>? = null
)

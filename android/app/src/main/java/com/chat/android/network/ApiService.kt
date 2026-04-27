package com.chat.android.network

import okhttp3.ResponseBody
import retrofit2.Response
import retrofit2.http.*

data class PushSubscribeResponse(
    val csrf: String,
    val pushContents: List<String>,
    val success: Boolean,
    val message: String
)

data class GoogleSignInResponse(
    val csrf: String,
    val success: Boolean,
    val message: String,
    val userId: String? = null,
    val nickname: String? = null
)

data class Channel(
    val id: String,
    val name: String,
    val description: String? = null
)

data class Message(
    val id: String,
    val channelId: String,
    val userId: String,
    val content: String,
    val timestamp: Long
)

interface ApiService {
    
    // Mobile push subscription (for Android FCM tokens)
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("PushSubscribeMobile/")
    suspend fun subscribeMobilePush(
        @Field("fcmToken") fcmToken: String,
        @Field("csrf") csrf: String
    ): Response<PushSubscribeResponse>
    
    // Legacy web push subscription (for browsers)
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("PushSubscribe/")
    suspend fun subscribeToPush(
        @Field("subscription") subscription: String,
        @Field("csrf") csrf: String
    ): Response<PushSubscribeResponse>
    
    // Google Sign-In (Dedicated mobile endpoint)
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("SignInGoogleMobile/")
    suspend fun signInWithGoogle(
        @Field("credential") idToken: String,
        @Field("g_csrf_token") csrf: String
    ): Response<ResponseBody>
    
    // Get user info
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("UserGet/")
    suspend fun getUser(
        @Field("csrf") csrf: String
    ): Response<GoogleSignInResponse>
    
    // Get channels
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ChannelGet/")
    suspend fun getChannels(
        @Field("csrf") csrf: String
    ): Response<List<Channel>>
    
    // Create channel
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ChannelPost/")
    suspend fun createChannel(
        @Field("csrf") csrf: String,
        @Field("name") name: String,
        @Field("description") description: String?
    ): Response<Channel>
    
    // Post message
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TweetPost/")
    suspend fun postMessage(
        @Field("csrf") csrf: String,
        @Field("channelId") channelId: String,
        @Field("content") content: String
    ): Response<Message>
    
    // Get messages
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TweetGet/")
    suspend fun getMessages(
        @Field("csrf") csrf: String,
        @Field("channelId") channelId: String
    ): Response<List<Message>>
}

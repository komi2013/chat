package com.chat.android.network

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

interface ApiService {
    
    // Mobile push subscription (for Android FCM tokens)
    @FormUrlEncoded
    @POST("PushSubscribeMobile/")
    suspend fun subscribeMobilePush(
        @Field("fcmToken") fcmToken: String,
        @Field("csrf") csrf: String
    ): Response<PushSubscribeResponse>
    
    // Legacy web push subscription (for browsers)
    @FormUrlEncoded
    @POST("PushSubscribe/")
    suspend fun subscribeToPush(
        @Field("subscription") subscription: String,
        @Field("csrf") csrf: String
    ): Response<PushSubscribeResponse>
    
    // Google Sign-In
    @FormUrlEncoded
    @POST("SignInGoogle/")
    suspend fun signInWithGoogle(
        @Field("idToken") idToken: String,
        @Field("csrf") csrf: String
    ): Response<GoogleSignInResponse>
    
    // Get user info
    @FormUrlEncoded
    @POST("UserGet/")
    suspend fun getUser(
        @Field("csrf") csrf: String
    ): Response<GoogleSignInResponse>
    
    // Get channels
    @FormUrlEncoded
    @POST("ChannelGet/")
    suspend fun getChannels(
        @Field("csrf") csrf: String
    ): Response<List<Any>>
    
    // Post message
    @FormUrlEncoded
    @POST("TweetPost/")
    suspend fun postMessage(
        @Field("csrf") csrf: String,
        @Field("channelId") channelId: String,
        @Field("content") content: String
    ): Response<Any>
    
    // Get messages
    @FormUrlEncoded
    @POST("TweetGet/")
    suspend fun getMessages(
        @Field("csrf") csrf: String,
        @Field("channelId") channelId: String
    ): Response<List<Any>>
}

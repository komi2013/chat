package com.chat.android.core.network

import com.chat.android.feature.channel.ChannelPayload
import okhttp3.RequestBody
import okhttp3.ResponseBody
import retrofit2.Response
import retrofit2.http.*

data class PushSubscribeResponse(
    val csrf: String,
    val pushContents: List<String>,
    /**
     * 所属チャネル一覧。Web版のPushSubscribe（controller/PushSubscribe.go）と
     * 同じ配列。サインイン直後（= PushSubscribeMobile）にローカルSQLiteを
     * 埋めるために使う。再インストールでSQLiteが消えても復元できる。
     */
    val channels: List<ChannelPayload>? = null,
    val success: Boolean,
    val message: String
)

data class GoogleSignInResponse(
    val csrf: String,
    val sessionId: String? = null,
    val success: Boolean? = null,
    val message: String? = null,
    val userId: String? = null,
    val nickname: String? = null,
    val user: UserResponse? = null,
    val nicknames: List<NicknameResponse>? = null,
    val pushContents: List<String>? = null,
    val error: String? = null
)

data class UserResponse(
    val name: String? = null,
    val mail: String? = null,
    val nickname: String? = null,
    val channelID: String? = null,
    val accessRight: String? = null,
    val admin: Boolean? = null,
    val telephone: String? = null,
    val walletAddress: String? = null,
    val latitude: Double? = null,
    val longitude: Double? = null,
    val nickImg: String? = null,
    val nickBio: String? = null
)

data class NicknameResponse(
    val nickname: String,
    val nickImg: String? = null,
    val nickBio: String? = null,
    val good: Int = 0,
    val bad: Int = 0,
    val createdAt: String? = null,
    val accessRight: String? = null
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

// Extended data models for Vue migration
data class ApiResponse<T>(
    val csrf: String? = null,
    val data: T? = null,
    val error: String? = null,
    val message: String? = null,
    val pushContents: List<String>? = null,
    val success: Boolean? = null
)

data class Tweet(
    val messageID: String,
    val channelID: String,
    val aliasName: String,
    val content: String,
    val timestamp: String,
    val good: Int = 0,
    val bad: Int = 0,
    val parentID: String? = null,
    val bookmark: Boolean = false,
    val aliasImg: String? = null
)

data class ChannelDetail(
    val id: String,
    val name: String,
    val description: String? = null,
    val myname: String? = null,
    val myimg: String? = null,
    val aliasImg: String? = null,
    val members: List<String>? = null
)

data class CalendarEvent(
    val id: String,
    val title: String,
    val date: String,
    val description: String? = null,
    val createdBy: String? = null
)

data class Reception(
    val id: String,
    val name: String,
    val description: String? = null,
    val isActive: Boolean = true
)

data class Ticket(
    val ticketID: String,
    val title: String,
    val description: String? = null,
    val status: String,
    val priority: String,
    val createdAt: String
)

data class TimestampEntry(
    val id: String,
    val code: String,
    val action: String,
    val timestamp: String,
    val aliasName: String
)

data class Advertisement(
    val id: String,
    val title: String,
    val content: String,
    val imageUrl: String? = null,
    val linkUrl: String? = null
)

data class Group(
    val id: String,
    val name: String,
    val description: String? = null,
    val members: List<String>? = null
)

data class People(
    val id: String,
    val name: String,
    val img: String? = null,
    val bio: String? = null
)

data class Thread(
    val channel_id: String,
    val parentID: String,
    val messages: List<Message>
)

data class EntryForm(
    val id: String,
    val title: String,
    val formJson: String,
    val isActive: Boolean = true
)

data class PushResponse(
    val error: String? = null,
    val csrf: String? = null,
    val pushContents: List<String>? = null
)

interface ApiService {
    
    // Mobile push subscription (for Android FCM tokens)
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("PushSubscribeMobile/")
    suspend fun subscribeMobilePush(
        @Field("pushToken") pushToken: String,
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
        @Field("csrf") csrf: String,
        // 端末識別子（UUIDv4）。サーバーは同一端末の古いセッションを削除するために使う。
        @Field("deviceID") deviceId: String
    ): Response<ResponseBody>
    
    // Get user info
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("UserGet/")
    suspend fun getUser(
        @Field("csrf") csrf: String
    ): Response<GoogleSignInResponse>
    
    @FormUrlEncoded
    @POST("UserEdit/")
    suspend fun editUser(
        @Field("csrf") csrf: String,
        @Field("nickname") nickname: String,
        @Field("nickImg") nickImg: String?,
        @Field("removeNickImg") removeNickImg: Boolean,
        @Field("nickBio") nickBio: String?,
        @Field("mail") mail: String?,
        @Field("telephone") telephone: String?,
        @Field("walletAddress") walletAddress: String?,
        @Field("latitude") latitude: Double?,
        @Field("longitude") longitude: Double?
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
        @Field("channelId") channelId: String,
        @Field("messageID") messageID: String? = null,
        @Field("date") date: String? = null,
        @Field("parentIDhtml") parentIDhtml: String? = null
    ): Response<List<Tweet>>

    // Get latest tweets
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TweetGetLatest/")
    suspend fun getLatestTweets(
        @Field("csrf") csrf: String
    ): Response<List<Tweet>>

    // Post tweet
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TweetPost/")
    suspend fun postTweet(
        @Field("csrf") csrf: String,
        @Field("content") content: String,
        @Field("channelID") channelID: String? = null,
        @Field("parentID") parentID: String? = null
    ): Response<ApiResponse<Tweet>>

    // Get channel details
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ChannelGet/")
    suspend fun getChannel(
        @Field("csrf") csrf: String,
        @Field("id") id: String? = null
    ): Response<ChannelDetail>

    // Edit channel
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ChannelEdit/")
    suspend fun editChannel(
        @Field("csrf") csrf: String,
        @Field("id") id: String,
        @Field("name") name: String? = null,
        @Field("description") description: String? = null,
        @Field("imgPath") imgPath: String? = null,
        @Field("myname") myname: String? = null,
        @Field("myimg") myimg: String? = null
    ): Response<ApiResponse<ChannelDetail>>

    // Join channel
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ChannelJoin/")
    suspend fun joinChannel(
        @Field("csrf") csrf: String,
        @Field("id") id: String,
        @Field("myname") myname: String,
        @Field("myimg") myimg: String? = null
    ): Response<ApiResponse<ChannelDetail>>

    // Get calendar events
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("CalendarGet/")
    suspend fun getCalendarEvents(
        @Field("csrf") csrf: String,
        @Field("date") date: String? = null
    ): Response<List<CalendarEvent>>

    // Edit calendar
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("CalendarEdit/")
    suspend fun editCalendar(
        @Field("csrf") csrf: String,
        @Field("id") id: String? = null,
        @Field("contents") contents: String
    ): Response<ApiResponse<CalendarEvent>>

    // Get reception
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionGet/")
    suspend fun getReception(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String? = null,
        @Field("channelID") channelID: String? = null,
        @Field("aliasName") aliasName: String? = null,
        @Field("code") code: String? = null,
        @Field("codeType") codeType: String? = null
    ): Response<ApiResponse<Reception>>

    // Reception order
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionOrder/")
    suspend fun receptionOrder(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String,
        @Field("code") code: String,
        @Field("receptionOrders") receptionOrders: String
    ): Response<ApiResponse<Any>>

    // Reception order delete
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionOrderDelete/")
    suspend fun receptionOrderDelete(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String,
        @Field("code") code: String
    ): Response<ApiResponse<Any>>

    // Reception shift
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionShift/")
    suspend fun receptionShift(
        @Field("csrf") csrf: String,
        @Field("updatedShifts") updatedShifts: String
    ): Response<ApiResponse<Any>>

    // Reception queue edit
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ReceptionQueueEdit/")
    suspend fun receptionQueueEdit(
        @Field("csrf") csrf: String,
        @Field("receptionID") receptionID: String,
        @Field("code") code: String,
        @Field("codeType") codeType: String,
        @Field("guestCount") guestCount: Int? = null,
        @Field("queueName") queueName: String? = null,
        @Field("editType") editType: Int
    ): Response<ApiResponse<Any>>

    // Get tickets
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TicketGet/")
    suspend fun getTickets(
        @Field("csrf") csrf: String,
        @Field("ticketID") ticketID: String? = null
    ): Response<List<Ticket>>

    // Get timestamp
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TimestampGet/")
    suspend fun getTimestamp(
        @Field("csrf") csrf: String,
        @Field("adminName") adminName: String,
        @Field("code") code: String
    ): Response<List<TimestampEntry>>

    // Timestamp code
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TimestampCode/")
    suspend fun getTimestampCode(
        @Field("csrf") csrf: String
    ): Response<ApiResponse<Any>>

    // Timestamp report
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("TimestampReport/")
    suspend fun getTimestampReport(
        @Field("csrf") csrf: String,
        @Field("admin") admin: String,
        @Field("month") month: String? = null,
        @Field("stamper") stamper: String? = null,
        @Field("contents") contents: String? = null
    ): Response<ApiResponse<Any>>

    // Get advertisements
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("AdPublicGet/")
    suspend fun getAdvertisements(
        @Field("csrf") csrf: String,
        @Field("date") date: String? = null
    ): Response<List<Advertisement>>

    // Get groups
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("GroupGet/")
    suspend fun getGroups(
        @Field("csrf") csrf: String,
        @Field("id") id: String
    ): Response<Group>

    // Get people
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("PeopleGet/")
    suspend fun getPeople(
        @Field("csrf") csrf: String,
        @Field("id") id: String,
        @Field("name") name: String
    ): Response<People>

    // Get thread
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ThreadGet/")
    suspend fun getThread(
        @Field("csrf") csrf: String,
        @Field("channel_id") channel_id: String,
        @Field("parentID") parentID: String
    ): Response<Thread>

    // Thread head
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("ThreadHeadGet/")
    suspend fun getThreadHead(
        @Field("csrf") csrf: String,
        @Field("parent_id") parent_id: String
    ): Response<ApiResponse<Any>>

    // Get entry form
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("EntryFormGet/")
    suspend fun getEntryForm(
        @Field("csrf") csrf: String,
        @Field("id") id: String
    ): Response<EntryForm>

    // Contents push (generic endpoint)
    // サーバーは channelID / updatedBy / pushNames を必ず要求し、
    // updatedBy が session の ChannelAliases に無いと "no true access right" で拒否する。
    //
    // 注意: @FormUrlEncoded ではなく **@Multipart** が必須。
    // controller/ContentsPush.go:40 が ParseMultipartForm を無条件に呼ぶため、
    // application/x-www-form-urlencoded では必ず
    // "request Content-Type isn't multipart/form-data;files more than 10MB" で失敗する
    // （Web 版の sendRequest は FormData をそのまま fetch に渡すので自動で multipart になる）。
    @Headers("Accept: application/json")
    @Multipart
    @POST("ContentsPush/")
    suspend fun contentsPush(
        @Part("csrf") csrf: RequestBody,
        @Part("channelID") channelID: RequestBody,
        @Part("updatedBy") updatedBy: RequestBody,
        @Part("pushNames") pushNames: RequestBody,
        @Part("contents") contents: RequestBody,
        @Part("pushTitle") pushTitle: RequestBody
    ): Response<ApiResponse<Any>>


    // Contents just push
    // こちらも controller/ContentsJustPush.go:37 が ParseMultipartForm を無条件に呼ぶため
    // @Multipart 必須（現状 Android 側からの呼び出しは無い）。
    @Headers("Accept: application/json")
    @Multipart
    @POST("ContentsJustPush/")
    suspend fun contentsJustPush(
        @Part("csrf") csrf: RequestBody,
        @Part("contents") contents: RequestBody,
        @Part("pushTitle") pushTitle: RequestBody? = null
    ): Response<ApiResponse<Any>>

    // WebRTC token for calls
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("WebRTCTokenGet/")
    suspend fun getWebRTCToken(
        @Field("csrf") csrf: String
    ): Response<ApiResponse<String>>

    // Get answers
    @Headers("Accept: application/json")
    @FormUrlEncoded
    @POST("AnswersGet/")
    suspend fun getAnswers(
        @Field("csrf") csrf: String,
        @Field("id") id: String,
        @Field("group") group: String? = null
    ): Response<ApiResponse<Any>>
}

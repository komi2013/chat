package com.chat.android.core.network

import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.converter.scalars.ScalarsConverterFactory
import com.chat.android.BuildConfig
import com.google.gson.GsonBuilder
import com.chat.android.core.data.SessionManager

object RetrofitClient {

    @Volatile
    private var sessionIdProvider: (() -> String?)? = null

    @Volatile
    private var sessionIdUpdater: ((String) -> Unit)? = null

    fun configureSessionManager(sessionManager: SessionManager) {
        sessionIdProvider = sessionManager::getSessionId
        sessionIdUpdater = sessionManager::setSessionId
    }

    // デバッグビルドでは BODY にして、POST パラメータ（リクエストボディ）とレスポンスボディを
    // Logcat でそのまま確認できるようにする。確認方法: adb logcat -s OkHttp
    // リリースビルドではトークン等の機密情報がログに残らないよう NONE のままにする。
    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = if (BuildConfig.DEBUG) HttpLoggingInterceptor.Level.BODY else HttpLoggingInterceptor.Level.NONE
        redactHeader("Authorization")
        redactHeader("X-Session-Token")
    }

    private val headerInterceptor = okhttp3.Interceptor { chain ->
        val requestBuilder = chain.request().newBuilder()
            .addHeader("Accept", "application/json")
            .addHeader("X-Requested-With", "XMLHttpRequest")
            .addHeader("User-Agent", "Mozilla/5.0 (Linux; Android 10; K) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Mobile Safari/537.36")
            .addHeader("Referer", BuildConfig.BASE_URL)

        sessionIdProvider?.invoke()?.takeIf(String::isNotBlank)?.let { sessionId ->
            requestBuilder.header("Authorization", "Bearer $sessionId")
        }

        val response = chain.proceed(requestBuilder.build())
        response.header("X-Session-Token")?.takeIf(String::isNotBlank)?.let { sessionId ->
            sessionIdUpdater?.invoke(sessionId)
        }
        response
    }

    private val okHttpClient = OkHttpClient.Builder()
        .addInterceptor(headerInterceptor)
        .addInterceptor(loggingInterceptor)
        .build()
    
    private val retrofit by lazy {
        val gson = GsonBuilder()
            .setLenient()
            .create()
            
        Retrofit.Builder()
            .baseUrl(BuildConfig.BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(ScalarsConverterFactory.create())
            .addConverterFactory(GsonConverterFactory.create(gson))
            .build()
    }
    
    val apiService: ApiService by lazy {
        retrofit.create(ApiService::class.java)
    }

    fun <T> createService(serviceClass: Class<T>): T {
        return retrofit.create(serviceClass)
    }
}

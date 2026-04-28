package com.chat.android.network

import okhttp3.CookieJar
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.converter.scalars.ScalarsConverterFactory
import com.chat.android.BuildConfig
import com.google.gson.GsonBuilder
import okhttp3.JavaNetCookieJar
import java.net.CookieManager
import java.net.CookiePolicy
import java.net.HttpCookie
import java.net.URI

object RetrofitClient {
    
    private val cookieManager = CookieManager().apply {
        setCookiePolicy(CookiePolicy.ACCEPT_ALL)
    }

    private val cookieJar = JavaNetCookieJar(cookieManager)

    fun setCsrfCookie(url: String, name: String, value: String) {
        try {
            android.util.Log.d("RetrofitClient", "Setting cookie for URL: $url")
            val uri = URI.create(url)
            val cookie = HttpCookie(name, value).apply {
                path = "/"
                domain = uri.host
                version = 0 
            }
            cookieManager.cookieStore.add(uri, cookie)
            android.util.Log.d("RetrofitClient", "Cookie set successfully: $name=$value")
        } catch (e: Exception) {
            android.util.Log.e("RetrofitClient", "Error setting CSRF cookie", e)
        }
    }

    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = HttpLoggingInterceptor.Level.BODY
    }

    private val headerInterceptor = okhttp3.Interceptor { chain ->
        val request = chain.request().newBuilder()
            .addHeader("Accept", "application/json")
            .addHeader("X-Requested-With", "XMLHttpRequest")
            .addHeader("User-Agent", "ChatAndroid/1.0 (Android)")
            .build()
        chain.proceed(request)
    }

    private val okHttpClient = OkHttpClient.Builder()
        .addInterceptor(headerInterceptor)
        .addInterceptor(loggingInterceptor)
        .cookieJar(cookieJar)
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
}

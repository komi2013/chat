package com.chat.android.network

import okhttp3.Cookie
import okhttp3.CookieJar
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import retrofit2.converter.scalars.ScalarsConverterFactory
import com.chat.android.BuildConfig
import com.google.gson.GsonBuilder

object RetrofitClient {
    
    private val cookieJar = object : CookieJar {
        private val cookieStore = mutableMapOf<String, List<Cookie>>()

        override fun saveFromResponse(url: HttpUrl, cookies: List<Cookie>) {
            cookieStore[url.host] = cookies
        }

        override fun loadForRequest(url: HttpUrl): List<Cookie> {
            return cookieStore[url.host] ?: listOf()
        }

        fun addCookie(url: HttpUrl, cookie: Cookie) {
            val host = url.host
            val cookies = cookieStore[host]?.toMutableList() ?: mutableListOf()
            cookies.add(cookie)
            cookieStore[host] = cookies
        }
    }

    fun setCsrfCookie(url: String, name: String, value: String) {
        try {
            android.util.Log.d("RetrofitClient", "Setting cookie for URL: $url")
            val httpUrl = url.toHttpUrl()
            val cookie = Cookie.Builder()
                .name(name)
                .value(value)
                .domain(httpUrl.host)
                .build()
            cookieJar.addCookie(httpUrl, cookie)
            android.util.Log.d("RetrofitClient", "Cookie set successfully")
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

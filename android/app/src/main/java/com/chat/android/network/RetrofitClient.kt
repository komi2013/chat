package com.chat.android.network

import okhttp3.Cookie
import okhttp3.CookieJar
import okhttp3.HttpUrl
import okhttp3.HttpUrl.Companion.toHttpUrl
import okhttp3.OkHttpClient
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import com.chat.android.BuildConfig

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
        val httpUrl = url.toHttpUrl()
        val cookie = Cookie.Builder()
            .name(name)
            .value(value)
            .domain(httpUrl.host)
            .build()
        cookieJar.addCookie(httpUrl, cookie)
    }

    private val okHttpClient = OkHttpClient.Builder()
        .cookieJar(cookieJar)
        .build()
    
    private val retrofit by lazy {
        Retrofit.Builder()
            .baseUrl(BuildConfig.BASE_URL)
            .client(okHttpClient)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }
    
    val apiService: ApiService by lazy {
        retrofit.create(ApiService::class.java)
    }
}

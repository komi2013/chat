package com.chat.android.network

import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import com.chat.android.BuildConfig

object RetrofitClient {
    
    private val retrofit by lazy {
        Retrofit.Builder()
            .baseUrl(BuildConfig.BASE_URL)
            .addConverterFactory(GsonConverterFactory.create())
            .build()
    }
    
    val apiService: ApiService by lazy {
        retrofit.create(ApiService::class.java)
    }
}

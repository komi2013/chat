package com.chat.android

import android.app.Application
import coil.ImageLoader
import coil.ImageLoaderFactory
import com.chat.android.core.network.RetrofitClient
import dagger.hilt.android.HiltAndroidApp

@HiltAndroidApp
class ChatApplication : Application(), ImageLoaderFactory {
    override fun newImageLoader(): ImageLoader = ImageLoader.Builder(this)
        .okHttpClient(RetrofitClient.authenticatedOkHttpClient())
        .build()
}

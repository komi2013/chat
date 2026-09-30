package com.chat.android.core.util

import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.util.Base64
import com.chat.android.BuildConfig
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import java.io.ByteArrayOutputStream
import java.net.HttpURLConnection
import java.net.URL

/**
 * 保存済み画像パスをサーバーが受け付ける DataURI へ再エンコードする。
 *
 * common.ImgSave は「data:image で始まる値」か「,絵文字,#色」以外は
 * "Emoji invalid:" で失敗するため（common/file.go:96）、
 * 保存済みの /img/... のようなパスをそのまま送ると保存が失敗する。
 * UserViewModel と同じ方法で、取得し直して DataURI にしてから送る。
 */
object ImageDataUri {

    private const val MAX_SIZE = 512
    private const val QUALITY = 90

    /** [reference]（相対/絶対パス、DataURI のいずれか）を DataURI に変換する。 */
    suspend fun reencodeStoredImageToDataUri(reference: String): String? = withContext(Dispatchers.IO) {
        runCatching {
            // すでに DataURI ならそのまま返す
            if (reference.startsWith("data:image")) return@runCatching reference

            val absolute = reference.toAbsoluteImageUrl()
            val connection = (URL(absolute).openConnection() as HttpURLConnection).apply {
                connectTimeout = 8000
                readTimeout = 8000
                instanceFollowRedirects = true
            }
            try {
                if (connection.responseCode !in 200..299) return@runCatching null
                val bytes = connection.inputStream.use { it.readBytes() }
                encode(bytes)
            } finally {
                connection.disconnect()
            }
        }.getOrNull()
    }

    private fun encode(bytes: ByteArray): String? {
        if (bytes.isEmpty()) return null
        val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
        BitmapFactory.decodeByteArray(bytes, 0, bytes.size, bounds)
        if (bounds.outWidth <= 0 || bounds.outHeight <= 0) return null

        val largestSide = maxOf(bounds.outWidth, bounds.outHeight)
        val options = BitmapFactory.Options().apply {
            inSampleSize = (largestSide / MAX_SIZE).coerceAtLeast(1)
        }
        val decoded = BitmapFactory.decodeByteArray(bytes, 0, bytes.size, options) ?: return null
        return try {
            val scale = MAX_SIZE.toFloat() / maxOf(decoded.width, decoded.height)
            val width = (decoded.width * minOf(1f, scale)).toInt().coerceAtLeast(1)
            val height = (decoded.height * minOf(1f, scale)).toInt().coerceAtLeast(1)
            val resized = if (width == decoded.width && height == decoded.height) decoded
            else Bitmap.createScaledBitmap(decoded, width, height, true)
            try {
                val isJpeg = bytes.size > 2 &&
                    bytes[0] == 0xFF.toByte() && bytes[1] == 0xD8.toByte()
                val format = if (isJpeg) Bitmap.CompressFormat.JPEG else Bitmap.CompressFormat.PNG
                val mime = if (isJpeg) "image/jpeg" else "image/png"
                val output = ByteArrayOutputStream()
                resized.compress(format, QUALITY, output)
                "data:$mime;base64,${Base64.encodeToString(output.toByteArray(), Base64.NO_WRAP)}"
            } finally {
                if (resized !== decoded) resized.recycle()
            }
        } finally {
            decoded.recycle()
        }
    }
}

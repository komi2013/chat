@file:OptIn(com.ntt.skyway.core.SkyWayOptIn::class)

package com.chat.android.feature.call

import android.Manifest
import android.graphics.BitmapFactory
import android.graphics.ImageFormat
import android.graphics.Rect
import android.graphics.YuvImage
import android.os.SystemClock
import android.util.Log
import android.widget.ImageView
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.hilt.navigation.compose.hiltViewModel
import com.ntt.skyway.core.content.VideoStream
import com.ntt.skyway.core.content.sink.CustomRenderer
import java.io.ByteArrayOutputStream
import java.util.concurrent.atomic.AtomicLong
import java.util.concurrent.atomic.AtomicReference

@Composable
fun ThreadCallScreen(
    code: String,
    onLeave: () -> Unit,
    viewModel: ThreadCallViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()
    val permissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        if (permissions[Manifest.permission.RECORD_AUDIO] == true) {
            // Camera is optional so people can still join an audio call when they
            // decline camera access or their device has no camera.
            viewModel.start(code, withVideo = permissions[Manifest.permission.CAMERA] == true)
        } else {
            viewModel.permissionDenied()
        }
    }

    LaunchedEffect(code) {
        permissionLauncher.launch(arrayOf(Manifest.permission.RECORD_AUDIO, Manifest.permission.CAMERA))
    }
    Box(Modifier.fillMaxSize().background(Color.Black)) {
        state.remoteVideo?.let { stream ->
            SkyWayVideoPreview(stream, Modifier.fillMaxSize(), "remote")
        }
        state.localVideo?.let { stream ->
            SkyWayVideoPreview(
                stream,
                modifier = if (state.remoteVideo == null) Modifier.fillMaxSize() else {
                    Modifier.align(Alignment.TopEnd).padding(16.dp).size(width = 112.dp, height = 160.dp)
                },
                label = "local"
            )
        }
        Column(
            Modifier.align(Alignment.BottomCenter).fillMaxWidth().padding(24.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            when {
                state.error != null -> Text(state.error.orEmpty(), color = Color.White)
                state.connecting -> CircularProgressIndicator(color = Color.White)
                state.remoteVideoError != null -> Text("相手の映像エラー: ${state.remoteVideoError}", color = Color.White)
                state.localVideoError != null -> Text("カメラエラー: ${state.localVideoError}", color = Color.White)
                state.remoteAudioConnected && !state.remoteVideoPublicationSeen ->
                    Text("音声は接続済みです。相手から映像が届いていません。", color = Color.White)
                state.remoteVideo == null -> Text("相手の参加を待っています…", color = Color.White)
            }
            Spacer(Modifier.height(16.dp))
            Button(
                onClick = { viewModel.leave(); onLeave() },
                colors = ButtonDefaults.buttonColors(containerColor = Color.Red)
            ) {
                Text("通話を終了")
            }
        }
    }
}

@Composable
private fun SkyWayVideoPreview(stream: VideoStream, modifier: Modifier, label: String) {
    val renderer = remember(stream) { CustomRenderer() }
    val imageViewRef = remember(stream) { AtomicReference<ImageView?>(null) }
    val lastFrameTime = remember(stream) { AtomicLong(0L) }
    val frameLogged = remember(stream) { AtomicLong(0L) }

    DisposableEffect(stream, renderer) {
        renderer.onFrameHandler = { frame ->
            val now = SystemClock.elapsedRealtime()
            val previous = lastFrameTime.get()
            if (now - previous >= 200 && lastFrameTime.compareAndSet(previous, now)) {
                if (frameLogged.compareAndSet(0L, 1L)) {
                    Log.d(TAG, "Custom $label preview received video frame ${frame.width}x${frame.height}")
                }
                frame.toBitmap()?.let { bitmap ->
                    imageViewRef.get()?.post { imageViewRef.get()?.setImageBitmap(bitmap) }
                }
            }
        }
        onDispose {
            stream.removeRenderer(renderer)
            renderer.dispose()
            imageViewRef.set(null)
        }
    }

    AndroidView(
        modifier = modifier,
        factory = { context ->
            ImageView(context).also { imageView ->
                imageView.setBackgroundColor(android.graphics.Color.BLACK)
                imageView.scaleType = ImageView.ScaleType.FIT_CENTER
                imageViewRef.set(imageView)
            }
        },
        update = {
            stream.removeRenderer(renderer)
            stream.addRenderer(renderer)
        }
    )
}

private fun CustomRenderer.VideoFrameBuffer.toBitmap(): android.graphics.Bitmap? = runCatching {
    val frameWidth = width
    val frameHeight = height
    if (frameWidth <= 0 || frameHeight <= 0) return null

    val nv21 = ByteArray(frameWidth * frameHeight * 3 / 2)
    val yPlane = dataY.duplicate()
    val uPlane = dataU.duplicate()
    val vPlane = dataV.duplicate()
    val yBase = yPlane.position()
    val uBase = uPlane.position()
    val vBase = vPlane.position()

    for (row in 0 until frameHeight) {
        for (column in 0 until frameWidth) {
            nv21[row * frameWidth + column] = yPlane.get(yBase + row * strideY + column)
        }
    }
    val chromaWidth = (frameWidth + 1) / 2
    val chromaHeight = (frameHeight + 1) / 2
    var output = frameWidth * frameHeight
    for (row in 0 until chromaHeight) {
        for (column in 0 until chromaWidth) {
            nv21[output++] = vPlane.get(vBase + row * strideV + column)
            nv21[output++] = uPlane.get(uBase + row * strideU + column)
        }
    }

    val jpeg = ByteArrayOutputStream()
    if (!YuvImage(nv21, ImageFormat.NV21, frameWidth, frameHeight, null)
            .compressToJpeg(Rect(0, 0, frameWidth, frameHeight), 75, jpeg)
    ) return null
    BitmapFactory.decodeByteArray(jpeg.toByteArray(), 0, jpeg.size())
}.getOrNull()

private const val TAG = "ThreadCall"

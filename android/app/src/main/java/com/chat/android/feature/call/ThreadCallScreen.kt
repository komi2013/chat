package com.chat.android.feature.call

import android.Manifest
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.viewinterop.AndroidView
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import com.ntt.skyway.core.content.local.LocalVideoStream
import com.ntt.skyway.core.content.remote.RemoteVideoStream
import com.ntt.skyway.core.content.sink.SurfaceViewRenderer

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
    DisposableEffect(Unit) { onDispose { viewModel.leave() } }

    Box(Modifier.fillMaxSize().background(Color.Black)) {
        state.remoteVideo?.let { stream ->
            SkyWayVideo(stream, Modifier.fillMaxSize(), isMediaOverlay = false)
        }
        state.localVideo?.let { stream ->
            SkyWayVideo(
                stream,
                modifier = if (state.remoteVideo == null) {
                    Modifier.fillMaxSize()
                } else {
                    Modifier.align(Alignment.TopEnd).padding(16.dp).size(width = 112.dp, height = 160.dp)
                },
                isMediaOverlay = state.remoteVideo != null
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
            Button(onClick = { viewModel.leave(); onLeave() }, colors = ButtonDefaults.buttonColors(containerColor = Color.Red)) {
                Text("通話を終了")
            }
        }
    }
}

@Composable
private fun SkyWayVideo(stream: Any, modifier: Modifier, isMediaOverlay: Boolean) {
    var renderer by remember { mutableStateOf<SurfaceViewRenderer?>(null) }
    AndroidView(
        modifier = modifier,
        factory = { context ->
            SurfaceViewRenderer(context).also {
                it.setup()
                // Keep a lone preview on the normal surface layer. Only lift the
                // local thumbnail above the remote surface when both are present.
                it.setZOrderMediaOverlay(isMediaOverlay)
                renderer = it
            }
        },
        update = { view ->
            when (stream) {
                is LocalVideoStream -> { stream.removeRenderer(view); stream.addRenderer(view) }
                is RemoteVideoStream -> { stream.removeRenderer(view); stream.addRenderer(view) }
            }
        }
    )
    DisposableEffect(stream, renderer) {
        onDispose {
            renderer?.let { view ->
                when (stream) {
                    is LocalVideoStream -> stream.removeRenderer(view)
                    is RemoteVideoStream -> stream.removeRenderer(view)
                }
                view.dispose()
            }
        }
    }
}

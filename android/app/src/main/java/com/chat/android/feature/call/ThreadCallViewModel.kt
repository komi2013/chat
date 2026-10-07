package com.chat.android.feature.call

import android.content.Context
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.network.ApiService
import com.chat.android.core.network.SessionManager
import com.ntt.skyway.core.SkyWayContext
import com.ntt.skyway.core.content.Stream
import com.ntt.skyway.core.content.local.LocalAudioStream
import com.ntt.skyway.core.content.local.LocalVideoStream
import com.ntt.skyway.core.content.local.source.AudioSource
import com.ntt.skyway.core.content.local.source.CameraSource
import com.ntt.skyway.core.content.remote.RemoteVideoStream
import com.ntt.skyway.room.Room
import com.ntt.skyway.room.RoomPublication
import com.ntt.skyway.room.member.LocalRoomMember
import com.ntt.skyway.room.member.RoomMember
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import java.util.UUID
import javax.inject.Inject

data class ThreadCallState(
    val connecting: Boolean = true,
    val localVideo: LocalVideoStream? = null,
    val remoteVideo: RemoteVideoStream? = null,
    val error: String? = null,
    val ended: Boolean = false
)

@HiltViewModel
class ThreadCallViewModel @Inject constructor(
    @ApplicationContext private val context: Context,
    private val api: ApiService,
    private val sessionManager: SessionManager
) : ViewModel() {
    private val _state = MutableStateFlow(ThreadCallState())
    val state = _state.asStateFlow()

    private var room: Room? = null
    private var member: LocalRoomMember? = null
    private var audio: LocalAudioStream? = null
    private var started = false

    fun start(code: String) {
        if (started) return
        started = true
        viewModelScope.launch {
            try {
                val csrf = sessionManager.getCsrf()?.takeIf { it.isNotBlank() }
                    ?: error("サインインが必要です")
                val response = api.getWebRTCToken(csrf)
                val body = response.body()
                if (!response.isSuccessful) error("通話トークンを取得できません (${response.code()})")
                val token = body?.token?.takeIf { it.isNotBlank() }
                    ?: error(body?.error ?: "通話トークンを取得できません")

                check(SkyWayContext.setup(context, token)) { "SkyWay に接続できません" }

                val camera = CameraSource.getFrontCameras(context).firstOrNull()
                    ?: error("カメラが見つかりません")
                CameraSource.startCapturing(context, camera, CameraSource.CapturingOptions(800, 800))
                val video = CameraSource.createStream()
                AudioSource.start()
                audio = AudioSource.createStream()
                _state.value = _state.value.copy(localVideo = video)

                room = Room.findOrCreate(name = code)
                member = room?.join(RoomMember.Init(name = "android_${UUID.randomUUID()}"))
                    ?: error("通話ルームに参加できません")
                val localMember = member ?: error("通話ルームに参加できません")

                room?.onStreamPublishedHandler = published@{ publication ->
                    if (publication.publisher?.id == localMember.id) return@published
                    subscribe(publication)
                }
                room?.publications?.forEach { publication ->
                    if (publication.publisher?.id != localMember.id) subscribe(publication)
                }
                val sfu = RoomPublication.Options(type = RoomPublication.Type.SFU)
                localMember.publish(video, sfu)
                audio?.let { localMember.publish(it, sfu) }
                _state.value = _state.value.copy(connecting = false)
            } catch (error: Exception) {
                _state.value = _state.value.copy(
                    connecting = false,
                    error = error.message ?: "通話を開始できませんでした"
                )
                leave()
            }
        }
    }

    private fun subscribe(publication: RoomPublication) {
        viewModelScope.launch {
            runCatching { member?.subscribe(publication) }
                .getOrNull()?.stream?.let { stream ->
                    if (stream.contentType == Stream.ContentType.VIDEO) {
                        _state.value = _state.value.copy(remoteVideo = stream as RemoteVideoStream)
                    }
                }
        }
    }

    fun leave() {
        if (_state.value.ended) return
        _state.value = _state.value.copy(ended = true)
        viewModelScope.launch(Dispatchers.IO) {
            runCatching { member?.leave() }
            runCatching { room?.dispose() }
            runCatching { AudioSource.stop() }
            runCatching { CameraSource.stopCapturing() }
            runCatching { SkyWayContext.dispose() }
            withContext(Dispatchers.Main) {
                member = null
                room = null
                audio = null
            }
        }
    }

    fun permissionDenied() {
        _state.value = _state.value.copy(
            connecting = false,
            error = "ビデオ通話にはカメラとマイクの許可が必要です"
        )
    }

    override fun onCleared() {
        leave()
        super.onCleared()
    }
}

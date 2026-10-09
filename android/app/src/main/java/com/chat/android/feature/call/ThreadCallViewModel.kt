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
    val remoteAudioConnected: Boolean = false,
    val remoteVideoPublicationSeen: Boolean = false,
    val localVideoError: String? = null,
    val remoteVideoError: String? = null,
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
    private var video: LocalVideoStream? = null
    private val subscribingPublicationIds = mutableSetOf<String>()
    private var started = false

    fun start(code: String, withVideo: Boolean = true) {
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

                AudioSource.start()
                audio = AudioSource.createStream()
                if (withVideo) {
                    runCatching {
                        val camera = CameraSource.getFrontCameras(context).firstOrNull()
                            ?: error("カメラが見つかりません")
                        CameraSource.startCapturing(context, camera, CameraSource.CapturingOptions(800, 800))
                        CameraSource.createStream()
                    }.onSuccess { localVideo ->
                        video = localVideo
                        _state.value = _state.value.copy(localVideo = localVideo)
                    }.onFailure { cameraError ->
                        val reason = cameraError.message ?: cameraError.javaClass.simpleName
                        android.util.Log.e(TAG, "Camera initialization failed", cameraError)
                        _state.value = _state.value.copy(localVideoError = reason)
                    }
                } else {
                    _state.value = _state.value.copy(localVideoError = "カメラの権限が許可されていません")
                }

                room = Room.findOrCreate(name = code)
                member = room?.join(RoomMember.Init(name = "android_${UUID.randomUUID()}"))
                    ?: error("通話ルームに参加できません")
                val localMember = member ?: error("通話ルームに参加できません")

                // Install the listener before inspecting existing publications so
                // a stream published during join cannot be missed.
                room?.onStreamPublishedHandler = published@{ publication ->
                    if (publication.publisher?.id == localMember.id) return@published
                    subscribe(publication)
                }
                room?.publications?.forEach { publication ->
                    if (publication.publisher?.id != localMember.id) subscribe(publication)
                }
                val sfu = RoomPublication.Options(type = RoomPublication.Type.SFU)
                android.util.Log.d(TAG, "Publishing local video=${video != null} and microphone=${audio != null}")
                video?.let { localMember.publish(it, sfu) }
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
        android.util.Log.d(
            TAG,
            "Remote publication received id=${publication.id} content=${publication.contentType}"
        )
        synchronized(subscribingPublicationIds) {
            if (!subscribingPublicationIds.add(publication.id)) return
        }
        if (publication.contentType == Stream.ContentType.VIDEO) {
            _state.value = _state.value.copy(remoteVideoPublicationSeen = true)
        }
        viewModelScope.launch {
            val result = runCatching {
                member?.subscribe(publication) ?: error("通話メンバーが未接続です")
            }
            result.onSuccess { subscription ->
                val stream = subscription.stream
                android.util.Log.d(TAG, "Subscribed publication=${publication.id} content=${stream?.contentType}")
                if (stream == null && publication.contentType == Stream.ContentType.VIDEO) {
                    _state.value = _state.value.copy(remoteVideoError = "相手の映像ストリームが空です")
                } else if (stream?.contentType == Stream.ContentType.AUDIO) {
                    _state.value = _state.value.copy(remoteAudioConnected = true)
                } else if (stream?.contentType == Stream.ContentType.VIDEO) {
                    val remoteVideo = stream as? RemoteVideoStream
                    _state.value = _state.value.copy(
                        remoteVideo = remoteVideo,
                        remoteVideoError = if (remoteVideo == null) "相手の映像を読み込めません" else null
                    )
                } else if (publication.contentType == Stream.ContentType.VIDEO) {
                    _state.value = _state.value.copy(remoteVideoError = "購読した映像の形式が不正です")
                }
            }.onFailure { subscriptionError ->
                android.util.Log.e(TAG, "Subscribe failed publication=${publication.id}", subscriptionError)
                if (publication.contentType == Stream.ContentType.VIDEO) {
                    _state.value = _state.value.copy(
                        remoteVideoError = subscriptionError.message ?: "相手の映像を購読できません"
                    )
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
                video = null
                subscribingPublicationIds.clear()
            }
        }
    }

    fun permissionDenied() {
        _state.value = _state.value.copy(
            connecting = false,
            error = "音声通話にはマイクの許可が必要です"
        )
    }

    override fun onCleared() {
        leave()
        super.onCleared()
    }

    private companion object {
        const val TAG = "ThreadCall"
    }
}

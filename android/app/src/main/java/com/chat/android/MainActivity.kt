package com.chat.android

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import com.chat.android.navigation.ChatNavigation
import com.chat.android.core.ui.theme.ChatAndroidTheme
import dagger.hilt.android.AndroidEntryPoint
import android.Manifest
import android.content.Intent
import android.content.pm.PackageManager
import android.util.Log
import androidx.activity.result.contract.ActivityResultContracts
import androidx.core.content.ContextCompat
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import com.chat.android.core.repository.UserRepository
import com.chat.android.navigation.ProfileRoute
import com.chat.android.firebase.PushNotificationManager
import com.chat.android.core.util.InvitationMapper
import com.android.installreferrer.api.InstallReferrerClient
import com.android.installreferrer.api.InstallReferrerStateListener
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    // Compose から購読できるよう mutableStateOf にする。
    // 通常の var にすると onNewIntent で更新しても再コンポジションが起きず、
    // 起動済みのアプリに-drop した招待リンクが無視される。
    private var inviteRequest by mutableStateOf<InviteRequest?>(null)
    private var callRequest by mutableStateOf<IncomingCallRequest?>(null)

    private var referrerHandled = false

    @Inject
    lateinit var userRepository: UserRepository

    private val notificationPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted ->
        Log.d(TAG, "POST_NOTIFICATIONS granted=$granted")
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        requestNotificationPermission()

        // 招待URL（chat://profile/{id}?code=...）で起動された場合
        val directInviteHandled = handleInviteIntent(intent)
        handleCallIntent(intent)

        if (!directInviteHandled) {
            checkInstallReferrer()
        }

        setContent {
            ChatAndroidTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    val isSignedIn by userRepository.signInState.collectAsState()
                    val currentChannelId by userRepository.currentChannelId.collectAsState()

                    ChatNavigation(
                        inviteRequest = inviteRequest,
                        callRequest = callRequest,
                        isSignedIn = isSignedIn,
                        initialChannelID = currentChannelId
                    )
                }
            }
        }
    }

    /**
     * launchMode=singleTask のため、既に起動済みのときは onNewIntent が呼ばれる。
     * inviteRequest は Compose の state なので、ここでの代入で再コンポジションし、
     * ChatNavigation 側の LaunchedEffect が改めて遷移する。
     */
    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        handleInviteIntent(intent)
        handleCallIntent(intent)
    }

    private fun handleCallIntent(intent: Intent?) {
        val code = intent?.getStringExtra(PushNotificationManager.EXTRA_CALL_CODE)
            ?.takeIf(String::isNotBlank) ?: return
        callRequest = IncomingCallRequest(code, System.nanoTime())
    }

    private fun handleInviteIntent(intent: Intent?): Boolean {
        val data = intent?.data ?: return false
        val route = InvitationMapper.parseUri(data) ?: return false
        
        // 同じURLを2回叩いても遷移できるよう token を付けて毎回別値にする
        val hasSession = userRepository.isSignedIn()
        inviteRequest = InviteRequest(route, System.nanoTime(), hasSession)
        
        // 未サインインなら参加できないので、Vue の Profile.vue と同じく
        // TO へ保存し、サインイン後に再開できるようにする。
        if (!hasSession) {
            val url = inviteUrlString(route)
            if (url.isNotBlank()) userRepository.setTO(url)
        }
        return true
    }

    private fun checkInstallReferrer() {
        if (referrerHandled) return
        
        val referrerClient = InstallReferrerClient.newBuilder(this).build()
        referrerClient.startConnection(object : InstallReferrerStateListener {
            override fun onInstallReferrerSetupFinished(responseCode: Int) {
                when (responseCode) {
                    InstallReferrerClient.InstallReferrerResponse.OK -> {
                        try {
                            val response = referrerClient.installReferrer
                            val referrerUrl = response.installReferrer
                            Log.d(TAG, "Install referrer: $referrerUrl")
                            
                            val route = InvitationMapper.parseReferrer(referrerUrl)
                            if (route != null) {
                                val hasSession = userRepository.isSignedIn()
                                inviteRequest = InviteRequest(route, System.nanoTime(), hasSession)
                                referrerHandled = true
                                if (!hasSession) {
                                    userRepository.setTO(inviteUrlString(route))
                                }
                            }
                        } catch (e: Exception) {
                            Log.e(TAG, "Failed to get install referrer", e)
                        } finally {
                            referrerClient.endConnection()
                        }
                    }
                    InstallReferrerClient.InstallReferrerResponse.FEATURE_NOT_SUPPORTED -> {
                        Log.w(TAG, "Install Referrer not supported")
                    }
                    InstallReferrerClient.InstallReferrerResponse.SERVICE_UNAVAILABLE -> {
                        Log.w(TAG, "Install Referrer service unavailable")
                    }
                }
            }

            override fun onInstallReferrerServiceDisconnected() {}
        })
    }

    /** TO に保存する、復元用のURL文字列。 */
    private fun inviteUrlString(route: Any): String = when (route) {
        is ProfileRoute -> buildString {
            append("chat://profile/")
            append(route.id.orEmpty())
            route.code?.takeIf { it.isNotBlank() }?.let { append("?code=").append(it) }
        }
        is com.chat.android.navigation.CalendarImportRoute -> route.uri
        else -> ""
    }

    private fun requestNotificationPermission() {
        if (android.os.Build.VERSION.SDK_INT >= android.os.Build.VERSION_CODES.TIRAMISU &&
            ContextCompat.checkSelfPermission(
                this,
                Manifest.permission.POST_NOTIFICATIONS
            ) != PackageManager.PERMISSION_GRANTED
        ) {
            notificationPermissionLauncher.launch(Manifest.permission.POST_NOTIFICATIONS)
        }
    }

    companion object {
        private const val TAG = "FCM_TEST"
    }
}

/**
 * 招待URLを受け取ったことを表す値。
 *
 * 同じURLを2回叩いた場合も別として扱うため token(呼び出し時刻)を持たせる。
 * Compose の state は構造比較で等価判定されるため、token が無いと
 * 2回目は state が変わらず遷移しない。
 *
 * @param hasSession 招待時点でサインイン済みか。false ならサインイン画面へ送り、
 *   TO（UserRepository）に保存したURLから参加を再開する。
 */
data class InviteRequest(
    val route: Any,
    val token: Long,
    val hasSession: Boolean
)

data class IncomingCallRequest(val code: String, val token: Long)

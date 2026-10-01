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
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import com.chat.android.core.repository.UserRepository
import com.chat.android.navigation.ProfileRoute
import javax.inject.Inject

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    // Compose から購読できるよう mutableStateOf にする。
    // 通常の var にすると onNewIntent で更新しても再コンポジションが起きず、
    // 起動済みのアプリに-drop した招待リンクが無視される。
    private var inviteRequest by mutableStateOf<InviteRequest?>(null)

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
        handleInviteIntent(intent)

        setContent {
            ChatAndroidTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    ChatNavigation(inviteRequest = inviteRequest)
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
    }

    private fun handleInviteIntent(intent: Intent?) {
        val route = parseInviteLink(intent) ?: return
        // ChannelJoin に渡せる csrf があるか（＝サインイン済みか）を先に判定する
        val hasSession = !userRepository.getCsrfToken().isNullOrBlank()
        // 同じURLを2回叩いても遷移できるよう token を付けて毎回別値にする
        inviteRequest = InviteRequest(route, System.nanoTime(), hasSession)
        // 未サインインなら参加できないので、Vue の Profile.vue と同じく
        // TO へ保存し、サインイン後に再開できるようにする。
        if (!hasSession) {
            userRepository.setTO(inviteUrlString(route))
        }
    }

    /** TO に保存する、復元用のURL文字列。 */
    private fun inviteUrlString(route: ProfileRoute): String = buildString {
        append("chat://profile/")
        append(route.id.orEmpty())
        route.code?.takeIf { it.isNotBlank() }?.let { append("?code=").append(it) }
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
    val route: ProfileRoute,
    val token: Long,
    val hasSession: Boolean
)

/**
 * 招待URLを ProfileRoute に分解する。
 */
private fun parseInviteLink(intent: Intent?): ProfileRoute? {
    val data = intent?.data ?: return null

    // カスタムスキーム: chat://profile/{channelID}?code=XXX
    //   host = "profile", pathSegments = [channelID]
    if (data.scheme == "chat") {
        if (data.host != "profile") return null
        val channelID = data.pathSegments?.firstOrNull().orEmpty()
        if (channelID.isBlank()) return null
        return ProfileRoute(id = channelID, code = data.getQueryParameter("code"))
    }

    // Web版URL: https://chat.quigen.info/profile/{channelID}/?code=XXX
    //   host = ドメイン, pathSegments = ["profile", channelID]
    if (data.scheme != "https" && data.scheme != "http") return null
    val segments = data.pathSegments ?: return null
    if (segments.size < 2 || segments[0] != "profile") return null
    val channelID = segments[1]
    if (channelID.isBlank()) return null
    return ProfileRoute(id = channelID, code = data.getQueryParameter("code"))
}

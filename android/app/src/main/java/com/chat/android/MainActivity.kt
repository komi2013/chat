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
import com.chat.android.navigation.ProfileRoute

@AndroidEntryPoint
class MainActivity : ComponentActivity() {

    private val notificationPermissionLauncher = registerForActivityResult(
        ActivityResultContracts.RequestPermission()
    ) { granted ->
        Log.d(TAG, "POST_NOTIFICATIONS granted=$granted")
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        requestNotificationPermission()

        // 招待URL（/profile/{id}/?code=...）で起動された場合
        initialInviteLink = parseInviteLink(intent)

        setContent {
            ChatAndroidTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    ChatNavigation(inviteLink = initialInviteLink)
                }
            }
        }
    }

    /**
     * launchMode=singleTask のため、既に起動済みのときは onNewIntent が呼ばれる。
     * Compose 側は startDestination として一度だけ使うため、ここでは state を
     * 更新するだけで、再コンポジションのたびに遷移しない。
     */
    override fun onNewIntent(intent: Intent) {
        super.onNewIntent(intent)
        setIntent(intent)
        initialInviteLink = parseInviteLink(intent)
    }
    /** 招待URLを ProfileRoute に分解する。 */
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

    /** 招待URLから読み取った ProfileRoute（起動時のみ）。 */
    private var initialInviteLink: ProfileRoute? = null
}

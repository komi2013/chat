package com.chat.android.ui.screen

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.ui.component.DrawerLayout
import com.chat.android.ui.viewmodel.TopViewModel
import com.chat.android.ui.viewmodel.TopUiState
import com.chat.android.navigation.UserRoute
import android.content.Context
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.text.AnnotatedString
import com.chat.android.BuildConfig
import com.google.firebase.messaging.FirebaseMessaging
import java.text.SimpleDateFormat
import java.util.Date
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TopScreen(
    navController: NavController,
    viewModel: TopViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    var fcmToken by remember { mutableStateOf<String?>(null) }
    var fcmError by remember { mutableStateOf<String?>(null) }
    val clipboardManager = LocalClipboardManager.current
    val context = LocalContext.current

    LaunchedEffect(Unit) {
        viewModel.refreshTopLinks()
        FirebaseMessaging.getInstance().token
            .addOnSuccessListener { token ->
                fcmToken = token
                context.getSharedPreferences("chat_prefs", Context.MODE_PRIVATE)
                    .edit()
                    .putString("fcm_token", token)
                    .apply()
            }
            .addOnFailureListener { error ->
                fcmError = error.message ?: "FCM token取得に失敗しました"
            }
    }

    DrawerLayout(navController = navController) { paddingValues ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(paddingValues)
                .padding(16.dp)
        ) {
            // Header (title only - menu is handled by DrawerLayout)
            Text(
                text = "ホーム",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 24.dp),
                textAlign = TextAlign.Center
            )

            // インストール中の APK を端末で目視確認するためのビルド情報
            BuildInfoText(modifier = Modifier.padding(bottom = 16.dp))

            // Loading indicator
            if (uiState.isLoading) {
                Box(
                    modifier = Modifier.fillMaxWidth(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }

            // Error message
            uiState.errorMessage?.let { error ->
                Card(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(vertical = 8.dp),
                    colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer)
                ) {
                    Text(
                        text = error,
                        modifier = Modifier.padding(16.dp),
                        color = MaterialTheme.colorScheme.onErrorContainer
                    )
                }
            }

            FcmTokenCard(
                token = fcmToken,
                errorMessage = fcmError,
                onCopy = { token -> clipboardManager.setText(AnnotatedString(token)) }
            )

            // Test button to quickly access Entry Forms
            Button(
                onClick = { navController.navigate("entryFormEdit") },
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 16.dp),
                colors = ButtonDefaults.buttonColors(containerColor = MaterialTheme.colorScheme.secondary)
            ) {
                Text("エントリーフォーム一覧 (テスト用)")
            }

            // Top links
            LazyColumn(
                modifier = Modifier.fillMaxWidth(),
                verticalArrangement = Arrangement.spacedBy(12.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                items(uiState.topLinks) { link ->
                    TopLinkCard(
                        link = link,
                        onClick = {
                            when (link.topLink) {
                                "/sign/" -> navController.navigate("login")
                                "/user/" -> navController.navigate(UserRoute)
                                "/tweets/" -> navController.navigate("tweets")
                                "/channel/" -> navController.navigate("channel")
                                "/setting/" -> navController.navigate("setting")
                                "/adSetting/" -> navController.navigate("adSetting")
                                "/html/rule/" -> navController.navigate("rule")
                                "/html/privacy/" -> navController.navigate("privacy")
                                else -> {
                                    // Handle other routes
                                    navController.navigate(link.topLink.removePrefix("/"))
                                }
                            }
                        }
                    )
                }
            }
        }
    }
}

@Composable
private fun FcmTokenCard(
    token: String?,
    errorMessage: String?,
    onCopy: (String) -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(bottom = 16.dp)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text("FCMテストトークン", style = MaterialTheme.typography.titleMedium)
            when {
                errorMessage != null -> Text(
                    text = errorMessage,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(top = 8.dp)
                )
                token == null -> CircularProgressIndicator(
                    modifier = Modifier
                        .padding(top = 12.dp)
                        .size(24.dp)
                )
                else -> {
                    Text(
                        text = token,
                        style = MaterialTheme.typography.bodySmall,
                        modifier = Modifier.padding(top = 8.dp)
                    )
                    Button(
                        onClick = { onCopy(token) },
                        modifier = Modifier.padding(top = 8.dp)
                    ) {
                        Text("トークンをコピー")
                    }
                }
            }
        }
    }
}

@Composable
private fun TopLinkCard(
    link: com.chat.android.data.model.TopLink,
    onClick: () -> Unit,
    modifier: Modifier = Modifier
) {
    Card(
        modifier = modifier
            .fillMaxWidth()
            .width(200.dp)
            .clickable { onClick() },
        elevation = CardDefaults.cardElevation(defaultElevation = 4.dp),
        colors = CardDefaults.cardColors(
            containerColor = MaterialTheme.colorScheme.primaryContainer
        )
    ) {
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(20.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text(
                text = link.topText,
                style = MaterialTheme.typography.bodyLarge,
                fontWeight = FontWeight.Medium,
                color = MaterialTheme.colorScheme.onPrimaryContainer,
                textAlign = TextAlign.Center
            )
        }
    }
}

/**
 * ホーム画面の先頭に表示するビルド情報。
 * 「修正を反映した APK が実際に入っているか」を端末上で確認できるようにするためのもの。
 *
 * 更新日時は APK のインストール（更新）時刻なので、再ビルドして入れ替えるだけで値が変わる。
 */
@Suppress("DEPRECATION")
@Composable
private fun BuildInfoText(modifier: Modifier = Modifier) {
    val context = LocalContext.current
    val updatedAt = remember {
        runCatching { context.packageManager.getPackageInfo(context.packageName, 0).lastUpdateTime }
            .getOrNull()
            ?.let { SimpleDateFormat("yyyy-MM-dd HH:mm", Locale.getDefault()).format(Date(it)) }
            ?: "不明"
    }

    Column(
        modifier = modifier.fillMaxWidth(),
        horizontalAlignment = Alignment.CenterHorizontally
    ) {
        Text(
            text = "バージョン ${BuildConfig.VERSION_NAME} (build ${BuildConfig.VERSION_CODE}) / ${BuildConfig.BUILD_TYPE}",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
        Text(
            text = "更新日時: $updatedAt / サーバー: ${BuildConfig.BASE_URL}",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant,
            textAlign = TextAlign.Center
        )
    }
}

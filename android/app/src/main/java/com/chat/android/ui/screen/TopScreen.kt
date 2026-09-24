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
import android.content.Context
import androidx.compose.ui.platform.LocalClipboardManager
import androidx.compose.ui.text.AnnotatedString
import com.google.firebase.messaging.FirebaseMessaging

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
                                "/user/" -> navController.navigate("user")
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

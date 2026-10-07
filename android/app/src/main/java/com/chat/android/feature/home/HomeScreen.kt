package com.chat.android.feature.home

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.BuildConfig
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.feature.channel.AliasAvatar
import com.chat.android.navigation.ChannelRoute

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun HomeScreen(
    navController: NavController,
    viewModel: HomeViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("チャネル一覧") }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = {
                navController.navigate(ChannelRoute(id = null))
            }) {
                Icon(Icons.Default.Add, contentDescription = "Add Channel")
            }
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding),
            contentPadding = PaddingValues(bottom = 24.dp)
        ) {
            if (uiState.isLoading) {
                item {
                    Box(
                        modifier = Modifier.fillMaxWidth().padding(16.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        CircularProgressIndicator()
                    }
                }
            } else if (uiState.channels.isEmpty()) {
                item {
                    Text(
                        "チャネルがありません。右下のボタンから作成してください。",
                        modifier = Modifier.padding(16.dp)
                    )
                }
            } else {
                items(uiState.channels) { channel ->
                    ListItem(
                        headlineContent = { Text(channel.channelName) },
                        supportingContent = { Text(channel.channelDescription) },
                        leadingContent = {
                            AliasAvatar(
                                aliasImg = channel.myimg.toAbsoluteImageUrl(),
                                modifier = Modifier.size(40.dp)
                            )
                        },
                        modifier = Modifier.clickable {
                            navController.navigate(ChannelRoute(id = channel.channelID))
                        }
                    )
                    Divider()
                }
            }

            // ビルド情報：端末にインストールされているのがどのバージョンか確認する用
            item { BuildInfoFooter() }
        }
    }
}

/**
 * バージョン名とビルド時刻を表示する。
 *
 * versionName だけだと同じままなので、APKに焼き込んだビルド時刻で
 * 「いま入れているビルド」を判別できるようにしている。
 */
@Composable
private fun BuildInfoFooter() {
    Column(
        modifier = Modifier.fillMaxWidth().padding(16.dp),
        horizontalAlignment = Alignment.CenterHorizontally,
        verticalArrangement = Arrangement.spacedBy(2.dp)
    ) {
        Text(
            "version ${BuildConfig.VERSION_NAME} (${BuildConfig.VERSION_CODE})",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        Text(
            "build ${BuildConfig.BUILD_TIME}",
            style = MaterialTheme.typography.bodySmall,
            fontFamily = FontFamily.Monospace,
            textAlign = TextAlign.Center,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
    }
}

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
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
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
        if (uiState.isLoading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else if (uiState.channels.isEmpty()) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                Text("チャネルがありません。右下のボタンから作成してください。")
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
            ) {
                items(uiState.channels) { channel ->
                    ListItem(
                        headlineContent = { Text(channel.channelName) },
                        supportingContent = { Text(channel.channelDescription) },
                        leadingContent = {
                            AliasAvatar(aliasImg = channel.myimg, modifier = Modifier.size(40.dp))
                        },
                        modifier = Modifier.clickable {
                            // In a real app, this might go to a ChatScreen.
                            // For now, let's go to the Channel Detail/Edit screen.
                            navController.navigate(ChannelRoute(id = channel.channelID))
                        }
                    )
                    Divider()
                }
            }
        }
    }
}

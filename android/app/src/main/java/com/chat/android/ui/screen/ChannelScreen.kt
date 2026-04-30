package com.chat.android.ui.screen

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.ui.viewmodel.ChannelViewModel
import com.chat.android.ui.viewmodel.ChannelUiState

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChannelScreen(
    navController: NavController,
    channelId: String? = null,
    viewModel: ChannelViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val scrollState = rememberScrollState()

    LaunchedEffect(channelId) {
        if (channelId != null) {
            viewModel.loadChannel(channelId)
        } else {
            viewModel.loadChannels()
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(scrollState)
            .padding(16.dp)
    ) {
        // Header
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            IconButton(onClick = { 
                if (channelId != null) navController.navigateUp() 
                else { /* Open Drawer - this requires passing the action */ }
            }) {
                Icon(
                    imageVector = if (channelId != null) Icons.Default.ArrowBack else Icons.Default.Menu,
                    contentDescription = if (channelId != null) "Back" else "Menu"
                )
            }
            Text(
                text = if (channelId != null) "チャネル詳細" else "チャネル一覧",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )
            Row {
                IconButton(onClick = { navController.navigate("user") }) {
                    Icon(
                        imageVector = Icons.Default.AccountCircle,
                        contentDescription = "User Page"
                    )
                }
                if (!uiState.isEditing && channelId != null) {
                    IconButton(onClick = viewModel::startEditing) {
                        Icon(
                            imageVector = Icons.Default.Edit,
                            contentDescription = "Edit"
                        )
                    }
                }
            }
        }

        // Loading indicator
        if (uiState.isLoading) {
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(200.dp),
                contentAlignment = Alignment.Center
            ) {
                CircularProgressIndicator()
            }
        }

        // Error/Success messages
        uiState.errorMessage?.let { error ->
            ErrorMessageCard(
                message = error,
                onDismiss = { viewModel.clearMessages() }
            )
        }

        uiState.successMessage?.let { success ->
            SuccessMessageCard(
                message = success,
                onDismiss = { viewModel.clearMessages() }
            )
        }

        // Channel content
        if (uiState.currentChannel != null) {
            ChannelContent(
                uiState = uiState,
                viewModel = viewModel,
                navController = navController
            )
        } else if (channelId == null) {
            // Channel list view
            ChannelList(
                channels = uiState.channels,
                onChannelClick = { channel ->
                    navController.navigate("channel/${channel.id}")
                }
            )
        }
    }
}

@Composable
private fun ChannelContent(
    uiState: ChannelUiState,
    viewModel: ChannelViewModel,
    navController: NavController
) {
    val channel = uiState.currentChannel!!

    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            if (uiState.isEditing) {
                // Edit mode
                EditChannelForm(
                    uiState = uiState,
                    viewModel = viewModel
                )
            } else {
                // View mode
                ChannelInfo(
                    channel = channel,
                    onJoinClick = {
                        viewModel.joinChannel(channel.id, channel.myname ?: "", channel.myimg)
                    }
                )
            }
        }
    }
}

@Composable
private fun ChannelInfo(
    channel: com.chat.android.network.ChannelDetail,
    onJoinClick: () -> Unit
) {
    Column(
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Channel name
        Text(
            text = channel.name,
            style = MaterialTheme.typography.headlineSmall,
            fontWeight = FontWeight.Bold
        )

        // Channel description
        if (!channel.description.isNullOrEmpty()) {
            Text(
                text = channel.description,
                style = MaterialTheme.typography.bodyMedium
            )
        }

        // User info in channel
        if (!channel.myname.isNullOrEmpty()) {
            Card(
                modifier = Modifier.fillMaxWidth(),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surfaceVariant
                )
            ) {
                Column(
                    modifier = Modifier.padding(16.dp),
                    verticalArrangement = Arrangement.spacedBy(8.dp)
                ) {
                    Text(
                        text = "あなたの情報",
                        style = MaterialTheme.typography.titleSmall,
                        fontWeight = FontWeight.Medium
                    )
                    Text("名前: ${channel.myname}")
                    if (!channel.members.isNullOrEmpty()) {
                        Text("メンバー数: ${channel.members.size}")
                    }
                }
            }
        }

        // Join button (if not joined)
        if (channel.myname.isNullOrEmpty()) {
            Button(
                onClick = onJoinClick,
                modifier = Modifier.fillMaxWidth()
            ) {
                Icon(
                    imageVector = Icons.Default.PersonAdd,
                    contentDescription = null,
                    modifier = Modifier.padding(end = 8.dp)
                )
                Text("チャネルに参加")
            }
        }
    }
}

@Composable
private fun EditChannelForm(
    uiState: ChannelUiState,
    viewModel: ChannelViewModel
) {
    Column(
        verticalArrangement = Arrangement.spacedBy(16.dp)
    ) {
        // Channel name
        OutlinedTextField(
            value = uiState.channelName,
            onValueChange = viewModel::updateChannelName,
            label = { Text("チャネル名") },
            modifier = Modifier.fillMaxWidth()
        )

        // Channel description
        OutlinedTextField(
            value = uiState.channelDescription,
            onValueChange = viewModel::updateChannelDescription,
            label = { Text("説明") },
            modifier = Modifier
                .fillMaxWidth()
                .height(120.dp),
            maxLines = 4
        )

        // User name
        OutlinedTextField(
            value = uiState.myname,
            onValueChange = viewModel::updateMyname,
            label = { Text("あなたの名前") },
            modifier = Modifier.fillMaxWidth()
        )

        // User image
        OutlinedTextField(
            value = uiState.myimg,
            onValueChange = viewModel::updateMyimg,
            label = { Text("画像URL") },
            modifier = Modifier.fillMaxWidth()
        )

        // Action buttons
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            OutlinedButton(
                onClick = viewModel::cancelEditing,
                modifier = Modifier.weight(1f)
            ) {
                Text("キャンセル")
            }
            Button(
                onClick = viewModel::saveChannel,
                modifier = Modifier.weight(1f)
            ) {
                Text("保存")
            }
        }
    }
}

@Composable
private fun ChannelList(
    channels: List<com.chat.android.network.ChannelDetail>,
    onChannelClick: (com.chat.android.network.ChannelDetail) -> Unit
) {
    LazyColumn(
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        items(channels) { channel ->
            ChannelListItem(
                channel = channel,
                onClick = { onChannelClick(channel) }
            )
        }
    }
}

@Composable
private fun ChannelListItem(
    channel: com.chat.android.network.ChannelDetail,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        elevation = CardDefaults.cardElevation(defaultElevation = 2.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                text = channel.name,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            
            if (!channel.description.isNullOrEmpty()) {
                Text(
                    text = channel.description,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant,
                    maxLines = 2
                )
            }
            
            if (!channel.members.isNullOrEmpty()) {
                Text(
                    text = "メンバー: ${channel.members.size}人",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

@Composable
private fun ErrorMessageCard(
    message: String,
    onDismiss: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.errorContainer)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = message,
                modifier = Modifier.weight(1f),
                color = MaterialTheme.colorScheme.onErrorContainer
            )
            IconButton(onClick = onDismiss) {
                Icon(
                    imageVector = Icons.Default.Close,
                    contentDescription = "Dismiss",
                    tint = MaterialTheme.colorScheme.onErrorContainer
                )
            }
        }
    }
}

@Composable
private fun SuccessMessageCard(
    message: String,
    onDismiss: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.primaryContainer)
    ) {
        Row(
            modifier = Modifier.padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = message,
                modifier = Modifier.weight(1f),
                color = MaterialTheme.colorScheme.onPrimaryContainer
            )
            IconButton(onClick = onDismiss) {
                Icon(
                    imageVector = Icons.Default.Close,
                    contentDescription = "Dismiss",
                    tint = MaterialTheme.colorScheme.onPrimaryContainer
                )
            }
        }
    }
}

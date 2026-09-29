package com.chat.android.feature.channel

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Save
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChannelScreen(
    navController: NavController,
    id: String? = null,
    viewModel: ChannelViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    LaunchedEffect(id) {
        viewModel.init(id)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (uiState.isCreateMode) "組織・チャネル登録" else uiState.channelName) },
                actions = {
                    if (!uiState.iamGuest) {
                        IconButton(onClick = { viewModel.save() }) {
                            Icon(Icons.Default.Save, contentDescription = "Save")
                        }
                    }
                    if (uiState.iamAdmin && !uiState.isCreateMode) {
                        IconButton(onClick = { viewModel.deleteChannel() }) {
                            Icon(Icons.Default.Delete, contentDescription = "Delete")
                        }
                    }
                }
            )
        }
    ) { padding ->
        if (uiState.isLoading) {
            Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding)
                    .padding(16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                item {
                    OutlinedTextField(
                        value = uiState.channelName,
                        onValueChange = { viewModel.onNameChange(it) },
                        label = { Text("チャネル名") },
                        modifier = Modifier.fillMaxWidth(),
                        readOnly = uiState.iamGuest
                    )
                }

                item {
                    OutlinedTextField(
                        value = uiState.myName,
                        onValueChange = { viewModel.onMyNameChange(it) },
                        label = { Text("表示名 (ニックネーム)") },
                        modifier = Modifier.fillMaxWidth(),
                        readOnly = !uiState.isCreateMode // Vue logic: often set at create
                    )
                }

                item {
                    OutlinedTextField(
                        value = uiState.channelDescription,
                        onValueChange = { viewModel.onDescriptionChange(it) },
                        label = { Text("説明") },
                        modifier = Modifier.fillMaxWidth().height(150.dp),
                        readOnly = uiState.iamGuest,
                        maxLines = 10
                    )
                }

                if (!uiState.isCreateMode) {
                    item {
                        Text("招待コード (一般): ${uiState.invitationCode}", style = MaterialTheme.typography.bodyMedium)
                        Text("招待コード (ゲスト): ${uiState.invitationGuestCode}", style = MaterialTheme.typography.bodyMedium)
                    }

                    item {
                        Text("メンバー", style = MaterialTheme.typography.titleMedium)
                    }

                    items(uiState.aliases) { alias ->
                        ListItem(
                            headlineContent = { Text(alias.aliasName) },
                            supportingContent = { Text(alias.accessRight) },
                            leadingContent = {
                                AliasAvatar(aliasImg = alias.aliasImg, modifier = Modifier.size(40.dp))
                            }
                        )
                    }

                    item {
                        Text("グループ", style = MaterialTheme.typography.titleMedium)
                    }

                    items(uiState.groups) { group ->
                        ListItem(
                            headlineContent = { Text(group.groupName) },
                            leadingContent = {
                                AliasAvatar(aliasImg = group.groupImg, modifier = Modifier.size(40.dp))
                            }
                        )
                    }
                }

                uiState.error?.let {
                    item {
                        Text(it, color = MaterialTheme.colorScheme.error)
                    }
                }
                
                uiState.successMessage?.let {
                    item {
                        Text(it, color = MaterialTheme.colorScheme.primary)
                    }
                }
            }
        }
    }
}

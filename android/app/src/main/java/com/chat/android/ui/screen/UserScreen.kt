package com.chat.android.ui.screen

import androidx.compose.foundation.*
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import coil.compose.AsyncImage
import coil.request.ImageRequest
import com.chat.android.ui.viewmodel.UserViewModel
import com.chat.android.ui.viewmodel.UserUiState

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun UserScreen(
    navController: NavController,
    viewModel: UserViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val scrollState = rememberScrollState()

    Column(
        modifier = Modifier
            .fillMaxSize()
            .verticalScroll(scrollState)
            .padding(16.dp)
    ) {
        // Header
        Text(
            text = "ユーザーページ",
            style = MaterialTheme.typography.headlineMedium,
            fontWeight = FontWeight.Bold,
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 16.dp),
            textAlign = TextAlign.Center
        )

        // TO invitation link
        if (uiState.isTO && uiState.toLink != null) {
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 16.dp),
                colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.tertiaryContainer)
            ) {
                Column(
                    modifier = Modifier.padding(16.dp)
                ) {
                    TextButton(
                        onClick = { /* Handle TO link navigation */ }
                    ) {
                        Text("招待参加ページ")
                    }
                    Text(
                        text = "必須項目を登録してから参加ページにお願いします",
                        style = MaterialTheme.typography.bodySmall,
                        color = MaterialTheme.colorScheme.onTertiaryContainer
                    )
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

        // User form (only show when fetched)
        if (uiState.fetched && uiState.user != null) {
            UserForm(
                uiState = uiState,
                viewModel = viewModel,
                navController = navController
            )
        }

        // Nickname selection
        if (uiState.nicknames.isNotEmpty()) {
            NicknameSelection(
                nicknames = uiState.nicknames,
                currentNickname = uiState.currentNickname,
                onNicknameSelected = viewModel::switchNickname
            )
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

@Composable
private fun UserForm(
    uiState: UserUiState,
    viewModel: UserViewModel,
    navController: NavController
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 8.dp)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            // Coordinates section
            CoordinateSection(
                coordinateInput = uiState.coordinateInput,
                onCoordinateChange = viewModel::updateCoordinateInput
            )

            // Optional fields section
            OptionalFieldsSection(
                user = uiState.user,
                onUserUpdate = { updatedUser ->
                    // This would update the user in the ViewModel
                }
            )

            // Nickname management
            NicknameManagementSection(
                uiState = uiState,
                viewModel = viewModel
            )

            // Wallet address
            OutlinedTextField(
                value = uiState.user?.walletAddress ?: "",
                onValueChange = { /* Update wallet address */ },
                label = { Text("JPYCアドレス") },
                modifier = Modifier.fillMaxWidth()
            )

            // Update button
            Button(
                onClick = viewModel::submitUser,
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(top = 16.dp)
            ) {
                Text("更新")
            }
        }
    }
}

@Composable
private fun CoordinateSection(
    coordinateInput: String,
    onCoordinateChange: (String) -> Unit
) {
    Column {
        Text(
            text = "経緯度:",
            style = MaterialTheme.typography.bodyLarge,
            fontWeight = FontWeight.Medium
        )
        
        Row(
            modifier = Modifier.fillMaxWidth(),
            horizontalArrangement = Arrangement.spacedBy(8.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            OutlinedTextField(
                value = coordinateInput,
                onValueChange = onCoordinateChange,
                placeholder = { Text("35.72300346964341, 139.52507136879356") },
                modifier = Modifier.weight(1f)
            )
            
            TextButton(onClick = { /* Get geolocation */ }) {
                Text("GEOデータ再取得")
            }
        }
        
        Text(
            text = "Googleマップの右クリックで取得できます",
            style = MaterialTheme.typography.bodySmall,
            color = MaterialTheme.colorScheme.onSurfaceVariant
        )
        
        // Show parsed coordinates
        val parts = coordinateInput.split(",").map { it.trim() }
        if (parts.size == 2) {
            val coords = try {
                val lat = parts[0].toDouble()
                val lng = parts[1].toDouble()
                lat to lng
            } catch (e: NumberFormatException) {
                null
            }

            if (coords != null) {
                Column(
                    modifier = Modifier.padding(top = 8.dp)
                ) {
                    Text("➤ 緯度: ${coords.first}", fontWeight = FontWeight.Medium)
                    Text("➤ 経度: ${coords.second}", fontWeight = FontWeight.Medium)
                }
            }
        }
    }
}

@Composable
private fun OptionalFieldsSection(
    user: com.chat.android.network.UserResponse?,
    onUserUpdate: (com.chat.android.network.UserResponse) -> Unit
) {
    Column {
        Text(
            text = "ーーー オプション ーーー",
            modifier = Modifier
                .fillMaxWidth()
                .padding(vertical = 8.dp),
            textAlign = TextAlign.Center,
            style = MaterialTheme.typography.bodyMedium
        )
        
        OutlinedTextField(
            value = user?.mail ?: "",
            onValueChange = { onUserUpdate(user?.copy(mail = it) ?: com.chat.android.network.UserResponse(mail = it)) },
            label = { Text("メール") },
            modifier = Modifier.fillMaxWidth()
        )
        
        OutlinedTextField(
            value = user?.telephone ?: "",
            onValueChange = { onUserUpdate(user?.copy(telephone = it) ?: com.chat.android.network.UserResponse(telephone = it)) },
            label = { Text("電話番号") },
            modifier = Modifier.fillMaxWidth()
        )
    }
}

@Composable
private fun NicknameManagementSection(
    uiState: UserUiState,
    viewModel: UserViewModel
) {
    if (!uiState.isNicknameFormVisible) {
        // Show nickname management buttons
        Column(
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            if (uiState.nicknames.isNotEmpty()) {
                Button(
                    onClick = { viewModel.openNicknameForm("edit") },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text("ニックネームの絵文字、画像を変更")
                }
            }
            
            Button(
                onClick = { viewModel.openNicknameForm("new") },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("新しいニックネームを作成")
            }
        }
    } else {
        // Show nickname form
        NicknameForm(
            uiState = uiState,
            viewModel = viewModel
        )
    }
}

@Composable
private fun NicknameForm(
    uiState: UserUiState,
    viewModel: UserViewModel
) {
    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = CardDefaults.cardColors(containerColor = MaterialTheme.colorScheme.surfaceVariant)
    ) {
        Column(
            modifier = Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            Text(
                text = if (uiState.editingMode == "edit") "ニックネームの編集" else "新しいニックネームを作成",
                style = MaterialTheme.typography.titleMedium
            )
            
            Text(
                text = "ニックネームは登録後は変更できません\n３つまで登録できます",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            
            OutlinedTextField(
                value = uiState.editableNickname,
                onValueChange = viewModel::updateEditableNickname,
                label = { Text("ニックネーム") },
                enabled = uiState.editingMode != "edit",
                modifier = Modifier.fillMaxWidth()
            )
            
            // Image selection (placeholder)
            OutlinedTextField(
                value = uiState.currentNickImg ?: "",
                onValueChange = viewModel::updateNickImg,
                label = { Text("画像URL") },
                modifier = Modifier.fillMaxWidth()
            )
            
            // Bio editor (simplified)
            OutlinedTextField(
                value = uiState.currentNickBio ?: "",
                onValueChange = viewModel::updateNickBio,
                label = { Text("自己紹介") },
                modifier = Modifier
                    .fillMaxWidth()
                    .height(120.dp),
                maxLines = 5
            )
            
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                OutlinedButton(
                    onClick = viewModel::closeNicknameForm,
                    modifier = Modifier.weight(1f)
                ) {
                    Text("キャンセル")
                }
            }
        }
    }
}

@Composable
private fun NicknameSelection(
    nicknames: List<com.chat.android.network.NicknameResponse>,
    currentNickname: String?,
    onNicknameSelected: (String) -> Unit
) {
    Column(
        modifier = Modifier.padding(vertical = 16.dp)
    ) {
        Text(
            text = "ニックネーム選択",
            style = MaterialTheme.typography.titleMedium,
            modifier = Modifier.padding(bottom = 8.dp)
        )
        
        LazyColumn(
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            items(nicknames) { nickname ->
                NicknameItem(
                    nickname = nickname,
                    isSelected = nickname.nickname == currentNickname,
                    onClick = { onNicknameSelected(nickname.nickname) }
                )
            }
        }
    }
}

@Composable
private fun NicknameItem(
    nickname: com.chat.android.network.NicknameResponse,
    isSelected: Boolean,
    onClick: () -> Unit
) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .clickable { onClick() },
        colors = CardDefaults.cardColors(
            containerColor = if (isSelected) 
                MaterialTheme.colorScheme.primaryContainer 
            else 
                MaterialTheme.colorScheme.surface
        )
    ) {
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(12.dp),
            verticalAlignment = Alignment.CenterVertically,
            horizontalArrangement = Arrangement.spacedBy(12.dp)
        ) {
            // Selection indicator
            Text(
                text = if (isSelected) "✅" else "⬜",
                modifier = Modifier.width(30.dp)
            )
            
            // Avatar
            if (!nickname.nickImg.isNullOrEmpty()) {
                if (nickname.nickImg.startsWith(",")) {
                    // Emoji avatar
                    val parts = nickname.nickImg.split(",")
                    if (parts.size >= 3) {
                        val emoji = parts[1]
                        val color = parts[2]
                        Box(
                            modifier = Modifier
                                .size(26.dp)
                                .clip(CircleShape)
                                .background(Color(android.graphics.Color.parseColor(color))),
                            contentAlignment = Alignment.Center
                        ) {
                            Text(
                                text = emoji,
                                fontSize = MaterialTheme.typography.bodyMedium.fontSize
                            )
                        }
                    }
                } else {
                    // Image avatar
                    AsyncImage(
                        model = ImageRequest.Builder(LocalContext.current)
                            .data(nickname.nickImg)
                            .crossfade(true)
                            .build(),
                        contentDescription = "Avatar",
                        modifier = Modifier
                            .size(26.dp)
                            .clip(CircleShape)
                    )
                }
            } else {
                // Default avatar
                Box(
                    modifier = Modifier
                        .size(26.dp)
                        .clip(CircleShape)
                        .background(MaterialTheme.colorScheme.primary),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = nickname.nickname.firstOrNull()?.uppercase() ?: "?",
                        color = MaterialTheme.colorScheme.onPrimary,
                        fontSize = MaterialTheme.typography.bodySmall.fontSize
                    )
                }
            }
            
            // Nickname
            Text(
                text = nickname.nickname,
                style = MaterialTheme.typography.bodyLarge,
                fontWeight = if (isSelected) FontWeight.Bold else FontWeight.Normal,
                modifier = Modifier.weight(1f)
            )
        }
    }
}

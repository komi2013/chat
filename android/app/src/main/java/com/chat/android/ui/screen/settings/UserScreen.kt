package com.chat.android.ui.screen.settings

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.ui.viewmodel.UserViewModel
import com.chat.android.ui.viewmodel.UserUiState
import com.chat.android.ui.viewmodel.toUserData
import com.chat.android.network.UserResponse
import com.chat.android.network.NicknameResponse

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun UserScreen(
    navController: NavController,
    viewModel: UserViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    
    LaunchedEffect(Unit) {
        // Data is already loaded in UserViewModel init
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
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
            IconButton(onClick = { navController.navigateUp() }) {
                Icon(
                    imageVector = Icons.Default.ArrowBack,
                    contentDescription = "Back"
                )
            }
            Text(
                text = "ユーザー設定",
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.width(48.dp))
        }

        when {
            uiState.isLoading -> {
                Box(
                    modifier = Modifier.fillMaxSize(),
                    contentAlignment = Alignment.Center
                ) {
                    CircularProgressIndicator()
                }
            }
            
            uiState.user != null -> {
                val user = uiState.user!!
                val userData = user.toUserData()
                LazyColumn(
                    verticalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    // Profile Section
                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Column(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(16.dp)
                            ) {
                                Text(
                                    text = "プロフィール",
                                    style = MaterialTheme.typography.titleLarge,
                                    fontWeight = FontWeight.Bold
                                )
                                Spacer(modifier = Modifier.height(12.dp))
                                
                                userData.name?.let { name ->
                                    ProfileInfoRow(Icons.Default.Person, "名前", name)
                                }
                                
                                userData.email?.let { email ->
                                    ProfileInfoRow(Icons.Default.Email, "メール", email)
                                }
                                
                                userData.nickname?.let { nickname ->
                                    ProfileInfoRow(Icons.Default.AccountCircle, "ニックネーム", nickname)
                                }
                                
                                userData.telephone?.let { telephone ->
                                    ProfileInfoRow(Icons.Default.Phone, "電話番号", telephone)
                                }
                            }
                        }
                    }
                    
                    // Account Settings
                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Column(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(16.dp)
                            ) {
                                Text(
                                    text = "アカウント設定",
                                    style = MaterialTheme.typography.titleLarge,
                                    fontWeight = FontWeight.Bold
                                )
                                Spacer(modifier = Modifier.height(12.dp))
                                
                                userData.channelId?.let { channelId ->
                                    ProfileInfoRow(Icons.Default.Business, "チャネルID", channelId)
                                }
                                
                                userData.accessRight?.let { accessRight ->
                                    ProfileInfoRow(Icons.Default.Security, "アクセス権", accessRight)
                                }
                                
                                userData.isAdmin?.let { isAdmin ->
                                    ProfileInfoRow(Icons.Default.AdminPanelSettings, "管理者", if (isAdmin) "はい" else "いいえ")
                                }
                                
                                userData.walletAddress?.let { wallet ->
                                    ProfileInfoRow(Icons.Default.AccountBalanceWallet, "ウォレットアドレス", wallet.take(12) + "...")
                                }
                            }
                        }
                    }
                    
                    // Location
                    if (userData.latitude != null && userData.longitude != null) {
                        item {
                            Card(
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Column(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(16.dp)
                                ) {
                                    Text(
                                        text = "位置情報",
                                        style = MaterialTheme.typography.titleLarge,
                                        fontWeight = FontWeight.Bold
                                    )
                                    Spacer(modifier = Modifier.height(12.dp))
                                    
                                    ProfileInfoRow(Icons.Default.LocationOn, "緯度", userData.latitude.toString())
                                    ProfileInfoRow(Icons.Default.LocationOn, "経度", userData.longitude.toString())
                                }
                            }
                        }
                    }
                    
                    // Nicknames
                    if (uiState.nicknames.isNotEmpty()) {
                        item {
                            Card(
                                modifier = Modifier.fillMaxWidth()
                            ) {
                                Column(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .padding(16.dp)
                                ) {
                                    Text(
                                        text = "ニックネーム一覧",
                                        style = MaterialTheme.typography.titleLarge,
                                        fontWeight = FontWeight.Bold
                                    )
                                    Spacer(modifier = Modifier.height(12.dp))
                                    
                                    uiState.nicknames.forEach { nickname ->
                                        NicknameInfoRow(nickname)
                                        Spacer(modifier = Modifier.height(4.dp))
                                    }
                                }
                            }
                        }
                    }
                    
                    // Actions
                    item {
                        Card(
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Column(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .padding(16.dp)
                            ) {
                                Text(
                                    text = "操作",
                                    style = MaterialTheme.typography.titleLarge,
                                    fontWeight = FontWeight.Bold
                                )
                                Spacer(modifier = Modifier.height(12.dp))
                                
                                Button(
                                    onClick = { viewModel.loadUserData() },
                                    modifier = Modifier.fillMaxWidth()
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.Refresh,
                                        contentDescription = "Refresh"
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text("データを更新")
                                }
                                
                                Spacer(modifier = Modifier.height(8.dp))
                                
                                Button(
                                    onClick = { viewModel.openNicknameForm("new") },
                                    modifier = Modifier.fillMaxWidth()
                                ) {
                                    Icon(
                                        imageVector = Icons.Default.Add,
                                        contentDescription = "Add Nickname"
                                    )
                                    Spacer(modifier = Modifier.width(8.dp))
                                    Text("ニックネームを追加")
                                }
                            }
                        }
                    }
                }
            }
            
            uiState.errorMessage != null -> {
                Card(
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Icon(
                            imageVector = Icons.Default.Error,
                            contentDescription = "Error",
                            tint = MaterialTheme.colorScheme.error,
                            modifier = Modifier.size(48.dp)
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Text(
                            text = "エラーが発生しました",
                            style = MaterialTheme.typography.titleLarge,
                            color = MaterialTheme.colorScheme.error
                        )
                        Spacer(modifier = Modifier.height(8.dp))
                        Text(
                            text = uiState.errorMessage ?: "",
                            style = MaterialTheme.typography.bodyMedium,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Button(
                            onClick = { 
                                viewModel.clearMessages()
                                viewModel.loadUserData()
                            }
                        ) {
                            Icon(
                                imageVector = Icons.Default.Refresh,
                                contentDescription = "Retry"
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("再試行")
                        }
                    }
                }
            }
            
            !uiState.fetched -> {
                Card(
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Column(
                        modifier = Modifier
                            .fillMaxWidth()
                            .padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(
                            text = "データを読み込んでいません",
                            style = MaterialTheme.typography.titleLarge
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        Button(
                            onClick = { viewModel.loadUserData() }
                        ) {
                            Icon(
                                imageVector = Icons.Default.Refresh,
                                contentDescription = "Load Data"
                            )
                            Spacer(modifier = Modifier.width(8.dp))
                            Text("データを読み込む")
                        }
                    }
                }
            }
        }
    }
}

@Composable
private fun ProfileInfoRow(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    value: String
) {
    UserInfoRow(icon, label, value)
}

@Composable
private fun NicknameInfoRow(nickname: NicknameResponse) {
    UserInfoRow(
        icon = Icons.Default.Person,
        label = nickname.nickname,
        value = nickname.accessRight ?: "一般"
    )
}

@Composable
private fun UserInfoRow(
    icon: androidx.compose.ui.graphics.vector.ImageVector,
    label: String,
    value: String
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Icon(
            imageVector = icon,
            contentDescription = null,
            tint = MaterialTheme.colorScheme.primary,
            modifier = Modifier.size(20.dp)
        )
        Spacer(modifier = Modifier.width(12.dp))
        Column(
            modifier = Modifier.weight(1f)
        ) {
            Text(
                text = label,
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
            Text(
                text = value,
                style = MaterialTheme.typography.bodyMedium,
                fontWeight = FontWeight.Medium
            )
        }
    }
}

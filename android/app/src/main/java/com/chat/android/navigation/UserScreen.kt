package com.chat.android.navigation

import androidx.compose.foundation.BorderStroke
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.text.style.TextOverflow
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import coil.compose.AsyncImage
import com.chat.android.BuildConfig
import com.chat.android.auth.UserUiState
import com.chat.android.auth.UserViewModel
import com.chat.android.network.NicknameResponse

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun UserScreen(
    navController: NavController,
    viewModel: UserViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("ユーザー設定") },
                navigationIcon = {
                    IconButton(onClick = { navController.popBackStack() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Box(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
        ) {
            when (val state = uiState) {
                is UserUiState.Loading -> {
                    CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                }
                is UserUiState.Error -> {
                    Column(
                        modifier = Modifier.align(Alignment.Center).padding(16.dp),
                        horizontalAlignment = Alignment.CenterHorizontally
                    ) {
                        Text(state.message, color = MaterialTheme.colorScheme.error)
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(onClick = { viewModel.fetchUser() }) {
                            Text("再試行")
                        }
                    }
                }
                is UserUiState.Success -> {
                    UserContent(state.data, viewModel)
                }
            }
        }
    }
}

@Composable
fun UserContent(data: com.chat.android.network.GoogleSignInResponse, viewModel: UserViewModel) {
    var mail by remember(data) { mutableStateOf(data.user?.mail ?: "") }
    var telephone by remember(data) { mutableStateOf(data.user?.telephone ?: "") }
    var walletAddress by remember(data) { mutableStateOf(data.user?.walletAddress ?: "") }
    var latitude by remember(data) { mutableStateOf(data.user?.latitude?.toString() ?: "") }
    var longitude by remember(data) { mutableStateOf(data.user?.longitude?.toString() ?: "") }
    
    val nicknames = data.nicknames ?: emptyList()
    var selectedNickname by remember(data) { mutableStateOf(data.nickname ?: "") }

    LazyColumn(
        modifier = Modifier.fillMaxSize().padding(horizontal = 16.dp),
        verticalArrangement = Arrangement.spacedBy(16.dp),
        contentPadding = PaddingValues(bottom = 32.dp)
    ) {
        item {
            Text("位置情報", style = MaterialTheme.typography.titleMedium)
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                OutlinedTextField(
                    value = latitude,
                    onValueChange = { latitude = it },
                    label = { Text("緯度") },
                    modifier = Modifier.weight(1f)
                )
                OutlinedTextField(
                    value = longitude,
                    onValueChange = { longitude = it },
                    label = { Text("経度") },
                    modifier = Modifier.weight(1f)
                )
            }
        }

        item {
            Text("基本情報", style = MaterialTheme.typography.titleMedium)
            OutlinedTextField(
                value = mail,
                onValueChange = { mail = it },
                label = { Text("メールアドレス") },
                modifier = Modifier.fillMaxWidth()
            )
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedTextField(
                value = telephone,
                onValueChange = { telephone = it },
                label = { Text("電話番号") },
                modifier = Modifier.fillMaxWidth()
            )
            Spacer(modifier = Modifier.height(8.dp))
            OutlinedTextField(
                value = walletAddress,
                onValueChange = { walletAddress = it },
                label = { Text("JPYCウォレットアドレス") },
                modifier = Modifier.fillMaxWidth()
            )
        }

        item {
            Text("ニックネームの選択", style = MaterialTheme.typography.titleMedium)
        }

        items(nicknames) { nick ->
            NicknameItem(nick, isSelected = nick.nickname == selectedNickname) {
                selectedNickname = nick.nickname
            }
        }

        item {
            Button(
                onClick = {
                    val currentNick = nicknames.find { it.nickname == selectedNickname }
                    viewModel.updateUser(
                        nickname = selectedNickname,
                        nickImg = currentNick?.nickImg,
                        nickBio = currentNick?.nickBio,
                        mail = mail,
                        telephone = telephone,
                        walletAddress = walletAddress,
                        latitude = latitude.toDoubleOrNull(),
                        longitude = longitude.toDoubleOrNull()
                    )
                },
                modifier = Modifier.fillMaxWidth().padding(top = 8.dp)
            ) {
                Text("プロフィールを更新")
            }
        }
    }
}

@Composable
fun NicknameItem(nick: NicknameResponse, isSelected: Boolean, onSelect: () -> Unit) {
    Surface(
        onClick = onSelect,
        shape = RoundedCornerShape(12.dp),
        color = if (isSelected) MaterialTheme.colorScheme.primaryContainer else MaterialTheme.colorScheme.surface,
        border = if (isSelected) BorderStroke(1.dp, MaterialTheme.colorScheme.primary) else BorderStroke(1.dp, MaterialTheme.colorScheme.outlineVariant),
        modifier = Modifier.fillMaxWidth()
    ) {
        Row(
            modifier = Modifier.padding(12.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            NicknameImage(nick.nickImg)
            
            Spacer(modifier = Modifier.width(16.dp))
            
            Column(modifier = Modifier.weight(1f)) {
                Text(nick.nickname, style = MaterialTheme.typography.titleMedium)
                nick.nickBio?.let {
                    if (it.isNotEmpty()) {
                        Text(
                            it, 
                            style = MaterialTheme.typography.bodySmall, 
                            color = MaterialTheme.colorScheme.onSurfaceVariant,
                            maxLines = 2,
                            overflow = TextOverflow.Ellipsis
                        )
                    }
                }
            }
            
            if (isSelected) {
                Icon(
                    Icons.Default.Check, 
                    contentDescription = "選択済み", 
                    tint = MaterialTheme.colorScheme.primary
                )
            }
        }
    }
}

@Composable
fun NicknameImage(nickImg: String?) {
    if (nickImg != null) {
        if (nickImg.startsWith(",")) {
            val parts = nickImg.split(",")
            if (parts.size >= 3) {
                val emoji = parts[1]
                val bgColor = try { Color(android.graphics.Color.parseColor(parts[2])) } catch (e: Exception) { Color.Gray }
                Box(
                    modifier = Modifier
                        .size(48.dp)
                        .clip(RoundedCornerShape(8.dp))
                        .background(bgColor),
                    contentAlignment = Alignment.Center
                ) {
                    Text(emoji, fontSize = 24.sp)
                }
                return
            }
        }
        
        val fullImgUrl = if (nickImg.startsWith("/")) {
            "${BuildConfig.BASE_URL.removeSuffix("/")}$nickImg"
        } else {
            nickImg
        }
        
        AsyncImage(
            model = fullImgUrl,
            contentDescription = null,
            modifier = Modifier
                .size(48.dp)
                .clip(RoundedCornerShape(8.dp)),
            contentScale = ContentScale.Crop
        )
    } else {
        Box(
            modifier = Modifier
                .size(48.dp)
                .clip(RoundedCornerShape(8.dp))
                .background(MaterialTheme.colorScheme.surfaceVariant),
            contentAlignment = Alignment.Center
        ) {
            Text("?", style = MaterialTheme.typography.titleMedium)
        }
    }
}

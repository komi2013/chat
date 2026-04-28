package com.chat.android.navigation

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
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
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import coil.compose.AsyncImage
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
                title = { Text("ユーザーページ") },
                navigationIcon = {
                    IconButton(onClick = { navController.popBackStack() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp)
        ) {
            when (val state = uiState) {
                is UserUiState.Loading -> {
                    Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
                        CircularProgressIndicator()
                    }
                }
                is UserUiState.Error -> {
                    Text(state.message, color = MaterialTheme.colorScheme.error)
                    Button(onClick = { viewModel.fetchUser() }) {
                        Text("Retry")
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
    var mail by remember { mutableStateOf(data.user?.mail ?: "") }
    var telephone by remember { mutableStateOf(data.user?.telephone ?: "") }
    var walletAddress by remember { mutableStateOf(data.user?.walletAddress ?: "") }
    var latitude by remember { mutableStateOf(data.user?.latitude?.toString() ?: "") }
    var longitude by remember { mutableStateOf(data.user?.longitude?.toString() ?: "") }
    
    // We'll just show the nicknames for now, editing them is more involved
    val nicknames = data.nicknames ?: emptyList()
    var selectedNickname by remember { mutableStateOf(data.nickname ?: "") }

    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.spacedBy(12.dp)
    ) {
        item {
            Text("経緯度", style = MaterialTheme.typography.titleMedium)
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
            Text("オプション", style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(top = 8.dp))
            OutlinedTextField(
                value = mail,
                onValueChange = { mail = it },
                label = { Text("メール") },
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
                label = { Text("JPYCアドレス") },
                modifier = Modifier.fillMaxWidth()
            )
        }

        item {
            Text("ニックネーム一覧", style = MaterialTheme.typography.titleMedium, modifier = Modifier.padding(top = 8.dp))
        }

        items(nicknames) { nick ->
            NicknameItem(nick, isSelected = nick.nickname == selectedNickname) {
                selectedNickname = nick.nickname
            }
        }

        item {
            Button(
                onClick = {
                    viewModel.updateUser(
                        nickname = selectedNickname,
                        nickImg = nicknames.find { it.nickname == selectedNickname }?.nickImg,
                        nickBio = nicknames.find { it.nickname == selectedNickname }?.nickBio,
                        mail = mail,
                        telephone = telephone,
                        walletAddress = walletAddress,
                        latitude = latitude.toDoubleOrNull(),
                        longitude = longitude.toDoubleOrNull()
                    )
                },
                modifier = Modifier.fillMaxWidth().padding(vertical = 16.dp)
            ) {
                Text("更新")
            }
        }
    }
}

@Composable
fun NicknameItem(nick: NicknameResponse, isSelected: Boolean, onSelect: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clip(RoundedCornerShape(8.dp))
            .background(if (isSelected) MaterialTheme.colorScheme.primaryContainer else Color.Transparent)
            .clickable { onSelect() }
            .padding(8.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        if (isSelected) {
            Icon(Icons.Default.Check, contentDescription = null, tint = MaterialTheme.colorScheme.primary)
        } else {
            Spacer(modifier = Modifier.size(24.dp))
        }
        
        Spacer(modifier = Modifier.width(8.dp))
        
        // Handle Emoji vs Image
        val nickImg = nick.nickImg
        if (nickImg != null) {
            if (nickImg.startsWith(",")) {
                val parts = nickImg.split(",")
                if (parts.size >= 3) {
                    val emoji = parts[1]
                    val bgColor = try { Color(android.graphics.Color.parseColor(parts[2])) } catch (e: Exception) { Color.Gray }
                    Box(
                        modifier = Modifier
                            .size(32.dp)
                            .clip(RoundedCornerShape(4.dp))
                            .background(bgColor),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(emoji, fontSize = 18.sp)
                    }
                }
            } else {
                AsyncImage(
                    model = nickImg,
                    contentDescription = null,
                    modifier = Modifier
                        .size(32.dp)
                        .clip(RoundedCornerShape(4.dp)),
                    contentScale = ContentScale.Crop
                )
            }
        } else {
            Box(modifier = Modifier.size(32.dp).background(Color.LightGray))
        }
        
        Spacer(modifier = Modifier.width(12.dp))
        Text(nick.nickname, style = MaterialTheme.typography.bodyLarge)
    }
}

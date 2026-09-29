package com.chat.android.core.ui.component

import androidx.compose.foundation.background
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.vector.ImageVector
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.core.repository.UserRepository
import com.chat.android.core.network.ChannelDetail
import com.chat.android.core.network.NicknameResponse
import com.chat.android.navigation.UserRoute
import com.chat.android.navigation.EntryFormListRoute
import com.chat.android.navigation.EntryFormEditRoute
import com.chat.android.navigation.ChannelRoute
import javax.inject.Inject

data class DrawerUiState(
    val isSignedIn: Boolean = false,
    val hasChannel: Boolean = false,
    val isGuest: Boolean = false
)

class DrawerViewModel @Inject constructor(
    private val userRepository: UserRepository
) : androidx.lifecycle.ViewModel() {
    private val _uiState = mutableStateOf(DrawerUiState())
    val uiState: State<DrawerUiState> = _uiState

    init {
        _uiState.value = DrawerUiState(
            isSignedIn = userRepository.isSignedIn(),
            hasChannel = userRepository.getCurrentChannelId() != null
        )
    }

    fun updateDrawerState(channel: ChannelDetail?, aliases: List<NicknameResponse>) {
        _uiState.value = _uiState.value.copy(
            isSignedIn = userRepository.isSignedIn(),
            hasChannel = channel != null || userRepository.getCurrentChannelId() != null
        )
    }
}

data class DrawerItem(
    val title: String,
    val route: String,
    val icon: ImageVector,
    val requiresAuth: Boolean = false,
    val requiresChannel: Boolean = false,
    val requiresNonGuest: Boolean = false
)

@Composable
fun AppDrawer(
    navController: NavController,
    channel: ChannelDetail? = null,
    aliases: List<NicknameResponse> = emptyList(),
    viewModel: DrawerViewModel = DrawerViewModel(hiltViewModel<com.chat.android.feature.user.UserViewModel>().userRepository),
    onClose: () -> Unit = {}
) {
    val uiState = viewModel.uiState.value
    
    LaunchedEffect(channel, aliases) {
        viewModel.updateDrawerState(channel, aliases)
    }

    ModalDrawerSheet(
        modifier = Modifier.width(280.dp)
    ) {
        // Header
        DrawerHeader(
            isSignedIn = uiState.isSignedIn,
            onClose = onClose
        )
        
        Divider()
        
        // Navigation items
        LazyColumn(
            modifier = Modifier.fillMaxWidth()
        ) {
            items(getDrawerItems(uiState)) { item ->
                DrawerItemRow(
                    item = item,
                    onClick = {
                        when (item.route) {
                            "user" -> navController.navigate(UserRoute)
                            "entryFormEdit" -> navController.navigate(EntryFormListRoute)
                            "channelCreate" -> navController.navigate(ChannelRoute(id = null))
                            else -> navController.navigate(item.route)
                        }
                        onClose()
                    }

                )
            }
        }
        
        // Advertisement section (like in Vue)
        if (uiState.isSignedIn) {
            Divider()
            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(
                    text = "Advertisement",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
            }
        }
    }
}

@Composable
private fun DrawerHeader(
    isSignedIn: Boolean,
    onClose: () -> Unit
) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        horizontalArrangement = Arrangement.SpaceBetween,
        verticalAlignment = Alignment.CenterVertically
    ) {
        Text(
            text = "Menu",
            style = MaterialTheme.typography.titleLarge,
            fontWeight = FontWeight.Bold
        )
        
        IconButton(onClick = onClose) {
            Icon(
                imageVector = Icons.Default.Close,
                contentDescription = "Close drawer"
            )
        }
    }
    
    if (isSignedIn) {
        Text(
            text = "Signed In",
            style = MaterialTheme.typography.bodyMedium,
            color = MaterialTheme.colorScheme.primary,
            modifier = Modifier.padding(start = 16.dp, bottom = 8.dp)
        )
    }
}

@Composable
private fun DrawerItemRow(
    item: DrawerItem,
    onClick: () -> Unit
) {
    NavigationDrawerItem(
        icon = {
            Icon(
                imageVector = item.icon,
                contentDescription = item.title
            )
        },
        label = {
            Text(
                text = item.title,
                fontWeight = FontWeight.Medium
            )
        },
        selected = false,
        onClick = onClick,
        modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)
    )
}

private fun getDrawerItems(uiState: DrawerUiState): List<DrawerItem> {
    val items = mutableListOf<DrawerItem>()
    
    // Authentication-dependent items
    if (uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "ユーザー設定", 
            route = "user",
            icon = Icons.Default.Person,
            requiresAuth = true
        ))
    }
    
    // Channel-dependent items (In this version, we always show entry form if signed in for simplicity, 
    // but following the original logic where it was under non-guest check)
    if (uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "チャネル登録",
            route = "channelCreate",
            icon = Icons.Default.AddBusiness,
            requiresAuth = true
        ))

        items.add(DrawerItem(
            title = "フォーム編集",
            route = "entryFormEdit",
            icon = Icons.Default.EditNote,
            requiresAuth = true
        ))
    }
    
    if (!uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "サインイン",
            route = "login",
            icon = Icons.Default.Login
        ))
    }
    
    return items
}

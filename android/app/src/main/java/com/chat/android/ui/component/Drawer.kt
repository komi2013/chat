package com.chat.android.ui.component

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
import com.chat.android.data.repository.UserRepository
import com.chat.android.network.ChannelDetail
import com.chat.android.network.NicknameResponse
import com.chat.android.ui.viewmodel.DrawerViewModel
import com.chat.android.ui.viewmodel.DrawerUiState
import javax.inject.Inject

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
    viewModel: DrawerViewModel = hiltViewModel(),
    onClose: () -> Unit = {}
) {
    val uiState by viewModel.uiState.collectAsState()
    
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
                        navController.navigate(item.route)
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
    
    // Always visible items
    items.add(DrawerItem(
        title = "Home",
        route = "top",
        icon = Icons.Default.Home
    ))
    
    // Authentication-dependent items
    if (uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "Channel Settings",
            route = "channel",
            icon = Icons.Default.Settings,
            requiresAuth = true
        ))
        
        items.add(DrawerItem(
            title = "User Settings", 
            route = "user",
            icon = Icons.Default.Person,
            requiresAuth = true
        ))
        
        items.add(DrawerItem(
            title = "Ad Settings",
            route = "adSetting", 
            icon = Icons.Default.AdUnits,
            requiresAuth = true
        ))
    }
    
    // Channel-dependent items
    if (uiState.hasChannel) {
        items.add(DrawerItem(
            title = "Calendar",
            route = "calendar",
            icon = Icons.Default.CalendarMonth,
            requiresChannel = true
        ))
        
        items.add(DrawerItem(
            title = "Tickets",
            route = "tickets",
            icon = Icons.Default.ConfirmationNumber,
            requiresChannel = true
        ))
        
        // Non-guest items
        if (!uiState.isGuest) {
            items.add(DrawerItem(
                title = "Timestamp Settings",
                route = "timestampCode",
                icon = Icons.Default.AccessTime,
                requiresChannel = true,
                requiresNonGuest = true
            ))
            
            items.add(DrawerItem(
                title = "Reception & Booking",
                route = "reception",
                icon = Icons.Default.Receipt,
                requiresChannel = true,
                requiresNonGuest = true
            ))
            
            items.add(DrawerItem(
                title = "Form Edit",
                route = "entryFormEdit",
                icon = Icons.Default.EditNote,
                requiresChannel = true,
                requiresNonGuest = true
            ))
            
            items.add(DrawerItem(
                title = "Homepage Edit",
                route = "topEdit",
                icon = Icons.Default.Edit,
                requiresChannel = true,
                requiresNonGuest = true
            ))
        }
    }
    
    // Always visible items
    items.add(DrawerItem(
        title = "Tweets",
        route = "tweets",
        icon = Icons.Default.Chat
    ))
    
    items.add(DrawerItem(
        title = "Settings",
        route = "setting",
        icon = Icons.Default.Settings
    ))
    
    if (!uiState.isSignedIn) {
        items.add(DrawerItem(
            title = "Sign In",
            route = "login",
            icon = Icons.Default.Login
        ))
        
        items.add(DrawerItem(
            title = "Rules",
            route = "rule",
            icon = Icons.Default.Gavel
        ))
        
        items.add(DrawerItem(
            title = "Privacy",
            route = "privacy",
            icon = Icons.Default.Security
        ))
    }
    
    return items
}

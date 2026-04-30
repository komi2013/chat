package com.chat.android.ui.component

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.navigation.NavController
import com.chat.android.network.ChannelDetail
import com.chat.android.network.NicknameResponse

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun DrawerLayout(
    navController: NavController,
    channel: ChannelDetail? = null,
    aliases: List<NicknameResponse> = emptyList(),
    content: @Composable (PaddingValues) -> Unit
) {
    val drawerState = rememberDrawerState(DrawerValue.Closed)
    
    ModalNavigationDrawer(
        drawerState = drawerState,
        drawerContent = {
            AppDrawer(
                navController = navController,
                channel = channel,
                aliases = aliases,
                onClose = { 
                    // Close drawer when item is clicked
                    // The drawer will close automatically when navigation happens
                }
            )
        }
    ) {
        Scaffold(
            topBar = {
                TopAppBar(
                    title = { },
                    navigationIcon = {
                        IconButton(
                            onClick = {
                                // Toggle drawer
                                if (drawerState.isClosed) {
                                    drawerState.open()
                                } else {
                                    drawerState.close()
                                }
                            }
                        ) {
                            Icon(
                                imageVector = androidx.compose.material.icons.Icons.Default.Menu,
                                contentDescription = "Menu"
                            )
                        }
                    },
                    colors = TopAppBarDefaults.topAppBarColors(
                        containerColor = Color.Transparent
                    )
                )
            }
        ) { paddingValues ->
            content(paddingValues)
        }
    }
}

package com.chat.android.core.ui.component

import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import kotlinx.coroutines.launch

val LocalDrawerState = compositionLocalOf<DrawerState?> { null }

@Composable
fun DrawerLayout(
    drawerContent: @Composable () -> Unit,
    gesturesEnabled: Boolean = false,
    content: @Composable () -> Unit
) {
    val drawerState = rememberDrawerState(DrawerValue.Closed)
    
    CompositionLocalProvider(LocalDrawerState provides drawerState) {
        ModalNavigationDrawer(
            drawerState = drawerState,
            drawerContent = drawerContent,
            gesturesEnabled = gesturesEnabled
        ) {
            content()
        }
    }
}

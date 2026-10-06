package com.chat.android.feature.timestamp

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController
import com.chat.android.core.ui.component.AppDrawer
import com.chat.android.core.ui.component.DrawerLayout
import com.chat.android.core.ui.component.LocalDrawerState
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimestampScreen(
    navController: NavController,
    adminName: String,
    code: String
) {
    DrawerLayout(
        drawerContent = {
            AppDrawer(navController = navController)
        }
    ) {
        val drawerState = LocalDrawerState.current
        val scope = rememberCoroutineScope()

        Scaffold(
            topBar = {
                TopAppBar(
                    title = { Text("タイムスタンプ") },
                    navigationIcon = {
                        IconButton(onClick = { scope.launch { drawerState?.open() } }) {
                            Icon(Icons.Default.Menu, contentDescription = "メニュー")
                        }
                    }
                )
            }
        ) { padding ->
            Box(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentAlignment = Alignment.Center
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text("管理者: $adminName", style = MaterialTheme.typography.bodyLarge)
                    Text("コード: $code", style = MaterialTheme.typography.bodyMedium)
                    Spacer(Modifier.height(16.dp))
                    Button(onClick = { /* TODO: 打刻処理 */ }) {
                        Text("打刻する")
                    }
                }
            }
        }
    }
}

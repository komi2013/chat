package com.chat.android.feature.timestamp

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimestampReportScreen(
    navController: NavController,
    admin: String,
    month: String?,
    stamper: String?
) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("勤怠レポート") }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text("管理者: $admin", style = MaterialTheme.typography.titleMedium)
            month?.let { Text("対象月: $it") }
            stamper?.let { Text("打刻者: $it") }
            
            Spacer(Modifier.height(16.dp))
            
            // TODO: レポート一覧の実装
            Text("レポートの詳細はここに表示されます。", style = MaterialTheme.typography.bodyMedium)
        }
    }
}

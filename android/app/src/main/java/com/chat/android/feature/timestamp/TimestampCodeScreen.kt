package com.chat.android.feature.timestamp

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimestampCodeScreen(navController: NavController) {
    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("タイムスタンプ設定") }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(padding)
                .padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Text("タイムスタンプ用のコードを入力してください")
            // TODO: コード入力の実装
            OutlinedTextField(
                value = "",
                onValueChange = {},
                label = { Text("管理用コード") },
                modifier = Modifier.fillMaxWidth()
            )
            Button(onClick = { /* TODO */ }) {
                Text("設定を保存")
            }
        }
    }
}

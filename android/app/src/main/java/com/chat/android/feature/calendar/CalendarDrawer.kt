package com.chat.android.feature.calendar

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Home
import androidx.compose.material.icons.filled.Search
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp

@Composable
fun CalendarDrawer(
    state: CalendarUiState,
    onToggle: (String, Boolean) -> Unit,
    onHome: () -> Unit,
    onClose: () -> Unit
) {
    var userSearch by remember { mutableStateOf("") }
    val aliasNames = remember(state.aliases, userSearch) {
        state.aliases.map { it.aliasName }
            .filter { it.isNotBlank() }
            .distinct()
            .filter { userSearch.isBlank() || it.contains(userSearch, true) }
    }

    ModalDrawerSheet(modifier = Modifier.width(300.dp)) {
        Column(Modifier.fillMaxSize()) {
            // Header
            Row(
                Modifier.fillMaxWidth().padding(16.dp),
                horizontalArrangement = Arrangement.SpaceBetween,
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text("表示設定", style = MaterialTheme.typography.titleLarge, fontWeight = FontWeight.Bold)
                IconButton(onClick = onClose) {
                    Icon(Icons.Default.Close, contentDescription = "閉じる")
                }
            }

            Divider()

            // ホームへの導線（共通ドロワーを出さない代わりにここへ置く）
            NavigationDrawerItem(
                label = { Text("ホーム") },
                icon = { Icon(Icons.Default.Home, contentDescription = null) },
                selected = false,
                onClick = onHome,
                modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)
            )

            Divider()

            // Search section (filters the alias list below)
            Column(Modifier.padding(16.dp)) {
                OutlinedTextField(
                    value = userSearch,
                    onValueChange = { userSearch = it },
                    label = { Text("エイリアスを絞り込み") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    leadingIcon = { Icon(Icons.Default.Search, null) }
                )
            }

            Divider()

            // Alias checklist: all channel aliases, checked (shown) by default
            Text(
                "カレンダー表示中",
                style = MaterialTheme.typography.titleSmall,
                modifier = Modifier.padding(16.dp),
                color = MaterialTheme.colorScheme.secondary
            )

            LazyColumn(Modifier.weight(1f)) {
                items(aliasNames) { aliasName ->
                    val checked = aliasName in state.selectedAliases
                    ListItem(
                        headlineContent = { Text(aliasName) },
                        leadingContent = { Checkbox(checked = checked, onCheckedChange = null) },
                        modifier = Modifier.clickable { onToggle(aliasName, !checked) }
                    )
                }
            }
        }
    }
}

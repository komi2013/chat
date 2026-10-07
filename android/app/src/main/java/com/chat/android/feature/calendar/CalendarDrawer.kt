package com.chat.android.feature.calendar

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Close
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
    searchedUsers: List<String>,
    onSearch: (String) -> Unit,
    onRemove: (String) -> Unit,
    onClose: () -> Unit
) {
    var userSearch by remember { mutableStateOf("") }

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

            // Search section
            Column(Modifier.padding(16.dp)) {
                OutlinedTextField(
                    value = userSearch,
                    onValueChange = { userSearch = it },
                    label = { Text("ユーザー・グループ検索") },
                    modifier = Modifier.fillMaxWidth(),
                    singleLine = true,
                    leadingIcon = { Icon(Icons.Default.Search, null) }
                )

                if (userSearch.isNotBlank()) {
                    val matches = (state.aliases.filter { it.aliasName.contains(userSearch, true) }.map { it.aliasName } +
                            state.groups.filter { it.groupName.contains(userSearch, true) }.flatMap { parseCalendarNames(it.aliasNamesJson) })
                        .distinct().filterNot { it in searchedUsers }

                    matches.take(6).forEach { aliasName ->
                        ListItem(
                            headlineContent = { Text(aliasName) },
                            modifier = Modifier.clickable {
                                onSearch(aliasName)
                                userSearch = ""
                            },
                            leadingContent = { Text("＋", color = MaterialTheme.colorScheme.primary) }
                        )
                    }
                }
            }

            Divider()

            // Selected users section
            Text(
                "表示中のカレンダー",
                style = MaterialTheme.typography.titleSmall,
                modifier = Modifier.padding(16.dp),
                color = MaterialTheme.colorScheme.secondary
            )

            LazyColumn(Modifier.weight(1f)) {
                items(searchedUsers) { aliasName ->
                    ListItem(
                        headlineContent = { Text(aliasName) },
                        trailingContent = {
                            IconButton(onClick = { onRemove(aliasName) }) {
                                Icon(Icons.Default.Close, null, modifier = Modifier.size(18.dp))
                            }
                        }
                    )
                }
            }
        }
    }
}

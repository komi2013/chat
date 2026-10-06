package com.chat.android.feature.calendar.ui

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.CalendarMonth
import androidx.compose.material.icons.filled.Clear
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Home
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.platform.LocalContext
import com.chat.android.feature.calendar.CalendarItem
import androidx.navigation.NavController
import com.chat.android.core.ui.component.LocalDrawerState
import com.chat.android.feature.calendar.CalendarViewModel
import com.chat.android.feature.calendar.parseCalendarNames
import com.chat.android.navigation.CalendarRoute
import com.chat.android.navigation.HomeRoute
import kotlinx.coroutines.launch

@OptIn(ExperimentalLayoutApi::class, ExperimentalMaterial3Api::class)
@Composable
fun CalendarDrawer(
    navController: NavController,
    viewModel: CalendarViewModel
) {
    val state by viewModel.state.collectAsState()
    val drawerState = LocalDrawerState.current
    val scope = rememberCoroutineScope()
    var userSearch by remember { mutableStateOf("") }
    
    ModalDrawerSheet(
        modifier = Modifier.width(280.dp)
    ) {
        // Header
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                text = "カレンダー検索",
                style = MaterialTheme.typography.titleLarge,
                fontWeight = FontWeight.Bold
            )
            
            IconButton(onClick = { scope.launch { drawerState?.close() } }) {
                Icon(Icons.Default.Close, contentDescription = "Close drawer")
            }
        }
        
        Divider()

        // Search Section
        Column(
            modifier = Modifier
                .fillMaxWidth()
                .padding(16.dp)
        ) {
            OutlinedTextField(
                value = userSearch,
                onValueChange = { userSearch = it },
                label = { Text("ユーザー・グループ検索") },
                modifier = Modifier.fillMaxWidth(),
                singleLine = true,
                trailingIcon = {
                    if (userSearch.isNotBlank()) {
                        IconButton(onClick = { userSearch = "" }) {
                            Icon(Icons.Default.Clear, contentDescription = "Clear")
                        }
                    }
                }
            )
            
            if (userSearch.isNotBlank()) {
                val matches = remember(userSearch, state.aliases, state.groups) {
                    (state.aliases.filter { it.aliasName.contains(userSearch, true) }.map { it.aliasName } +
                            state.groups.filter { it.groupName.contains(userSearch, true) }
                                .flatMap { parseCalendarNames(it.aliasNamesJson) })
                        .distinct()
                        .take(5)
                }
                
                matches.forEach { aliasName ->
                    Text(
                        text = "＋ $aliasName",
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable {
                                viewModel.searchCalendar(aliasName, java.time.LocalDate.now().toString())
                                userSearch = ""
                                scope.launch { drawerState?.close() }
                            }
                            .padding(vertical = 8.dp, horizontal = 4.dp),
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }
        }
        
        Divider()

        // Active filters (Added back)
        val searchedUsers = remember(state.events, state.aliasName) {
            state.events.map { it.aliasName }.distinct().filter { it != state.aliasName }
        }

        if (searchedUsers.isNotEmpty()) {
            Text(
                "検索中:",
                style = MaterialTheme.typography.labelMedium,
                modifier = Modifier.padding(horizontal = 16.dp, vertical = 4.dp)
            )
            FlowRow(
                modifier = Modifier.padding(horizontal = 16.dp).fillMaxWidth(),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                searchedUsers.forEach { name ->
                    AssistChip(
                        onClick = { viewModel.removeSearchedCalendar(name) },
                        label = { Text(name) },
                        trailingIcon = { Icon(Icons.Default.Close, null, Modifier.size(16.dp)) }
                    )
                }
            }
            Divider(Modifier.padding(vertical = 8.dp))
        }
        
        // Navigation items (Link to Home as requested)
        NavigationDrawerItem(
            icon = { Icon(Icons.Default.Home, contentDescription = null) },
            label = { Text("ホーム") },
            selected = false,
            onClick = {
                navController.navigate(HomeRoute) {
                    popUpTo(HomeRoute) { inclusive = true }
                }
                scope.launch { drawerState?.close() }
            },
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)
        )

        NavigationDrawerItem(
            icon = { Icon(Icons.Default.CalendarMonth, contentDescription = null) },
            label = { Text("自分のカレンダー") },
            selected = true,
            onClick = {
                // If already on calendar, just close
                scope.launch { drawerState?.close() }
            },
            modifier = Modifier.padding(horizontal = 12.dp, vertical = 4.dp)
        )
    }
}

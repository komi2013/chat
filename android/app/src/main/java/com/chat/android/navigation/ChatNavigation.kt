package com.chat.android.navigation

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import androidx.navigation.NavType
import androidx.navigation.navArgument
import com.chat.android.ui.screen.*
import com.chat.android.ui.screen.settings.UserScreen
import com.chat.android.ui.screen.calendar.CalendarScreen
import com.chat.android.ui.screen.calendar.CalendarEditScreen
import com.chat.android.ui.screen.tickets.TicketsScreen
import com.chat.android.ui.screen.tickets.TicketScreen
import com.chat.android.ui.screen.static.RuleScreen
import com.chat.android.ui.screen.static.PrivacyScreen
import kotlinx.coroutines.launch

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChatNavigation() {
    val navController = rememberNavController()
    val drawerState = rememberDrawerState(initialValue = DrawerValue.Closed)
    val scope = rememberCoroutineScope()

    ModalNavigationDrawer(
        drawerState = drawerState,
        drawerContent = {
            ModalDrawerSheet {
                Spacer(modifier = Modifier.height(12.dp))
                NavigationDrawerItem(
                    label = { Text("ホーム") },
                    selected = false,
                    onClick = {
                        navController.navigate("top")
                        scope.launch { drawerState.close() }
                    },
                    icon = { Icon(Icons.Default.Home, contentDescription = null) }
                )
                NavigationDrawerItem(
                    label = { Text("ユーザー設定") },
                    selected = false,
                    onClick = {
                        navController.navigate("user")
                        scope.launch { drawerState.close() }
                    },
                    icon = { Icon(Icons.Default.Person, contentDescription = null) }
                )
                NavigationDrawerItem(
                    label = { Text("チャネル一覧") },
                    selected = false,
                    onClick = {
                        navController.navigate("channel")
                        scope.launch { drawerState.close() }
                    },
                    icon = { Icon(Icons.Default.List, contentDescription = null) }
                )
                NavigationDrawerItem(
                    label = { Text("ツイート一覧") },
                    selected = false,
                    onClick = {
                        navController.navigate("tweets")
                        scope.launch { drawerState.close() }
                    },
                    icon = { Icon(Icons.Default.Chat, contentDescription = null) }
                )
                Divider(modifier = Modifier.padding(vertical = 8.dp))
                NavigationDrawerItem(
                    label = { Text("設定") },
                    selected = false,
                    onClick = {
                        navController.navigate("setting")
                        scope.launch { drawerState.close() }
                    },
                    icon = { Icon(Icons.Default.Settings, contentDescription = null) }
                )
            }
        }
    ) {
        NavHost(navController = navController, startDestination = "top") {
        // Main navigation
        composable("top") {
            TopScreen(navController = navController)
        }
        
        composable("login") {
            LoginScreen(navController = navController)
        }
        
        composable("user") {
            UserScreen(navController = navController)
        }
        
        // Tweet navigation
        composable("tweets") {
            TweetScreen(navController = navController)
        }
        
        composable(
            "tweet/{date?}/{parentIDhtml?}",
            arguments = listOf(
                navArgument("date") { type = NavType.StringType; nullable = true },
                navArgument("parentIDhtml") { type = NavType.StringType; nullable = true }
            )
        ) { backStackEntry ->
            TweetScreen(
                navController = navController,
                date = backStackEntry.arguments?.getString("date"),
                parentIDhtml = backStackEntry.arguments?.getString("parentIDhtml")
            )
        }
        
        // Channel navigation
        composable("channel") {
            ChannelScreen(navController = navController)
        }
        
        composable(
            "channel/{id}",
            arguments = listOf(navArgument("id") { type = NavType.StringType })
        ) { backStackEntry ->
            ChannelScreen(
                navController = navController,
                channelId = backStackEntry.arguments?.getString("id")
            )
        }
        
        // Calendar navigation
        composable(
            "calendar/{date?}",
            arguments = listOf(navArgument("date") { type = NavType.StringType; nullable = true })
        ) { backStackEntry ->
            CalendarScreen(
                navController = navController,
                date = backStackEntry.arguments?.getString("date")
            )
        }
        
        composable(
            "calendarEdit/{id?}",
            arguments = listOf(navArgument("id") { type = NavType.StringType; nullable = true })
        ) { backStackEntry ->
            CalendarEditScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id")
            )
        }
        
        // Reception navigation
        composable(
            "reception/{id?}",
            arguments = listOf(navArgument("id") { type = NavType.StringType; nullable = true })
        ) { backStackEntry ->
            ReceptionScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id")
            )
        }
        
        composable(
            "receptionBook/{id?}",
            arguments = listOf(navArgument("id") { type = NavType.StringType; nullable = true })
        ) { backStackEntry ->
            ReceptionBookScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id")
            )
        }
        
        composable(
            "receptionMenu/{id}/{code}/{codeType}",
            arguments = listOf(
                navArgument("id") { type = NavType.StringType },
                navArgument("code") { type = NavType.StringType },
                navArgument("codeType") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            ReceptionMenuScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: "",
                code = backStackEntry.arguments?.getString("code") ?: "",
                codeType = backStackEntry.arguments?.getString("codeType") ?: ""
            )
        }
        
        composable(
            "receptionEnter/{id}/{passkey}",
            arguments = listOf(
                navArgument("id") { type = NavType.StringType },
                navArgument("passkey") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            ReceptionEnterScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: "",
                passkey = backStackEntry.arguments?.getString("passkey") ?: ""
            )
        }
        
        composable(
            "receptionOrder/{id}/{code}",
            arguments = listOf(
                navArgument("id") { type = NavType.StringType },
                navArgument("code") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            ReceptionOrderScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: "",
                code = backStackEntry.arguments?.getString("code") ?: ""
            )
        }
        
        composable(
            "receptionShift/{id}",
            arguments = listOf(navArgument("id") { type = NavType.StringType })
        ) { backStackEntry ->
            ReceptionShiftScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: ""
            )
        }
        
        composable(
            "receptionThread/{channelID}/{code}",
            arguments = listOf(
                navArgument("channelID") { type = NavType.StringType },
                navArgument("code") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            ReceptionThreadScreen(
                navController = navController,
                channelID = backStackEntry.arguments?.getString("channelID") ?: "",
                code = backStackEntry.arguments?.getString("code") ?: ""
            )
        }
        
        // Thread navigation
        composable(
            "thread/{channel_id}/{parentID}",
            arguments = listOf(
                navArgument("channel_id") { type = NavType.StringType },
                navArgument("parentID") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            ThreadScreen(
                navController = navController,
                channelId = backStackEntry.arguments?.getString("channel_id") ?: "",
                parentId = backStackEntry.arguments?.getString("parentID") ?: ""
            )
        }
        
        composable(
            "threadCall/{channelName}",
            arguments = listOf(navArgument("channelName") { type = NavType.StringType })
        ) { backStackEntry ->
            ThreadCallScreen(
                navController = navController,
                channelName = backStackEntry.arguments?.getString("channelName") ?: ""
            )
        }
        
        composable(
            "threadHead/{parent_id}",
            arguments = listOf(navArgument("parent_id") { type = NavType.StringType })
        ) { backStackEntry ->
            ThreadHeadScreen(
                navController = navController,
                parentId = backStackEntry.arguments?.getString("parent_id") ?: ""
            )
        }
        
        // Ticket navigation
        composable(
            "ticket/{ticketID?}",
            arguments = listOf(navArgument("ticketID") { type = NavType.StringType; nullable = true })
        ) { backStackEntry ->
            TicketScreen(
                navController = navController,
                ticketId = backStackEntry.arguments?.getString("ticketID")
            )
        }
        
        composable("tickets") {
            TicketsScreen(navController = navController)
        }
        
        // Timestamp navigation
        composable(
            "timestamp/{adminName}/{code}",
            arguments = listOf(
                navArgument("adminName") { type = NavType.StringType },
                navArgument("code") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            TimestampScreen(
                navController = navController,
                adminName = backStackEntry.arguments?.getString("adminName") ?: "",
                code = backStackEntry.arguments?.getString("code") ?: ""
            )
        }
        
        composable("timestampCode") {
            TimestampCodeScreen(navController = navController)
        }
        
        composable(
            "timestampReport/{admin}",
            arguments = listOf(navArgument("admin") { type = NavType.StringType })
        ) { backStackEntry ->
            TimestampReportScreen(
                navController = navController,
                admin = backStackEntry.arguments?.getString("admin") ?: ""
            )
        }
        
        // Other screens
        composable("setting") {
            SettingScreen(navController = navController)
        }
        
        composable(
            "adSetting/{date?}",
            arguments = listOf(navArgument("date") { type = NavType.StringType; nullable = true })
        ) { backStackEntry ->
            AdSettingScreen(
                navController = navController,
                date = backStackEntry.arguments?.getString("date")
            )
        }
        
        composable(
            "answers/{id}",
            arguments = listOf(navArgument("id") { type = NavType.StringType })
        ) { backStackEntry ->
            AnswersScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: ""
            )
        }
        
        composable(
            "entryForm/{id}",
            arguments = listOf(navArgument("id") { type = NavType.StringType })
        ) { backStackEntry ->
            EntryFormScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: ""
            )
        }
        
        composable(
            "entryFormEdit/{id?}",
            arguments = listOf(navArgument("id") { type = NavType.StringType; nullable = true })
        ) { backStackEntry ->
            EntryFormEditScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id")
            )
        }
        
        composable(
            "group/{id}",
            arguments = listOf(navArgument("id") { type = NavType.StringType })
        ) { backStackEntry ->
            GroupScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: ""
            )
        }
        
        composable(
            "nickname/{name}",
            arguments = listOf(navArgument("name") { type = NavType.StringType })
        ) { backStackEntry ->
            NicknameScreen(
                navController = navController,
                name = backStackEntry.arguments?.getString("name") ?: ""
            )
        }
        
        composable(
            "people/{id}/{name}",
            arguments = listOf(
                navArgument("id") { type = NavType.StringType },
                navArgument("name") { type = NavType.StringType }
            )
        ) { backStackEntry ->
            PeopleScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: "",
                name = backStackEntry.arguments?.getString("name") ?: ""
            )
        }
        
        composable(
            "profile/{id}",
            arguments = listOf(navArgument("id") { type = NavType.StringType })
        ) { backStackEntry ->
            ProfileScreen(
                navController = navController,
                id = backStackEntry.arguments?.getString("id") ?: ""
            )
        }
        
        composable("topEdit") {
            TopEditScreen(navController = navController)
        }
        
        // Static HTML pages
        composable("rule") {
            RuleScreen(navController = navController)
        }
        
        composable("privacy") {
            PrivacyScreen(navController = navController)
        }
        }
    }
}

package com.chat.android.ui.screen

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.navigation.NavController

// Calendar Edit Screen
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun CalendarEditScreen(
    navController: NavController,
    id: String? = null
) {
    BasicScreen(
        navController = navController,
        title = "カレンダー編集",
        subtitle = if (id != null) "ID: $id" else "新規作成"
    )
}

// Reception Screens
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionScreen(
    navController: NavController,
    id: String? = null
) {
    BasicScreen(
        navController = navController,
        title = "受付",
        subtitle = if (id != null) "ID: $id" else "一覧"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionBookScreen(
    navController: NavController,
    id: String? = null
) {
    BasicScreen(
        navController = navController,
        title = "受付予約",
        subtitle = if (id != null) "ID: $id" else "一覧"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionMenuScreen(
    navController: NavController,
    id: String,
    code: String,
    codeType: String
) {
    BasicScreen(
        navController = navController,
        title = "受付メニュー",
        subtitle = "ID: $id, Code: $code, Type: $codeType"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionEnterScreen(
    navController: NavController,
    id: String,
    passkey: String
) {
    BasicScreen(
        navController = navController,
        title = "受付入場",
        subtitle = "ID: $id, Passkey: $passkey"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionOrderScreen(
    navController: NavController,
    id: String,
    code: String
) {
    BasicScreen(
        navController = navController,
        title = "受付注文",
        subtitle = "ID: $id, Code: $code"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionShiftScreen(
    navController: NavController,
    id: String
) {
    BasicScreen(
        navController = navController,
        title = "受付シフト",
        subtitle = "ID: $id"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionThreadScreen(
    navController: NavController,
    channelID: String,
    code: String
) {
    BasicScreen(
        navController = navController,
        title = "受付スレッド",
        subtitle = "Channel: $channelID, Code: $code"
    )
}

// Thread Screens
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ThreadScreen(
    navController: NavController,
    channelId: String,
    parentId: String
) {
    BasicScreen(
        navController = navController,
        title = "スレッド",
        subtitle = "Channel: $channelId, Parent: $parentId"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ThreadCallScreen(
    navController: NavController,
    channelName: String
) {
    BasicScreen(
        navController = navController,
        title = "スレッド通話",
        subtitle = "Channel: $channelName"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ThreadHeadScreen(
    navController: NavController,
    parentId: String
) {
    BasicScreen(
        navController = navController,
        title = "スレッドヘッド",
        subtitle = "Parent: $parentId"
    )
}

// Ticket Screens
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TicketScreen(
    navController: NavController,
    ticketId: String? = null
) {
    BasicScreen(
        navController = navController,
        title = "チケット",
        subtitle = if (ticketId != null) "ID: $ticketId" else "一覧"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TicketsScreen(navController: NavController) {
    BasicScreen(
        navController = navController,
        title = "チケット一覧",
        subtitle = "すべてのチケット"
    )
}

// Timestamp Screens
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimestampScreen(
    navController: NavController,
    adminName: String,
    code: String
) {
    BasicScreen(
        navController = navController,
        title = "タイムスタンプ",
        subtitle = "Admin: $adminName, Code: $code"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimestampCodeScreen(navController: NavController) {
    BasicScreen(
        navController = navController,
        title = "タイムスタンプコード",
        subtitle = "コード管理"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TimestampReportScreen(
    navController: NavController,
    admin: String
) {
    BasicScreen(
        navController = navController,
        title = "タイムスタンプレポート",
        subtitle = "Admin: $admin"
    )
}

// Other Screens
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun SettingScreen(navController: NavController) {
    BasicScreen(
        navController = navController,
        title = "設定",
        subtitle = "アプリ設定"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AdSettingScreen(
    navController: NavController,
    date: String? = null
) {
    BasicScreen(
        navController = navController,
        title = "広告設定",
        subtitle = if (date != null) "日付: $date" else "全般"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun AnswersScreen(
    navController: NavController,
    id: String
) {
    BasicScreen(
        navController = navController,
        title = "回答",
        subtitle = "ID: $id"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun GroupScreen(
    navController: NavController,
    id: String
) {
    BasicScreen(
        navController = navController,
        title = "グループ",
        subtitle = "ID: $id"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun NicknameScreen(
    navController: NavController,
    name: String
) {
    BasicScreen(
        navController = navController,
        title = "ニックネーム",
        subtitle = "Name: $name"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PeopleScreen(
    navController: NavController,
    id: String,
    name: String
) {
    BasicScreen(
        navController = navController,
        title = "人物",
        subtitle = "ID: $id, Name: $name"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(
    navController: NavController,
    id: String
) {
    BasicScreen(
        navController = navController,
        title = "プロフィール",
        subtitle = "ID: $id"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TopEditScreen(navController: NavController) {
    BasicScreen(
        navController = navController,
        title = "トップ編集",
        subtitle = "トップページ編集"
    )
}

// Static HTML pages
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun RuleScreen(navController: NavController) {
    BasicScreen(
        navController = navController,
        title = "規則",
        subtitle = "利用規則"
    )
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PrivacyScreen(navController: NavController) {
    BasicScreen(
        navController = navController,
        title = "個人情報遵守",
        subtitle = "プライバシーポリシー"
    )
}

// Login Screen (placeholder)
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun LoginScreen(navController: NavController) {
    BasicScreen(
        navController = navController,
        title = "サインイン",
        subtitle = "Googleサインイン"
    )
}

// Basic reusable screen component
@OptIn(ExperimentalMaterial3Api::class)
@Composable
private fun BasicScreen(
    navController: NavController,
    title: String,
    subtitle: String
) {
    Column(
        modifier = Modifier
            .fillMaxSize()
            .padding(16.dp)
    ) {
        // Header
        Row(
            modifier = Modifier
                .fillMaxWidth()
                .padding(bottom = 16.dp),
            horizontalArrangement = Arrangement.SpaceBetween,
            verticalAlignment = Alignment.CenterVertically
        ) {
            IconButton(onClick = { navController.navigateUp() }) {
                Icon(
                    imageVector = Icons.Default.ArrowBack,
                    contentDescription = "Back"
                )
            }
            Text(
                text = title,
                style = MaterialTheme.typography.headlineMedium,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.width(48.dp)) // Balance the back button
        }

        // Content
        Card(
            modifier = Modifier.fillMaxWidth()
        ) {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = title,
                    style = MaterialTheme.typography.titleLarge
                )
                Spacer(modifier = Modifier.height(8.dp))
                Text(
                    text = subtitle,
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )
                Spacer(modifier = Modifier.height(16.dp))
                Text(
                    text = "Vueから移行された画面",
                    style = MaterialTheme.typography.bodySmall,
                    color = MaterialTheme.colorScheme.outline
                )
            }
        }
    }
}

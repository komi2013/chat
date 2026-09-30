package com.chat.android.feature.profile

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.core.util.RandomAvatar
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.feature.channel.AliasAvatar

/**
 * チャネル参加 / プロフィール編集画面（vue/src/views/Profile.vue に対応）。
 *
 * 招待URL（/profile/{id}/?code=...）をスキャンするとここが開く。
 * code があれば「参加」、なければ参加済みチャネルのプロフィール編集になる。
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ProfileScreen(
    navController: NavController,
    id: String? = null,
    code: String? = null,
    viewModel: ProfileViewModel = hiltViewModel()
) {
    val state by viewModel.uiState.collectAsState()
    var showImgPicker by remember { mutableStateOf(false) }
    var initialized by rememberSaveable { mutableStateOf(false) }

    LaunchedEffect(id, code) {
        if (!initialized) {
            initialized = true
            viewModel.init(id, code)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (state.isJoinMode) "チャネルに参加" else "プロフィール編集") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                },
                actions = {
                    IconButton(onClick = { viewModel.onImgChange(RandomAvatar.random()) }) {
                        Icon(Icons.Default.Refresh, contentDescription = "アイコンをランダムに")
                    }
                }
            )
        }
    ) { padding ->
        LazyColumn(
            modifier = Modifier.fillMaxSize().padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(16.dp)
        ) {
            if (state.isLoading) {
                item { LinearProgressIndicator(modifier = Modifier.fillMaxWidth()) }
            }
            state.error?.let { item { Banner(it, MaterialTheme.colorScheme.error) } }
            state.successMessage?.let { item { Banner(it, MaterialTheme.colorScheme.primary) } }

            item {
                val channel = state.channel
                if (channel != null) {
                    TextButton(onClick = { navController.navigateUp() }) {
                        Text(channel.channelName)
                    }
                } else {
                    Text("チャネル内ニックネーム設定", style = MaterialTheme.typography.titleMedium)
                }
            }

            item {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    AliasAvatar(
                        aliasImg = state.aliasImg.toAbsoluteImageUrl(),
                        modifier = Modifier.size(64.dp)
                    )
                    Spacer(Modifier.width(16.dp))
                    if (state.isJoinMode) {
                        OutlinedTextField(
                            value = state.aliasName,
                            onValueChange = viewModel::onNameChange,
                            label = { Text("このチャネルのニックネーム") },
                            singleLine = true,
                            modifier = Modifier.weight(1f)
                        )
                    } else {
                        Text(
                            state.aliasName,
                            style = MaterialTheme.typography.titleMedium,
                            modifier = Modifier.weight(1f)
                        )
                    }
                }
            }

            item {
                OutlinedButton(
                    onClick = { showImgPicker = true },
                    modifier = Modifier.fillMaxWidth()
                ) { Text("アイコンを変更") }
            }

            // 自己紹介は参加後に編集できる
            if (!state.isJoinMode) {
                item {
                    OutlinedTextField(
                        value = state.aliasBio,
                        onValueChange = viewModel::onBioChange,
                        label = { Text("自己紹介を入力してください") },
                        modifier = Modifier.fillMaxWidth().heightIn(min = 100.dp),
                        minLines = 4
                    )
                }
            }

            item {
                Button(
                    onClick = { viewModel.submit() },
                    enabled = !state.isLoading && state.aliasName.isNotBlank(),
                    modifier = Modifier.fillMaxWidth()
                ) { Text(if (state.isJoinMode) "参加" else "保存") }
            }

            if (state.joinGroups.isNotEmpty()) {
                item { SectionTitle("参加グループ一覧") }
                items(state.joinGroups, key = { it.groupID }) { group ->
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        AliasAvatar(
                            aliasImg = group.groupImg.toAbsoluteImageUrl(),
                            modifier = Modifier.size(32.dp)
                        )
                        Spacer(Modifier.width(8.dp))
                        Text(group.groupName)
                    }
                }
            }

            if (state.sameUserAliases.isNotEmpty()) {
                item { SectionTitle("マイニックネーム一覧") }
                items(state.sameUserAliases, key = { it.aliasID }) { alias ->
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        AliasAvatar(
                            aliasImg = alias.aliasImg.toAbsoluteImageUrl(),
                            modifier = Modifier.size(32.dp)
                        )
                        Spacer(Modifier.width(8.dp))
                        Text(alias.aliasName, modifier = Modifier.weight(1f))
                        // Profile.vue の 🔀 と同じ切り替え
                        Text(
                            "🔀",
                            modifier = Modifier
                                .clickable { viewModel.switchAlias(alias.aliasName) }
                                .padding(8.dp)
                        )
                    }
                }
            }
        }
    }

    if (showImgPicker) {
        AvatarPickerDialog(
            current = state.aliasImg,
            onSelect = { viewModel.onImgChange(it); showImgPicker = false },
            onDismiss = { showImgPicker = false }
        )
    }
}

@Composable
private fun Banner(message: String, color: Color) {
    Card(
        colors = CardDefaults.cardColors(containerColor = color.copy(alpha = 0.12f)),
        modifier = Modifier.fillMaxWidth()
    ) { Text(message, color = color, modifier = Modifier.padding(12.dp)) }
}

@Composable
private fun SectionTitle(text: String) {
    Text(text, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
}

/** ",絵文字,#色" 形式のアイコン選択ダイアログ。 */
@Composable
private fun AvatarPickerDialog(
    current: String,
    onSelect: (String) -> Unit,
    onDismiss: () -> Unit
) {
    val colors = RandomAvatar.COLOR_PRESETS
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("アイコンを選択") },
        text = {
            Column(
                modifier = Modifier.verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                RandomAvatar.EMOJI_PRESETS.chunked(6).forEachIndexed { rowIndex, rowEmojis ->
                    Row(horizontalArrangement = Arrangement.spacedBy(6.dp)) {
                        rowEmojis.forEachIndexed { columnIndex, emoji ->
                            val color = colors[(rowIndex * 6 + columnIndex) % colors.size]
                            val image = ",$emoji,$color"
                            val isCurrent = image == current
                            Box(
                                modifier = Modifier
                                    .size(40.dp)
                                    .background(
                                        MaterialTheme.colorScheme.surfaceVariant,
                                        RoundedCornerShape(8.dp)
                                    )
                                    .border(
                                        if (isCurrent) 2.dp else 0.dp,
                                        MaterialTheme.colorScheme.primary,
                                        RoundedCornerShape(8.dp)
                                    )
                                    .clickable { onSelect(image) },
                                contentAlignment = Alignment.Center
                            ) { Text(emoji) }
                        }
                    }
                }
            }
        },
        confirmButton = { TextButton(onClick = onDismiss) { Text("閉じる") } }
    )
}

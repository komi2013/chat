package com.chat.android.feature.group

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.feature.channel.AliasAvatar
import com.chat.android.feature.channel.DbAlias

/**
 * グループ編集画面（vue/src/views/Group.vue に対応）。
 *
 * チャネル画面から遷移し、グループの作成・削除・メンバー変更を行い、
 * ChannelEdit/ へ差分（groups）を送る。
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun GroupScreen(
    navController: NavController,
    id: String? = null,
    viewModel: GroupViewModel = hiltViewModel()
) {
    val state by viewModel.uiState.collectAsState()
    var initialized by rememberSaveable { mutableStateOf(false) }
    var searchFor by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(id) {
        if (!initialized) {
            initialized = true
            viewModel.init(id)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("グループ編集") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                },
                actions = {
                    TextButton(onClick = { viewModel.save() }, enabled = !state.isLoading) {
                        Text("保存")
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

            state.channel?.let { channel ->
                item { Text(channel.channelName, style = MaterialTheme.typography.titleMedium) }
            }

            if (state.groups.isEmpty()) {
                item { Text("グループがありません", style = MaterialTheme.typography.bodyMedium) }
            }

            items(state.groups, key = { it.groupID.ifBlank { "new-${it.groupName}" } }) { group ->
                GroupCard(
                    group = group,
                    aliases = state.aliases,
                    searchOpen = searchFor == group.groupID,
                    onToggleSearch = {
                        searchFor = if (searchFor == group.groupID) null else group.groupID
                    },
                    onNameChange = { viewModel.updateName(group.groupID, it) },
                    onBioChange = { viewModel.updateBio(group.groupID, it) },
                    onRandomImg = { viewModel.randomImage(group.groupID) },
                    onAddMember = { viewModel.addMember(group.groupID, it) },
                    onRemoveMember = { viewModel.removeMember(group.groupID, it) },
                    onToggleRemoved = { viewModel.toggleRemoved(group.groupID) }
                )
            }

            item {
                OutlinedButton(
                    onClick = { viewModel.addGroup() },
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Icon(Icons.Default.Add, contentDescription = null)
                    Spacer(Modifier.width(8.dp))
                    Text("グループを追加")
                }
            }
        }
    }
}

/** 1つのグループの編集カード。 */
@Composable
private fun GroupCard(
    group: EditableGroup,
    aliases: List<DbAlias>,
    searchOpen: Boolean,
    onToggleSearch: () -> Unit,
    onNameChange: (String) -> Unit,
    onBioChange: (String) -> Unit,
    onRandomImg: () -> Unit,
    onAddMember: (String) -> Unit,
    onRemoveMember: (String) -> Unit,
    onToggleRemoved: () -> Unit
) {
    var search by remember(group.groupID) { mutableStateOf("") }
    val editable = group.editable

    Card(
        modifier = Modifier.fillMaxWidth(),
        colors = if (group.removed) {
            CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.errorContainer.copy(alpha = 0.3f)
            )
        } else {
            CardDefaults.cardColors()
        }
    ) {
        Column(Modifier.padding(16.dp), verticalArrangement = Arrangement.spacedBy(8.dp)) {
            if (group.isNew || editable) {
                OutlinedTextField(
                    value = group.groupName,
                    onValueChange = onNameChange,
                    label = { Text("グループ名") },
                    singleLine = true,
                    // 既存グループの名前は変更不可（Vue と同じ）
                    enabled = group.isNew,
                    modifier = Modifier.fillMaxWidth()
                )
            } else {
                Text(group.groupName, style = MaterialTheme.typography.titleMedium)
            }

            Row(verticalAlignment = Alignment.CenterVertically) {
                AliasAvatar(
                    aliasImg = group.groupImg.toAbsoluteImageUrl(),
                    modifier = Modifier.size(40.dp)
                )
                Spacer(Modifier.width(12.dp))
                if (editable) {
                    IconButton(onClick = onRandomImg) {
                        Icon(Icons.Default.Refresh, contentDescription = "アイコンをランダムに")
                    }
                }
            }

            if (editable) {
                OutlinedTextField(
                    value = group.groupBio,
                    onValueChange = onBioChange,
                    label = { Text("グループ説明") },
                    modifier = Modifier.fillMaxWidth(),
                    minLines = 2
                )

                // メンバー選択（SelectAlias.vue 相当）
                if (searchOpen) {
                    OutlinedTextField(
                        value = search,
                        onValueChange = { search = it },
                        label = { Text("ユーザー検索") },
                        singleLine = true,
                        modifier = Modifier.fillMaxWidth()
                    )
                    val candidates = aliases.filter { alias ->
                        alias.aliasName.contains(search, ignoreCase = true) &&
                            alias.aliasName !in group.aliasNames
                    }
                    if (search.isNotBlank() && candidates.isEmpty()) {
                        Text("該当なし", style = MaterialTheme.typography.bodySmall)
                    }
                    candidates.take(8).forEach { alias ->
                        MemberRow(alias) { onAddMember(alias.aliasName) }
                    }
                }

                SelectedMembers(
                    members = group.aliasNames,
                    aliases = aliases,
                    onRemove = onRemoveMember
                )

                OutlinedButton(onClick = onToggleSearch, modifier = Modifier.fillMaxWidth()) {
                    Text(if (searchOpen) "検索を閉じる" else "メンバーを追加")
                }

                OutlinedButton(
                    onClick = onToggleRemoved,
                    modifier = Modifier.fillMaxWidth()
                ) {
                    Text(if (group.removed) "削除を取り消す" else "削除")
                }
            }
        }
    }
}

/** 選択済みメンバーの一覧（× で除外）。 */
@Composable
private fun SelectedMembers(
    members: List<String>,
    aliases: List<DbAlias>,
    onRemove: (String) -> Unit
) {
    if (members.isEmpty()) return
    Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
        Text("メンバー", style = MaterialTheme.typography.labelMedium)
        members.forEach { name ->
            val alias = aliases.find { it.aliasName == name }
            Row(verticalAlignment = Alignment.CenterVertically) {
                AliasAvatar(
                    aliasImg = (alias?.aliasImg ?: "").toAbsoluteImageUrl(),
                    modifier = Modifier.size(28.dp)
                )
                Spacer(Modifier.width(8.dp))
                Text(name, modifier = Modifier.weight(1f))
                TextButton(onClick = { onRemove(name) }) { Text("×") }
            }
        }
    }
}

@Composable
private fun MemberRow(alias: DbAlias, onClick: () -> Unit) {
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .clickable(onClick = onClick)
            .padding(vertical = 4.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        AliasAvatar(
            aliasImg = alias.aliasImg.toAbsoluteImageUrl(),
            modifier = Modifier.size(28.dp)
        )
        Spacer(Modifier.width(8.dp))
        Text(alias.aliasName)
    }
}

@Composable
private fun Banner(message: String, color: Color) {
    Card(
        colors = CardDefaults.cardColors(containerColor = color.copy(alpha = 0.12f)),
        modifier = Modifier.fillMaxWidth()
    ) { Text(message, color = color, modifier = Modifier.padding(12.dp)) }
}

package com.chat.android.feature.people

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.feature.channel.AliasAvatar
import com.chat.android.navigation.ChannelRoute
import com.chat.android.navigation.GroupRoute
import com.chat.android.navigation.PeopleRoute

/**
 * ニックネーム一覧画面（vue/src/views/People.vue に対応）。
 *
 * name にエイリアス名が来れば個人、グループ名と一致するものがあれば
 * グループとして表示する（Vue と同じ判定）。
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun PeopleScreen(
    navController: NavController,
    id: String? = null,
    name: String? = null,
    viewModel: PeopleViewModel = hiltViewModel()
) {
    val state by viewModel.uiState.collectAsState()
    var initialized by rememberSaveable { mutableStateOf(false) }

    LaunchedEffect(id, name) {
        if (!initialized) {
            initialized = true
            viewModel.init(id, name)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("ニックネーム一覧") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        }
    ) { padding ->
        if (state.isLoading) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
            return@Scaffold
        }

        state.error?.let { error ->
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                Text(error, color = MaterialTheme.colorScheme.error)
            }
            return@Scaffold
        }

        LazyColumn(
            modifier = Modifier.fillMaxSize().padding(padding),
            contentPadding = PaddingValues(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            state.channel?.let { channel ->
                item {
                    Column {
                        Text(
                            channel.channelName,
                            style = MaterialTheme.typography.titleMedium,
                            color = MaterialTheme.colorScheme.primary,
                            modifier = Modifier.clickable {
                                navController.navigate(ChannelRoute(id = channel.channelID))
                            }
                        )
                        if (state.isGroup) {
                            Text(
                                "⬅️ グループ編集",
                                style = MaterialTheme.typography.bodySmall,
                                modifier = Modifier.clickable {
                                    navController.navigate(GroupRoute(id = channel.channelID))
                                }
                            )
                        }
                    }
                }
            }

            item {
                Row(verticalAlignment = Alignment.CenterVertically) {
                    AliasAvatar(
                        aliasImg = state.personImage.toAbsoluteImageUrl(),
                        modifier = Modifier.size(56.dp)
                    )
                    Spacer(Modifier.width(12.dp))
                    Text(state.personName, style = MaterialTheme.typography.headlineSmall)
                }
            }

            if (state.personBio.isNotBlank()) {
                item {
                    Surface(
                        color = MaterialTheme.colorScheme.surfaceVariant,
                        modifier = Modifier.fillMaxWidth()
                    ) {
                        Text(
                            state.personBio,
                            style = MaterialTheme.typography.bodyMedium,
                            modifier = Modifier.padding(12.dp)
                        )
                    }
                }
            }

            item { SectionTitle("ニックネーム一覧") }
            if (state.myAliases.isEmpty()) {
                item { Text("ありません", style = MaterialTheme.typography.bodyMedium) }
            }
            items(state.myAliases, key = { it.aliasID }) { alias ->
                Row(
                    modifier = Modifier.fillMaxWidth().clickable {
                        navController.navigate(PeopleRoute(id = id, name = alias.aliasName))
                    }.padding(vertical = 4.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    AliasAvatar(
                        aliasImg = alias.aliasImg.toAbsoluteImageUrl(),
                        modifier = Modifier.size(32.dp)
                    )
                    Spacer(Modifier.width(8.dp))
                    Text(alias.aliasName, modifier = Modifier.weight(1f))
                }
            }
            // 参加グループ一覧は個人表示のときだけ（Vue と同じ）
            if (!state.isGroup) {
                item { SectionTitle("参加グループ一覧") }
                if (state.joinGroups.isEmpty()) {
                    item { Text("ありません", style = MaterialTheme.typography.bodyMedium) }
                }
                items(state.joinGroups, key = { it.groupID }) { group ->
                    Row(
                        modifier = Modifier.fillMaxWidth().clickable {
                            navController.navigate(PeopleRoute(id = id, name = group.groupName))
                        }.padding(vertical = 4.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        AliasAvatar(
                            aliasImg = group.groupImg.toAbsoluteImageUrl(),
                            modifier = Modifier.size(32.dp)
                        )
                        Spacer(Modifier.width(8.dp))
                        Text(group.groupName, modifier = Modifier.weight(1f))
                    }
                }
            }
        }
    }
}

@Composable
private fun SectionTitle(text: String) {
    Text(text, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold)
}
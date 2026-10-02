package com.chat.android.feature.thread

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.Mood
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.core.util.EmojiCatalog
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.feature.channel.AliasAvatar
import com.chat.android.navigation.ThreadHeadRoute

/**
 * スレッド画面（vue/src/views/Thread.vue に対応）。
 *
 * メッセージ一覧・絵文字リアクション・入力欄（EditBox）を提供する。
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ThreadScreen(
    navController: NavController,
    channelID: String? = null,
    parentID: String? = null,
    viewModel: ThreadViewModel = hiltViewModel()
) {
    val state by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val uriHandler = LocalUriHandler.current
    var initialized by rememberSaveable { mutableStateOf(false) }

    var editingId by remember { mutableStateOf<String?>(null) }
    var editingText by remember { mutableStateOf("") }
    var emojiTarget by remember { mutableStateOf<String?>(null) }
    var emojiedTarget by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(channelID, parentID) {
        if (!initialized) {
            initialized = true
            viewModel.init(channelID, parentID)
        }
    }

    val emojis = remember(emojiTarget) { EmojiCatalog.masterEmojis(context) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        state.title.ifBlank { "スレッド" },
                        modifier = if (state.head != null) Modifier.clickable {
                            navController.navigate(ThreadHeadRoute(parentID = parentID))
                        } else Modifier
                    )
                },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        }
    ) { padding ->
        Column(Modifier.fillMaxSize().padding(padding)) {
            state.error?.let {
                Text(it, color = MaterialTheme.colorScheme.error, modifier = Modifier.padding(8.dp))
            }

            LazyColumn(
                modifier = Modifier.weight(1f).fillMaxWidth(),
                contentPadding = PaddingValues(8.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                state.head?.let { head ->
                    item(key = "head") {
                        val isEditing = editingId == head.parentID
                        val headAlias = state.aliases.find { it.aliasName == head.aliasName }
                        MessageRow(
                            aliasName = head.aliasName,
                            aliasImg = headAlias?.aliasImg?.toAbsoluteImageUrl().orEmpty(),
                            createdAt = head.updatedAt,
                            emojisJson = head.emojisJson,
                            canEdit = head.aliasName == state.myname,
                            isEditing = isEditing,
                            onEdit = {
                                editingId = head.parentID
                                editingText = head.messageTxt
                            },
                            onReact = { emojiTarget = head.parentID },
                            onShowReactions = { emojiedTarget = head.parentID }
                        ) {
                            if (isEditing) {
                                EditBox(
                                    editingMessageId = editingId,
                                    initialText = editingText,
                                    canPost = state.canPost,
                                    isPosting = state.isPosting,
                                    isDm = state.isDm,
                                    isNewThread = false,
                                    myname = state.myname,
                                    aliases = state.aliases,
                                    groups = state.groups,
                                    onPost = { text, editingMessageID, asDelete, task, group, fileUris ->
                                        // parentID is messageID for threadHead
                                        viewModel.postMessage(text, editingMessageID, asDelete, task, group, fileUris)
                                        editingId = null
                                        editingText = ""
                                    },
                                    onCall = {
                                        viewModel.postCall { url -> runCatching { uriHandler.openUri(url) } }
                                    },
                                    onCancelEdit = {
                                        editingId = null
                                        editingText = ""
                                    }
                                )
                            } else {
                                ThreadMessageText(
                                    text = head.messageTxt,
                                    modifier = Modifier.fillMaxWidth().padding(start = 40.dp, top = 2.dp)
                                )
                            }
                        }
                        if (!isEditing) {
                            Divider(Modifier.padding(vertical = 8.dp))
                        }
                    }
                }

                items(state.messages, key = { it.messageID }) { message ->
                    val isEditing = editingId == message.messageID
                    MessageRow(
                        aliasName = message.aliasName,
                        aliasImg = message.aliasImg.toAbsoluteImageUrl(),
                        createdAt = message.createdAt,
                        emojisJson = message.emojisJson,
                        canEdit = message.aliasName == state.myname,
                        isEditing = isEditing,
                        onEdit = {
                            editingId = message.messageID
                            editingText = message.messageTxt
                        },
                        onReact = { emojiTarget = message.messageID },
                        onShowReactions = { emojiedTarget = message.messageID }
                    ) {
                        if (isEditing) {
                            EditBox(
                                editingMessageId = editingId,
                                initialText = editingText,
                                canPost = state.canPost,
                                isPosting = state.isPosting,
                                isDm = state.isDm,
                                isNewThread = false,
                                myname = state.myname,
                                aliases = state.aliases,
                                groups = state.groups,
                                onPost = { text, editingMessageID, asDelete, task, group, fileUris ->
                                    viewModel.postMessage(text, editingMessageID, asDelete, task, group, fileUris)
                                    editingId = null
                                    editingText = ""
                                },
                                onCall = {
                                    viewModel.postCall { url -> runCatching { uriHandler.openUri(url) } }
                                },
                                onCancelEdit = {
                                    editingId = null
                                    editingText = ""
                                }
                            )
                        } else {
                            ThreadMessageText(
                                text = message.messageTxt,
                                modifier = Modifier.fillMaxWidth().padding(start = 40.dp, top = 2.dp)
                            )
                        }
                    }
                }
            }

            if (editingId == null) {
                EditBox(
                    editingMessageId = null,
                    initialText = "",
                    canPost = state.canPost,
                    isPosting = state.isPosting,
                    isDm = state.isDm,
                    isNewThread = state.isNewThread,
                    myname = state.myname,
                    aliases = state.aliases,
                    groups = state.groups,
                    onPost = { text, editingMessageID, asDelete, task, group, fileUris ->
                        viewModel.postMessage(text, editingMessageID, asDelete, task, group, fileUris)
                    },
                    onCall = {
                        viewModel.postCall { url -> runCatching { uriHandler.openUri(url) } }
                    },
                    onCancelEdit = {}
                )
            }
        }
    }

    emojiTarget?.let { messageID ->
        EmojiModal(
            emojis = emojis,
            errorMessage = null,
            onSelect = { emoji ->
                EmojiCatalog.rotateEmoji(context, emoji)
                viewModel.postEmoji(messageID, emoji, state.parentID, delete = false)
                emojiTarget = null
            },
            onDismiss = { emojiTarget = null }
        )
    }

    emojiedTarget?.let { messageID ->
        val json = if (messageID == state.parentID) {
            state.head?.emojisJson ?: "[]"
        } else {
            state.messages.firstOrNull { it.messageID == messageID }?.emojisJson ?: "[]"
        }
        EmojiedModal(emojisJson = json, onDismiss = { emojiedTarget = null })
    }
}
@Composable
private fun MessageRow(
    aliasName: String,
    aliasImg: String,
    createdAt: String,
    emojisJson: String,
    canEdit: Boolean,
    isEditing: Boolean,
    onEdit: () -> Unit,
    onReact: () -> Unit,
    onShowReactions: () -> Unit,
    content: @Composable () -> Unit
) {
    Column(Modifier.fillMaxWidth()) {
        Row(verticalAlignment = Alignment.CenterVertically) {
            AliasAvatar(aliasImg = aliasImg, modifier = Modifier.size(32.dp))
            Spacer(Modifier.width(8.dp))
            Text(aliasName, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
            Spacer(Modifier.width(8.dp))
            Text(createdAt, style = MaterialTheme.typography.labelSmall)
        }
        
        content()

        // リアクションの集約表示（vue の calcEmoji 相当）
        val counts = EmojiCatalog.calcEmojis(emojisJson, aliasName)
        if (counts.isNotEmpty() && !isEditing) {
            Row(
                modifier = Modifier.padding(start = 40.dp, top = 4.dp),
                horizontalArrangement = Arrangement.spacedBy(4.dp)
            ) {
                counts.forEach { item ->
                    Surface(
                        color = if (item.selected) MaterialTheme.colorScheme.primaryContainer
                        else MaterialTheme.colorScheme.surfaceVariant,
                        shape = MaterialTheme.shapes.small,
                        modifier = Modifier.clickable(onClick = onShowReactions)
                    ) {
                        Text(
                            "${item.emoji} ${item.count}",
                            style = MaterialTheme.typography.labelSmall,
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp)
                        )
                    }
                }
            }
        }

        if (!isEditing) {
            Row(modifier = Modifier.padding(start = 40.dp)) {
                IconButton(onClick = onReact, modifier = Modifier.size(32.dp)) {
                    Icon(Icons.Default.Mood, contentDescription = "リアクション", modifier = Modifier.size(18.dp))
                }
                if (canEdit) {
                    IconButton(onClick = onEdit, modifier = Modifier.size(32.dp)) {
                        Icon(Icons.Default.Edit, contentDescription = "編集", modifier = Modifier.size(18.dp))
                    }
                }
            }
        }
    }
}
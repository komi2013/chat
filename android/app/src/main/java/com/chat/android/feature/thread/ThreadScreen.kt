package com.chat.android.feature.thread

import androidx.compose.foundation.ExperimentalFoundationApi
import androidx.compose.foundation.clickable
import androidx.compose.foundation.combinedClickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Bookmark
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.layout.ContentScale
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.core.util.Base62
import com.chat.android.core.util.EmojiCatalog
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.feature.channel.AliasAvatar
import com.chat.android.navigation.ThreadHeadRoute
import com.chat.android.navigation.ThreadRoute
import coil.compose.AsyncImage
import java.net.URLDecoder
import java.nio.charset.StandardCharsets
import java.text.SimpleDateFormat
import java.util.Locale

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
    backID: String? = null,
    viewModel: ThreadViewModel = hiltViewModel()
) {
    val state by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val uriHandler = LocalUriHandler.current

    var editingId by remember { mutableStateOf<String?>(null) }
    var editingText by remember { mutableStateOf("") }
    var emojiTarget by remember { mutableStateOf<String?>(null) }
    var emojiedTarget by remember { mutableStateOf<Pair<String, String>?>(null) }
    var optionsTarget by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(channelID, parentID, backID) {
        viewModel.init(channelID, parentID, backID)
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
                    IconButton(onClick = {
                        val backID = state.head?.backID.orEmpty()
                        if (backID.isNotBlank() && channelID != null) {
                            navController.navigate(ThreadRoute(channelID = channelID, parentID = backID)) {
                                popUpTo<ThreadRoute> { inclusive = true }
                            }
                        } else {
                            navController.navigateUp()
                        }
                    }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        }
    ) { padding ->
        Column(Modifier.fillMaxSize().padding(padding)) {
            if (state.isLoading) {
                LinearProgressIndicator(Modifier.fillMaxWidth())
            }
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
                            myAliasName = state.myname,
                            createdAt = head.updatedAt,
                            emojisJson = head.emojisJson,
                            isEditing = isEditing,
                            isBookmarked = head.isBookmarked,
                            onEmojiToggle = { emoji ->
                                viewModel.postEmoji(head.parentID, emoji, state.parentID)
                            },
                            onLongClickEmoji = { emoji -> emojiedTarget = head.parentID to emoji },
                            onLongClickMessage = { optionsTarget = head.parentID }
                        ) {
                            if (isEditing) {
                                EditBox(
                                    editingMessageId = editingId,
                                    initialText = editingText,
                                    canPost = state.canPost,
                                    isPosting = state.isPosting,
                                    postError = state.error,
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
                        myAliasName = state.myname,
                        createdAt = message.createdAt,
                        emojisJson = message.emojisJson,
                        isEditing = isEditing,
                        isBookmarked = message.isBookmarked,
                        onEmojiToggle = { emoji ->
                            viewModel.postEmoji(message.messageID, emoji, state.parentID)
                        },
                        onLongClickEmoji = { emoji -> emojiedTarget = message.messageID to emoji },
                        onLongClickMessage = { optionsTarget = message.messageID }
                    ) {
                        if (isEditing) {
                            EditBox(
                                editingMessageId = editingId,
                                initialText = editingText,
                                canPost = state.canPost,
                                isPosting = state.isPosting,
                                postError = state.error,
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
                    postError = state.error,
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
                viewModel.postEmoji(messageID, emoji, state.parentID)
                emojiTarget = null
            },
            onDismiss = { emojiTarget = null }
        )
    }

    emojiedTarget?.let { (messageID, emoji) ->
        val json = if (messageID == state.parentID) {
            state.head?.emojisJson ?: "[]"
        } else {
            state.messages.firstOrNull { it.messageID == messageID }?.emojisJson ?: "[]"
        }
        EmojiedModal(
            emojisJson = json,
            emojiFilter = emoji,
            onDismiss = { emojiedTarget = null }
        )
    }

    optionsTarget?.let { messageID ->
        val head = state.head
        val msg = if (messageID == head?.parentID) head else state.messages.find { it.messageID == messageID }
        if (msg == null) {
            optionsTarget = null
            return@let
        }

        // Get properties polymorphically
        val msgAliasName = when (msg) {
            is com.chat.android.feature.channel.DbThreadHead -> msg.aliasName
            is com.chat.android.feature.channel.DbThread -> msg.aliasName
            else -> ""
        }
        val msgText = when (msg) {
            is com.chat.android.feature.channel.DbThreadHead -> msg.messageTxt
            is com.chat.android.feature.channel.DbThread -> msg.messageTxt
            else -> ""
        }
        val msgBackID = when (msg) {
            is com.chat.android.feature.channel.DbThreadHead -> msg.backID
            is com.chat.android.feature.channel.DbThread -> msg.backID
            else -> ""
        }

        // vue/src/views/Messages.vue: editable() logic
        val createdAtUnix = Base62.decode(if (messageID.length > 1) messageID.dropLast(1) else messageID)
        val tenMinsMillis = 10 * 60 * 1000L
        val isTimeOk = (System.currentTimeMillis() - (createdAtUnix * 1000)) < tenMinsMillis

        val isOwner = msgAliasName == state.myname
        val isHead = messageID == state.parentID
        val editable = (isTimeOk && isOwner) || isHead
        val replyable = !isHead && msgBackID.isEmpty()
        val isBookmarked = when (msg) {
            is com.chat.android.feature.channel.DbThreadHead -> msg.isBookmarked
            is com.chat.android.feature.channel.DbThread -> msg.isBookmarked
            else -> false
        }

        MessageOptionsModal(
            show = true,
            editable = editable,
            replyable = replyable,
            onEmoji = { emojiTarget = messageID },
            onReply = {
                val replyMessage = msg as? com.chat.android.feature.channel.DbThread
                if (channelID != null && replyMessage != null) {
                    navController.navigate(
                        ThreadRoute(
                            channelID = channelID,
                            parentID = replyMessage.messageID,
                            backID = replyMessage.parentID
                        )
                    )
                }
            },
            onEdit = {
                editingId = messageID
                editingText = msgText
            },
            onDelete = {
                viewModel.postMessage("", editingMessageID = messageID, asDelete = true)
            },
            onBookmark = {
                viewModel.toggleBookmark(messageID)
            },
            isBookmarked = isBookmarked,
            onClose = { optionsTarget = null }
        )
    }
}

@OptIn(ExperimentalFoundationApi::class)
@Composable
private fun MessageRow(
    aliasName: String,
    aliasImg: String,
    myAliasName: String,
    createdAt: String,
    emojisJson: String,
    isEditing: Boolean,
    isBookmarked: Boolean = false,
    onEmojiToggle: (String) -> Unit,
    onLongClickEmoji: (String) -> Unit,
    onLongClickMessage: () -> Unit,
    content: @Composable () -> Unit
) {
    val displayTime = remember(createdAt) {
        runCatching {
            SimpleDateFormat("yyyy/MM/dd HH:mm:ss", Locale.getDefault())
                .parse(createdAt)
                ?.let { SimpleDateFormat("MM-dd HH:mm", Locale.getDefault()).format(it) }
        }.getOrNull() ?: createdAt
    }
    Column(
        modifier = Modifier
            .fillMaxWidth()
            .combinedClickable(onClick = {}, onLongClick = onLongClickMessage)
    ) {
        Column(
            modifier = if (isEditing) Modifier.fillMaxWidth() else {
                Modifier
                    .padding(vertical = 6.dp)
                    .fillMaxWidth()
                    .defaultMinSize(minHeight = 56.dp)
                    .padding(vertical = 6.dp)
            }
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                AliasAvatar(aliasImg = aliasImg, modifier = Modifier.size(32.dp))
                Spacer(Modifier.width(8.dp))
                Text(aliasName, style = MaterialTheme.typography.bodyMedium, fontWeight = FontWeight.SemiBold)
                if (isBookmarked) {
                    Spacer(Modifier.width(4.dp))
                    Icon(
                        Icons.Default.Bookmark,
                        contentDescription = null,
                        modifier = Modifier.size(14.dp),
                        tint = MaterialTheme.colorScheme.primary
                    )
                }
                Spacer(Modifier.width(8.dp))
                Text(displayTime, style = MaterialTheme.typography.labelSmall)
            }
            content()
        }

        // リアクションの集約表示（vue の calcEmoji 相当）
        val counts = EmojiCatalog.calcEmojis(emojisJson, myAliasName)
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
                        modifier = Modifier.combinedClickable(
                            onClick = { onEmojiToggle(item.emoji) },
                            onLongClick = { onLongClickEmoji(item.emoji) }
                        )
                    ) {
                        Row(
                            modifier = Modifier.padding(horizontal = 6.dp, vertical = 2.dp),
                            verticalAlignment = Alignment.CenterVertically,
                            horizontalArrangement = Arrangement.spacedBy(4.dp)
                        ) {
                            if (EmojiCatalog.isImageEmoji(item.emoji)) {
                                AsyncImage(
                                    model = item.emoji.toAbsoluteImageUrl(),
                                    contentDescription = "Reaction",
                                    modifier = Modifier.size(20.dp),
                                    contentScale = ContentScale.Fit
                                )
                            } else {
                                Text(item.emoji, style = MaterialTheme.typography.labelSmall)
                            }
                            Text(item.count.toString(), style = MaterialTheme.typography.labelSmall)
                        }
                    }
                }
            }
        }

        }
}

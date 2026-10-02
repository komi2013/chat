package com.chat.android.feature.thread

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.grid.GridCells
import androidx.compose.foundation.lazy.grid.LazyVerticalGrid
import androidx.compose.foundation.lazy.grid.items
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import coil.compose.AsyncImage
import com.chat.android.core.util.EmojiCatalog
import com.chat.android.core.util.toAbsoluteImageUrl
import org.json.JSONArray

/**
 * 絵文字ピッカー（vue/src/components/EmojiModal.vue に対応）。
 *
 * 選んだ絵文字はそのままサーバーへ送り、他クライアントにも push で届く。
 * 直近使った絵文字は先頭に再配置される（vue の rotateEmoji と同じ）。
 */
@Composable
fun EmojiModal(
    emojis: List<String>,
    errorMessage: String?,
    onSelect: (String) -> Unit,
    onDismiss: () -> Unit
) {
    var typed by remember { mutableStateOf("") }
    var error by remember(errorMessage) { mutableStateOf(errorMessage) }

    Dialog(onDismissRequest = onDismiss) {
        Surface(
            shape = RoundedCornerShape(12.dp),
            color = MaterialTheme.colorScheme.surface
        ) {
            Column(
                modifier = Modifier.padding(16.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                LazyVerticalGrid(
                    columns = GridCells.Fixed(8),
                    modifier = Modifier.height(240.dp),
                    horizontalArrangement = Arrangement.spacedBy(4.dp),
                    verticalArrangement = Arrangement.spacedBy(4.dp)
                ) {
                    items(emojis) { emoji ->
                        if (EmojiCatalog.isImageEmoji(emoji)) {
                            // 画像絵文字（/img/...）は Coil で読み込む
                            AsyncImage(
                                model = emoji.toAbsoluteImageUrl(),
                                contentDescription = emoji,
                                modifier = Modifier
                                    .size(32.dp)
                                    .clip(RoundedCornerShape(6.dp))
                                    .clickable { onSelect(emoji) }
                            )
                        } else {
                            Text(
                                text = emoji,
                                fontSize = 20.sp,
                                modifier = Modifier
                                    .clip(RoundedCornerShape(6.dp))
                                    .clickable { onSelect(emoji) }
                                    .padding(4.dp)
                            )
                        }
                    }
                }

                OutlinedTextField(
                    value = typed,
                    onValueChange = { if (it.length <= 2048) typed = it },
                    singleLine = true,
                    isError = error != null,
                    label = { Text("絵文字または画像URL") },
                    modifier = Modifier.fillMaxWidth()
                )
                if (EmojiCatalog.isImageEmoji(typed.trim())) {
                    AsyncImage(
                        model = typed.trim().toAbsoluteImageUrl(),
                        contentDescription = "画像リアクションのプレビュー",
                        modifier = Modifier.size(30.dp),
                        contentScale = androidx.compose.ui.layout.ContentScale.Fit
                    )
                }
                TextButton(onClick = {
                    val value = typed.trim()
                    if (EmojiCatalog.validateEmoji(value)) {
                        error = null
                        onSelect(value)
                    } else {
                        error = "絵文字か画像URLを入力してください"
                    }
                }) { Text("決定") }

                TextButton(onClick = onDismiss) { Text("閉じる") }

                error?.let {
                    Text(it, color = MaterialTheme.colorScheme.error, style = MaterialTheme.typography.bodySmall)
                }
            }
        }
    }
}

/**
 * リアクション一覧（vue/src/components/EmojiedModal.vue に対応）。
 * 誰がどの絵文字を付けたかを読み取り専用で表示する。
 */
@Composable
fun EmojiedModal(
    emojisJson: String,
    emojiFilter: String? = null,
    onDismiss: () -> Unit
) {
    val entries = remember(emojisJson, emojiFilter) {
        val array = runCatching { JSONArray(emojisJson) }.getOrNull()
        (0 until (array?.length() ?: 0)).mapNotNull { i ->
            array!!.optJSONObject(i)?.let {
                it.optString("aliasName", "") to it.optString("emoji", "")
            }
        }.filter { emojiFilter == null || it.second == emojiFilter }
    }

    Dialog(onDismissRequest = onDismiss) {
        Surface(shape = RoundedCornerShape(12.dp), color = MaterialTheme.colorScheme.surface) {
            Column(
                modifier = Modifier.padding(16.dp).heightIn(max = 400.dp)
            ) {
                emojiFilter?.let { emoji ->
                    Row(
                        modifier = Modifier.fillMaxWidth().padding(bottom = 8.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        if (EmojiCatalog.isImageEmoji(emoji)) {
                            AsyncImage(
                                model = emoji.toAbsoluteImageUrl(),
                                contentDescription = "Selected reaction",
                                modifier = Modifier.size(28.dp),
                                contentScale = androidx.compose.ui.layout.ContentScale.Fit
                            )
                        } else {
                            Text(emoji, fontSize = 22.sp)
                        }
                        Spacer(Modifier.width(8.dp))
                        Text("このリアクションを付けた人")
                    }
                }
                if (entries.isEmpty()) {
                    Text(
                        if (emojiFilter == null) "リアクションはありません" else "このリアクションはありません",
                        style = MaterialTheme.typography.bodyMedium
                    )
                } else {
                    entries.forEach { (aliasName, emoji) ->
                        Row(
                            modifier = Modifier.fillMaxWidth().padding(vertical = 6.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            if (emojiFilter == null && EmojiCatalog.isImageEmoji(emoji)) {
                                AsyncImage(
                                    model = emoji.toAbsoluteImageUrl(),
                                    contentDescription = "Reaction by $aliasName",
                                    modifier = Modifier.size(30.dp).padding(end = 8.dp),
                                    contentScale = androidx.compose.ui.layout.ContentScale.Fit
                                )
                            } else if (emojiFilter == null) {
                                Text(emoji, fontSize = 22.sp, modifier = Modifier.width(40.dp))
                            }
                            Text(aliasName, fontWeight = FontWeight.Normal)
                        }
                    }
                }
                TextButton(onClick = onDismiss, modifier = Modifier.fillMaxWidth()) {
                    Text("閉じる")
                }
            }
        }
    }
}

/**
 * 汎用モーダル枠（vue/src/components/EditOptionModal.vue に対応）。
 * 外側をタップすると閉じ、中に任意のコンテンツを差し込む。
 */
@Composable
fun EditOptionModal(
    show: Boolean,
    onClose: () -> Unit,
    content: @Composable () -> Unit
) {
    if (!show) return
    Dialog(onDismissRequest = onClose) {
        Surface(
            shape = RoundedCornerShape(12.dp),
            color = MaterialTheme.colorScheme.surface
        ) {
            Column(
                modifier = Modifier.padding(16.dp).width(260.dp)
            ) {
                content()
                TextButton(onClick = onClose, modifier = Modifier.fillMaxWidth()) {
                    Text("閉じる")
                }
            }
        }
    }
}

/**
 * メッセージの長押しメニュー（vue/src/views/Messages.vue の @contextmenu に相当）。
 * 返信、絵文字、編集、削除、ブックマークなどの操作を提供する。
 */
@Composable
fun MessageOptionsModal(
    show: Boolean,
    editable: Boolean,
    replyable: Boolean,
    onEmoji: () -> Unit,
    onReply: () -> Unit,
    onEdit: () -> Unit,
    onDelete: () -> Unit,
    onBookmark: () -> Unit,
    isBookmarked: Boolean,
    onClose: () -> Unit
) {
    EditOptionModal(show = show, onClose = onClose) {
        if (replyable) {
            TextButton(
                onClick = { onReply(); onClose() },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("返信する")
            }
        }
        // 絵文字リアクションも長押しメニューから選べるようにする。
        TextButton(
            onClick = { onEmoji(); onClose() },
            modifier = Modifier.fillMaxWidth()
        ) {
            Text("絵文字を追加")
        }
        if (editable) {
            TextButton(
                onClick = { onEdit(); onClose() },
                modifier = Modifier.fillMaxWidth()
            ) {
                Text("編集する")
            }
            TextButton(
                onClick = { onDelete(); onClose() },
                modifier = Modifier.fillMaxWidth(),
                colors = ButtonDefaults.textButtonColors(contentColor = MaterialTheme.colorScheme.error)
            ) {
                Text("削除する")
            }
        }
        TextButton(
            onClick = { onBookmark(); onClose() },
            modifier = Modifier.fillMaxWidth()
        ) {
            Text(if (isBookmarked) "ブックマーク解除" else "ブックマーク")
        }
    }
}

package com.chat.android.feature.thread

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import com.chat.android.feature.channel.MarkdownCodec

/**
 * メッセージ入力欄（vue/src/components/EditBox.vue に対応）。
 *
 * 編集時は元の本文を読み込み、削除ボタンで本文を空にして threadEdit を送る
 * （サーバーは messageTxt が空の threadEdit を削除として扱う）。
 */
@Composable
fun EditBox(
    editingMessageId: String?,
    initialText: String,
    canPost: Boolean,
    isPosting: Boolean,
    onPost: (text: String, editingMessageID: String?, asDelete: Boolean) -> Unit,
    onCancelEdit: () -> Unit
) {
    var text by remember(editingMessageId) { mutableStateOf(initialText) }
    var editMode by remember(editingMessageId) { mutableStateOf(editingMessageId != null) }

    if (!canPost) {
        Text(
            "ブロードキャスト設定のため管理者のみ投稿できます",
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.padding(8.dp)
        )
        return
    }

    Column(modifier = Modifier.fillMaxWidth().padding(8.dp)) {
        OutlinedTextField(
            value = text,
            onValueChange = { text = it },
            modifier = Modifier.fillMaxWidth(),
            minLines = 2,
            placeholder = { Text(if (editMode) "メッセージを編集" else "メッセージを入力") }
        )
        Row(
            modifier = Modifier.fillMaxWidth().padding(top = 4.dp),
            horizontalArrangement = Arrangement.End,
            verticalAlignment = Alignment.CenterVertically
        ) {
            if (editMode) {
                TextButton(onClick = onCancelEdit) { Text("編集をやめる") }
                TextButton(
                    onClick = {
                        editMode = false
                        onPost("", editingMessageId, true)
                    }
                ) {
                    Icon(Icons.Default.Delete, contentDescription = "削除")
                    Spacer(Modifier.width(4.dp))
                    Text("削除")
                }
            }
            Button(
                onClick = {
                    editMode = false
                    onPost(text, editingMessageId, false)
                },
                enabled = !isPosting && (text.isNotBlank() || editMode)
            ) {
                if (isPosting) {
                    CircularProgressIndicator(modifier = Modifier.size(16.dp), strokeWidth = 2.dp)
                } else {
                    Icon(Icons.Default.Send, contentDescription = null)
                    Spacer(Modifier.width(4.dp))
                    Text(if (editMode) "更新" else "投稿")
                }
            }
        }
    }
}

/** 本文をComposeのテキストへ描画する（MarkdownCodec を利用）。 */
@Composable
fun ThreadMessageText(text: String, modifier: Modifier = Modifier) {
    Text(text = MarkdownCodec.toAnnotatedString(text), modifier = modifier)
}
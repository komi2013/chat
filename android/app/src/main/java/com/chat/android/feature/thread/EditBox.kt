package com.chat.android.feature.thread

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.AttachFile
import androidx.compose.material.icons.filled.Bookmark
import androidx.compose.material.icons.filled.BookmarkBorder
import androidx.compose.material.icons.filled.Call
import androidx.compose.material.icons.filled.Close
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Groups
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage
import com.chat.android.core.util.toAbsoluteImageUrl
import androidx.compose.foundation.text.ClickableText
import androidx.compose.ui.platform.LocalUriHandler
import com.chat.android.feature.channel.AliasAvatar
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
import com.chat.android.feature.channel.MarkdownCodec

/** 「グループなし」を表す選択肢（vue の emptyGroup と同じ）。 */
private val NO_GROUP = DbGroup(
    groupID = "",
    channelID = "",
    groupName = "グループなし",
    groupImg = "",
    aliasNamesJson = "[]",
    groupBio = ""
)

/** 装飾ツールバーのボタン（Vue の Quill ツールバー相当）。
 * markdown.js と同じ記号を使う。Triple(表示名, 開始記号, 終了記号)。
 * リンクだけは選択文字列を「」で挟み、（）内にURLを書かせるため特別扱いする。 */
private val FORMAT_BUTTONS = listOf(
    Triple("太字", "＊太＊", "・＊太＊"),
    Triple("取消", "〜〜", "・〜〜"),
    Triple("引用", "＜quote＞", "＜・quote＞"),
    Triple("コード", "｀｀｀", "・｀｀｀"),
    Triple("赤", "色＊赤", "赤＊色")
)

/**
 * メッセージ入力欄（vue/src/components/EditBox.vue に対応）。
 *
 * 装飾ツールバー・添付（🌄）・ブックマーク（🔖）・グループ投稿（👥）・
 * 通話（☎️）・送信（▶️）・削除（🗑）を提供する。
 * 編集時は元の本文を読み込み、削除ボタンで本文を空にして threadEdit を送る
 * （サーバーは messageTxt が空の threadEdit を削除として扱う）。
 */
@Composable
fun EditBox(
    editingMessageId: String?,
    initialText: String,
    canPost: Boolean,
    isPosting: Boolean,
    isDm: Boolean,
    isNewThread: Boolean,
    myname: String,
    aliases: List<DbAlias>,
    groups: List<DbGroup>,
    onPost: (
        text: String,
        editingMessageID: String?,
        asDelete: Boolean,
        task: Boolean,
        group: DbGroup?,
        fileUris: List<Uri>
    ) -> Unit,
    onCall: () -> Unit,
    onCancelEdit: () -> Unit
) {
    // Vue と同じく、編集時は素のテキストを入力欄に読み込む。
    // markdown のタグを含めたまま編集できるようにするため removeMark は使用しない。
    val plainInitial = remember(editingMessageId) { initialText }
    var field by remember(editingMessageId) { mutableStateOf(TextFieldValue(plainInitial)) }
    var editMode by remember(editingMessageId) { mutableStateOf(editingMessageId != null) }
    var task by remember(editingMessageId) { mutableStateOf(false) }
    var selectedGroup by remember(editingMessageId) { mutableStateOf<DbGroup?>(null) }
    var showGroupModal by remember { mutableStateOf(false) }
    var files by remember(editingMessageId) { mutableStateOf<List<Uri>>(emptyList()) }
    var fileNames by remember(editingMessageId) { mutableStateOf<List<String>>(emptyList()) }
    var showMentions by remember { mutableStateOf(false) }
    var mentionQuery by remember { mutableStateOf("") }
    val context = LocalContext.current

    // 自分が所属するグループだけ選べる（vue の myGroups と同じ）。先頭は「グループなし」。
    val myGroups = remember(groups, myname) {
        listOf(NO_GROUP) + groups.filter { it.aliasNamesJson.toNameList().contains(myname) }
    }
    // ＠メンションの候補（エイリアス＋グループ。vue の mention source と同じ）。
    val mentionTargets = remember(aliases, groups) {
        val aliasItems = aliases.map { MentionTarget(it.aliasName, it.aliasImg) }
        val groupItems = groups.map { MentionTarget(it.groupName, it.groupImg) }
        aliasItems + groupItems
    }
    val mentionSuggestions = remember(mentionQuery, mentionTargets) {
        if (mentionQuery.isBlank()) mentionTargets
        else mentionTargets.filter { it.name.contains(mentionQuery, ignoreCase = true) }
    }

    fun refreshMentionState(value: TextFieldValue) {
        // 直前の ＠ からカーソルまでの文字列を候補の絞り込みに使う。
        val cursor = value.selection.start.coerceIn(0, value.text.length)
        val before = value.text.take(cursor)
        val at = before.lastIndexOf('＠')
        if (at < 0) {
            showMentions = false
            mentionQuery = ""
            return
        }
        val query = before.substring(at + 1, cursor)
        // 空白や改行を含む場合はメンション入力中とみなさない（quill-mention 相当）。
        if (query.contains(' ') || query.contains('　') || query.contains('\n')) {
            showMentions = false
            mentionQuery = ""
            return
        }
        showMentions = true
        mentionQuery = query
    }

    fun insertMention(target: MentionTarget) {
        // カーソル直前の ＠... を ＠＠名前・＠＠ に置き換える。
        val cursor = field.selection.start.coerceIn(0, field.text.length)
        val before = field.text.take(cursor)
        val at = before.lastIndexOf('＠')
        val start = if (at >= 0) at else cursor
        val mention = "＠＠${target.name}・＠＠"
        val newText = field.text.replaceRange(start, cursor, mention)
        val newCursor = start + mention.length
        field = field.copy(text = newText, selection = TextRange(newCursor))
        showMentions = false
        mentionQuery = ""
    }

    val filePicker = rememberLauncherForActivityResult(
        ActivityResultContracts.GetMultipleContents()
    ) { uris ->
        if (uris.isEmpty()) return@rememberLauncherForActivityResult
        files = files + uris
        val resolver = context.contentResolver
        val names = uris.map { uri ->
            runCatching {
                resolver.query(uri, null, null, null, null)?.use { cursor ->
                    val index = cursor.getColumnIndex(android.provider.OpenableColumns.DISPLAY_NAME)
                    if (index >= 0 && cursor.moveToFirst()) cursor.getString(index) else null
                }
            }.getOrNull() ?: uri.lastPathSegment.orEmpty().ifBlank { "file" }
        }
        // Vue の fileInfo と同じく、選んだ順にファイル名を並べて残す。
        fileNames = fileNames + names
    }

    if (!canPost) {
        Text(
            "ブロードキャスト設定のため管理者のみ投稿できます",
            style = MaterialTheme.typography.bodySmall,
            modifier = Modifier.padding(8.dp)
        )
        return
    }

    Column(modifier = Modifier.fillMaxWidth().padding(8.dp)) {
        // ツールバー（vue の EditBox toolbar 相当）
        Row(
            modifier = Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(4.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            FORMAT_BUTTONS.forEach { (label, open, close) ->
                OutlinedButton(
                    onClick = { field = wrapSelection(field, open, close) },
                    contentPadding = PaddingValues(horizontal = 8.dp, vertical = 4.dp)
                ) { Text(label, style = MaterialTheme.typography.labelSmall) }
            }
            // リンクは「選択文字列」（URL）の形（markdown.js の applyLinks の逆）にする。
            // 選択なしのときはカーソル位置に「」（）を挿入し、（）の中にカーソルを置く。
            OutlinedButton(
                onClick = { field = wrapLink(field) },
                contentPadding = PaddingValues(horizontal = 8.dp, vertical = 4.dp)
            ) { Text("リンク", style = MaterialTheme.typography.labelSmall) }
            IconButton(onClick = { filePicker.launch("*/*") }) {
                Icon(Icons.Default.AttachFile, contentDescription = "添付")
            }
            if (editMode) {
                IconButton(onClick = {
                    editMode = false
                    onPost("", editingMessageId, true, false, null, emptyList())
                }) {
                    Icon(Icons.Default.Delete, contentDescription = "削除")
                }
            }
            IconButton(onClick = { task = !task }) {
                Icon(
                    imageVector = if (task) Icons.Default.Bookmark else Icons.Default.BookmarkBorder,
                    contentDescription = "タスク",
                    tint = if (task) MaterialTheme.colorScheme.primary else LocalContentColor.current
                )
            }
            IconButton(onClick = { showGroupModal = true }) {
                Icon(
                    imageVector = Icons.Default.Groups,
                    contentDescription = "グループ",
                    tint = if (selectedGroup != null) MaterialTheme.colorScheme.primary
                    else LocalContentColor.current
                )
            }
            if (isDm && !isNewThread) {
                IconButton(onClick = onCall) {
                    Icon(Icons.Default.Call, contentDescription = "通話")
                }
            }
            Button(
                onClick = {
                    editMode = false
                    onPost(field.text, editingMessageId, false, task, selectedGroup, files)
                },
                enabled = !isPosting &&
                    (field.text.isNotBlank() || editMode || files.isNotEmpty())
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

        OutlinedTextField(
            value = field,
            onValueChange = { newValue ->
                field = newValue
                refreshMentionState(newValue)
            },
            modifier = Modifier.fillMaxWidth(),
            minLines = 2,
            placeholder = { Text(if (editMode) "メッセージを編集" else "メッセージを入力") },
            visualTransformation = { text ->
                androidx.compose.ui.text.input.TransformedText(
                    MarkdownCodec.styleTags(text.text),
                    androidx.compose.ui.text.input.OffsetMapping.Identity
                )
            }
        )
        if (showMentions && mentionSuggestions.isNotEmpty()) {
            LazyColumn(
                modifier = Modifier
                    .fillMaxWidth()
                    .heightIn(max = 180.dp)
                    .background(
                        MaterialTheme.colorScheme.surfaceVariant,
                        RoundedCornerShape(8.dp)
                    )
            ) {
                items(mentionSuggestions) { target ->
                    Row(
                        modifier = Modifier
                            .fillMaxWidth()
                            .clickable { insertMention(target) }
                            .padding(horizontal = 8.dp, vertical = 6.dp),
                        verticalAlignment = Alignment.CenterVertically
                    ) {
                        AliasAvatar(aliasImg = target.icon, modifier = Modifier.size(24.dp))
                        Spacer(Modifier.width(8.dp))
                        Text(target.name, style = MaterialTheme.typography.bodyMedium)
                    }
                }
            }
        }
        if (editMode) {
            TextButton(onClick = onCancelEdit) { Text("編集をやめる") }
        }

        if (files.isNotEmpty()) {
            Row(
                modifier = Modifier.fillMaxWidth().horizontalScroll(rememberScrollState()),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                files.forEachIndexed { index, uri ->
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Box {
                            AsyncImage(
                                model = uri,
                                contentDescription = null,
                                modifier = Modifier.size(48.dp)
                            )
                            IconButton(
                                onClick = {
                                    files = files.filterIndexed { i, _ -> i != index }
                                    fileNames = fileNames.filterIndexed { i, _ -> i != index }
                                },
                                modifier = Modifier.size(18.dp)
                            ) {
                                Icon(
                                    Icons.Default.Close,
                                    contentDescription = "添付を削除",
                                    modifier = Modifier.size(14.dp)
                                )
                            }
                        }
                        Text(
                            fileNames.getOrNull(index).orEmpty(),
                            style = MaterialTheme.typography.labelSmall,
                            maxLines = 1
                        )
                    }
                }
            }
        }
    }

    // グループ選択（vue の EditOptionModal + SelectGroup 相当）
    EditOptionModal(show = showGroupModal, onClose = { showGroupModal = false }) {
        Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
            myGroups.forEach { group ->
                Row(
                    modifier = Modifier
                        .fillMaxWidth()
                        .clickable {
                            selectedGroup = if (group.groupID.isEmpty()) null else group
                            showGroupModal = false
                        }
                        .padding(8.dp),
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Icon(Icons.Default.Groups, contentDescription = null)
                    Spacer(Modifier.width(8.dp))
                    Text(
                        group.groupName,
                        fontWeight = if (selectedGroup?.groupID == group.groupID) FontWeight.Bold
                        else FontWeight.Normal
                    )
                }
            }
        }
    }
}

/** 選択範囲を装飾記号で挟む。選択が空ならカーソル位置（末尾ではない）に挿入する。 */
private fun wrapSelection(value: TextFieldValue, open: String, close: String): TextFieldValue {
    val text = value.text
    val start = value.selection.start.coerceIn(0, text.length)
    val end = value.selection.end.coerceIn(0, text.length)
    val (from, to) = if (start <= end) start to end else end to start
    if (from == to) {
        // 空選択：カーソル位置に open + close を挿入し、間にカーソルを置く。
        val newText = text.substring(0, from) + open + close + text.substring(from)
        val cursor = from + open.length
        return value.copy(text = newText, selection = TextRange(cursor))
    }
    val selected = text.substring(from, to)
    val newText = text.substring(0, from) + open + selected + close + text.substring(to)
    val newStart = from + open.length
    return value.copy(
        text = newText,
        selection = TextRange(newStart, newStart + selected.length)
    )
}

/** メンション候補の1件（エイリアス名またはグループ名とそのアイコン）。 */
private data class MentionTarget(val name: String, val icon: String)

/**
 * 選択文字列を「選択」（）のリンク形にする。
 * 空選択ならカーソル位置に「」（）を入れて（）の中へカーソルを移動する。
 */
private fun wrapLink(value: TextFieldValue): TextFieldValue {
    val text = value.text
    val start = value.selection.start.coerceIn(0, text.length)
    val end = value.selection.end.coerceIn(0, text.length)
    val (from, to) = if (start <= end) start to end else end to start
    if (from == to) {
        val insert = "「」（）"
        val newText = text.substring(0, from) + insert + text.substring(from)
        // （ の直後（＝ ）の前）にカーソルを置く
        val cursor = from + "「」".length + 1
        return value.copy(text = newText, selection = TextRange(cursor))
    }
    val selected = text.substring(from, to)
    val insert = "「$selected」（）"
    val newText = text.substring(0, from) + insert + text.substring(to)
    // （）の中にカーソル
    val cursor = from + insert.length - 1
    return value.copy(text = newText, selection = TextRange(cursor))
}

/** カーソル位置へ文字列を挿入する（絵文字・メンション用）。 */
private fun insertAtCursor(value: TextFieldValue, insert: String): TextFieldValue {
    val text = value.text
    val start = value.selection.start.coerceIn(0, text.length)
    val end = value.selection.end.coerceIn(0, text.length)
    val (from, to) = if (start <= end) start to end else end to start
    val newText = text.substring(0, from) + insert + text.substring(to)
    val cursor = from + insert.length
    return value.copy(text = newText, selection = TextRange(cursor))
}

/** 本文をComposeのテキストへ描画する（MarkdownCodec を利用）。リンクのタップに対応。 */
@Composable
fun ThreadMessageText(text: String, modifier: Modifier = Modifier) {
    val annotatedString = MarkdownCodec.toAnnotatedString(text)
    val uriHandler = LocalUriHandler.current

    ClickableText(
        text = annotatedString,
        modifier = modifier,
        style = LocalTextStyle.current.copy(color = LocalContentColor.current),
        onClick = { offset ->
            annotatedString.getStringAnnotations("URL", offset, offset)
                .firstOrNull()?.let { annotation ->
                    runCatching { uriHandler.openUri(annotation.item) }
                }
        }
    )
}

package com.chat.android.feature.thread

import android.net.Uri
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.ClickableText
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
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.layout.ContentScale
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import coil.compose.AsyncImage
import com.chat.android.core.util.MarkdownFormatRepair
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.feature.channel.AliasAvatar
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
import com.chat.android.feature.channel.MarkdownCodec
import java.net.URLDecoder
import java.nio.charset.StandardCharsets

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
 * emoji は表示アイコン、open/close は挿入する記号（markdown.js と同じ）。
 * 引用だけは選択行の先頭に「> 」を付ける専用処理にするため isQuote で区別する。
 * リンクだけは選択文字列を「」で挟み、（）内にURLを書かせるため特別扱いする。 */
private data class FormatButton(
    val emoji: String,
    val open: String,
    val close: String,
    val isQuote: Boolean = false
)

private val FORMAT_BUTTONS = listOf(
    FormatButton("🅱️", "**", "**"),
    FormatButton("🆂", "~~", "~~"),
    FormatButton("💬", "> ", "\n", isQuote = true),
    FormatButton("💻", "```\n", "\n```"),
    FormatButton("🔴", "色＊赤", "赤＊色")
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
    postError: String? = null,
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
    // initialText が変わったとき（返信時など）にもリセットされるようにする。
    val plainInitial = remember(editingMessageId, initialText) { initialText }
    var field by remember(editingMessageId, initialText) { mutableStateOf(TextFieldValue(plainInitial)) }
    var editMode by remember(editingMessageId) { mutableStateOf(editingMessageId != null) }
    var task by remember(editingMessageId) { mutableStateOf(false) }
    var selectedGroup by remember(editingMessageId) { mutableStateOf<DbGroup?>(null) }
    var showGroupModal by remember { mutableStateOf(false) }
    var files by remember(editingMessageId) { mutableStateOf<List<Uri>>(emptyList()) }
    var fileNames by remember(editingMessageId) { mutableStateOf<List<String>>(emptyList()) }
    var showMentions by remember { mutableStateOf(false) }
    var mentionQuery by remember { mutableStateOf("") }
    var wasPosting by remember { mutableStateOf(false) }
    val context = LocalContext.current

    LaunchedEffect(isPosting, postError) {
        if (wasPosting && !isPosting && postError == null) {
            field = TextFieldValue("")
            task = false
            selectedGroup = null
            files = emptyList()
            fileNames = emptyList()
            showMentions = false
            mentionQuery = ""
        }
        wasPosting = isPosting
    }

    // 自分が所属するグループだけ選べる（vue の myGroups と同じ）。先頭は「グループなし」。
    val myGroups = remember(groups, myname) {
        listOf(NO_GROUP) + groups.filter { it.aliasNamesJson.toNameList().contains(myname) }
    }
    // ＠メンションの候補（エイリアス＋グループ。vue の mention source と同じ）。
    val mentionTargets = remember(aliases, groups) {
        val aliasItems = aliases.map { MentionTarget(it.aliasName, it.aliasImg, isGroup = false) }
        val groupItems = groups.map { MentionTarget(it.groupName, it.groupImg, isGroup = true) }
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
        var at = before.lastIndexOfAny(charArrayOf('@', '＠'))
        while (at > 0 && (before[at - 1] == '＠' || before[at - 1] == '・')) {
            at = before.lastIndexOfAny(charArrayOf('@', '＠'), at - 1)
        }
        if (at < 0) {
            showMentions = false
            mentionQuery = ""
            return
        }
        val query = before.substring(at + 1, cursor)
        // 空白や改行を含む場合はメンション入力中とみなさない（quill-mention 相当）。
        if (query.contains(' ') || query.contains('　') || query.contains('\n') || query.contains('・')) {
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
        val at = before.lastIndexOfAny(charArrayOf('@', '＠'))
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
            // 投稿ボタンは右端ではなく左端に置く（横スクロールで隠れないように）。
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
            FORMAT_BUTTONS.forEach { button ->
                OutlinedButton(
                    onClick = { field = if (button.isQuote) quoteSelection(field) else wrapSelection(field, button.open, button.close) },
                    contentPadding = PaddingValues(horizontal = 8.dp, vertical = 4.dp)
                ) { Text(button.emoji, style = MaterialTheme.typography.labelSmall) }
            }
            // リンクは「選択文字列」（URL）の形（markdown.js の applyLinks の逆）にする。
            // 選択なしのときはカーソル位置に「」（）を挿入し、（）の中にカーソルを置く。
            OutlinedButton(
                onClick = { field = wrapLink(field) },
                contentPadding = PaddingValues(horizontal = 8.dp, vertical = 4.dp)
            ) { Text("🔗", style = MaterialTheme.typography.labelSmall) }
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
        }

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
                        Text(
                            target.name,
                            modifier = Modifier.weight(1f),
                            style = MaterialTheme.typography.bodyMedium
                        )
                        Text(
                            if (target.isGroup) "グループ" else "ユーザー",
                            style = MaterialTheme.typography.labelSmall,
                            color = MaterialTheme.colorScheme.onSurfaceVariant
                        )
                    }
                }
            }
        }

        OutlinedTextField(
            value = field,
            onValueChange = { newValue ->
                val oldText = field.text
                field = MarkdownFormatRepair.repair(newValue)
                val newText = field.text

                // 「＠」が入力されたとき、または「＠」以降で入力中のときにメンションを表示
                if (newText.length > oldText.length) {
                    val cursor = field.selection.start
                    if (cursor > 0) {
                        val added = newText.substring(cursor - 1, cursor)
                        if (added == "＠" || added == "@") {
                            // @ も ＠ に変換して扱う（vue parity）
                            if (added == "@") {
                                field = field.copy(text = newText.replaceRange(cursor - 1, cursor, "＠"))
                            }
                            showMentions = true
                        }
                    }
                }
                refreshMentionState(field)
            },
            modifier = Modifier.fillMaxWidth(),
            minLines = 2,
            placeholder = { Text(if (editMode) "メッセージを編集" else "メッセージを入力") },
            visualTransformation = { text -> MarkdownCodec.editorText(text.text) }
        )
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

private fun quoteSelection(value: TextFieldValue): TextFieldValue {
    val text = value.text
    val start = value.selection.start.coerceIn(0, text.length)
    val end = value.selection.end.coerceIn(0, text.length)
    val (from, to) = if (start <= end) start to end else end to start
    if (from == to) {
        val newText = text.substring(0, from) + "> " + text.substring(from)
        return value.copy(text = newText, selection = TextRange(from + 2))
    }
    val selected = text.substring(from, to)
    val quoted = selected.lineSequence().joinToString("\n") { "> $it" }
    val newText = text.substring(0, from) + quoted + text.substring(to)
    val insertedPrefixLength = quoted.length - selected.length
    return value.copy(
        text = newText,
        selection = TextRange(from + 2, to + insertedPrefixLength)
    )
}

/** メンション候補の1件（エイリアス名またはグループ名とそのアイコン）。 */
private data class MentionTarget(val name: String, val icon: String, val isGroup: Boolean)

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
        val insert = "[]()"
        val newText = text.substring(0, from) + insert + text.substring(from)
        // Put the cursor inside the destination parentheses.
        val cursor = from + "[](".length
        return value.copy(text = newText, selection = TextRange(cursor))
    }
    val selected = text.substring(from, to)
    val insert = "[$selected]()"
    val newText = text.substring(0, from) + insert + text.substring(to)
    // Place the cursor inside the destination parentheses.
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
    val parts = remember(text) { parseThreadMessageParts(text) }
    val uriHandler = LocalUriHandler.current

    Column(modifier = modifier) {
        parts.forEach { part ->
            when (part.kind) {
                ThreadMessagePartKind.TEXT -> {
                    val annotated = MarkdownCodec.toAnnotatedString(part.value)
                    ClickableText(
                        text = annotated,
                        modifier = Modifier.fillMaxWidth(),
                        style = LocalTextStyle.current.copy(color = LocalContentColor.current),
                        onClick = { offset ->
                            annotated.getStringAnnotations("URL", offset, offset)
                                .firstOrNull()?.let { annotation ->
                                    val url = annotation.item.trim()
                                    if (url.startsWith("https://", true) || url.startsWith("http://", true) || url.startsWith("mailto:", true) || url.startsWith("tel:", true) || url.startsWith("/")) {
                                        runCatching { uriHandler.openUri(url.toAbsoluteImageUrl()) }
                                    }
                                }
                        }
                    )
                }
                ThreadMessagePartKind.IMAGE -> {
                    AsyncImage(
                        model = part.value.toAbsoluteImageUrl(),
                        contentDescription = "Attached image",
                        modifier = Modifier
                            .padding(vertical = 4.dp)
                            .sizeIn(maxWidth = 250.dp, maxHeight = 250.dp)
                            .wrapContentSize(),
                        contentScale = ContentScale.Fit
                    )
                }
                ThreadMessagePartKind.FILE -> {
                    val fileName = runCatching {
                        URLDecoder.decode(
                            part.value.substringBefore('?').substringAfterLast('/'),
                            StandardCharsets.UTF_8.name()
                        )
                    }.getOrDefault("Download file").ifBlank { "Download file" }
                    Text(
                        text = "📎 $fileName",
                        color = MaterialTheme.colorScheme.primary,
                        modifier = Modifier.clickable {
                            if (isSafeMessageUrl(part.value)) {
                                runCatching { uriHandler.openUri(part.value.toAbsoluteImageUrl()) }
                            }
                        }.padding(vertical = 4.dp)
                    )
                }
            }
        }
    }
}

private enum class ThreadMessagePartKind { TEXT, IMAGE, FILE }

private data class ThreadMessagePart(val kind: ThreadMessagePartKind, val value: String)

private fun parseThreadMessageParts(text: String): List<ThreadMessagePart> {
    val parts = mutableListOf<ThreadMessagePart>()
    val attachmentPattern = Regex("＊(img|f)＊(.*?)・＊\\1＊|!\\[[^\\]]*\\]\\(([^)]+)\\)", RegexOption.DOT_MATCHES_ALL)
    var cursor = 0
    attachmentPattern.findAll(text).forEach { match ->
        if (match.range.first > cursor) {
            parts += ThreadMessagePart(
                ThreadMessagePartKind.TEXT,
                text.substring(cursor, match.range.first)
            )
        }
        val markdownImageUrl = match.groups[3]?.value
        val attachmentUrl = (markdownImageUrl ?: match.groupValues[2]).trim()
        val isImage = markdownImageUrl != null || match.groupValues[1] == "img"
        if (isImage && !isSafeMessageUrl(attachmentUrl, allowMailAndTel = false)) {
            parts += ThreadMessagePart(ThreadMessagePartKind.TEXT, match.value)
            cursor = match.range.last + 1
            return@forEach
        }
        val kind = if (isImage) {
            ThreadMessagePartKind.IMAGE
        } else {
            ThreadMessagePartKind.FILE
        }
        parts += ThreadMessagePart(kind, attachmentUrl)
        cursor = match.range.last + 1
    }
    if (cursor < text.length || parts.isEmpty()) {
        parts += ThreadMessagePart(ThreadMessagePartKind.TEXT, text.substring(cursor))
    }
    return parts
}

private fun isSafeMessageUrl(value: String, allowMailAndTel: Boolean = true): Boolean {
    val url = value.trim()
    if (url.startsWith("https://", true) || url.startsWith("http://", true)) return true
    if (url.startsWith("/") && !url.startsWith("//")) return true
    if (allowMailAndTel && (url.startsWith("mailto:", true) || url.startsWith("tel:", true))) return true
    return !Regex("^[a-z][a-z0-9+.-]*:", RegexOption.IGNORE_CASE).containsMatchIn(url) && !url.startsWith("//")
}

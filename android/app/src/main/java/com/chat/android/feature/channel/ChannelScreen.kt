package com.chat.android.feature.channel

import android.content.ClipData
import android.content.ClipboardManager
import android.content.Context
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.LazyListScope
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.ContentCopy
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material.icons.filled.Groups
import androidx.compose.material.icons.filled.Save
import androidx.compose.material.icons.filled.Send
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.runtime.saveable.rememberSaveable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.BuildConfig
import com.chat.android.core.ui.component.QrCodeImage
import com.chat.android.core.util.toAbsoluteImageUrl
import com.chat.android.navigation.GroupRoute

/**
 * チャネル設定画面（vue/src/views/Channel.vue に対応）。
 *
 * Vue と同じく 1 画面に「詳細・編集」と「チャネル一覧」を置く。
 * 詳細・編集フォームを上に、ローカルSQLite（channel テーブル）から読んだ
 * チャネル一覧を下に並べ、一項目のタップで詳細が切り替わる。
 * 「保存」で ChannelEdit/ へ POST する（ChannelRepository 経由）。
 */
@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ChannelScreen(
    navController: NavController,
    id: String? = null,
    viewModel: ChannelViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()

    // ルート引数は初期選択としてだけ使う。以降は一覧のタップで切り替える。
    var initialized by rememberSaveable { mutableStateOf(false) }
    LaunchedEffect(id) {
        if (!initialized) {
            initialized = true
            viewModel.init(id)
        }
    }

    var guest by rememberSaveable { mutableStateOf(false) }
    var showDeleteDialog by remember { mutableStateOf(false) }
    var showAvatarPicker by remember { mutableStateOf(false) }

    // 説明は TextFieldValue で持ち、ツールバーで選択範囲を装飾できるようにする。
    var descriptionField by remember(uiState.channelID, uiState.isCreateMode) {
        mutableStateOf(TextFieldValue(uiState.channelDescription))
    }
    LaunchedEffect(uiState.channelDescription) {
        if (uiState.channelDescription != descriptionField.text) {
            descriptionField = descriptionField.copy(text = uiState.channelDescription)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        when {
                            uiState.isCreateMode -> "組織・チャネル登録"
                            else -> uiState.channelName.ifBlank { "チャネル設定" }
                        },
                        maxLines = 1
                    )
                },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                },
                actions = {
                    if (!uiState.isCreateMode && !uiState.iamGuest) {
                        IconButton(
                            onClick = { viewModel.generateInvitation(guest) },
                            enabled = !uiState.isLoading
                        ) {
                            Icon(Icons.Default.Send, contentDescription = "招待URL")
                        }
                    }
                    if (!uiState.iamGuest) {
                        IconButton(
                            onClick = { viewModel.save() },
                            enabled = !uiState.isLoading
                        ) {
                            Icon(Icons.Default.Save, contentDescription = "保存")
                        }
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
            if (uiState.isLoading) {
                item { LinearProgressIndicator(modifier = Modifier.fillMaxWidth()) }
            }
            uiState.error?.let { item { MessageBanner(it, MaterialTheme.colorScheme.error) } }
            uiState.successMessage?.let {
                item { MessageBanner(it, MaterialTheme.colorScheme.primary) }
            }

            if (uiState.isCreateMode) {
                item { SectionCard("新規チャネル") { Text("チャネル情報を入力してください") } }
            } else {
                item { ChannelHeaderCard(uiState) }
            }
            item {
                OutlinedTextField(
                    value = uiState.channelName,
                    onValueChange = viewModel::onNameChange,
                    label = { Text("組織・チャネル名") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth(),
                    readOnly = uiState.iamGuest
                )
            }

            if (uiState.isCreateMode) {
                item {
                    OutlinedTextField(
                        value = uiState.myName,
                        onValueChange = viewModel::onMyNameChange,
                        label = { Text("このチャネルのニックネーム") },
                        singleLine = true,
                        modifier = Modifier.fillMaxWidth()
                    )
                }
                item { AvatarPickerRow(uiState.myImg) { showAvatarPicker = true } }
            } else {
                item {
                    Text(
                        "このチャネルのニックネーム: ${uiState.myName}",
                        style = MaterialTheme.typography.bodyMedium
                    )
                }
            }

            item {
                MarkdownToolbar(enabled = !uiState.iamGuest) { open, close ->
                    descriptionField = wrapSelection(descriptionField, open, close)
                }
            }

            item {
                OutlinedTextField(
                    value = descriptionField,
                    onValueChange = {
                        descriptionField = it
                        viewModel.onDescriptionChange(it.text)
                    },
                    label = { Text("説明") },
                    modifier = Modifier.fillMaxWidth().heightIn(min = 150.dp),
                    readOnly = uiState.iamGuest,
                    minLines = 6
                )
            }

            if (descriptionField.text.isNotBlank()) {
                item {
                    SectionCard("プレビュー") {
                        Text(MarkdownCodec.toAnnotatedString(descriptionField.text))
                    }
                }
            }

            if (!uiState.isCreateMode) {
                item {
                    InvitationCard(
                        state = uiState,
                        guest = guest,
                        onGuestChange = { guest = it },
                        onGenerate = { viewModel.generateInvitation(it) }
                    )
                }

                MemberSection(uiState)
                GroupSection(uiState, navController)

                if (uiState.iamAdmin) {
                    item {
                        OutlinedButton(
                            onClick = { showDeleteDialog = true },
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Icon(Icons.Default.Delete, contentDescription = null)
                            Spacer(Modifier.width(8.dp))
                            Text("このチャネルの削除")
                        }
                    }
                }
            }

            item { Divider() }

            // ローカルSQLite の channel テーブルから読んだ一覧。
            // タップで詳細が切り替わり、そのまま編集・保存できる。
            ChannelListSection(
                channels = uiState.channels,
                selectedId = uiState.channelID,
                isCreateMode = uiState.isCreateMode,
                onSelect = { viewModel.selectChannel(it) },
                onCreate = { viewModel.startCreateMode() }
            )
        }
    }

    if (showDeleteDialog) {
        AlertDialog(
            onDismissRequest = { showDeleteDialog = false },
            title = { Text("このチャネルの削除") },
            text = { Text("本当に削除しますか？この操作は取り消せません。") },
            confirmButton = {
                TextButton(onClick = {
                    showDeleteDialog = false
                    viewModel.deleteChannel()
                    navController.navigateUp()
                }) { Text("削除") }
            },
            dismissButton = {
                TextButton(onClick = { showDeleteDialog = false }) { Text("キャンセル") }
            }
        )
    }

    if (showAvatarPicker) {
        AvatarPickerDialog(
            current = uiState.myImg,
            onSelect = { viewModel.onMyImgChange(it); showAvatarPicker = false },
            onDismiss = { showAvatarPicker = false }
        )
    }
}

/**
 * チャネル一覧（SQLite の channel テーブルから読んだもの）。
 * 項目をタップすると詳細が切り替わり、そのまま編集・保存できる。
 */
private fun LazyListScope.ChannelListSection(
    channels: List<DbChannel>,
    selectedId: String?,
    isCreateMode: Boolean,
    onSelect: (String) -> Unit,
    onCreate: () -> Unit
) {
    item { SectionTitle("チャネル一覧") }

    if (channels.isEmpty()) {
        item { Text("チャネルがありません", style = MaterialTheme.typography.bodyMedium) }
    }

    items(channels, key = { it.channelID }) { channel ->
        val selected = !isCreateMode && channel.channelID == selectedId
        Surface(
            tonalElevation = if (selected) 4.dp else 0.dp,
            color = if (selected) MaterialTheme.colorScheme.secondaryContainer
            else MaterialTheme.colorScheme.surface,
            modifier = Modifier.fillMaxWidth(),
            onClick = { onSelect(channel.channelID) }
        ) {
            Row(
                modifier = Modifier.padding(12.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                AliasAvatar(
                    aliasImg = channel.myimg.toAbsoluteImageUrl(),
                    modifier = Modifier.size(40.dp)
                )
                Spacer(Modifier.width(12.dp))
                Column(Modifier.weight(1f)) {
                    Text(channel.channelName, style = MaterialTheme.typography.titleSmall)
                    if (channel.channelDescription.isNotBlank()) {
                        Text(
                            channel.channelDescription.take(40),
                            style = MaterialTheme.typography.bodySmall,
                            maxLines = 1
                        )
                    }
                }
                if (selected) {
                    Icon(Icons.Default.Check, contentDescription = "選択中")
                }
            }
        }
    }

    item {
        OutlinedButton(onClick = onCreate, modifier = Modifier.fillMaxWidth()) {
            Icon(Icons.Default.Add, contentDescription = null)
            Spacer(Modifier.width(8.dp))
            Text("新規")
        }
    }
}

/** 選択中チャネルの概要カード。 */
@Composable
private fun ChannelHeaderCard(state: ChannelUiState) {
    SectionCard("チャネル情報") {
        Row(verticalAlignment = Alignment.CenterVertically) {
            AliasAvatar(
                aliasImg = state.myImg.toAbsoluteImageUrl(),
                modifier = Modifier.size(56.dp)
            )
            Spacer(Modifier.width(12.dp))
            Column(Modifier.weight(1f)) {
                Text(
                    state.channelName.ifBlank { "(名称未設定)" },
                    style = MaterialTheme.typography.titleMedium
                )
                Text("ID: ${state.channelID}", style = MaterialTheme.typography.bodySmall)
                Text("表示名: ${state.myName}", style = MaterialTheme.typography.bodySmall)
                if (state.iamAdmin) {
                    Text(
                        "管理者",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
                if (state.iamGuest) {
                    Text(
                        "ゲスト（保存できません）",
                        style = MaterialTheme.typography.labelMedium,
                        color = MaterialTheme.colorScheme.error
                    )
                }
            }
        }
    }
}

/**
 * 招待URLの表示・コピー・QRコード。
 *
 * コードはDBに保存しないため、「招待する」を押したときだけ生成され、
 * 画面内にだけ存在する。画面を再訪すると消える（＝共有済みのURLは失効する）。
 */
@Composable
private fun InvitationCard(
    state: ChannelUiState,
    guest: Boolean,
    onGuestChange: (Boolean) -> Unit,
    onGenerate: (Boolean) -> Unit
) {
    var showConfirm by remember { mutableStateOf(false) }
    val code = if (guest) state.invitationGuestCode else state.invitationCode

    SectionCard("招待URL") {
        Row(verticalAlignment = Alignment.CenterVertically) {
            Checkbox(checked = guest, onCheckedChange = onGuestChange)
            Text("ゲスト用")
        }

        if (code.isBlank()) {
            Text(
                "未生成です。「招待する」を押すとリンクとQRコードが発行されます。",
                style = MaterialTheme.typography.bodyMedium
            )
        } else {
            val url = invitationUrl(state.channelID, code)
            val webUrl = webInvitationUrl(state.channelID, code)
            val context = LocalContext.current
            Text(url, style = MaterialTheme.typography.bodySmall)

            TextButton(onClick = {
                val clipboard = context
                    .getSystemService(Context.CLIPBOARD_SERVICE) as? ClipboardManager
                clipboard?.setPrimaryClip(ClipData.newPlainText("channel", url))
            }) {
                Icon(Icons.Default.ContentCopy, contentDescription = null)
                Spacer(Modifier.width(4.dp))
                Text("コピー")
            }

            // このURLのQRコード。fillMaxWidth() を付けると親幅で伸びて歪むため
            // 固定サイズの正方形にする。
            Box(Modifier.fillMaxWidth(), contentAlignment = Alignment.Center) {
                QrCodeImage(content = url, size = 260.dp)
            }

            Text(
                "このQRをスキャンすると、このアプリが開きます。" +
                    "アプリを使わない場合は下のWeb版URLをコピーしてください。",
                style = MaterialTheme.typography.bodySmall
            )
            TextButton(onClick = {
                val clipboard = context
                    .getSystemService(Context.CLIPBOARD_SERVICE) as? ClipboardManager
                clipboard?.setPrimaryClip(ClipData.newPlainText("channel", webUrl))
            }) {
                Text("Web版URLをコピー: $webUrl")
            }
        }

        Button(
            onClick = { showConfirm = true },
            enabled = !state.isLoading && !state.iamGuest,
            modifier = Modifier.fillMaxWidth()
        ) {
            Icon(Icons.Default.Send, contentDescription = null)
            Spacer(Modifier.width(8.dp))
            Text("招待する")
        }

        if (code.isNotBlank()) {
            Text(
                "注意: 「招待する」を押すとコードが更新され、以前共有したURLは無効になります。",
                style = MaterialTheme.typography.bodySmall,
                color = MaterialTheme.colorScheme.error
            )
        }
    }

    if (showConfirm) {
        AlertDialog(
            onDismissRequest = { showConfirm = false },
            title = { Text("招待リンクを生成") },
            text = {
                Text(
                    if (code.isBlank()) {
                        "招待用のリンクとQRコードを生成します。"
                    } else {
                        "招待リンクを生成し直します。\n" +
                            "以前共有したURLはすべて無効になります。よろしいですか？"
                    }
                )
            },
            confirmButton = {
                TextButton(onClick = {
                    showConfirm = false
                    onGenerate(guest)
                }) { Text("生成") }
            },
            dismissButton = {
                TextButton(onClick = { showConfirm = false }) { Text("キャンセル") }
            }
        )
    }
}

/** 招待用のカスタムスキームURL。QRはこれを埋め込む。 */
private fun invitationUrl(channelID: String?, code: String): String =
    "chat://profile/${channelID.orEmpty()}?code=$code"

private fun webInvitationUrl(channelID: String?, code: String): String =
    BuildConfig.BASE_URL.trimEnd('/') + "/profile/${channelID.orEmpty()}/?code=$code"

/** メンバー一覧。 */
private fun LazyListScope.MemberSection(state: ChannelUiState) {
    item { SectionTitle("ユーザー一覧") }
    if (state.aliases.isEmpty()) {
        item { Text("メンバーがいません", style = MaterialTheme.typography.bodyMedium) }
    }
    items(state.aliases, key = { it.aliasID }) { alias ->
        Surface(
            color = MaterialTheme.colorScheme.surface,
            modifier = Modifier.fillMaxWidth()
        ) {
            Row(
                modifier = Modifier.padding(12.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                AliasAvatar(
                    aliasImg = alias.aliasImg.toAbsoluteImageUrl(),
                    modifier = Modifier.size(40.dp)
                )
                Spacer(Modifier.width(12.dp))
                Column(Modifier.weight(1f)) {
                    Text(alias.aliasName, style = MaterialTheme.typography.bodyLarge)
                    if (alias.aliasBio.isNotBlank()) {
                        Text(alias.aliasBio, style = MaterialTheme.typography.bodySmall)
                    }
                }
                if (alias.accessRight.isNotBlank()) {
                    Text(
                        alias.accessRight,
                        style = MaterialTheme.typography.labelSmall,
                        color = MaterialTheme.colorScheme.primary
                    )
                }
            }
        }
    }
}

/** グループ一覧。 */
private fun LazyListScope.GroupSection(
    state: ChannelUiState,
    navController: NavController
) {
    item { SectionTitle("グループ") }

    // グループ編集画面への導線（Vue の「👪 グループアカウント作成・編集」相当）
    item {
        OutlinedButton(
            onClick = { navController.navigate(GroupRoute(id = state.channelID)) },
            modifier = Modifier.fillMaxWidth()
        ) {
            Icon(Icons.Default.Groups, contentDescription = null)
            Spacer(Modifier.width(8.dp))
            Text("グループの作成・編集")
        }
    }

    if (state.groups.isEmpty()) {
        item { Text("グループがありません", style = MaterialTheme.typography.bodyMedium) }
    }
    items(state.groups, key = { it.groupID }) { group ->
        Surface(
            color = MaterialTheme.colorScheme.surface,
            modifier = Modifier.fillMaxWidth()
        ) {
            Row(
                modifier = Modifier.padding(12.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Icon(Icons.Default.Groups, contentDescription = null)
                Spacer(Modifier.width(12.dp))
                Column(Modifier.weight(1f)) {
                    Text(group.groupName, style = MaterialTheme.typography.bodyLarge)
                    if (group.groupBio.isNotBlank()) {
                        Text(group.groupBio, style = MaterialTheme.typography.bodySmall)
                    }
                }
            }
        }
    }
}

/** 説明文の装飾ツールバー（Vue の toolbar 相当）。 */
@Composable
private fun MarkdownToolbar(enabled: Boolean, onWrap: (String, String) -> Unit) {
    val buttons = listOf(
        "＊太＊" to "・＊太＊",
        "〜〜" to "・〜〜",
        "＜quote＞" to "＜・quote＞",
        "｀｀｀" to "・｀｀｀",
        "「" to "」（）",
        "色＊赤" to "赤＊色"
    )
    Row(
        modifier = Modifier
            .fillMaxWidth()
            .horizontalScroll(rememberScrollState()),
        horizontalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        buttons.forEach { (open, close) ->
            OutlinedButton(onClick = { onWrap(open, close) }, enabled = enabled) {
                Text(open.take(3), style = MaterialTheme.typography.labelSmall)
            }
        }
    }
}

/** 新規作成時のアイコン選択行。 */
@Composable
private fun AvatarPickerRow(myImg: String, onPick: () -> Unit) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        AliasAvatar(
            aliasImg = myImg.toAbsoluteImageUrl(),
            modifier = Modifier.size(48.dp)
        )
        Spacer(Modifier.width(12.dp))
        OutlinedButton(onClick = onPick) { Text("アイコンを変更") }
    }
}

/** アイコン（,絵文字,#色）選択ダイアログ。 */
@Composable
private fun AvatarPickerDialog(
    current: String,
    onSelect: (String) -> Unit,
    onDismiss: () -> Unit
) {
    val emojis = listOf("A", "B", "C", "D", "E", "F", "あ", "い", "う")
    val colors = listOf(
        "#E53935", "#8E24AA", "#3949AB", "#00897B",
        "#F9A825", "#FB8C00", "#6D4C41", "#546E7A"
    )
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("アイコンを選択") },
        text = {
            Column(
                modifier = Modifier.verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                colors.chunked(4).forEachIndexed { rowIndex, rowColors ->
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        rowColors.forEachIndexed { columnIndex, color ->
                            val emoji = emojis[(rowIndex * 4 + columnIndex) % emojis.size]
                            val image = ",$emoji,$color"
                            Box(
                                modifier = Modifier
                                    .size(44.dp)
                                    .background(
                                        MaterialTheme.colorScheme.surfaceVariant,
                                        RoundedCornerShape(8.dp)
                                    )
                                    .border(
                                        if (image == current) 2.dp else 0.dp,
                                        MaterialTheme.colorScheme.primary,
                                        RoundedCornerShape(8.dp)
                                    )
                                    .clickable { onSelect(image) },
                                contentAlignment = Alignment.Center
                            ) {
                                Text(emoji)
                            }
                        }
                    }
                }
            }
        },
        confirmButton = { TextButton(onClick = onDismiss) { Text("閉じる") } }
    )
}

/** 選択範囲を装飾記号で挟む（選択が空なら末尾に追記）。 */
private fun wrapSelection(value: TextFieldValue, open: String, close: String): TextFieldValue {
    val text = value.text
    val selection = value.selection
    if (selection.collapsed) {
        val newText = text + open + close
        return value.copy(text = newText, selection = TextRange(newText.length))
    }
    val selected = text.substring(selection.start, selection.end)
    val newText = text.replaceRange(selection.start, selection.end, open + selected + close)
    val newStart = selection.start + open.length
    return value.copy(
        text = newText,
        selection = TextRange(newStart, newStart + selected.length)
    )
}

@Composable
private fun SectionCard(title: String, content: @Composable ColumnScope.() -> Unit) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(
            Modifier.padding(16.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Text(
                title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.SemiBold
            )
            content()
        }
    }
}

@Composable
private fun SectionTitle(text: String) {
    Text(
        text,
        style = MaterialTheme.typography.titleMedium,
        fontWeight = FontWeight.SemiBold,
        modifier = Modifier.padding(top = 4.dp, bottom = 4.dp)
    )
}

@Composable
private fun MessageBanner(message: String, color: Color) {
    Card(
        colors = CardDefaults.cardColors(containerColor = color.copy(alpha = 0.12f)),
        modifier = Modifier.fillMaxWidth()
    ) {
        Text(message, color = color, modifier = Modifier.padding(12.dp))
    }
}

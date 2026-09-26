package com.chat.android.feature.user

import android.Manifest
import android.graphics.Color as AndroidColor
import android.net.Uri
import android.os.Build
import android.util.Base64
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Check
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Edit
import androidx.compose.material.icons.filled.LocationOn
import androidx.compose.material.icons.filled.Map
import androidx.compose.material.icons.filled.PhotoLibrary
import androidx.compose.material.icons.filled.Refresh
import androidx.compose.material3.AlertDialog
import androidx.compose.material3.Button
import androidx.compose.material3.CircularProgressIndicator
import androidx.compose.material3.Divider
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.RadioButton
import androidx.compose.material3.Scaffold
import androidx.compose.material3.TopAppBar
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalUriHandler
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.core.content.ContextCompat
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import coil.compose.AsyncImage
import com.chat.android.data.database.entities.UserNicknameEntity
import java.util.Locale

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun UserScreen(
    navController: NavController,
    viewModel: UserViewModel = hiltViewModel()
) {
    val state by viewModel.uiState.collectAsState()
    val context = LocalContext.current
    val uriHandler = LocalUriHandler.current
    var colorPickerVisible by remember { mutableStateOf(false) }

    val locationPermissionLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.RequestMultiplePermissions()
    ) { permissions ->
        if (permissions.values.any { it }) viewModel.refreshLocation()
    }
    val imagePickerLauncher = rememberLauncherForActivityResult(ActivityResultContracts.GetContent()) { uri: Uri? ->
        uri?.let(viewModel::setAvatarFromUri)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("ユーザー設定") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        }
    ) { padding ->
        when {
            state.isLoading && state.profile == null -> {
                Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            }
            state.profile == null && !state.isLoading -> {
                Column(
                    Modifier.fillMaxSize().padding(padding).padding(24.dp),
                    verticalArrangement = Arrangement.Center,
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(state.errorMessage ?: "ユーザー情報を表示できません")
                    Button(onClick = viewModel::loadUserData) { Text("再読み込み") }
                }
            }
            else -> {
                LazyColumn(
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentPadding = androidx.compose.foundation.layout.PaddingValues(horizontal = 20.dp, vertical = 16.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    item {
                        Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.End) {
                            IconButton(onClick = viewModel::loadUserData, enabled = !state.isLoading) {
                                Icon(Icons.Default.Refresh, contentDescription = "更新")
                            }
                        }
                    }
                    if (state.isTO && !state.toLink.isNullOrBlank()) {
                        item {
                            Column(verticalArrangement = Arrangement.spacedBy(4.dp)) {
                                TextButton(onClick = { runCatching { uriHandler.openUri(state.toLink!!) } }) {
                                    Text("招待参加ページ")
                                }
                                Text("必須項目を登録してから参加ページに進んでください", style = MaterialTheme.typography.bodySmall)
                            }
                        }
                    }
                    state.errorMessage?.let { message ->
                        item { MessageLine(message, MaterialTheme.colorScheme.error, viewModel::clearMessages) }
                    }
                    state.successMessage?.let { message ->
                        item { MessageLine(message, MaterialTheme.colorScheme.primary, viewModel::clearMessages) }
                    }
                    item {
                        SectionTitle("位置情報")
                        Text("緯度と経度を入力してください。位置情報から取得するか、Googleマップで確認できます。", style = MaterialTheme.typography.bodySmall)
                        OutlinedTextField(
                            value = state.coordinateInput,
                            onValueChange = viewModel::updateCoordinateInput,
                            modifier = Modifier.fillMaxWidth(),
                            label = { Text("緯度, 経度") },
                            placeholder = { Text("35.723003, 139.525071") },
                            singleLine = true
                        )
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            OutlinedButton(
                                onClick = {
                                    val permissions = arrayOf(Manifest.permission.ACCESS_FINE_LOCATION, Manifest.permission.ACCESS_COARSE_LOCATION)
                                    val hasPermission = permissions.any {
                                        ContextCompat.checkSelfPermission(context, it) == android.content.pm.PackageManager.PERMISSION_GRANTED
                                    }
                                    if (hasPermission) viewModel.refreshLocation()
                                    else locationPermissionLauncher.launch(permissions)
                                },
                                enabled = !state.isLocating
                            ) {
                                if (state.isLocating) CircularProgressIndicator(Modifier.size(18.dp), strokeWidth = 2.dp)
                                else Icon(Icons.Default.LocationOn, contentDescription = null)
                                Spacer(Modifier.width(6.dp))
                                Text("現在地を取得")
                            }
                            OutlinedButton(onClick = { uriHandler.openUri("https://maps.google.com/") }) {
                                Icon(Icons.Default.Map, contentDescription = null)
                                Spacer(Modifier.width(6.dp))
                                Text("Googleマップ")
                            }
                        }
                    }
                    item {
                        Divider()
                        SectionTitle("連絡先・ウォレット")
                        OutlinedTextField(state.mail, viewModel::updateMail, Modifier.fillMaxWidth(), label = { Text("メール") }, singleLine = true)
                        OutlinedTextField(state.telephone, viewModel::updateTelephone, Modifier.fillMaxWidth(), label = { Text("電話番号") }, singleLine = true)
                        OutlinedTextField(state.walletAddress, viewModel::updateWalletAddress, Modifier.fillMaxWidth(), label = { Text("JPYCアドレス") }, singleLine = true)
                    }
                    item {
                        Divider()
                        SectionTitle("ニックネーム")
                        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                            OutlinedButton(onClick = { viewModel.openNicknameForm(UserViewModel.MODE_EDIT) }, enabled = state.nicknames.isNotEmpty()) {
                                Icon(Icons.Default.Edit, contentDescription = null)
                                Spacer(Modifier.width(6.dp))
                                Text("画像・自己紹介を編集")
                            }
                            OutlinedButton(onClick = { viewModel.openNicknameForm(UserViewModel.MODE_NEW) }, enabled = state.nicknames.size < UserViewModel.MAX_NICKNAMES) {
                                Text("新規作成 (${state.nicknames.size}/3)")
                            }
                        }
                    }
                    if (state.isNicknameFormVisible) {
                        item {
                            NicknameEditor(
                                state = state,
                                onNameChange = viewModel::updateEditableNickname,
                                onEmojiChange = viewModel::updateEmoji,
                                onImageMode = { imagePickerLauncher.launch("image/*"); viewModel.selectImageAvatar() },
                                onRemoveImage = viewModel::removeNicknameImage,
                                onChooseColor = { colorPickerVisible = true },
                                onBioChange = viewModel::updateNickBio,
                                onFormat = viewModel::formatNickBio,
                                onCancel = viewModel::closeNicknameForm
                            )
                        }
                    }
                    items(state.nicknames, key = UserNicknameEntity::nickname) { nickname ->
                        NicknameRow(
                            nickname = nickname,
                            selected = nickname.nickname == state.currentNickname,
                            onClick = { viewModel.switchNickname(nickname.nickname) }
                        )
                    }
                    item {
                        Button(
                            onClick = viewModel::submitUser,
                            modifier = Modifier.fillMaxWidth(),
                            enabled = !state.isSaving
                        ) {
                            if (state.isSaving) CircularProgressIndicator(Modifier.size(18.dp), strokeWidth = 2.dp)
                            else Text("更新")
                        }
                    }
                }
            }
        }
    }

    if (colorPickerVisible) {
        AvatarColorDialog(
            onColorSelected = { viewModel.updateAvatarColor(it); colorPickerVisible = false },
            onDismiss = { colorPickerVisible = false }
        )
    }
}

@Composable
private fun NicknameEditor(
    state: UserUiState,
    onNameChange: (String) -> Unit,
    onEmojiChange: (String) -> Unit,
    onImageMode: () -> Unit,
    onRemoveImage: () -> Unit,
    onChooseColor: () -> Unit,
    onBioChange: (TextFieldValue) -> Unit,
    onFormat: (String, String) -> Unit,
    onCancel: () -> Unit
) {
    Column(verticalArrangement = Arrangement.spacedBy(10.dp)) {
        SectionTitle(if (state.editingMode == UserViewModel.MODE_EDIT) "ニックネームの編集" else "新しいニックネーム")
        Text("ニックネームは登録後に変更できません。最大3件まで登録できます。", style = MaterialTheme.typography.bodySmall)
        OutlinedTextField(
            value = state.editableNickname,
            onValueChange = onNameChange,
            modifier = Modifier.fillMaxWidth(),
            label = { Text("ニックネーム") },
            enabled = state.editingMode != UserViewModel.MODE_EDIT,
            singleLine = true
        )
        Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
            OutlinedButton(onClick = onChooseColor, enabled = state.isEmojiAvatar) {
                AvatarPreview(state.nickImg, Modifier.size(32.dp))
                Spacer(Modifier.width(8.dp))
                Text("絵文字・色")
            }
            OutlinedButton(onClick = onImageMode) {
                Icon(Icons.Default.PhotoLibrary, contentDescription = null)
                Spacer(Modifier.width(6.dp))
                Text("画像を選択")
            }
        }
        if (state.isEmojiAvatar) {
            OutlinedTextField(
                value = state.nickImg.split(',').getOrNull(1).orEmpty(),
                onValueChange = onEmojiChange,
                label = { Text("絵文字 (1文字)") },
                singleLine = true,
                modifier = Modifier.fillMaxWidth()
            )
        } else {
            AvatarPreview(state.nickImg, Modifier.size(56.dp))
        }
        TextButton(onClick = onRemoveImage, enabled = state.nickImg.isNotBlank() && !state.removeNickImg) {
            Text("アイコンを削除")
        }
        Text("自己紹介 (${state.nickBio.text.length}/200)", style = MaterialTheme.typography.titleSmall)
        Row(horizontalArrangement = Arrangement.spacedBy(4.dp)) {
            TextButton(onClick = { onFormat("＊太＊", "・＊太＊") }) { Text("太字") }
            TextButton(onClick = { onFormat("〜〜", "・〜〜") }) { Text("取消") }
            TextButton(onClick = { onFormat("＜quote＞", "＜・quote＞") }) { Text("引用") }
            TextButton(onClick = { onFormat("｀｀｀", "・｀｀｀") }) { Text("コード") }
            TextButton(onClick = { onFormat("色＊赤", "赤＊色") }) { Text("赤") }
        }
        OutlinedTextField(
            value = state.nickBio,
            onValueChange = onBioChange,
            modifier = Modifier.fillMaxWidth().height(150.dp),
            placeholder = { Text("自己紹介") },
            minLines = 4
        )
        OutlinedButton(onClick = onCancel, modifier = Modifier.fillMaxWidth()) { Text("キャンセル") }
    }
}

@Composable
private fun NicknameRow(nickname: UserNicknameEntity, selected: Boolean, onClick: () -> Unit) {
    Row(
        modifier = Modifier.fillMaxWidth().clickable(onClick = onClick).padding(vertical = 6.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        AvatarPreview(nickname.nickImg.orEmpty(), Modifier.size(32.dp))
        Spacer(Modifier.width(10.dp))
        Text(nickname.nickname, modifier = Modifier.weight(1f), fontWeight = if (selected) FontWeight.Bold else FontWeight.Normal)
        RadioButton(selected = selected, onClick = onClick)
    }
}

@Composable
private fun AvatarPreview(image: String, modifier: Modifier = Modifier) {
    val emojiParts = image.split(',')
    if (image.startsWith(",") && emojiParts.size >= 3) {
        val color = runCatching { Color(AndroidColor.parseColor(emojiParts[2])) }.getOrDefault(MaterialTheme.colorScheme.surfaceVariant)
        Box(modifier.clip(CircleShape).background(color), contentAlignment = Alignment.Center) {
            Text(emojiParts[1])
        }
    } else if (image.startsWith("data:image")) {
        val bytes = remember(image) { runCatching { Base64.decode(image.substringAfter(','), Base64.DEFAULT) }.getOrNull() }
        AsyncImage(model = bytes, contentDescription = "ニックネーム画像", modifier = modifier.clip(CircleShape))
    } else if (image.isNotBlank()) {
        AsyncImage(model = image, contentDescription = "ニックネーム画像", modifier = modifier.clip(CircleShape))
    } else {
        Box(modifier.clip(CircleShape).background(MaterialTheme.colorScheme.surfaceVariant), contentAlignment = Alignment.Center) {
            Text("🙂")
        }
    }
}

@Composable
private fun AvatarColorDialog(onColorSelected: (String) -> Unit, onDismiss: () -> Unit) {
    val colors = listOf("#E57373", "#FFB74D", "#FFF176", "#81C784", "#4FC3F7", "#7986CB", "#BA68C8", "#BDBDBD")
    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("背景色") },
        text = {
            Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                colors.chunked(4).forEach { rowColors ->
                    Row(horizontalArrangement = Arrangement.spacedBy(12.dp)) {
                        rowColors.forEach { value ->
                            val color = Color(AndroidColor.parseColor(value))
                            IconButton(onClick = { onColorSelected(value) }, modifier = Modifier.size(40.dp).background(color, CircleShape)) {
                                Icon(Icons.Default.Check, contentDescription = value, tint = Color.Transparent)
                            }
                        }
                    }
                }
            }
        },
        confirmButton = { TextButton(onClick = onDismiss) { Text("閉じる") } }
    )
}

@Composable
private fun SectionTitle(text: String) {
    Text(text, style = MaterialTheme.typography.titleMedium, fontWeight = FontWeight.SemiBold, modifier = Modifier.padding(top = 4.dp, bottom = 4.dp))
}

@Composable
private fun MessageLine(message: String, color: Color, onDismiss: () -> Unit) {
    Row(Modifier.fillMaxWidth(), verticalAlignment = Alignment.CenterVertically) {
        Text(message, modifier = Modifier.weight(1f), color = color)
        TextButton(onClick = onDismiss) { Text("閉じる") }
    }
}
package com.chat.android.feature.user

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.net.Uri
import android.util.Base64
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.database.entities.UserProfileEntity
import com.chat.android.core.repository.UserRepository
import com.google.android.gms.location.LocationServices
import com.google.android.gms.location.Priority
import com.google.android.gms.tasks.CancellationTokenSource
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import java.io.ByteArrayOutputStream
import java.net.HttpURLConnection
import java.net.URL
import java.util.Locale
import javax.inject.Inject
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.input.TextFieldValue
import com.chat.android.feature.user.UserUiState.Companion.AVATAR_MODE_EMOJI
import com.chat.android.feature.user.UserUiState.Companion.AVATAR_MODE_IMAGE
import com.chat.android.feature.user.UserUiState.Companion.DEFAULT_EMOJI_AVATAR
import com.chat.android.core.util.toAbsoluteImageUrl

@HiltViewModel
class UserViewModel @Inject constructor(
    val userRepository: UserRepository,
    @ApplicationContext private val context: Context
) : ViewModel() {
    private val _uiState = MutableStateFlow(UserUiState())
    val uiState = _uiState.asStateFlow()

    init {
        observeUserData()
        checkInvitationLink()
        loadUserData()
    }

    private fun observeUserData() {
        viewModelScope.launch {
            combine(userRepository.observeUserProfile(), userRepository.observeNicknames()) { profile, nicknames ->
                profile to nicknames
            }.collect { (profile, nicknames) ->
                val previous = _uiState.value
                val storedNickname = userRepository.getNickname()
                val selectedNickname = previous.currentNickname
                    ?.takeIf { selected -> nicknames.any { it.nickname == selected } }
                    ?: storedNickname?.takeIf { selected -> nicknames.any { it.nickname == selected } }
                    ?: profile?.nickname?.takeIf { selected -> nicknames.any { it.nickname == selected } }
                    ?: nicknames.firstOrNull()?.nickname
                val selected = nicknames.find { it.nickname == selectedNickname }
                val storedAvatar = selected?.nickImg ?: profile?.nickImg.orEmpty()

                _uiState.update { state ->
                    state.copy(
                        profile = profile,
                        nicknames = nicknames,
                        currentNickname = selectedNickname,
                        mail = if (!state.isDraftDirty && profile != null) profile.mail.orEmpty() else state.mail,
                        telephone = if (!state.isDraftDirty && profile != null) profile.telephone.orEmpty() else state.telephone,
                        walletAddress = if (!state.isDraftDirty && profile != null) profile.walletAddress.orEmpty() else state.walletAddress,
                        coordinateInput = if (!state.isDraftDirty && profile != null) {
                            if (profile.latitude != null && profile.longitude != null) {
                                "${profile.latitude}, ${profile.longitude}"
                            } else ""
                        } else state.coordinateInput,
                        nickImg = if (state.isNicknameFormVisible) state.nickImg else storedAvatar,
                        nickBio = if (state.isNicknameFormVisible) state.nickBio else TextFieldValue(selected?.nickBio ?: profile?.nickBio.orEmpty()),
                        avatarMode = if (state.isNicknameFormVisible) state.avatarMode else avatarModeFor(storedAvatar),
                        emojiAvatar = if (state.isNicknameFormVisible) state.emojiAvatar else emojiDraftOf(storedAvatar),
                        imageAvatar = if (state.isNicknameFormVisible) state.imageAvatar else imageDraftOf(storedAvatar),
                        originalAvatar = if (state.isNicknameFormVisible) state.originalAvatar else storedAvatar,
                        removeNickImg = if (state.isNicknameFormVisible) state.removeNickImg else false,
                        fetched = state.fetched || profile != null
                    )
                }

                if (selectedNickname != null && selectedNickname != storedNickname) {
                    userRepository.setNickname(selectedNickname)
                }
            }
        }
    }

    private fun checkInvitationLink() {
        userRepository.getTO()?.let { link ->
            _uiState.update { it.copy(isTO = true, toLink = link) }
        }
    }

    fun loadUserData() {
        viewModelScope.launch {
            _uiState.update { it.copy(isLoading = true, errorMessage = null) }
            val error = runCatching { userRepository.refreshUser() }
                .getOrElse { it.message ?: "ユーザー情報を取得できませんでした" }
            _uiState.update {
                it.copy(
                    isLoading = false,
                    fetched = it.profile != null || error == null,
                    errorMessage = error
                )
            }
        }
    }

    fun updateMail(value: String) = updateDraft { copy(mail = value) }
    fun updateTelephone(value: String) = updateDraft { copy(telephone = value) }
    fun updateWalletAddress(value: String) = updateDraft { copy(walletAddress = value) }

    fun updateCoordinateInput(value: String) {
        updateDraft { copy(coordinateInput = value) }
    }

    fun openNicknameForm(mode: String) {
        val state = _uiState.value
        if (mode == MODE_NEW && state.nicknames.size >= MAX_NICKNAMES) {
            _uiState.update { it.copy(errorMessage = "ニックネームは3つまで作成できます。") }
            return
        }
        val storedNickname = userRepository.getNickname().orEmpty()
        val editTarget = state.currentNickname.orEmpty()
            .ifBlank { storedNickname }
            .takeIf { nickname -> state.nicknames.any { it.nickname == nickname } }
        // 編集対象が未登録（まだニックネームが1件もない等）の場合は新規作成として扱う
        val effectiveMode = if (mode == MODE_EDIT && editTarget == null) MODE_NEW else mode
        val selected = state.nicknames.find { it.nickname == editTarget }
        val storedAvatar = if (effectiveMode == MODE_EDIT) selected?.nickImg.orEmpty() else ""
        val initialName = when {
            effectiveMode == MODE_EDIT -> editTarget.orEmpty()
            mode == MODE_EDIT -> state.profile?.nickname.orEmpty().ifBlank { storedNickname }
            else -> ""
        }
        val initialBio = when {
            effectiveMode != MODE_EDIT && mode == MODE_NEW -> ""
            else -> selected?.nickBio ?: state.profile?.nickBio.orEmpty()
        }
        _uiState.update {
            it.copy(
                isNicknameFormVisible = true,
                editingMode = effectiveMode,
                editableNickname = initialName,
                removeNickImg = false,
                nickBio = TextFieldValue(initialBio),
                originalAvatar = storedAvatar,
                errorMessage = null
            ).withAvatar(
                mode = avatarModeFor(storedAvatar),
                emoji = emojiDraftOf(storedAvatar),
                image = imageDraftOf(storedAvatar)
            )
        }
    }

    fun closeNicknameForm() {
        _uiState.update { state ->
            val selected = state.nicknames.find { it.nickname == state.currentNickname }
            val storedAvatar = selected?.nickImg ?: state.profile?.nickImg.orEmpty()
            state.copy(
                isNicknameFormVisible = false,
                editingMode = "",
                editableNickname = "",
                nickImg = storedAvatar,
                removeNickImg = false,
                nickBio = TextFieldValue(selected?.nickBio ?: state.profile?.nickBio.orEmpty()),
                avatarMode = avatarModeFor(storedAvatar),
                emojiAvatar = emojiDraftOf(storedAvatar),
                imageAvatar = imageDraftOf(storedAvatar),
                originalAvatar = storedAvatar
            )
        }
    }

    fun updateEditableNickname(value: String) {
        _uiState.update { it.copy(editableNickname = value, isDraftDirty = true) }
    }

    fun updateNickBio(value: TextFieldValue) {
        val text = value.text.take(MAX_BIO_LENGTH)
        val selection = TextRange(
            value.selection.start.coerceAtMost(text.length),
            value.selection.end.coerceAtMost(text.length)
        )
        _uiState.update { it.copy(nickBio = TextFieldValue(text, selection), isDraftDirty = true) }
    }

    fun formatNickBio(prefix: String, suffix: String) {
        val current = _uiState.value.nickBio
        val start = minOf(current.selection.start, current.selection.end)
        val end = maxOf(current.selection.start, current.selection.end)
        val selected = current.text.substring(start, end)
        val updated = current.text.replaceRange(start, end, "$prefix$selected$suffix").take(MAX_BIO_LENGTH)
        val cursor = (start + prefix.length + selected.length).coerceAtMost(updated.length)
        updateNickBio(TextFieldValue(updated, TextRange(cursor)))
    }

    fun updateEmoji(value: String) = updateDraft {
        withAvatar(mode = AVATAR_MODE_EMOJI, emoji = buildAvatarEmoji(sanitizeAvatarEmoji(value), avatarColor(emojiAvatar)))
    }

    fun selectEmoji(emoji: String) = updateDraft {
        withAvatar(mode = AVATAR_MODE_EMOJI, emoji = buildAvatarEmoji(sanitizeAvatarEmoji(emoji), avatarColor(emojiAvatar)))
    }

    fun updateAvatarColor(value: String) = updateDraft {
        withAvatar(mode = AVATAR_MODE_EMOJI, emoji = buildAvatarEmoji(avatarEmoji(emojiAvatar), value))
    }

    /** 絵文字アイコン／画像アイコンの切り替え（画像未選択でも絵文字へ戻れる） */
    fun selectAvatarMode(mode: String) = updateDraft {
        val target = if (mode == AVATAR_MODE_IMAGE) AVATAR_MODE_IMAGE else AVATAR_MODE_EMOJI
        withAvatar(
            mode = target,
            emoji = emojiAvatar.ifBlank { DEFAULT_EMOJI_AVATAR },
            image = imageAvatar.ifBlank { imageDraftOf(originalAvatar) }
        )
    }

    fun selectImageAvatar() = updateDraft {
        withAvatar(mode = AVATAR_MODE_IMAGE, image = imageAvatar.ifBlank { imageDraftOf(originalAvatar) })
    }

    fun removeNicknameImage() = updateDraft {
        withAvatar(
            mode = AVATAR_MODE_EMOJI,
            emoji = "",
            image = "",
            removed = true
        )
    }

    /** 端末で選んだ画像をサーバーが受け付けるDataURI（50px以内）へ変換して下書きに反映する */
    fun setAvatarFromUri(uri: Uri) {
        viewModelScope.launch {
            val result = withContext(Dispatchers.IO) {
                runCatching {
                    // decodeStream は inJustDecodeBounds 時に必ず null を返すため、一度バイト列へ読み出してから判定する。
                    // 併せて、ContentProvider によっては同じストリームを2回開けない問題も回避できる。
                    val bytes = context.contentResolver.openInputStream(uri)?.use { it.readBytes() }
                    if (bytes == null || bytes.isEmpty()) error(IMAGE_LOAD_ERROR)
                    encodeImageToDataUri(bytes) ?: error(IMAGE_FORMAT_ERROR)
                }
            }
            result.onSuccess { dataUrl ->
                updateDraft { withAvatar(mode = AVATAR_MODE_IMAGE, image = dataUrl) }
            }.onFailure { error ->
                _uiState.update { it.copy(errorMessage = error.message ?: IMAGE_LOAD_ERROR) }
            }
        }
    }

    fun switchNickname(nickname: String) {
        val selected = _uiState.value.nicknames.find { it.nickname == nickname } ?: return
        userRepository.setNickname(selected.nickname)
        val storedAvatar = selected.nickImg.orEmpty()
        _uiState.update {
            it.copy(
                currentNickname = selected.nickname,
                nickImg = storedAvatar,
                removeNickImg = false,
                nickBio = TextFieldValue(selected.nickBio.orEmpty()),
                avatarMode = avatarModeFor(storedAvatar),
                emojiAvatar = emojiDraftOf(storedAvatar),
                imageAvatar = imageDraftOf(storedAvatar),
                originalAvatar = storedAvatar,
                isDraftDirty = false,
                isNicknameFormVisible = false,
                editingMode = ""
            )
        }
    }

    fun submitUser() {
        val state = _uiState.value
        if (state.isSaving) return
        val nickname = if (state.editingMode == MODE_NEW) state.editableNickname.trim()
            else state.currentNickname.orEmpty()
                .ifBlank { userRepository.getNickname().orEmpty() }
                .ifBlank { state.profile?.nickname.orEmpty() }
        if (nickname.isBlank()) {
            _uiState.update { it.copy(errorMessage = "ニックネームを入力してください") }
            return
        }
        if (state.editingMode == MODE_NEW && state.nicknames.size >= MAX_NICKNAMES) {
            _uiState.update { it.copy(errorMessage = "ニックネームは3つまで作成できます。") }
            return
        }
        val avatar = state.nickImg.trim()
        if (state.avatarMode == AVATAR_MODE_EMOJI && avatar.isNotBlank() && !AVATAR_EMOJI_REGEX.matches(avatar)) {
            _uiState.update { it.copy(errorMessage = "絵文字アイコンは「絵文字,#rrggbb」の形式で入力してください") }
            return
        }
        if (state.avatarMode == AVATAR_MODE_IMAGE && avatar.isNotBlank() && !AVATAR_IMAGE_REGEX.matches(avatar)) {
            _uiState.update { it.copy(errorMessage = "アイコン画像の形式が正しくありません。画像を選び直してください") }
            return
        }
        // 画像モードで画像が未選択の場合、保存済みアイコンが画像なら削除操作が無い限り保存を拒否する
        val storedImageAvatar = state.originalAvatar.trim()
            .takeIf { it.isNotBlank() && !it.startsWith(",") }
            .orEmpty()
        val imageNotSelected = state.avatarMode == AVATAR_MODE_IMAGE && avatar.isBlank()
        if (imageNotSelected && storedImageAvatar.isNotBlank() && !state.removeNickImg) {
            _uiState.update { it.copy(errorMessage = "画像を選択するか、「アイコンを削除」を押してください") }
            return
        }
        // 画像が未選択で保存済みアイコンが絵文字の場合は、絵文字アイコンを維持して意図しない消去を防ぐ
        val effectiveAvatar = if (imageNotSelected && !state.removeNickImg) {
            state.emojiAvatar.ifBlank { DEFAULT_EMOJI_AVATAR }
        } else avatar
        // 座標が未設定のユーザーでもアイコンだけ変更できるよう、入力がない場合は保存済みの値で送信する
        val coords = parseCoordinates(state.coordinateInput) ?: storedCoordinates(state.profile)
        // 既存の画像URLをサーバーへ再送すると「Emoji invalid:」エラーになるため、未変更の画像パスはDataURIへ再エンコードして送信する
        val unchangedStoredImage = state.avatarMode == AVATAR_MODE_IMAGE &&
            effectiveAvatar.isNotBlank() && effectiveAvatar == state.originalAvatar.trim() &&
            !effectiveAvatar.startsWith("data:image/")

        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true, errorMessage = null, successMessage = null) }
            val requestNickImg = when {
                state.removeNickImg -> ""
                unchangedStoredImage -> {
                    val encoded = reencodeStoredImageToDataUri(effectiveAvatar)
                    if (encoded == null) {
                        _uiState.update { it.copy(isSaving = false, errorMessage = "アイコン画像の取得に失敗しました。画像を選び直してください") }
                        return@launch
                    }
                    encoded
                }
                else -> effectiveAvatar
            }

            val result = runCatching {
                userRepository.updateUser(
                    nickname = nickname,
                    nickImg = requestNickImg,
                    removeNickImg = state.removeNickImg,
                    nickBio = normalizeLinks(state.nickBio.text),
                    mail = state.mail,
                    telephone = state.telephone,
                    walletAddress = state.walletAddress,
                    latitude = coords.first,
                    longitude = coords.second
                )
            }.getOrElse { exception ->
                _uiState.update { it.copy(isSaving = false, errorMessage = exception.message ?: "更新に失敗しました") }
                return@launch
            }
            if (result.error != null) {
                _uiState.update { it.copy(isSaving = false, errorMessage = result.error) }
                return@launch
            }
            userRepository.setNickname(nickname)
            // サーバー応答に新しい画像パスが含まれないため、送信した値をそのままローカルに保持する。
            // 既存画像の再エンコード送信時は元のURLパスを保持し、新規選択時のみキャッシュキーを更新する。
            val isNewUploadedImage = requestNickImg.startsWith("data:image/") && !unchangedStoredImage
            val savedAvatar = when {
                state.removeNickImg -> ""
                unchangedStoredImage -> avatar
                else -> requestNickImg.ifBlank { effectiveAvatar }
            }
            val cacheStamp = if (isNewUploadedImage) System.currentTimeMillis() else state.avatarCacheStamp
            _uiState.update {
                it.copy(
                    isSaving = false,
                    isDraftDirty = false,
                    isNicknameFormVisible = false,
                    editingMode = "",
                    editableNickname = "",
                    currentNickname = nickname,
                    nickImg = savedAvatar,
                    originalAvatar = savedAvatar,
                    avatarMode = avatarModeFor(savedAvatar),
                    emojiAvatar = emojiDraftOf(savedAvatar),
                    imageAvatar = imageDraftOf(savedAvatar),
                    avatarCacheStamp = cacheStamp,
                    successMessage = result.message,
                    errorMessage = null
                )
            }
        }
    }

    fun refreshLocation() {
        _uiState.update { it.copy(isLocating = true, errorMessage = null) }
        try {
            val client = LocationServices.getFusedLocationProviderClient(context)
            client.getCurrentLocation(Priority.PRIORITY_HIGH_ACCURACY, CancellationTokenSource().token)
                .addOnSuccessListener { location ->
                    if (location == null) {
                        _uiState.update { it.copy(isLocating = false, errorMessage = "現在地を取得できませんでした") }
                    } else {
                        val coordinates = String.format(Locale.US, "%.6f, %.6f", location.latitude, location.longitude)
                        updateCoordinateInput(coordinates)
                        _uiState.update { it.copy(isLocating = false) }
                    }
                }
                .addOnFailureListener { error ->
                    _uiState.update { it.copy(isLocating = false, errorMessage = error.message ?: "現在地を取得できませんでした") }
                }
        } catch (error: SecurityException) {
            _uiState.update { it.copy(isLocating = false, errorMessage = "位置情報の権限が必要です") }
        }
    }

    fun clearMessages() {
        _uiState.update { it.copy(errorMessage = null, successMessage = null) }
    }

    private fun updateDraft(change: UserUiState.() -> UserUiState) {
        _uiState.update { change(it).copy(isDraftDirty = true) }
    }

    private fun parseCoordinates(input: String): Pair<Double, Double>? {
        val parts = input.split(",").map { it.trim() }
        if (parts.size != 2) return null
        val latitude = parts[0].toDoubleOrNull() ?: return null
        val longitude = parts[1].toDoubleOrNull() ?: return null
        if (latitude !in -90.0..90.0 || longitude !in -180.0..180.0) return null
        return latitude to longitude
    }

    private fun normalizeLinks(text: String): String = BARE_URL_REGEX.replace(text) { match ->
        "「${match.value}」（${match.value}）"
    }

    private fun avatarEmoji(image: String): String = image.split(',').getOrNull(1).orEmpty()

    private fun avatarColor(image: String): String = image.split(',').getOrNull(2)?.takeIf { it.matches(HEX_COLOR_REGEX) }
        ?: DEFAULT_AVATAR_COLOR

    /** サーバーが保存する値（先頭がカンマなら絵文字、それ以外は画像パス）から編集モードを決める */
    private fun avatarModeFor(image: String): String =
        if (image.isNotBlank() && !image.startsWith(",")) AVATAR_MODE_IMAGE else AVATAR_MODE_EMOJI

    private fun emojiDraftOf(image: String): String =
        if (image.startsWith(",")) image else DEFAULT_EMOJI_AVATAR

    private fun imageDraftOf(image: String): String =
        if (image.isNotBlank() && !image.startsWith(",")) image else ""

    /** サーバー（^,.{1,2},#[0-9a-fA-F]{6}$）が受け付ける ",絵文字,#色" 形式にする */
    private fun buildAvatarEmoji(emoji: String, color: String): String {
        val normalizedColor = if (color.startsWith("#")) color else "#$color"
        return ",$emoji,$normalizedColor"
    }

    /** カンマを除き、コードポイント単位で最大2文字に丸める（サロゲートペアを壊さない） */
    private fun sanitizeAvatarEmoji(value: String): String {
        val builder = StringBuilder()
        var index = 0
        while (index < value.length && builder.codePointCount(0, builder.length) < MAX_AVATAR_EMOJI) {
            val codePoint = value.codePointAt(index)
            if (codePoint != ','.code) builder.appendCodePoint(codePoint)
            index += Character.charCount(codePoint)
        }
        return builder.toString()
    }

    private fun storedCoordinates(profile: UserProfileEntity?): Pair<Double, Double> =
        profile?.latitude?.let { latitude ->
            profile.longitude?.let { longitude -> latitude to longitude }
        } ?: (0.0 to 0.0)

    /** 絵文字・画像のドラフトから、送信する nickImg と削除フラグを導出する */
    private fun UserUiState.withAvatar(
        mode: String = avatarMode,
        emoji: String = emojiAvatar,
        image: String = imageAvatar,
        removed: Boolean = false
    ): UserUiState {
        val derived = if (mode == AVATAR_MODE_IMAGE) image else emoji
        return copy(
            avatarMode = mode,
            emojiAvatar = emoji,
            imageAvatar = image,
            nickImg = derived,
            removeNickImg = removed
        )
    }

    /** 既存の画像URLから画像バイトを取得し、サーバーが受け付けるDataURIへ再エンコードする */
    private suspend fun reencodeStoredImageToDataUri(reference: String): String? = withContext(Dispatchers.IO) {
        runCatching {
            val absolute = reference.toAbsoluteImageUrl()
            val connection = (URL(absolute).openConnection() as HttpURLConnection).apply {
                connectTimeout = 8000
                readTimeout = 8000
                instanceFollowRedirects = true
            }
            try {
                if (connection.responseCode !in 200..299) return@withContext null
                val bytes = connection.inputStream.use { it.readBytes() }
                encodeImageToDataUri(bytes)
            } finally {
                connection.disconnect()
            }
        }.getOrNull()
    }

    /**
     * 画像バイト列を AVATAR_MAX_SIZE 以内に縮小し、common.ImgSave が受け付ける DataURI へ変換する。
     * デコードできない形式（非対応コーデック等）は null を返す。
     */
    private fun encodeImageToDataUri(bytes: ByteArray): String? {
        if (bytes.isEmpty()) return null
        // inJustDecodeBounds の結果は null になるため、寸法は bounds.outWidth/outHeight で判定する。
        val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
        BitmapFactory.decodeByteArray(bytes, 0, bytes.size, bounds)
        if (bounds.outWidth <= 0 || bounds.outHeight <= 0) return null

        val largestSide = maxOf(bounds.outWidth, bounds.outHeight)
        val options = BitmapFactory.Options().apply {
            inSampleSize = (largestSide / AVATAR_MAX_SIZE).coerceAtLeast(1)
        }
        val decoded = BitmapFactory.decodeByteArray(bytes, 0, bytes.size, options) ?: return null
        return try {
            val scale = AVATAR_MAX_SIZE.toFloat() / maxOf(decoded.width, decoded.height)
            val width = (decoded.width * minOf(1f, scale)).toInt().coerceAtLeast(1)
            val height = (decoded.height * minOf(1f, scale)).toInt().coerceAtLeast(1)
            val resized = if (width == decoded.width && height == decoded.height) decoded
                else Bitmap.createScaledBitmap(decoded, width, height, true)
            try {
                // 透過を保持したい形式は PNG、それ以外（JPEG写真など）は JPEG で圧縮する。
                val isJpeg = bytes.size > 2 &&
                    bytes[0] == 0xFF.toByte() && bytes[1] == 0xD8.toByte()
                val format = if (isJpeg) Bitmap.CompressFormat.JPEG else Bitmap.CompressFormat.PNG
                val mime = if (isJpeg) "image/jpeg" else "image/png"
                val output = ByteArrayOutputStream()
                resized.compress(format, AVATAR_QUALITY, output)
                "data:$mime;base64,${Base64.encodeToString(output.toByteArray(), Base64.NO_WRAP)}"
            } finally {
                if (resized !== decoded) resized.recycle()
            }
        } finally {
            decoded.recycle()
        }
    }

    companion object {
        const val MODE_NEW = "new"
        const val MODE_EDIT = "edit"
        const val MAX_NICKNAMES = 3
        const val MAX_BIO_LENGTH = 200
        val AVATAR_EMOJI_PRESETS = listOf(
            "🙂", "😀", "😆", "😎", "🥳", "🤔",
            "🐱", "🐶", "🦊", "🐼", "🐨", "🦁",
            "🌸", "🍀", "🌟", "🌙", "🔥", "☕",
            "⚽", "🎮", "🎵", "📚", "🚀", "💎"
        )
        private const val AVATAR_MAX_SIZE = 50
        private const val AVATAR_QUALITY = 92
        private const val MAX_AVATAR_EMOJI = 2
        private const val DEFAULT_AVATAR_COLOR = "#cccccc"
        private const val IMAGE_LOAD_ERROR = "画像を読み込めませんでした"
        private const val IMAGE_FORMAT_ERROR = "この形式の画像は読み込めません。別の画像を選んでください"
        private val HEX_COLOR_REGEX = Regex("^#[0-9a-fA-F]{6}$")
        private val AVATAR_EMOJI_REGEX = Regex("^,.{1,2},#[0-9a-fA-F]{6}$")
        // Kotlin の Regex.matches は全体一致のため、値全体を消費するパターンにする必要がある
        private val AVATAR_IMAGE_REGEX = Regex("^(https?://\\S+|/[^\\s]*|data:image/[a-zA-Z0-9.+-]+;base64,[-A-Za-z0-9+/=]+)$")
        private val BARE_URL_REGEX = Regex("(?<!（)https?://[^\\s）]+")
    }
}
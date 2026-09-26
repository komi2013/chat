package com.chat.android.feature.user

import android.content.Context
import android.graphics.Bitmap
import android.graphics.BitmapFactory
import android.net.Uri
import android.util.Base64
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.data.repository.UserRepository
import com.google.android.gms.location.LocationServices
import com.google.android.gms.location.Priority
import com.google.android.gms.tasks.CancellationTokenSource
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import java.io.ByteArrayOutputStream
import java.util.Locale
import javax.inject.Inject
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.input.TextFieldValue

@HiltViewModel
class UserViewModel @Inject constructor(
    private val userRepository: UserRepository,
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
                        nickImg = if (state.isNicknameFormVisible) state.nickImg else selected?.nickImg ?: profile?.nickImg.orEmpty(),
                        nickBio = if (state.isNicknameFormVisible) state.nickBio else TextFieldValue(selected?.nickBio ?: profile?.nickBio.orEmpty()),
                        isEmojiAvatar = if (state.isNicknameFormVisible) state.isEmojiAvatar
                            else (selected?.nickImg ?: profile?.nickImg).isNullOrBlank() ||
                                (selected?.nickImg ?: profile?.nickImg)?.startsWith(",") == true,
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
        val selected = state.nicknames.find { it.nickname == state.currentNickname }
        val image = if (mode == MODE_NEW) ",🙂,#cccccc" else selected?.nickImg.orEmpty()
        _uiState.update {
            it.copy(
                isNicknameFormVisible = true,
                editingMode = mode,
                editableNickname = if (mode == MODE_EDIT) state.currentNickname.orEmpty() else "",
                nickImg = image,
                nickBio = TextFieldValue(if (mode == MODE_NEW) "" else selected?.nickBio.orEmpty()),
                isEmojiAvatar = image.isBlank() || image.startsWith(","),
                errorMessage = null
            )
        }
    }

    fun closeNicknameForm() {
        _uiState.update { state ->
            val selected = state.nicknames.find { it.nickname == state.currentNickname }
            val image = selected?.nickImg ?: state.profile?.nickImg.orEmpty()
            state.copy(
                isNicknameFormVisible = false,
                editingMode = "",
                editableNickname = "",
                nickImg = image,
                nickBio = TextFieldValue(selected?.nickBio ?: state.profile?.nickBio.orEmpty()),
                isEmojiAvatar = image.isBlank() || image.startsWith(",")
            )
        }
    }

    fun updateEditableNickname(value: String) {
        _uiState.update { it.copy(editableNickname = value) }
    }

    fun updateNickBio(value: TextFieldValue) {
        val text = value.text.take(MAX_BIO_LENGTH)
        val selection = TextRange(
            value.selection.start.coerceAtMost(text.length),
            value.selection.end.coerceAtMost(text.length)
        )
        _uiState.update { it.copy(nickBio = TextFieldValue(text, selection)) }
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

    fun updateEmoji(value: String) {
        val emoji = value.take(2)
        val color = avatarColor(_uiState.value.nickImg)
        _uiState.update { it.copy(nickImg = ",$emoji,$color", isEmojiAvatar = true) }
    }

    fun updateAvatarColor(value: String) {
        val emoji = avatarEmoji(_uiState.value.nickImg).ifBlank { "🙂" }
        _uiState.update { it.copy(nickImg = ",$emoji,$value", isEmojiAvatar = true) }
    }

    fun selectImageAvatar() {
        _uiState.update { it.copy(isEmojiAvatar = false) }
    }

    fun setAvatarFromUri(uri: Uri) {
        viewModelScope.launch {
            runCatching {
                val resolver = context.contentResolver
                val bounds = BitmapFactory.Options().apply { inJustDecodeBounds = true }
                resolver.openInputStream(uri)?.use { BitmapFactory.decodeStream(it, null, bounds) }
                    ?: error("画像を読み込めませんでした")
                val largestSide = maxOf(bounds.outWidth, bounds.outHeight)
                val sampleSize = (largestSide / AVATAR_MAX_SIZE).coerceAtLeast(1)
                val options = BitmapFactory.Options().apply { inSampleSize = sampleSize }
                val decoded = resolver.openInputStream(uri)?.use { BitmapFactory.decodeStream(it, null, options) }
                    ?: error("画像を読み込めませんでした")
                val scale = AVATAR_MAX_SIZE.toFloat() / maxOf(decoded.width, decoded.height)
                val width = (decoded.width * minOf(1f, scale)).toInt().coerceAtLeast(1)
                val height = (decoded.height * minOf(1f, scale)).toInt().coerceAtLeast(1)
                val resized = Bitmap.createScaledBitmap(decoded, width, height, true)
                val output = ByteArrayOutputStream()
                resized.compress(Bitmap.CompressFormat.JPEG, 92, output)
                if (resized !== decoded) resized.recycle()
                decoded.recycle()
                "data:image/jpeg;base64,${Base64.encodeToString(output.toByteArray(), Base64.NO_WRAP)}"
            }.onSuccess { dataUrl ->
                _uiState.update { it.copy(nickImg = dataUrl, isEmojiAvatar = false) }
            }.onFailure { error ->
                _uiState.update { it.copy(errorMessage = error.message ?: "画像を読み込めませんでした") }
            }
        }
    }

    fun switchNickname(nickname: String) {
        val selected = _uiState.value.nicknames.find { it.nickname == nickname } ?: return
        userRepository.setNickname(selected.nickname)
        _uiState.update {
            it.copy(
                currentNickname = selected.nickname,
                nickImg = selected.nickImg.orEmpty(),
                nickBio = TextFieldValue(selected.nickBio.orEmpty()),
                isEmojiAvatar = selected.nickImg.isNullOrBlank() || selected.nickImg.startsWith(","),
                isNicknameFormVisible = false,
                editingMode = ""
            )
        }
    }

    fun submitUser() {
        val state = _uiState.value
        if (state.isSaving) return
        val coords = parseCoordinates(state.coordinateInput)
        if (coords == null) {
            _uiState.update { it.copy(errorMessage = "緯度と経度を「35.77, 139.57」の形式で入力してください") }
            return
        }
        val nickname = if (state.editingMode == MODE_NEW) state.editableNickname.trim()
            else state.currentNickname.orEmpty()
        if (state.isNicknameFormVisible && nickname.isBlank()) {
            _uiState.update { it.copy(errorMessage = "ニックネームを入力してください") }
            return
        }
        if (state.editingMode == MODE_NEW && state.nicknames.size >= MAX_NICKNAMES) {
            _uiState.update { it.copy(errorMessage = "ニックネームは3つまで作成できます。") }
            return
        }

        viewModelScope.launch {
            _uiState.update { it.copy(isSaving = true, errorMessage = null, successMessage = null) }
            val result = runCatching {
                userRepository.updateUser(
                    nickname = nickname,
                    nickImg = state.nickImg,
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
            if (nickname.isNotBlank()) userRepository.setNickname(nickname)
            _uiState.update {
                it.copy(
                    isSaving = false,
                    isDraftDirty = false,
                    isNicknameFormVisible = false,
                    editingMode = "",
                    editableNickname = "",
                    successMessage = result.message,
                    errorMessage = null
                )
            }
            loadUserData()
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

    companion object {
        const val MODE_NEW = "new"
        const val MODE_EDIT = "edit"
        const val MAX_NICKNAMES = 3
        const val MAX_BIO_LENGTH = 200
        private const val AVATAR_MAX_SIZE = 50
        private const val DEFAULT_AVATAR_COLOR = "#cccccc"
        private val HEX_COLOR_REGEX = Regex("^#[0-9a-fA-F]{6}$")
        private val BARE_URL_REGEX = Regex("(?<!（)https?://[^\\s）]+")
    }
}
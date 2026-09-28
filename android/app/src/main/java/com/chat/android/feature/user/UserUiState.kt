package com.chat.android.feature.user

import androidx.compose.ui.text.input.TextFieldValue
import com.chat.android.data.database.entities.UserNicknameEntity
import com.chat.android.data.database.entities.UserProfileEntity

data class UserUiState(
    val isLoading: Boolean = true,
    val isSaving: Boolean = false,
    val isLocating: Boolean = false,
    val profile: UserProfileEntity? = null,
    val nicknames: List<UserNicknameEntity> = emptyList(),
    val currentNickname: String? = null,
    val mail: String = "",
    val telephone: String = "",
    val walletAddress: String = "",
    val coordinateInput: String = "",
    val isDraftDirty: Boolean = false,
    val isNicknameFormVisible: Boolean = false,
    val editingMode: String = "",
    val editableNickname: String = "",
    val nickImg: String = "",
    val removeNickImg: Boolean = false,
    val nickBio: TextFieldValue = TextFieldValue(),
    val avatarMode: String = AVATAR_MODE_EMOJI,
    val emojiAvatar: String = "",
    val imageAvatar: String = "",
    val originalAvatar: String = "",
    val avatarCacheStamp: Long = 0L,
    val isTO: Boolean = false,
    val toLink: String? = null,
    val fetched: Boolean = false,
    val errorMessage: String? = null,
    val successMessage: String? = null
) {
    val isEmojiAvatar: Boolean
        get() = avatarMode == AVATAR_MODE_EMOJI

    companion object {
        const val AVATAR_MODE_EMOJI = "emoji"
        const val AVATAR_MODE_IMAGE = "image"
        const val DEFAULT_EMOJI_AVATAR = ",🙂,#cccccc"
    }
}

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
    val isEmojiAvatar: Boolean = true,
    val isTO: Boolean = false,
    val toLink: String? = null,
    val fetched: Boolean = false,
    val errorMessage: String? = null,
    val successMessage: String? = null
)
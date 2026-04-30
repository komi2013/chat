package com.chat.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.data.repository.UserRepository
import com.chat.android.network.ApiService
import com.chat.android.network.UserResponse
import com.chat.android.network.NicknameResponse
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class UserUiState(
    val isLoading: Boolean = false,
    val user: UserResponse? = null,
    val nicknames: List<NicknameResponse> = emptyList(),
    val currentNickname: String? = null,
    val currentNickImg: String? = null,
    val currentNickBio: String? = null,
    val coordinateInput: String = "",
    val isNicknameFormVisible: Boolean = false,
    val editingMode: String = "", // "new" or "edit"
    val editableNickname: String = "",
    val errorMessage: String? = null,
    val successMessage: String? = null,
    val fetched: Boolean = false,
    val isTO: Boolean = false,
    val toLink: String? = null
)

@HiltViewModel
class UserViewModel @Inject constructor(
    private val userRepository: UserRepository,
    private val apiService: ApiService
) : ViewModel() {

    private val _uiState = MutableStateFlow(UserUiState())
    val uiState: StateFlow<UserUiState> = _uiState.asStateFlow()

    init {
        loadUserData()
        checkTOStatus()
    }

    private fun loadUserData() {
        viewModelScope.launch {
            _uiState.value = _uiState.value.copy(isLoading = true)
            
            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Not authenticated"
                    )
                    return@launch
                }

                val response = apiService.getUser(csrf)
                if (response.isSuccessful) {
                    val body = response.body()
                    if (body?.csrf != null) {
                        userRepository.setCsrfToken(body.csrf)
                        
                        // Handle push contents
                        body.pushContents?.forEach { content ->
                            // Handle push content if needed
                        }
                        
                        if (body.error != null) {
                            _uiState.value = _uiState.value.copy(
                                isLoading = false,
                                errorMessage = body.error
                            )
                        } else {
                            _uiState.value = _uiState.value.copy(
                                isLoading = false,
                                user = body.user,
                                nicknames = body.nicknames ?: emptyList(),
                                fetched = true
                            )
                            
                            // Set current nickname and bio
                            setCurrentNickname()
                            
                            // Set coordinate input if user has location
                            body.user?.let { user ->
                                if (user.latitude != null && user.longitude != null) {
                                    _uiState.value = _uiState.value.copy(
                                        coordinateInput = "${user.latitude}, ${user.longitude}"
                                    )
                                }
                            }
                        }
                    } else {
                        _uiState.value = _uiState.value.copy(
                            isLoading = false,
                            errorMessage = "Invalid response"
                        )
                    }
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Failed to load user data"
                    )
                }
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    isLoading = false,
                    errorMessage = "Error: ${e.message}"
                )
            }
        }
    }

    private fun setCurrentNickname() {
        val storedName = userRepository.getNickname()
        val nicknames = _uiState.value.nicknames
        
        val found = nicknames.find { it.nickname == storedName }
        if (found != null) {
            _uiState.value = _uiState.value.copy(
                currentNickname = found.nickname,
                currentNickImg = found.nickImg,
                currentNickBio = found.nickBio
            )
        } else if (nicknames.isNotEmpty()) {
            val first = nicknames.first()
            _uiState.value = _uiState.value.copy(
                currentNickname = first.nickname,
                currentNickImg = first.nickImg,
                currentNickBio = first.nickBio
            )
            userRepository.setNickname(first.nickname)
        }
    }

    private fun checkTOStatus() {
        val toLink = userRepository.getTO()
        if (toLink != null) {
            _uiState.value = _uiState.value.copy(
                isTO = true,
                toLink = toLink
            )
        }
    }

    fun openNicknameForm(mode: String) {
        val currentState = _uiState.value
        if (mode == "new" && currentState.nicknames.size >= 3) {
            _uiState.value = currentState.copy(
                errorMessage = "ニックネームは3つまで作成できます。"
            )
            return
        }

        _uiState.value = currentState.copy(
            isNicknameFormVisible = true,
            editingMode = mode,
            editableNickname = if (mode == "edit") currentState.currentNickname ?: "" else "",
            currentNickImg = if (mode == "new") "" else currentState.currentNickImg,
            currentNickBio = if (mode == "new") "" else currentState.currentNickBio
        )
    }

    fun closeNicknameForm() {
        _uiState.value = _uiState.value.copy(
            isNicknameFormVisible = false,
            editingMode = "",
            editableNickname = ""
        )
    }

    fun updateCoordinateInput(input: String) {
        _uiState.value = _uiState.value.copy(coordinateInput = input)
        parseCoordinates()
    }

    private fun parseCoordinates() {
        val input = _uiState.value.coordinateInput
        if (input.isEmpty()) return

        val parts = input.split(",").map { it.trim() }
        if (parts.size != 2) {
            updateUserLocation("", "")
            return
        }

        try {
            val lat = parts[0].toDouble()
            val lng = parts[1].toDouble()
            updateUserLocation(lat.toString(), lng.toString())
        } catch (e: NumberFormatException) {
            updateUserLocation("", "")
        }
    }

    private fun updateUserLocation(latitude: String, longitude: String) {
        val currentUser = _uiState.value.user
        val updatedUser = currentUser?.copy(
            latitude = latitude.toDoubleOrNull(),
            longitude = longitude.toDoubleOrNull()
        )
        _uiState.value = _uiState.value.copy(user = updatedUser)
    }

    fun updateEditableNickname(nickname: String) {
        _uiState.value = _uiState.value.copy(editableNickname = nickname)
    }

    fun updateNickImg(img: String) {
        _uiState.value = _uiState.value.copy(currentNickImg = img)
    }

    fun updateNickBio(bio: String) {
        _uiState.value = _uiState.value.copy(currentNickBio = bio)
    }

    fun submitUser() {
        viewModelScope.launch {
            val currentState = _uiState.value
            if (currentState.isNicknameFormVisible && currentState.editableNickname.isEmpty()) {
                _uiState.value = currentState.copy(
                    errorMessage = "ニックネームの入力してください"
                )
                return@launch
            }

            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) {
                    _uiState.value = currentState.copy(
                        errorMessage = "Not authenticated"
                    )
                    return@launch
                }

                val nickname = if (currentState.editingMode == "new") {
                    currentState.editableNickname
                } else {
                    currentState.currentNickname ?: ""
                }

                val response = apiService.editUser(
                    csrf = csrf,
                    nickname = nickname,
                    nickImg = currentState.currentNickImg,
                    nickBio = currentState.currentNickBio,
                    mail = currentState.user?.mail,
                    telephone = currentState.user?.telephone,
                    walletAddress = currentState.user?.walletAddress,
                    latitude = currentState.user?.latitude,
                    longitude = currentState.user?.longitude
                )

                if (response.isSuccessful) {
                    val body = response.body()
                    if (body?.csrf != null) {
                        userRepository.setCsrfToken(body.csrf)
                        
                        // Handle push contents
                        body.pushContents?.forEach { content ->
                            // Handle push content if needed
                        }
                        
                        if (body.error != null) {
                            _uiState.value = currentState.copy(errorMessage = body.error)
                        } else {
                            userRepository.setNickname(body.nickname ?: nickname)
                            _uiState.value = currentState.copy(
                                isNicknameFormVisible = false,
                                successMessage = body.message,
                                currentNickname = body.nickname
                            )
                            
                            // Reload user data to get updated nicknames
                            loadUserData()
                        }
                    }
                } else {
                    _uiState.value = currentState.copy(
                        errorMessage = "Failed to update user"
                    )
                }
            } catch (e: Exception) {
                _uiState.value = currentState.copy(
                    errorMessage = "Error: ${e.message}"
                )
            }
        }
    }

    fun switchNickname(selectedName: String) {
        val selected = _uiState.value.nicknames.find { it.nickname == selectedName }
        if (selected != null) {
            _uiState.value = _uiState.value.copy(
                currentNickname = selected.nickname,
                currentNickImg = selected.nickImg,
                currentNickBio = selected.nickBio,
                isNicknameFormVisible = false
            )
            userRepository.setNickname(selected.nickname)
        }
    }

    fun clearMessages() {
        _uiState.value = _uiState.value.copy(
            errorMessage = null,
            successMessage = null
        )
    }
}

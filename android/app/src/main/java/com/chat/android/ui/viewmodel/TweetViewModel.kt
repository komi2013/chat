package com.chat.android.ui.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.data.repository.UserRepository
import com.chat.android.network.ApiService
import com.chat.android.network.Tweet
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class TweetUiState(
    val isLoading: Boolean = false,
    val tweets: List<Tweet> = emptyList(),
    val newTweetContent: String = "",
    val errorMessage: String? = null,
    val successMessage: String? = null,
    val selectedDate: String? = null,
    val parentIDhtml: String? = null,
    val messageID: String? = null
)

@HiltViewModel
class TweetViewModel @Inject constructor(
    private val userRepository: UserRepository,
    private val apiService: ApiService
) : ViewModel() {

    private val _uiState = MutableStateFlow(TweetUiState())
    val uiState: StateFlow<TweetUiState> = _uiState.asStateFlow()

    init {
        loadTweets()
    }

    fun loadTweets(date: String? = null, parentIDhtml: String? = null, messageID: String? = null) {
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

                val response = apiService.getMessages(
                    csrf = csrf,
                    channelId = "", // Will be filled based on context
                    date = date,
                    parentIDhtml = parentIDhtml,
                    messageID = messageID
                )

                if (response.isSuccessful) {
                    val tweets = response.body() ?: emptyList()
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        tweets = tweets,
                        selectedDate = date,
                        parentIDhtml = parentIDhtml,
                        messageID = messageID
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Failed to load tweets"
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

    fun loadLatestTweets() {
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

                val response = apiService.getLatestTweets(csrf)
                if (response.isSuccessful) {
                    val tweets = response.body() ?: emptyList()
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        tweets = tweets
                    )
                } else {
                    _uiState.value = _uiState.value.copy(
                        isLoading = false,
                        errorMessage = "Failed to load latest tweets"
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

    fun updateNewTweetContent(content: String) {
        _uiState.value = _uiState.value.copy(newTweetContent = content)
    }

    fun postTweet(channelID: String? = null, parentID: String? = null) {
        viewModelScope.launch {
            val currentState = _uiState.value
            if (currentState.newTweetContent.isBlank()) {
                _uiState.value = currentState.copy(
                    errorMessage = "Tweet content cannot be empty"
                )
                return@launch
            }

            _uiState.value = currentState.copy(isLoading = true)

            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) {
                    _uiState.value = currentState.copy(
                        isLoading = false,
                        errorMessage = "Not authenticated"
                    )
                    return@launch
                }

                val response = apiService.postTweet(
                    csrf = csrf,
                    content = currentState.newTweetContent,
                    channelID = channelID,
                    parentID = parentID
                )

                if (response.isSuccessful) {
                    val body = response.body()
                    if (body?.csrf != null) {
                        userRepository.setCsrfToken(body.csrf)
                    }
                    
                    _uiState.value = currentState.copy(
                        isLoading = false,
                        newTweetContent = "",
                        successMessage = "Tweet posted successfully"
                    )
                    
                    // Reload tweets to show the new one
                    loadTweets(currentState.selectedDate, currentState.parentIDhtml, currentState.messageID)
                } else {
                    _uiState.value = currentState.copy(
                        isLoading = false,
                        errorMessage = "Failed to post tweet"
                    )
                }
            } catch (e: Exception) {
                _uiState.value = currentState.copy(
                    isLoading = false,
                    errorMessage = "Error: ${e.message}"
                )
            }
        }
    }

    fun toggleBookmark(tweetId: String, currentState: Boolean) {
        viewModelScope.launch {
            try {
                val csrf = userRepository.getCsrfToken()
                if (csrf.isNullOrEmpty()) return@launch

                // This would call a bookmark toggle endpoint
                // For now, we'll simulate it by updating local state
                val updatedTweets = _uiState.value.tweets.map { tweet ->
                    if (tweet.messageID == tweetId) {
                        tweet.copy(bookmark = !currentState)
                    } else tweet
                }
                
                _uiState.value = _uiState.value.copy(tweets = updatedTweets)
            } catch (e: Exception) {
                _uiState.value = _uiState.value.copy(
                    errorMessage = "Failed to toggle bookmark: ${e.message}"
                )
            }
        }
    }

    fun clearMessages() {
        _uiState.value = _uiState.value.copy(
            errorMessage = null,
            successMessage = null
        )
    }

    fun refresh() {
        loadTweets(
            _uiState.value.selectedDate,
            _uiState.value.parentIDhtml,
            _uiState.value.messageID
        )
    }
}

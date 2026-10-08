package com.chat.android.feature.reception

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.feature.thread.ThreadViewModel
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ReceptionThreadViewModel @Inject constructor(
    private val repository: ReceptionRepository,
    private val threadViewModel: ThreadViewModel
) : ViewModel() {

    private val _reception = MutableStateFlow<Reception?>(null)
    val reception: StateFlow<Reception?> = _reception

    private val _nickname = MutableStateFlow("")
    val nickname: StateFlow<String> = _nickname

    private val _errorMessage = MutableStateFlow<String?>(null)
    val errorMessage: StateFlow<String?> = _errorMessage

    private val _isFetched = MutableStateFlow(false)
    val isFetched: StateFlow<Boolean> = _isFetched

    fun init(channelID: String, code: String, backID: String?) {
        viewModelScope.launch {
            repository.getReception(channelID, code).onSuccess { response ->
                _reception.value = response.reception
                _nickname.value = response.nickname ?: ""
                
                val parentID = "@" + (response.nickname ?: "")
                threadViewModel.init(channelID, parentID, backID)
                
                _isFetched.value = true
            }.onFailure {
                _errorMessage.value = it.message
            }
        }
    }
    
    val threadUiState = threadViewModel.uiState
    
    fun postMessage(text: String) {
        threadViewModel.postMessage(text)
    }
}

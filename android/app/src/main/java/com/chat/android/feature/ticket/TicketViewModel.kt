package com.chat.android.feature.ticket

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.network.Ticket
import com.chat.android.core.repository.UserRepository
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class TicketViewModel @Inject constructor(
    private val repository: TicketRepository,
    private val userRepository: UserRepository
) : ViewModel() {

    private val _tickets = MutableStateFlow<List<Ticket>>(emptyList())
    val tickets: StateFlow<List<Ticket>> = _tickets.asStateFlow()

    private val _isLoading = MutableStateFlow(false)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()

    private val _error = MutableStateFlow<String?>(null)
    val error: StateFlow<String?> = _error.asStateFlow()

    init {
        loadTickets()
        viewModelScope.launch {
            repository.dbUpdateFlow.collect {
                loadTickets()
            }
        }
    }

    fun loadTickets() {
        viewModelScope.launch {
            _tickets.value = repository.getTickets()
            _isLoading.value = true
            repository.fetchTickets().onFailure {
                _error.value = it.message
            }
            _tickets.value = repository.getTickets()
            _isLoading.value = false
        }
    }

    fun createTicket(title: String, description: String, status: String, priority: String) {
        val channelID = userRepository.getCurrentChannelId() ?: return
        val updatedBy = userRepository.getNickname() ?: ""
        viewModelScope.launch {
            _isLoading.value = true
            repository.createTicket(channelID, updatedBy, title, description, status, priority)
                .onFailure { _error.value = it.message }
            _isLoading.value = false
        }
    }

    fun updateTicket(ticketID: String, title: String?, description: String?, status: String?, priority: String?) {
        val channelID = userRepository.getCurrentChannelId() ?: return
        val updatedBy = userRepository.getNickname() ?: ""
        viewModelScope.launch {
            _isLoading.value = true
            repository.updateTicket(channelID, updatedBy, ticketID, title, description, status, priority)
                .onFailure { _error.value = it.message }
            _isLoading.value = false
        }
    }

    fun deleteTicket(ticketID: String) {
        val channelID = userRepository.getCurrentChannelId() ?: return
        val updatedBy = userRepository.getNickname() ?: ""
        viewModelScope.launch {
            _isLoading.value = true
            repository.deleteTicket(channelID, updatedBy, ticketID)
                .onFailure { _error.value = it.message }
            _isLoading.value = false
        }
    }
    
    fun clearError() {
        _error.value = null
    }
}

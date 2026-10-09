package com.chat.android.feature.reception

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*
import javax.inject.Inject

@HiltViewModel
class ReceptionEnterViewModel @Inject constructor(
    private val repository: ReceptionRepository
) : ViewModel() {

    var reception by mutableStateOf<Reception?>(null)
        private set

    var waitMinutes by mutableStateOf(0)
        private set

    var inputGuests by mutableStateOf(1)
        private set

    var assignedFacility by mutableStateOf<Facility?>(null)
        private set

    var errorMessage by mutableStateOf<String?>(null)
        private set

    var isLoading by mutableStateOf(false)
        private set

    fun onGuestCountChange(count: Int) {
        inputGuests = count
        calculateWaitMinutes()
    }

    fun loadReception(id: String, passkey: String) {
        viewModelScope.launch {
            isLoading = true
            errorMessage = null
            repository.getReception(id, passkey).onSuccess { response ->
                reception = response.reception
                calculateWaitMinutes()
            }.onFailure {
                errorMessage = it.message
            }
            isLoading = false
        }
    }

    fun calculateWaitMinutes() {
        val currentReception = reception ?: return
        val guestCount = inputGuests
        val queues = currentReception.queues.toMutableList()
        val facilities = currentReception.facilities
        val now = Date()

        assignedFacility = null

        // --- すでに待機列がある場合はそれを優先 ---
        if (queues.isNotEmpty()) {
            queues.add(Queue(waitingGuest = guestCount))
            var totalWait = 0
            queues.forEach { queue ->
                val guests = queue.waitingGuest
                val config = currentReception.waitConfigs.find { cfg ->
                    val min = cfg.guestRange.getOrNull(0) ?: 0
                    val max = cfg.guestRange.getOrNull(1) ?: 0
                    guests >= min && (max == 0 || guests <= max)
                }
                val waitRatio = config?.waitRatio ?: 1
                totalWait += waitRatio
            }
            waitMinutes = totalWait
            return
        }

        // --- 待機列が無い場合のみ、新規割り当てを試みる ---
        if (guestCount > 0) {
            val df = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss", Locale.getDefault())
            
            val availableFacilities = facilities.filter { f ->
                // Vue code: if (f.currentCode) return false
                // But Facility model doesn't have currentCode in Kotlin yet? 
                // Let's check the Vue logic and Reception.kt
                
                val config = currentReception.waitConfigs.find { cfg ->
                    val min = cfg.guestRange.getOrNull(0) ?: 0
                    val max = cfg.guestRange.getOrNull(1) ?: 0
                    guestCount >= min && (max == 0 || guestCount <= max)
                }
                val waitRatio = config?.waitRatio ?: 1
                val leaveTime = now.time + waitRatio * 60000L

                val conflicts = currentReception.books.filter { it.bookID.isNotEmpty() }.any { bt ->
                    try {
                        val start = df.parse(bt.bookStart)?.time ?: 0
                        val end = df.parse(bt.bookEnd)?.time ?: 0
                        val inReservedNow = now.time in start..end
                        val overlapWithFutureBooking = leaveTime > start && now.time < end
                        inReservedNow || overlapWithFutureBooking
                    } catch (e: Exception) {
                        false
                    }
                }
                if (conflicts) return@filter false

                if (!f.currentCode.isNullOrEmpty()) return@filter false

                val available = f.capacity - (f.currentGuests ?: 0)
                available > 0
            }

            if (availableFacilities.isNotEmpty()) {
                assignedFacility = availableFacilities
                    .filter { it.capacity >= guestCount }
                    .minByOrNull { it.capacity }
                    ?: availableFacilities.minByOrNull { it.capacity }
            }
        }

        // --- 待機時間を再計算 ---
        var totalWait = 0
        queues.forEach { queue ->
            val guests = queue.waitingGuest
            val config = currentReception.waitConfigs.find { cfg ->
                val min = cfg.guestRange.getOrNull(0) ?: 0
                val max = cfg.guestRange.getOrNull(1) ?: 0
                guests >= min && (max == 0 || guests <= max)
            }
            val waitRatio = config?.waitRatio ?: 1
            totalWait += waitRatio
        }
        waitMinutes = totalWait
    }

    fun queueUp(receptionID: String, passkey: String, onSuccess: () -> Unit) {
        viewModelScope.launch {
            errorMessage = null
            repository.queueEdit(
                receptionID = receptionID,
                code = passkey,
                guestCount = inputGuests,
                queueName = null,
                editType = 1
            ).onSuccess {
                onSuccess()
            }.onFailure {
                errorMessage = it.message
            }
        }
    }

    fun queueRemove(receptionID: String, queueName: String, onSuccess: () -> Unit) {
        viewModelScope.launch {
            errorMessage = null
            repository.queueEdit(
                receptionID = receptionID,
                code = null,
                guestCount = null,
                queueName = queueName,
                editType = 2
            ).onSuccess {
                onSuccess()
            }.onFailure {
                errorMessage = it.message
            }
        }
    }
}

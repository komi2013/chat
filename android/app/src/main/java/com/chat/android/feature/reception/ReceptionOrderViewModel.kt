package com.chat.android.feature.reception

import android.content.Context
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.push.PushReceiveDispatcher
import dagger.hilt.android.lifecycle.HiltViewModel
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.launch
import javax.inject.Inject

@HiltViewModel
class ReceptionOrderViewModel @Inject constructor(
    @ApplicationContext private val context: Context,
    private val repository: ReceptionRepository,
    private val dbHelper: ReceptionDbHelper,
    private val pushDispatcher: PushReceiveDispatcher
) : ViewModel() {

    var reception by mutableStateOf<Reception?>(null)
        private set

    var receptionOrders = mutableStateListOf<ReceptionOrder>()
        private set

    var errorMessage by mutableStateOf<String?>(null)
        private set

    var totalPrice by mutableStateOf(0)
        private set

    var seatName by mutableStateOf("")
        private set

    var iamStaff by mutableStateOf(false)
        private set

    var isLoading by mutableStateOf(false)
        private set

    private var updateJob: kotlinx.coroutines.Job? = null

    fun init(id: String, code: String) {
        viewModelScope.launch {
            isLoading = true
            errorMessage = null
            
            repository.getReception(id, code).onSuccess { res ->
                reception = res.reception
                seatName = res.facilityName ?: ""
                
                val myname = repository.getAliasName()
                iamStaff = res.reception?.joinNames?.contains(myname) == true
                
                fetchOrders()
            }.onFailure {
                errorMessage = it.message
            }
            isLoading = false
        }

        updateJob?.cancel()
        updateJob = viewModelScope.launch {
            pushDispatcher.onUpdated.collect { updatedChannelID ->
                if (updatedChannelID == reception?.channelID) {
                    fetchOrders()
                }
            }
        }
    }

    private fun fetchOrders() {
        val allOrders = dbHelper.getAllOrders()
        val seatOrders = allOrders.filter { it.seatName == seatName }
        receptionOrders.clear()
        receptionOrders.addAll(seatOrders)
        totalPrice = seatOrders.sumOf { it.price }
    }

    fun getItemDetail(itemID: Int): ItemDetail? {
        return reception?.itemDetails?.find { it.itemID == itemID }
    }

    fun deleteOrder(id: String, code: String) {
        viewModelScope.launch {
            isLoading = true
            repository.deleteOrder(id, code).onSuccess {
                dbHelper.deleteOrdersForSeat(seatName)
                fetchOrders()
            }.onFailure {
                errorMessage = it.message
            }
            isLoading = false
        }
    }

    override fun onCleared() {
        updateJob?.cancel()
        super.onCleared()
    }
}

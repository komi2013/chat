package com.chat.android.feature.reception

import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.launch
import javax.inject.Inject

data class SelectedOptions(
    val paidOptions: MutableList<Int> = mutableStateListOf(),
    val freeOptions: MutableList<Int?> = mutableStateListOf(),
    val freeMultiOptions: MutableList<Int> = mutableStateListOf()
)

@HiltViewModel
class ReceptionMenuViewModel @Inject constructor(
    private val repository: ReceptionRepository
) : ViewModel() {

    var reception by mutableStateOf<Reception?>(null)
        private set

    var seatName by mutableStateOf("")
        private set

    var nickname by mutableStateOf("")
        private set

    var errorMessage by mutableStateOf<String?>(null)
        private set

    var isLoading by mutableStateOf(false)
        private set

    // menuID -> (key -> choice)
    val selectedChoices = mutableMapOf<Int, MutableMap<String, String>>()
    
    // menuID -> options
    val selectedOptions = mutableMapOf<Int, SelectedOptions>()

    val receptionOrders = mutableStateListOf<ReceptionOrder>()

    val totalPrice: Int
        get() = receptionOrders.sumOf { it.price }

    private var cartItemSeq = 0

    fun loadReception(id: String, code: String, codeType: String, onRedirect: (String) -> Unit) {
        if (codeType == "2") {
            onRedirect("3")
            return
        }

        viewModelScope.launch {
            isLoading = true
            errorMessage = null
            repository.getReception(id, code).onSuccess { response ->
                val r = response.reception ?: return@onSuccess
                
                // Merge menus and items from response.menu if available
                val menus = response.menu?.menus ?: r.menus
                val items = response.menu?.itemDetails ?: r.itemDetails
                
                reception = r.copy(menus = menus, itemDetails = items)
                seatName = response.reception.facilities.firstOrNull()?.facilityName ?: "" // Vue uses res.facilityName, assuming it's in response
                // nickname = response.nickname // Need to add nickname to ReceptionGetResponse if it exists
                
                initSelections(menus, items)
            }.onFailure {
                errorMessage = it.message
            }
            isLoading = false
        }
    }

    private fun initSelections(menus: List<Menu>, items: List<ItemDetail>) {
        val itemMap = items.associateBy { it.itemID }
        
        menus.forEach { menu ->
            val choicesMap = mutableStateMapOf<String, String>()
            menu.items.forEach { itemID ->
                itemMap[itemID]?.choices?.forEachIndexed { groupIndex, group ->
                    val key = "item-$itemID-$groupIndex"
                    choicesMap[key] = group.firstOrNull() ?: ""
                }
            }
            selectedChoices[menu.menuID] = choicesMap

            val freeOptionsDefaults = menu.freeOptions.map { it.firstOrNull() }
            selectedOptions[menu.menuID] = SelectedOptions(
                freeOptions = mutableStateListOf(*freeOptionsDefaults.toTypedArray())
            )
        }
    }

    fun getItemDetail(itemID: Int): ItemDetail? {
        return reception?.itemDetails?.find { it.itemID == itemID }
    }

    fun addCart(menu: Menu) {
        val seq = ++cartItemSeq
        val opts = selectedOptions[menu.menuID] ?: SelectedOptions()
        val choices = selectedChoices[menu.menuID] ?: emptyMap<String, String>()

        val paidOptions = opts.paidOptions.map { id ->
            val optDef = menu.paidOptions.find { it.itemID == id }
            ReceptionOrderPaidOption(itemID = id, price = optDef?.price ?: 0)
        }
        val paidOptionsPrice = paidOptions.sumOf { it.price }
        val price = menu.price + paidOptionsPrice

        val itemDetails = menu.items.mapNotNull { getItemDetail(it) }

        val order = ReceptionOrder(
            receptionOrderID = "${menu.menuID}${nickname}${seq}",
            seatName = seatName,
            menuID = menu.menuID,
            menuName = menu.menuName,
            itemChoices = choices.toMap(),
            freeOptions = opts.freeOptions.filterNotNull(),
            paidOptions = paidOptions,
            freeMultiOptions = opts.freeMultiOptions.toList(),
            price = price,
            items = itemDetails.map { d ->
                ReceptionOrderItem(
                    itemID = d.itemID,
                    itemName = d.itemName,
                    imgPath = d.imgPath,
                    choices = d.choices
                )
            }
        )
        receptionOrders.add(order)
    }

    fun removeOrder(orderID: String) {
        receptionOrders.removeAll { it.receptionOrderID == orderID }
    }

    fun submitOrder(receptionID: String, code: String, onSuccess: () -> Unit) {
        viewModelScope.launch {
            isLoading = true
            errorMessage = null
            repository.orderReception(receptionID, receptionOrders.toList(), code).onSuccess {
                onSuccess()
            }.onFailure {
                errorMessage = it.message
            }
            isLoading = false
        }
    }
}

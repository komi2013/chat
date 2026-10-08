package com.chat.android.feature.reception

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import javax.inject.Inject

data class ReceptionUiState(
    val reception: Reception = Reception(),
    val aliases: List<DbAlias> = emptyList(),
    val groups: List<DbGroup> = emptyList(),
    val loading: Boolean = false,
    val saving: Boolean = false,
    val error: String? = null,
    val success: Boolean = false
)

@HiltViewModel
class ReceptionViewModel @Inject constructor(
    private val repository: ReceptionRepository
) : ViewModel() {

    private val _state = MutableStateFlow(ReceptionUiState())
    val state = _state.asStateFlow()

    fun loadReception(id: String, code: String? = null) {
        viewModelScope.launch {
            _state.value = _state.value.copy(loading = true, error = null)
            val channelID = repository.getChannelId()
            val aliases = repository.getAliases(channelID)
            val groups = repository.getGroups(channelID)

            val result = repository.getReception(id, code)
            result.onSuccess { response ->
                _state.value = _state.value.copy(
                    reception = response.reception ?: Reception(receptionID = id, channelID = channelID),
                    aliases = aliases,
                    groups = groups,
                    loading = false
                )
            }.onFailure { error ->
                _state.value = _state.value.copy(
                    aliases = aliases,
                    groups = groups,
                    loading = false,
                    error = error.message ?: "データの取得に失敗しました"
                )
            }
        }
    }

    fun updateReception(transform: Reception.() -> Reception) {
        _state.value = _state.value.copy(reception = _state.value.reception.transform())
    }

    fun submit() {
        viewModelScope.launch {
            _state.value = _state.value.copy(saving = true, error = null, success = false)
            val reception = _state.value.reception
            val result = repository.editReception(
                reception = reception,
                menus = reception.menus,
                itemDetails = reception.itemDetails
            )
            result.onSuccess {
                _state.value = _state.value.copy(saving = false, success = true)
            }.onFailure { error ->
                _state.value = _state.value.copy(
                    saving = false,
                    error = error.message ?: "送信に失敗しました"
                )
            }
        }
    }

    // Helper methods for list manipulations
    fun addSkill() = updateReception { copy(skills = skills + "") }
    fun removeSkill(index: Int) = updateReception { copy(skills = skills.filterIndexed { i, _ -> i != index }) }
    fun updateSkill(index: Int, value: String) = updateReception {
        copy(skills = skills.mapIndexed { i, s -> if (i == index) value else s })
    }

    fun addFacility() = updateReception { copy(facilities = facilities + Facility()) }
    fun removeFacility(index: Int) = updateReception { copy(facilities = facilities.filterIndexed { i, _ -> i != index }) }
    
    fun addMenu() = updateReception { 
        copy(menus = menus + Menu(menuID = menus.size + 1)) 
    }
    fun removeMenu(index: Int) = updateReception { copy(menus = menus.filterIndexed { i, _ -> i != index }) }

    fun addItemDetail() = updateReception {
        copy(itemDetails = itemDetails + ItemDetail(itemID = itemDetails.size + 1))
    }
    fun removeItemDetail(index: Int) = updateReception { 
        copy(itemDetails = itemDetails.filterIndexed { i, _ -> i != index }) 
    }

    fun addPasscode() = updateReception { copy(passcodes = passcodes + Passcode()) }
    fun removePasscode(index: Int) = updateReception { copy(passcodes = passcodes.filterIndexed { i, _ -> i != index }) }

    fun addAsk() = updateReception { copy(asks = asks + Ask(sequence = asks.size)) }
    fun removeAsk(index: Int) = updateReception { copy(asks = asks.filterIndexed { i, _ -> i != index }) }
    fun addAskChoice() = updateReception { copy(askChoices = askChoices + Ask(sequence = askChoices.size)) }
    fun removeAskChoice(index: Int) = updateReception { copy(askChoices = askChoices.filterIndexed { i, _ -> i != index }) }
    fun addAskMultiChoice() = updateReception { copy(askMultiChoices = askMultiChoices + Ask(sequence = askMultiChoices.size)) }
    fun removeAskMultiChoice(index: Int) = updateReception { copy(askMultiChoices = askMultiChoices.filterIndexed { i, _ -> i != index }) }

    fun addOpenTime() = updateReception { copy(openTimes = openTimes + OpenTime()) }
    fun removeOpenTime(index: Int) = updateReception { copy(openTimes = openTimes.filterIndexed { i, _ -> i != index }) }
    fun addShift() = updateReception { copy(shifts = shifts + Shift(aliasNames = listOf(""))) }
    fun removeShift(index: Int) = updateReception { copy(shifts = shifts.filterIndexed { i, _ -> i != index }) }
    fun addWaitConfig() = updateReception { copy(waitConfigs = waitConfigs + WaitConfig()) }

    // Nested updates for Menus
    fun addMenuPaidOption(menuIndex: Int) = updateReception {
        copy(menus = menus.toMutableList().apply {
            this[menuIndex] = this[menuIndex].copy(
                paidOptions = this[menuIndex].paidOptions + PaidOption()
            )
        })
    }
    fun removeMenuPaidOption(menuIndex: Int, optIndex: Int) = updateReception {
        copy(menus = menus.toMutableList().apply {
            this[menuIndex] = this[menuIndex].copy(
                paidOptions = this[menuIndex].paidOptions.filterIndexed { i, _ -> i != optIndex }
            )
        })
    }

    fun addMenuFreeOptionGroup(menuIndex: Int) = updateReception {
        copy(menus = menus.toMutableList().apply {
            this[menuIndex] = this[menuIndex].copy(
                freeOptions = this[menuIndex].freeOptions + listOf(listOf(0))
            )
        })
    }

    fun addStaffSkill() = updateReception {
        copy(staffSkills = staffSkills + StaffSkill(skills = listOf("")))
    }
    fun removeStaffSkill(index: Int) = updateReception { copy(staffSkills = staffSkills.filterIndexed { i, _ -> i != index }) }
    
    fun addStaffSkillEntry(staffIndex: Int) = updateReception {
        copy(staffSkills = staffSkills.toMutableList().apply {
            this[staffIndex] = this[staffIndex].copy(
                skills = this[staffIndex].skills + ""
            )
        })
    }
    fun removeStaffSkillEntry(staffIndex: Int, skillIndex: Int) = updateReception {
        copy(staffSkills = staffSkills.toMutableList().apply {
            this[staffIndex] = this[staffIndex].copy(
                skills = this[staffIndex].skills.filterIndexed { i, _ -> i != skillIndex }
            )
        })
    }

    fun addItemDetailChoiceGroup(itemIndex: Int) = updateReception {
        copy(itemDetails = itemDetails.toMutableList().apply {
            this[itemIndex] = this[itemIndex].copy(
                choices = this[itemIndex].choices + listOf(listOf("硬い", "普通", "柔らかい"))
            )
        })
    }

    fun addItemDetailChoice(itemIndex: Int, groupIndex: Int) = updateReception {
        copy(itemDetails = itemDetails.toMutableList().apply {
            val updatedChoices = this[itemIndex].choices.toMutableList().apply {
                this[groupIndex] = this[groupIndex] + ""
            }
            this[itemIndex] = this[itemIndex].copy(choices = updatedChoices)
        })
    }
}

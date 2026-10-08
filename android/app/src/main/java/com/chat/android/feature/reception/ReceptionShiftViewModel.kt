package com.chat.android.feature.reception

import androidx.compose.runtime.State
import androidx.compose.runtime.mutableStateOf
import androidx.lifecycle.SavedStateHandle
import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.core.network.SessionManager
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.launch
import java.text.SimpleDateFormat
import java.util.*
import javax.inject.Inject

@HiltViewModel
class ReceptionShiftViewModel @Inject constructor(
    val repository: ReceptionRepository,
    private val sessionManager: SessionManager,
    savedStateHandle: SavedStateHandle
) : ViewModel() {

    private val receptionID: String = savedStateHandle["receptionID"] ?: ""
    private val initialDate: String? = savedStateHandle["date"]

    private val _uiState = MutableStateFlow<ReceptionShiftUiState>(ReceptionShiftUiState.Loading)
    val uiState: StateFlow<ReceptionShiftUiState> = _uiState

    private val _schedules = mutableStateOf<List<ScheduleItem>>(emptyList())
    val schedules: State<List<ScheduleItem>> = _schedules

    private val _availableSkills = mutableStateOf<List<String>>(emptyList())
    val availableSkills: State<List<String>> = _availableSkills

    private val _iamAdmin = mutableStateOf(false)
    val iamAdmin: State<Boolean> = _iamAdmin

    private val _reception = mutableStateOf<Reception?>(null)
    val reception: State<Reception?> = _reception

    private val _errorMessage = mutableStateOf<String?>(null)
    val errorMessage: State<String?> = _errorMessage

    private val sdf = SimpleDateFormat("yyyy-MM-dd", Locale.getDefault())
    private val dateTimeSdf = SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault())

    val monthDates: List<Date> by lazy {
        val today = if (initialDate != null) sdf.parse(initialDate) ?: Date() else Date()
        val calendar = Calendar.getInstance()
        calendar.time = today
        List(30) {
            val date = calendar.time
            calendar.add(Calendar.DAY_OF_YEAR, 1)
            date
        }
    }

    val hours = (0..23).toList()

    init {
        loadReception()
    }

    private fun loadReception() {
        viewModelScope.launch {
            _uiState.value = ReceptionShiftUiState.Loading
            repository.getReception(receptionID, null).onSuccess { response ->
                val rec = response.reception ?: return@onSuccess
                _reception.value = rec
                val myName = repository.getAliasName()
                _iamAdmin.value = rec.adminNames.contains(myName)

                val matchedStaff = rec.staffSkills.find { it.aliasName == myName }
                _availableSkills.value = matchedStaff?.skills ?: emptyList()

                generateSchedules(rec, myName)
                _uiState.value = ReceptionShiftUiState.Success
            }.onFailure {
                _errorMessage.value = it.message
                _uiState.value = ReceptionShiftUiState.Error(it.message ?: "Unknown error")
            }
        }
    }

    private fun generateSchedules(rec: Reception, myName: String) {
        val items = mutableListOf<ScheduleItem>()
        rec.shifts.forEach { shift ->
            if (availableSkills.value.contains(shift.skill) || iamAdmin.value) {
                items.add(
                    ScheduleItem(
                        key = "${shift.shiftStart}_${shift.skill}",
                        aliasNames = shift.aliasNames,
                        timeStart = shift.shiftStart,
                        timeEnd = shift.shiftEnd,
                        skill = shift.skill,
                        open = shift.open,
                        start = shift.shiftStart,
                        end = shift.shiftEnd,
                        fix = shift.fix
                    )
                )
            }
        }
        _schedules.value = items
    }

    fun getEventsForDayAndHour(day: Date, hour: Int): List<ScheduleItem> {
        val dayCalendar = Calendar.getInstance().apply { time = day }
        return _schedules.value.filter { event ->
            val eventStart = dateTimeSdf.parse(event.timeStart) ?: return@filter false
            val eventCalendar = Calendar.getInstance().apply { time = eventStart }
            eventCalendar.get(Calendar.YEAR) == dayCalendar.get(Calendar.YEAR) &&
                    eventCalendar.get(Calendar.MONTH) == dayCalendar.get(Calendar.MONTH) &&
                    eventCalendar.get(Calendar.DAY_OF_MONTH) == dayCalendar.get(Calendar.DAY_OF_MONTH) &&
                    eventCalendar.get(Calendar.HOUR_OF_DAY) == hour
        }
    }

    fun moveStaffToTop(scheduleKey: String, aliasName: String) {
        if (!iamAdmin.value) return
        val list = _schedules.value.toMutableList()
        val index = list.indexOfFirst { it.key == scheduleKey }
        if (index != -1) {
            val item = list[index]
            val aliasList = item.aliasNames.toMutableList()
            val staffIndex = aliasList.indexOf(aliasName)
            if (staffIndex > 0) {
                aliasList.add(0, aliasList.removeAt(staffIndex))
                list[index] = item.copy(aliasNames = aliasList)
                _schedules.value = list
            }
        }
    }

    fun onOffOK(scheduleKey: String) {
        val myName = repository.getAliasName()
        val list = _schedules.value.toMutableList()
        val index = list.indexOfFirst { it.key == scheduleKey }
        if (index != -1) {
            val item = list[index]
            if (!iamAdmin.value && item.fix) return
            
            val aliasList = item.aliasNames.toMutableList()
            val staffIndex = aliasList.indexOf(myName)
            if (staffIndex == -1) {
                aliasList.add(myName)
            } else {
                aliasList.removeAt(staffIndex)
            }
            list[index] = item.copy(aliasNames = aliasList)
            _schedules.value = list
        }
    }

    fun toggleFix(scheduleKey: String) {
        if (!iamAdmin.value) return
        val list = _schedules.value.toMutableList()
        val index = list.indexOfFirst { it.key == scheduleKey }
        if (index != -1) {
            val item = list[index]
            list[index] = item.copy(fix = !item.fix)
            _schedules.value = list
        }
    }

    fun submitShift() {
        val currentReception = _reception.value ?: return
        val updatedShifts = currentReception.shifts.mapNotNull { shift ->
            val key = "${shift.shiftStart}_${shift.skill}"
            val schedule = _schedules.value.find { it.key == key }
            if (schedule != null && (schedule.aliasNames != shift.aliasNames || schedule.fix != shift.fix)) {
                Shift(
                    aliasNames = schedule.aliasNames,
                    shiftStart = schedule.start,
                    shiftEnd = schedule.end,
                    skill = schedule.skill,
                    open = schedule.open,
                    fix = schedule.fix
                )
            } else {
                null
            }
        }

        if (updatedShifts.isNotEmpty()) {
            viewModelScope.launch {
                repository.submitShift(receptionID, updatedShifts).onSuccess {
                    loadReception() // Reload to get fresh data
                }.onFailure {
                    _errorMessage.value = it.message
                }
            }
        }
    }
}

data class ScheduleItem(
    val key: String,
    val aliasNames: List<String>,
    val timeStart: String,
    val timeEnd: String,
    val skill: String,
    val open: Int,
    val start: String,
    val end: String,
    var fix: Boolean
)

sealed class ReceptionShiftUiState {
    object Loading : ReceptionShiftUiState()
    object Success : ReceptionShiftUiState()
    data class Error(val message: String) : ReceptionShiftUiState()
}

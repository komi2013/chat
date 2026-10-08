package com.chat.android.feature.reception

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.chat.android.feature.channel.DbAlias
import com.chat.android.feature.channel.DbGroup
import com.google.gson.Gson
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter
import java.time.temporal.ChronoUnit
import javax.inject.Inject

data class ScheduleSlot(
    val timeStart: String,
    val timeEnd: String,
    val facility: String? = null,
    val index: Int = 0
)

data class ReceptionBookUiState(
    val reception: Reception? = null,
    val menus: List<Menu> = emptyList(),
    val itemDetails: List<ItemDetail> = emptyList(),
    val slots: List<ScheduleSlot> = emptyList(),
    val selectedMenuID: Int? = null,
    val loading: Boolean = false,
    val error: String? = null,
    val bookingSuccess: Boolean = false
)

@HiltViewModel
class ReceptionBookViewModel @Inject constructor(
    private val repository: ReceptionRepository
) : ViewModel() {

    private val _state = MutableStateFlow(ReceptionBookUiState())
    val state = _state.asStateFlow()

    private val isoFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm:ss")
    private val dateOnlyFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd")

    fun loadReception(id: String, menuID: Int? = null) {
        viewModelScope.launch {
            _state.value = _state.value.copy(loading = true, error = null)
            val result = repository.getReception(id, null)
            result.onSuccess { response ->
                val reception = response.reception
                if (reception != null) {
                    val menus = response.menu?.menus ?: emptyList()
                    val itemDetails = response.menu?.itemDetails ?: emptyList()
                    _state.value = _state.value.copy(
                        reception = reception,
                        menus = menus,
                        itemDetails = itemDetails,
                        selectedMenuID = menuID,
                        loading = false
                    )
                    generateSchedules(menuID)
                } else {
                    _state.value = _state.value.copy(loading = false, error = "受信データがありません")
                }
            }.onFailure { error ->
                _state.value = _state.value.copy(loading = false, error = error.message)
            }
        }
    }

    fun selectMenu(menuID: Int) {
        _state.value = _state.value.copy(selectedMenuID = menuID)
        generateSchedules(menuID)
    }

    private fun generateSchedules(menuID: Int?) {
        val reception = _state.value.reception ?: return
        val today = LocalDateTime.now()
        val days = (0 until 30).map { today.plusDays(it.toLong()) }
        
        val selectedMenu = if (menuID != null) _state.value.menus.find { it.menuID == menuID } else null
        val needSkill = selectedMenu?.needSkill
        val needFacility = selectedMenu?.needFacility

        val tempSlots = mutableListOf<ScheduleSlot>()

        if (!reception.workStaffNeed) {
            days.forEach { day ->
                val dateStr = day.format(dateOnlyFormatter)
                reception.openTimes.forEach { open ->
                    if (!open.limitStart.startsWith(dateStr)) return@forEach
                    
                    val start = LocalDateTime.parse(open.limitStart.substring(0, 19), isoFormatter)
                    val end = LocalDateTime.parse(open.limitEnd.substring(0, 19), isoFormatter)

                    if (reception.facilities.isNotEmpty()) {
                        reception.facilities.forEach { facility ->
                            if (!facility.bookable) return@forEach
                            val blocks = splitBlock(start, end, reception.books, facility.capacity)
                            for (i in 0 until (if (facility.capacity > 0) facility.capacity else 1)) {
                                blocks.forEach { block ->
                                    tempSlots.add(ScheduleSlot(block.start.format(isoFormatter), block.end.format(isoFormatter), facility.facilityName, i))
                                }
                            }
                        }
                    } else {
                        val blocks = splitBlock(start, end, reception.books, 1)
                        blocks.forEach { block ->
                            tempSlots.add(ScheduleSlot(block.start.format(isoFormatter), block.end.format(isoFormatter)))
                        }
                    }
                }
            }
        } else {
            // workStaffNeed == true
            days.forEach { day ->
                val dateStr = day.format(dateOnlyFormatter)
                reception.workStaffs.forEach { work ->
                    if (!work.workStart.startsWith(dateStr)) return@forEach
                    
                    if (needSkill != null) {
                        val valid = reception.staffSkills.any { it.aliasName == work.aliasName && it.skills.contains(needSkill) }
                        if (!valid) return@forEach
                    }

                    val start = LocalDateTime.parse(work.workStart.substring(0, 19), isoFormatter)
                    val end = LocalDateTime.parse(work.workEnd.substring(0, 19), isoFormatter)
                    
                    val blocks = splitBlock(start, end, reception.books, 1)
                    blocks.forEach { block ->
                        tempSlots.add(ScheduleSlot(block.start.format(isoFormatter), block.end.format(isoFormatter)))
                    }
                }
            }
        }

        // Apply facility constraints if needed
        var finalSlots = tempSlots
        if (reception.workStaffNeed && !needFacility.isNullOrEmpty()) {
            val unavailable = makeUnavailableTime(needFacility, reception)
            finalSlots = mutableListOf<ScheduleSlot>()
            tempSlots.forEach { slot ->
                val start = LocalDateTime.parse(slot.timeStart, isoFormatter)
                val end = LocalDateTime.parse(slot.timeEnd, isoFormatter)
                val blocks = splitBlock(start, end, unavailable, 1)
                blocks.forEach { block ->
                    finalSlots.add(ScheduleSlot(block.start.format(isoFormatter), block.end.format(isoFormatter)))
                }
            }
        }

        _state.value = _state.value.copy(slots = mergeSchedules(finalSlots))
    }

    private data class Block(val start: LocalDateTime, val end: LocalDateTime, var free: Int)

    private fun splitBlock(start: LocalDateTime, end: LocalDateTime, books: List<Book>, capacity: Int): List<Block> {
        var blocks = mutableListOf(Block(start, end, capacity))
        val relevantBooks = books.filter { isOverlapping(start, end, it.bookStart, it.bookEnd) }
            .sortedBy { it.bookStart }

        relevantBooks.forEach { book ->
            val bookStart = LocalDateTime.parse(book.bookStart.substring(0, 19), isoFormatter)
            val bookEnd = LocalDateTime.parse(book.bookEnd.substring(0, 19), isoFormatter)
            val nextBlocks = mutableListOf<Block>()

            blocks.forEach { b ->
                if (bookStart >= b.end || bookEnd <= b.start) {
                    nextBlocks.add(b)
                    return@forEach
                }
                if (bookStart > b.start) {
                    nextBlocks.add(Block(b.start, bookStart, b.free))
                }
                val overlapStart = if (b.start.isAfter(bookStart)) b.start else bookStart
                val overlapEnd = if (b.end.isBefore(bookEnd)) b.end else bookEnd
                if (overlapStart.isBefore(overlapEnd)) {
                    nextBlocks.add(Block(overlapStart, overlapEnd, b.free - 1))
                }
                if (bookEnd < b.end) {
                    nextBlocks.add(Block(bookEnd, b.end, b.free))
                }
            }
            blocks = nextBlocks
        }
        return blocks.filter { it.free > 0 }
    }
    
    // For unavailable facility time, we treat them as books with capacity 1
    private fun splitBlock(start: LocalDateTime, end: LocalDateTime, unavailable: List<Book>, capacity: Int): List<Block> {
        return splitBlock(start, end, unavailable, 1) 
    }

    private fun isOverlapping(startA: LocalDateTime, endA: LocalDateTime, startBStr: String, endBStr: String): Boolean {
        val startB = LocalDateTime.parse(startBStr.substring(0, 19), isoFormatter)
        val endB = LocalDateTime.parse(endBStr.substring(0, 19), isoFormatter)
        return !(startA.isAfter(endB) || endA.isBefore(startB))
    }

    private fun makeUnavailableTime(needFacility: String, reception: Reception): List<Book> {
        val facilityBooks = reception.books.filter { book ->
            val menu = reception.menus.find { it.menuID == book.menuID }
            menu?.needFacility == needFacility
        }.sortedBy { it.bookStart }

        val merged = mutableListOf<Book>()
        var currentStart: LocalDateTime? = null
        var currentEnd: LocalDateTime? = null

        facilityBooks.forEach { book ->
            val start = LocalDateTime.parse(book.bookStart.substring(0, 19), isoFormatter)
            val end = LocalDateTime.parse(book.bookEnd.substring(0, 19), isoFormatter)

            if (currentStart == null || start.isAfter(currentEnd)) {
                if (currentStart != null) {
                    merged.add(Book(bookStart = currentStart!!.format(isoFormatter), bookEnd = currentEnd!!.format(isoFormatter)))
                }
                currentStart = start
                currentEnd = end
            } else {
                if (end.isAfter(currentEnd)) currentEnd = end
            }
        }
        if (currentStart != null) {
            merged.add(Book(bookStart = currentStart!!.format(isoFormatter), bookEnd = currentEnd!!.format(isoFormatter)))
        }
        return merged
    }

    private fun mergeSchedules(slots: List<ScheduleSlot>): List<ScheduleSlot> {
        if (slots.isEmpty()) return emptyList()
        
        val filtered = slots.filter { it.timeStart != it.timeEnd }
        val grouped = filtered.groupBy { "${it.facility ?: "no_facility"}-${it.index}" }
        
        val mergedSlots = mutableListOf<ScheduleSlot>()
        grouped.forEach { (_, group) ->
            val sorted = group.sortedBy { it.timeStart }
            val merged = mutableListOf<ScheduleSlot>()
            sorted.forEach { curr ->
                if (merged.isEmpty()) {
                    merged.add(curr)
                    return@forEach
                }
                val last = merged.last()
                val lastEnd = LocalDateTime.parse(last.timeEnd, isoFormatter)
                val currStart = LocalDateTime.parse(curr.timeStart, isoFormatter)
                if (lastEnd >= currStart) {
                    val currEnd = LocalDateTime.parse(curr.timeEnd, isoFormatter)
                    val newEnd = if (lastEnd.isAfter(currEnd)) lastEnd else currEnd
                    merged[merged.size - 1] = last.copy(timeEnd = newEnd.format(isoFormatter))
                } else {
                    merged.add(curr)
                }
            }
            mergedSlots.addAll(merged)
        }
        return mergedSlots
    }

    fun submitBooking(start: String, end: String, answers: List<String>) {
        val reception = _state.value.reception ?: return
        val menuID = _state.value.selectedMenuID ?: return
        
        viewModelScope.launch {
            _state.value = _state.value.copy(loading = true, error = null)
            val result = repository.bookReception(
                receptionID = reception.receptionID,
                aliasName = repository.getAliasName(),
                bookStart = start,
                bookEnd = end,
                menuID = menuID,
                answers = answers
            )
            result.onSuccess {
                _state.value = _state.value.copy(loading = false, bookingSuccess = true)
                loadReception(reception.receptionID, menuID)
            }.onFailure { error ->
                _state.value = _state.value.copy(loading = false, error = error.message)
            }
        }
    }
}

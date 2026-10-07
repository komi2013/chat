package com.chat.android.feature.calendar

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import com.chat.android.core.push.PushReceiveDispatcher
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.collect
import kotlinx.coroutines.launch
import org.json.JSONArray
import org.json.JSONException
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter
import javax.inject.Inject

data class CalendarUiState(
    val loading: Boolean = true,
    val channelID: String = "",
    val aliasName: String = "",
    val events: List<CalendarItem> = emptyList(),
    val aliases: List<com.chat.android.feature.channel.DbAlias> = emptyList(),
    val groups: List<com.chat.android.feature.channel.DbGroup> = emptyList(),
    val error: String? = null,
    val notice: String? = null
)

@HiltViewModel
class CalendarViewModel @Inject constructor(
    private val repository: CalendarRepository,
    private val pushDispatcher: PushReceiveDispatcher
) : ViewModel() {
    private val mutableState = MutableStateFlow(CalendarUiState())
    val state = mutableState.asStateFlow()
    private var initializedChannel = ""

    fun load() {
        viewModelScope.launch {
            pushDispatcher.onUpdated.collect { channelID ->
                if (channelID == initializedChannel) refresh()
            }
        }
        viewModelScope.launch {
            val id = repository.currentChannelId().orEmpty()
            if (id.isBlank()) { mutableState.value = CalendarUiState(loading = false, error = "チャネルを選択してください"); return@launch }
            initializedChannel = id
            mutableState.value = CalendarUiState(loading = true, channelID = id)
            runCatching {
                CalendarUiState(false, id, repository.currentAlias(id), repository.events(id), repository.aliases(id), repository.groups(id))
            }.onSuccess { mutableState.value = it }.onFailure { mutableState.value = CalendarUiState(false, id, error = it.message ?: "カレンダーを読み込めませんでした") }
        }
    }

    fun refresh() {
        val id = initializedChannel
        if (id.isBlank()) return
        viewModelScope.launch { mutableState.value = mutableState.value.copy(events = repository.events(id)) }
    }

    suspend fun getEvent(id: String) = repository.event(id)

    suspend fun save(item: CalendarItem, attendees: List<String>, repeat: String): String? {
        val occurrences = repeatEvents(item, repeat)
        val error = repository.send(item, attendees, occurrences)
        if (error == null) {
            refresh()
            mutableState.value = mutableState.value.copy(notice = "カレンダーを保存しました")
        }
        return error
    }

    suspend fun delete(item: CalendarItem, attendees: List<String>): String? {
        val error = repository.send(item, attendees, listOf(item), delete = true)
        if (error == null) {
            refresh()
            mutableState.value = mutableState.value.copy(notice = "カレンダーを削除しました")
        }
        return error
    }

    fun clearNotice() { mutableState.value = mutableState.value.copy(notice = null) }

    fun searchCalendar(aliasName: String, date: String) {
        val id = initializedChannel
        if (id.isBlank()) return
        viewModelScope.launch {
            val error = repository.requestUserCalendar(id, aliasName, date)
            mutableState.value = mutableState.value.copy(
                error = error,
                notice = if (error == null) "検索ユーザーのカレンダーを取得しました" else error
            )
        }
    }

    fun removeSearchedCalendar(aliasName: String) {
        val id = initializedChannel
        if (id.isBlank()) return
        viewModelScope.launch {
            repository.removeUserCalendar(id, aliasName)
            refresh()
            mutableState.value = mutableState.value.copy(notice = "検索ユーザーが変更されました")
        }
    }

    private fun repeatEvents(item: CalendarItem, repeat: String): List<CalendarItem> {
        val start = runCatching { LocalDateTime.parse(item.timeStart) }.getOrNull() ?: return emptyList()
        val end = runCatching { LocalDateTime.parse(item.timeEnd) }.getOrNull() ?: return emptyList()
        val offsets = when (repeat) { "weekly" -> 0..25; "monthly" -> 1..6; else -> 0..0 }
        return offsets.map { offset ->
            val nextStart = when (repeat) { "weekly" -> start.plusWeeks(offset.toLong()); "monthly" -> start.plusMonths(offset.toLong()); else -> start }
            val nextEnd = when (repeat) { "weekly" -> end.plusWeeks(offset.toLong()); "monthly" -> end.plusMonths(offset.toLong()); else -> end }
            item.copy(
                calendarID = if (offset == 0 || repeat == "none") item.calendarID else "",
                timeStart = nextStart.format(FORMAT), timeEnd = nextEnd.format(FORMAT), repeatOption = repeat
            )
        }
    }
    companion object { val FORMAT: DateTimeFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm") }
}

fun parseCalendarNames(json: String): List<String> = try {
    val array = JSONArray(json)
    (0 until array.length()).map { array.optString(it) }.filter(String::isNotBlank)
} catch (_: JSONException) { emptyList() }

fun calendarDateTime(value: String): LocalDateTime? = runCatching {
    LocalDateTime.parse(value, CalendarViewModel.FORMAT)
}.getOrNull()

fun dayStart(date: LocalDate, hour: Int): String = date.atTime(hour, 0).format(CalendarViewModel.FORMAT)

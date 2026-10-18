package com.chat.android.feature.calendar

import android.net.Uri
import android.widget.Toast
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.itemsIndexed
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavController
import kotlinx.coroutines.launch
import org.json.JSONArray
import java.time.Instant
import java.time.LocalDateTime
import java.time.ZoneId
import java.time.ZoneOffset
import java.time.format.DateTimeFormatter

data class ImportSchedule(
    val title: String,
    val startEpochMillis: Long,
    val endEpochMillis: Long,
    val isSelected: Boolean = true
)

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ScheduleImportScreen(
    navController: NavController,
    uriString: String,
    viewModel: CalendarViewModel = hiltViewModel()
) {
    val context = LocalContext.current
    val state by viewModel.state.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()
    val schedules = remember(uriString) {
        mutableStateListOf<ImportSchedule>().apply {
            addAll(parseScheduleUri(uriString))
        }
    }

    val selectedCount = schedules.count { it.isSelected }
    var saving by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) { viewModel.load() }

    val save = {
        val targets = schedules.filter { it.isSelected }
        when {
            state.channelID.isBlank() -> Toast.makeText(context, "チャネルを選択してください", Toast.LENGTH_SHORT).show()
            state.aliasName.isBlank() -> Toast.makeText(context, "送信者名が設定されていません", Toast.LENGTH_SHORT).show()
            else -> {
                saving = true
                scope.launch {
                    var error: String? = null
                    // vue の CalendarEdit.vue と同じく、予定ごとに 1 回 ContentsPush する。
                    for (schedule in targets) {
                        val item = schedule.toCalendarItem(state.channelID, state.aliasName)
                        error = viewModel.save(item, listOf(state.aliasName), "none")
                        if (error != null) break
                    }
                    saving = false
                    if (error == null) {
                        Toast.makeText(context, "${targets.size}件の予定を保存しました", Toast.LENGTH_SHORT).show()
                        navController.popBackStack()
                    } else {
                        Toast.makeText(context, "保存に失敗しました: $error", Toast.LENGTH_LONG).show()
                    }
                }
            }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("予定をインポート") },
                navigationIcon = {
                    IconButton(onClick = { navController.popBackStack() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        },
        bottomBar = {
            Surface(tonalElevation = 8.dp) {
                Button(
                    onClick = { save() },
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(16.dp),
                    enabled = selectedCount > 0 && !saving
                ) {
                    Text(if (saving) "送信中…" else "保存 (${selectedCount}件)")
                }
            }
        }
    ) { padding ->
        if (schedules.isEmpty()) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                Text("インポート可能な予定が見つかりませんでした")
            }
        } else {
            LazyColumn(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(padding),
                contentPadding = PaddingValues(bottom = 80.dp)
            ) {
                itemsIndexed(schedules) { index, schedule ->
                    ScheduleImportRow(
                        schedule = schedule,
                        onCheckedChange = { checked ->
                            schedules[index] = schedule.copy(isSelected = checked)
                        }
                    )
                    Divider(modifier = Modifier.padding(horizontal = 16.dp))
                }
            }
        }
    }
}

@Composable
fun ScheduleImportRow(
    schedule: ImportSchedule,
    onCheckedChange: (Boolean) -> Unit
) {
    val startDateTime = remember(schedule.startEpochMillis) {
        LocalDateTime.ofInstant(Instant.ofEpochMilli(schedule.startEpochMillis), ZoneId.systemDefault())
    }
    val endDateTime = remember(schedule.endEpochMillis) {
        LocalDateTime.ofInstant(Instant.ofEpochMilli(schedule.endEpochMillis), ZoneId.systemDefault())
    }
    val dateFormatter = DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm")

    Row(
        modifier = Modifier
            .fillMaxWidth()
            .padding(16.dp),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Checkbox(
            checked = schedule.isSelected,
            onCheckedChange = onCheckedChange
        )
        Spacer(Modifier.width(8.dp))
        Column {
            Text(
                text = schedule.title,
                style = MaterialTheme.typography.titleMedium,
                fontWeight = FontWeight.Bold
            )
            Text(
                text = "${startDateTime.format(dateFormatter)} - ${endDateTime.format(dateFormatter)}",
                style = MaterialTheme.typography.bodyMedium,
                color = MaterialTheme.colorScheme.onSurfaceVariant
            )
        }
    }
}

private fun parseScheduleUri(uriString: String): List<ImportSchedule> {
    val uri = Uri.parse(uriString)
    val result = mutableListOf<ImportSchedule>()
    var i = 0
    val formatter = DateTimeFormatter.ofPattern("yyyyMMdd'T'HHmmss'Z'").withZone(ZoneOffset.UTC)

    while (true) {
        val text = uri.getQueryParameter("text[$i]") ?: uri.getQueryParameter("text")?.takeIf { i == 0 } ?: break
        val dates = uri.getQueryParameter("dates[$i]") ?: uri.getQueryParameter("dates")?.takeIf { i == 0 } ?: break
        
        val parts = dates.split("/")
        if (parts.size == 2) {
            try {
                val start = Instant.from(formatter.parse(parts[0])).toEpochMilli()
                val end = Instant.from(formatter.parse(parts[1])).toEpochMilli()
                result.add(ImportSchedule(text, start, end))
            } catch (e: Exception) {
                e.printStackTrace()
            }
        }
        i++
    }
    return result
}

/** 取り込んだ予定を、Vue のカレンダー保存と同じ形式の CalendarItem に変換する。 */
private fun ImportSchedule.toCalendarItem(channelID: String, aliasName: String): CalendarItem {
    val zone = ZoneId.systemDefault()
    val formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm")
    return CalendarItem(
        calendarID = "",
        channelID = channelID,
        aliasName = aliasName,
        timeStart = LocalDateTime.ofInstant(Instant.ofEpochMilli(startEpochMillis), zone).format(formatter),
        timeEnd = LocalDateTime.ofInstant(Instant.ofEpochMilli(endEpochMillis), zone).format(formatter),
        todo = title,
        aliasNamesJson = JSONArray().toString(),
        repeatOption = "none"
    )
}

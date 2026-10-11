package com.chat.android.feature.calendar

import androidx.compose.foundation.background
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ChevronLeft
import androidx.compose.material.icons.filled.ChevronRight
import androidx.compose.material.icons.filled.Close
import androidx.compose.material3.*
import androidx.compose.runtime.*
import kotlinx.coroutines.launch
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clipToBounds
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.IntOffset
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.lifecycle.compose.collectAsStateWithLifecycle
import androidx.navigation.NavController
import com.chat.android.navigation.CalendarEditRoute
import androidx.compose.foundation.ScrollState
import androidx.compose.ui.text.style.TextAlign
import java.time.LocalDate
import java.time.format.DateTimeFormatter
import java.util.Locale

@Composable
fun CalendarScreen(
    navController: NavController,
    date: String?,
    viewModel: CalendarViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    var firstDay by remember(date) { mutableStateOf(runCatching { LocalDate.parse(date) }.getOrDefault(LocalDate.now())) }
    val dayScroll = rememberScrollState()
    var searchedUsers by remember { mutableStateOf(emptyList<String>()) }
    val drawerState = rememberDrawerState(initialValue = DrawerValue.Closed)
    val scope = rememberCoroutineScope()

    LaunchedEffect(Unit) { viewModel.load() }
    LaunchedEffect(state.notice) {
        if (state.notice != null) { kotlinx.coroutines.delay(2400); viewModel.clearNotice() }
    }
    val days = remember(firstDay) { (0..29).map { firstDay.plusDays(it.toLong()) } }
    val eventsByDayHour = remember(state.events) {
        state.events.groupBy { event ->
            calendarDateTime(event.timeStart)?.let { it.toLocalDate() to it.hour }
        }
    }

    ModalNavigationDrawer(
        drawerState = drawerState,
        drawerContent = {
            CalendarDrawer(
                state = state,
                searchedUsers = searchedUsers,
                onSearch = { aliasName ->
                    searchedUsers = searchedUsers + aliasName
                    viewModel.searchCalendar(aliasName, firstDay.toString())
                },
                onRemove = { aliasName ->
                    searchedUsers = searchedUsers - aliasName
                    viewModel.removeSearchedCalendar(aliasName)
                },
                onClose = { scope.launch { drawerState.close() } }
            )
        },
        gesturesEnabled = drawerState.isOpen // Prevent drawer from being dragged open if it interferes with main navigation
    ) {
        Column(Modifier.fillMaxSize().padding(horizontal = 8.dp)) {
            if (state.channelID.isNotBlank()) {
                Box(Modifier.weight(1f)) {
                    LazyColumn(Modifier.fillMaxSize()) {
                        item {
                            // Calendar range controls (scrollable header)
                            Surface(
                                color = MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.3f),
                                modifier = Modifier.fillMaxWidth().padding(vertical = 4.dp),
                                shape = RoundedCornerShape(8.dp)
                            ) {
                                Column {
                                    Row(
                                        Modifier.fillMaxWidth(),
                                        verticalAlignment = Alignment.CenterVertically,
                                        horizontalArrangement = Arrangement.Center
                                    ) {
                                        IconButton(onClick = { firstDay = firstDay.minusDays(30) }) {
                                            Icon(Icons.Default.ChevronLeft, "前の30日")
                                        }
                                        Text(
                                            "< ${firstDay.format(DateTimeFormatter.ofPattern("yyyy/MM/dd"))}から30日 >",
                                            style = MaterialTheme.typography.titleMedium,
                                            fontWeight = FontWeight.Bold
                                        )
                                        IconButton(onClick = { firstDay = firstDay.plusDays(30) }) {
                                            Icon(Icons.Default.ChevronRight, "次の30日")
                                        }
                                    }
                                    if (state.loading) LinearProgressIndicator(Modifier.fillMaxWidth())
                                }
                            }
                            state.error?.let { Text(it, color = MaterialTheme.colorScheme.error, modifier = Modifier.padding(8.dp)) }
                            state.notice?.let { Text(it, Modifier.fillMaxWidth().padding(8.dp), color = MaterialTheme.colorScheme.primary) }
                        }
                        item {
                            CalendarDateHeader(days = days, dayScroll = dayScroll, isBottom = false, onMenuClick = {
                                scope.launch { drawerState.open() }
                            })
                        }
                        items((0..23).toList()) { hour ->
                            Row(Modifier.height(62.dp), verticalAlignment = Alignment.Top) {
                                Text("%02d:00".format(hour), modifier = Modifier.width(52.dp).padding(top = 4.dp), textAlign = androidx.compose.ui.text.style.TextAlign.Center, style = MaterialTheme.typography.labelSmall)
                                Row(Modifier.horizontalScroll(dayScroll)) {
                                    days.forEach { day ->
                                        val events = eventsByDayHour[day to hour].orEmpty()
                                        Box(Modifier.width(112.dp).fillMaxHeight().padding(1.dp).background(
                                            if (day.dayOfWeek.value == 7) MaterialTheme.colorScheme.error.copy(alpha = .06f) else MaterialTheme.colorScheme.surfaceVariant.copy(alpha = .28f)
                                        ).clickable { navController.navigate(CalendarEditRoute(dateTime = dayStart(day, hour))) }) {
                                            Column(Modifier.fillMaxSize()) {
                                                events.take(2).forEach { event ->
                                                    Text(event.title, Modifier.fillMaxWidth().padding(2.dp).background(MaterialTheme.colorScheme.primaryContainer, RoundedCornerShape(4.dp)).clickable {
                                                        navController.navigate(CalendarEditRoute(id = event.calendarID))
                                                    }.padding(horizontal = 4.dp, vertical = 3.dp), maxLines = 2, style = MaterialTheme.typography.labelSmall)
                                                }
                                                if (events.size > 2) Text("+${events.size - 2}", style = MaterialTheme.typography.labelSmall)
                                            }
                                        }
                                    }
                                }
                            }
                        }
                        item {
                            CalendarDateHeader(days = days, dayScroll = dayScroll, isBottom = true, onMenuClick = {
                                scope.launch { drawerState.open() }
                            })
                        }
                    }

                    FloatingActionButton(
                        onClick = { navController.navigate(CalendarEditRoute(dateTime = dayStart(LocalDate.now(), java.time.LocalTime.now().hour))) },
                        modifier = Modifier.align(Alignment.BottomEnd).padding(16.dp)
                    ) {
                        Icon(Icons.Default.Add, "予定を追加")
                    }
                }
            }
        }
    }
}

@Composable
private fun CalendarDateHeader(
    days: List<LocalDate>,
    dayScroll: ScrollState,
    isBottom: Boolean = false,
    onMenuClick: () -> Unit = {}
) {
    Row(
        Modifier
            .fillMaxWidth()
            .background(MaterialTheme.colorScheme.surfaceVariant.copy(alpha = 0.5f)),
        verticalAlignment = Alignment.CenterVertically
    ) {
        Box(
            modifier = Modifier.width(52.dp).clickable { onMenuClick() },
            contentAlignment = Alignment.Center
        ) {
            Text("≡", style = MaterialTheme.typography.titleMedium)
        }
        Row(Modifier.horizontalScroll(dayScroll)) {
            days.forEach { day ->
                Column(
                    Modifier.width(112.dp).padding(vertical = 4.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    if (!isBottom) {
                        Text(
                            day.dayOfWeek.getDisplayName(java.time.format.TextStyle.SHORT, Locale.JAPAN),
                            style = MaterialTheme.typography.labelSmall
                        )
                        Text(
                            day.format(DateTimeFormatter.ofPattern("MM/dd")),
                            fontWeight = FontWeight.Bold,
                            style = MaterialTheme.typography.labelMedium
                        )
                    } else {
                        Text(
                            day.format(DateTimeFormatter.ofPattern("MM/dd")),
                            fontWeight = FontWeight.Bold,
                            style = MaterialTheme.typography.labelMedium
                        )
                        Text(
                            day.dayOfWeek.getDisplayName(java.time.format.TextStyle.SHORT, Locale.JAPAN),
                            style = MaterialTheme.typography.labelSmall
                        )
                    }
                }
            }
        }
    }
}

@Composable
@OptIn(ExperimentalMaterial3Api::class)
fun CalendarEditScreen(
    navController: NavController,
    id: String?,
    dateTime: String?,
    viewModel: CalendarViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsStateWithLifecycle()
    val scope = rememberCoroutineScope()
    var start by remember { mutableStateOf(dateTime ?: LocalDate.now().atTime(java.time.LocalTime.now().hour, 0).format(CalendarViewModel.FORMAT)) }
    var end by remember { mutableStateOf(calendarDateTime(start)?.plusMinutes(30)?.format(CalendarViewModel.FORMAT).orEmpty()) }
    var todo by remember { mutableStateOf("") }
    var attendees by remember { mutableStateOf(emptyList<String>()) }
    var repeat by remember { mutableStateOf("none") }
    var search by remember { mutableStateOf("") }
    var loading by remember { mutableStateOf(true) }
    var saving by remember { mutableStateOf(false) }
    var error by remember { mutableStateOf<String?>(null) }
    var confirmAction by remember { mutableStateOf<String?>(null) }

    LaunchedEffect(Unit) { viewModel.load() }
    LaunchedEffect(state.aliases, id) {
        if (state.channelID.isNotBlank()) {
            if (id != null) {
                val saved = viewModel.getEvent(id)
                if (saved != null) {
                    start = saved.timeStart; end = saved.timeEnd; todo = saved.todo
                    attendees = parseCalendarNames(saved.aliasNamesJson); repeat = saved.repeatOption
                }
            } else if (attendees.isEmpty() && state.aliasName.isNotBlank()) attendees = listOf(state.aliasName)
            loading = false
        }
    }

    val candidates = remember(search, state.aliases, state.groups, attendees) {
        if (search.isBlank()) emptyList() else {
            val aliases = state.aliases.filter { it.aliasName.contains(search, true) }.map { it.aliasName }
            val groups = state.groups.filter { it.groupName.contains(search, true) }.flatMap { parseCalendarNames(it.aliasNamesJson) }
            (aliases + groups).distinct().filterNot { it in attendees }
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (id == null) "予定を作成" else "予定を編集") },
                navigationIcon = {
                    IconButton(onClick = { navController.popBackStack() }) {
                        Icon(Icons.Default.Close, "閉じる")
                    }
                }
            )
        }
    ) { padding ->
        Column(Modifier.fillMaxSize().verticalScroll(rememberScrollState()).padding(padding).padding(16.dp), verticalArrangement = Arrangement.spacedBy(10.dp)) {
            if (loading) LinearProgressIndicator(Modifier.fillMaxWidth()) else {
                OutlinedTextField(start, { start = it }, label = { Text("開始 (YYYY-MM-DDTHH:mm)") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
                OutlinedTextField(end, { end = it }, label = { Text("終了 (YYYY-MM-DDTHH:mm)") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
                OutlinedTextField(todo, { todo = it }, label = { Text("Todo") }, modifier = Modifier.fillMaxWidth(), minLines = 2)
                Text("参加ユーザー", style = MaterialTheme.typography.titleSmall)
                OutlinedTextField(search, { search = it }, label = { Text("ユーザーまたはグループを検索") }, modifier = Modifier.fillMaxWidth(), singleLine = true)
                candidates.take(8).forEach { name ->
                    ListItem(
                        headlineContent = { Text(name) },
                        modifier = Modifier.clickable { attendees = attendees + name; search = "" },
                        leadingContent = { Text("＋", color = MaterialTheme.colorScheme.primary) }
                    )
                }
                Row(Modifier.fillMaxWidth().horizontalScroll(rememberScrollState())) {
                    attendees.forEach { name -> AssistChip(onClick = { attendees = attendees - name }, label = { Text("$name ×") }, modifier = Modifier.padding(end = 4.dp)) }
                }
                var expanded by remember { mutableStateOf(false) }
                ExposedDropdownMenuBox(expanded = expanded, onExpandedChange = { expanded = it }) {
                    OutlinedTextField(
                        value = when (repeat) { "weekly" -> "毎週同じ曜日"; "monthly" -> "毎月同じ日"; else -> "繰り返しなし" },
                        onValueChange = {}, readOnly = true, label = { Text("繰り返し") },
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded) }, modifier = Modifier.menuAnchor().fillMaxWidth()
                    )
                    ExposedDropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                        listOf("none" to "繰り返しなし", "weekly" to "毎週同じ曜日", "monthly" to "毎月同じ日").forEach { (value, label) ->
                            DropdownMenuItem(text = { Text(label) }, onClick = { repeat = value; expanded = false })
                        }
                    }
                }
                error?.let { Text(it, color = MaterialTheme.colorScheme.error) }
                Row(Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                    Button(onClick = { confirmAction = "save" }, enabled = todo.isNotBlank() && !saving, modifier = Modifier.weight(1f)) { Text(if (saving) "送信中…" else "保存") }
                    if (id != null) OutlinedButton(onClick = { confirmAction = "delete" }, enabled = !saving) { Text("削除") }
                }
            }
        }
    }
    confirmAction?.let { action ->
        AlertDialog(
            onDismissRequest = { confirmAction = null },
            title = { Text(if (action == "delete") "予定を削除しますか？" else "予定を保存しますか？") },
            confirmButton = { TextButton(onClick = {
                confirmAction = null
                val parsedStart = calendarDateTime(start); val parsedEnd = calendarDateTime(end)
                if (parsedStart == null || parsedEnd == null || !parsedEnd.isAfter(parsedStart)) { error = "開始・終了日時を YYYY-MM-DDTHH:mm 形式で入力してください"; return@TextButton }
                val item = CalendarItem(id.orEmpty(), state.channelID, state.aliasName, start, end, todo, org.json.JSONArray(attendees).toString(), repeat)
                saving = true
                scope.launch {
                    error = if (action == "delete") viewModel.delete(item, attendees) else viewModel.save(item, attendees, repeat)
                    saving = false
                    if (error == null) navController.popBackStack()
                }
            }) { Text("実行") } },
            dismissButton = { TextButton(onClick = { confirmAction = null }) { Text("キャンセル") } }
        )
    }
}

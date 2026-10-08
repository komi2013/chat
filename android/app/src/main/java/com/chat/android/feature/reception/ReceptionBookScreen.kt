package com.chat.android.feature.reception

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.ArrowForward
import androidx.compose.material.icons.filled.Menu
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import java.time.LocalDate
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionBookScreen(
    receptionID: String,
    initialDate: String? = null,
    initialMenuID: Int? = null,
    onBack: () -> Unit,
    viewModel: ReceptionBookViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()
    val drawerState = rememberDrawerState(initialValue = DrawerValue.Closed)

    LaunchedEffect(receptionID) {
        viewModel.loadReception(receptionID, initialMenuID)
    }

    var showModal by remember { mutableStateOf(false) }
    var selectedSlot by remember { mutableStateOf<ScheduleSlot?>(null) }

    val isoFormatter = DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm:ss")
    val dayOfWeekFormatter = DateTimeFormatter.ofPattern("E")

    val today = LocalDate.now()
    val monthDates = (0 until 30).map { today.plusDays(it.toLong()) }
    val hours = (0 until 24).toList()

    ModalNavigationDrawer(
        drawerState = drawerState,
        drawerContent = {
            ModalDrawerSheet {
                Spacer(Modifier.height(12.dp))
                Text("メニュー", modifier = Modifier.padding(16.dp), fontWeight = FontWeight.Bold)
                state.menus.filter { it.forBookType.isNotEmpty() }.forEach { menu ->
                    NavigationDrawerItem(
                        label = { Text("${menu.menuName} - ${menu.price}円") },
                        selected = state.selectedMenuID == menu.menuID,
                        onClick = {
                            viewModel.selectMenu(menu.menuID)
                        },
                        modifier = Modifier.padding(NavigationDrawerItemDefaults.ItemPadding)
                    )
                }
            }
        }
    ) {
        Scaffold(
            topBar = {
                TopAppBar(
                    title = { Text(state.reception?.receptionTitle ?: "受付予約") },
                    navigationIcon = {
                        IconButton(onClick = onBack) {
                            Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                        }
                    },
                    actions = {
                        IconButton(onClick = { /* TODO: Open drawer logic if needed outside scaffold */ }) {
                            Icon(Icons.Default.Menu, contentDescription = "メニュー")
                        }
                    }
                )
            }
        ) { padding ->
            Box(modifier = Modifier.padding(padding).fillMaxSize()) {
                if (state.loading) {
                    CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                }

                state.error?.let {
                    Text(it, color = Color.Red, modifier = Modifier.align(Alignment.Center))
                }

                val horizontalScrollState = rememberScrollState()
                val verticalScrollState = rememberScrollState()

                Column(modifier = Modifier.fillMaxSize()) {
                    // Calendar Header
                    Row(modifier = Modifier.horizontalScroll(horizontalScrollState)) {
                        Box(modifier = Modifier.width(40.dp).height(50.dp).background(Color(0xFFF0F0F0)).border(0.5.dp, Color.LightGray), contentAlignment = Alignment.Center) {
                            Text("≡", fontSize = 20.sp)
                        }
                        monthDates.forEach { date ->
                            Column(
                                modifier = Modifier.width(120.dp).height(50.dp).background(Color(0xFFF0F0F0)).border(0.5.dp, Color.LightGray),
                                horizontalAlignment = Alignment.CenterHorizontally,
                                verticalArrangement = Arrangement.Center
                            ) {
                                Text(date.format(dayOfWeekFormatter), fontSize = 12.sp)
                                Text(date.format(DateTimeFormatter.ofPattern("MM/dd")), fontSize = 14.sp, fontWeight = FontWeight.Bold)
                            }
                        }
                    }

                    // Calendar Grid
                    Box(modifier = Modifier.weight(1f).verticalScroll(verticalScrollState).horizontalScroll(horizontalScrollState)) {
                        Column {
                            hours.forEach { hour ->
                                Row {
                                    // Hour label / Back column
                                    Box(
                                        modifier = Modifier.width(40.dp).height(60.dp).background(Color(0xFFFAFAFA)).border(0.5.dp, Color.LightGray).clickable { /* jump(-30) */ },
                                        contentAlignment = Alignment.Center
                                    ) {
                                        if (hour % 6 == 5) Icon(Icons.Default.ArrowBack, "", modifier = Modifier.size(16.dp))
                                    }

                                    monthDates.forEach { date ->
                                        val isSunday = date.dayOfWeek.value == 7
                                        Box(
                                            modifier = Modifier
                                                .width(120.dp)
                                                .height(60.dp)
                                                .background(if (isSunday) Color(0xFFFFEBEE) else Color.Gray)
                                                .border(0.5.dp, Color.LightGray)
                                                .clickable {
                                                    val start = date.atTime(hour, 0).format(isoFormatter)
                                                    val end = date.atTime(hour, 0).plusHours(1).format(isoFormatter)
                                                    selectedSlot = ScheduleSlot(start, end)
                                                    showModal = true
                                                }
                                        ) {
                                            Text(
                                                text = hour.toString(),
                                                fontSize = 10.sp,
                                                modifier = Modifier.padding(2.dp).align(Alignment.TopStart)
                                            )

                                            // Events (Available slots)
                                            state.slots.filter { slot ->
                                                val start = LocalDateTime.parse(slot.timeStart.substring(0, 19), isoFormatter)
                                                val end = LocalDateTime.parse(slot.timeEnd.substring(0, 19), isoFormatter)
                                                start.toLocalDate() == date && start.hour <= hour && (end.hour > hour || (end.hour == hour && end.minute > 0))
                                            }.forEach { slot ->
                                                val start = LocalDateTime.parse(slot.timeStart.substring(0, 19), isoFormatter)
                                                val end = LocalDateTime.parse(slot.timeEnd.substring(0, 19), isoFormatter)
                                                
                                                val top = if (start.hour < hour) 0 else start.minute
                                                val bottom = if (end.hour > hour) 60 else end.minute
                                                val height = bottom - top
                                                
                                                if (height > 0) {
                                                    Box(
                                                        modifier = Modifier
                                                            .padding(start = (slot.index * 60).dp) // Adjust width to fit multiple
                                                            .offset(y = top.dp)
                                                            .width(60.dp) // narrower to see overlaps
                                                            .height(height.dp)
                                                            .background(Color.White)
                                                            .border(1.dp, Color.Blue.copy(alpha = 0.3f))
                                                            .clickable {
                                                                selectedSlot = slot
                                                                showModal = true
                                                            }
                                                    )
                                                }
                                            }
                                        }
                                    }

                                    // Forward column
                                    Box(
                                        modifier = Modifier.width(40.dp).height(60.dp).background(Color(0xFFFAFAFA)).border(0.5.dp, Color.LightGray).clickable { /* jump(30) */ },
                                        contentAlignment = Alignment.Center
                                    ) {
                                        if (hour % 6 == 5) Icon(Icons.Default.ArrowForward, "", modifier = Modifier.size(16.dp))
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    if (showModal && selectedSlot != null && state.reception != null) {
        BookModal(
            timeStart = selectedSlot!!.timeStart,
            timeEnd = selectedSlot!!.timeEnd,
            reception = state.reception!!,
            menus = state.menus,
            initialMenuID = state.selectedMenuID,
            onDismiss = { showModal = false },
            onSubmit = { start, end, menuID, answers ->
                viewModel.submitBooking(start, end, answers)
                showModal = false
            }
        )
    }

    if (state.bookingSuccess) {
        AlertDialog(
            onDismissRequest = { /* Reset success state if needed */ },
            confirmButton = { TextButton(onClick = { /* Reset */ }) { Text("OK") } },
            title = { Text("予約完了") },
            text = { Text("予約が完了しました。") }
        )
    }
}

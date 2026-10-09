package com.chat.android.feature.reception

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material.Button
import androidx.compose.material.ButtonDefaults
import androidx.compose.material.Card
import androidx.compose.material.CircularProgressIndicator
import androidx.compose.material.Icon
import androidx.compose.material.IconButton
import androidx.compose.material.Scaffold
import androidx.compose.material.Text
import androidx.compose.material.TopAppBar
import androidx.compose.material.rememberScaffoldState
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import java.text.SimpleDateFormat
import java.util.*

@Composable
fun ReceptionShiftScreen(
    receptionID: String,
    onBack: () -> Unit,
    viewModel: ReceptionShiftViewModel = hiltViewModel()
) {
    val uiState by viewModel.uiState.collectAsState()
    val scaffoldState = rememberScaffoldState()

    Scaffold(
        scaffoldState = scaffoldState,
        topBar = {
            TopAppBar(
                title = { Text("シフト設定") },
                navigationIcon = {
                    IconButton(onClick = onBack) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    Button(onClick = { viewModel.submitShift() }) {
                        Text("提出", color = Color.White)
                    }
                }
            )
        }
    ) { padding ->
        Box(modifier = Modifier.padding(padding).fillMaxSize()) {
            when (uiState) {
                is ReceptionShiftUiState.Loading -> CircularProgressIndicator(modifier = Modifier.align(Alignment.Center))
                is ReceptionShiftUiState.Error -> Text((uiState as ReceptionShiftUiState.Error).message, color = Color.Red, modifier = Modifier.align(Alignment.Center))
                is ReceptionShiftUiState.Success -> {
                    ShiftCalendar(viewModel)
                }
            }
            
            viewModel.errorMessage.value?.let {
                Text(it, color = Color.Red, modifier = Modifier.align(Alignment.BottomCenter).padding(16.dp))
            }
        }
    }
}

@Composable
fun ShiftCalendar(viewModel: ReceptionShiftViewModel) {
    val monthDates = viewModel.monthDates
    val hours = viewModel.hours
    val horizontalScrollState = rememberScrollState()
    val verticalScrollState = rememberScrollState()
    
    val dayWidth = 120.dp
    val hourHeight = 60.dp

    Column(modifier = Modifier.fillMaxSize()) {
        // Header
        Row(modifier = Modifier.horizontalScroll(horizontalScrollState)) {
            Box(modifier = Modifier.width(40.dp).height(40.dp).background(Color(0xFFF0F0F0)), contentAlignment = Alignment.Center) {
                Text("≡", fontSize = 20.sp)
            }
            monthDates.forEach { day ->
                Column(
                    modifier = Modifier.width(dayWidth).height(40.dp).background(Color(0xFFF0F0F0)).border(0.5.dp, Color.LightGray),
                    horizontalAlignment = Alignment.CenterHorizontally,
                    verticalArrangement = Arrangement.Center
                ) {
                    Text(SimpleDateFormat("EEE", Locale.getDefault()).format(day), fontSize = 10.sp)
                    Text(SimpleDateFormat("MM/dd", Locale.getDefault()).format(day), fontSize = 12.sp, fontWeight = FontWeight.Bold)
                }
            }
        }

        // Body
        Box(modifier = Modifier.weight(1f)) {
            Column(modifier = Modifier.verticalScroll(verticalScrollState)) {
                hours.forEach { hour ->
                    Row(modifier = Modifier.horizontalScroll(horizontalScrollState)) {
                        // Time column
                        Box(
                            modifier = Modifier.width(40.dp).height(hourHeight).background(Color(0xFFFAFAFA)).border(0.5.dp, Color.LightGray),
                            contentAlignment = Alignment.TopCenter
                        ) {
                            Text(hour.toString(), fontSize = 8.sp, modifier = Modifier.padding(top = 2.dp))
                        }

                        // Day slots
                        monthDates.forEach { day ->
                            val calendar = Calendar.getInstance().apply { time = day }
                            val isSunday = calendar.get(Calendar.DAY_OF_WEEK) == Calendar.SUNDAY
                            
                            Box(
                                modifier = Modifier
                                    .width(dayWidth)
                                    .height(hourHeight)
                                    .background(if (isSunday) Color(0xFFFF0000).copy(alpha = 0.1f) else Color.Transparent)
                                    .border(0.5.dp, Color.LightGray)
                            ) {
                                val events = viewModel.getEventsForDayAndHour(day, hour)
                                events.forEachIndexed { index, event ->
                                    EventItem(event, index, events.size, viewModel)
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

@Composable
fun EventItem(event: ScheduleItem, index: Int, total: Int, viewModel: ReceptionShiftViewModel) {
    val dateTimeSdf = SimpleDateFormat("yyyy-MM-dd HH:mm:ss", Locale.getDefault())
    val start = try { dateTimeSdf.parse(event.timeStart) } catch(e: Exception) { null } ?: Date()
    
    val startCal = Calendar.getInstance().apply { time = start }
    
    val durationMinutes = 60
    val topOffset = startCal.get(Calendar.MINUTE).dp
    val height = durationMinutes.dp
    
    val width = 118.dp / total
    val leftOffset = (index * width.value).dp
    
    val colors = listOf(
        Color(255, 0, 0, 64),
        Color(255, 255, 0, 64),
        Color(128, 0, 128, 64),
        Color(0, 0, 255, 64)
    )
    val myName = viewModel.repository.getAliasName()
    val nameCount = event.aliasNames.size
    val backgroundColor = if (nameCount > 0) colors.getOrElse(nameCount - 1) { colors.last() } else Color.LightGray.copy(alpha = 0.3f)

    Card(
        modifier = Modifier
            .padding(1.dp)
            .offset(x = leftOffset, y = topOffset)
            .width(width - 2.dp)
            .height(height - 4.dp),
        backgroundColor = backgroundColor,
        elevation = 2.dp
    ) {
        Column(modifier = Modifier.padding(2.dp).verticalScroll(rememberScrollState())) {
            Text("${event.skill}:", fontSize = 8.sp, fontWeight = FontWeight.Bold)
            event.aliasNames.forEachIndexed { openIndex, aliasName ->
                val isMe = aliasName == myName
                val isWithinOpen = openIndex < event.open
                Button(
                    onClick = { viewModel.moveStaffToTop(event.key, aliasName) },
                    colors = ButtonDefaults.buttonColors(
                        backgroundColor = if (isWithinOpen) Color.Blue else Color.LightGray,
                        contentColor = if (isMe) Color.Black else Color.White
                    ),
                    contentPadding = PaddingValues(2.dp),
                    modifier = Modifier.height(20.dp).fillMaxWidth().padding(vertical = 1.dp)
                ) {
                    Text(aliasName, fontSize = 7.sp)
                }
            }
        }
    }
}

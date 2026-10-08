package com.chat.android.feature.reception

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import java.text.SimpleDateFormat
import java.util.*

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionEnterScreen(
    id: String,
    passkey: String,
    onBack: () -> Unit,
    viewModel: ReceptionEnterViewModel = hiltViewModel()
) {
    val reception = viewModel.reception
    val waitMinutes = viewModel.waitMinutes
    val assignedFacility = viewModel.assignedFacility
    val errorMessage = viewModel.errorMessage
    val isLoading = viewModel.isLoading

    LaunchedEffect(id, passkey) {
        viewModel.loadReception(id, passkey)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("受付待ち時間") },
                navigationIcon = {
                    TextButton(onClick = onBack) {
                        Text("戻る")
                    }
                }
            )
        }
    ) { padding ->
        Column(
            modifier = Modifier
                .padding(padding)
                .padding(16.dp)
                .fillMaxSize()
        ) {
            if (errorMessage != null) {
                Text(
                    text = errorMessage,
                    color = MaterialTheme.colorScheme.error,
                    modifier = Modifier.padding(bottom = 8.dp)
                )
            }

            if (isLoading) {
                CircularProgressIndicator(modifier = Modifier.align(Alignment.CenterHorizontally))
            } else if (reception != null) {
                Column(verticalArrangement = Arrangement.spacedBy(16.dp)) {
                    Row(
                        verticalAlignment = Alignment.CenterVertically,
                        horizontalArrangement = Arrangement.spacedBy(8.dp)
                    ) {
                        Text("何名さまですか？")
                        OutlinedTextField(
                            value = viewModel.inputGuests.toString(),
                            onValueChange = {
                                val count = it.toIntOrNull() ?: 0
                                viewModel.onGuestCountChange(count)
                            },
                            keyboardOptions = KeyboardOptions(keyboardType = KeyboardType.Number),
                            modifier = Modifier.width(80.dp)
                        )
                        Button(onClick = { viewModel.calculateWaitMinutes() }) {
                            Text("再計算")
                        }
                    }

                    if (assignedFacility != null) {
                        Surface(
                            color = Color(0xFFE8F5E9),
                            shape = MaterialTheme.shapes.medium,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Column(Modifier.padding(16.dp)) {
                                Text(
                                    "✅ ご案内可能: ${assignedFacility.facilityName}",
                                    fontWeight = FontWeight.Bold,
                                    color = Color(0xFF2E7D32)
                                )
                                Text("(定員 ${assignedFacility.capacity} 名)")
                            }
                        }
                    } else if (waitMinutes > 0) {
                        Text(
                            text = "お待ち時間は 約 $waitMinutes 分 です",
                            fontSize = 18.sp,
                            fontWeight = FontWeight.Bold
                        )
                    }

                    if (assignedFacility == null || reception.adminNames.isNotEmpty()) {
                        Button(
                            onClick = {
                                viewModel.queueUp(id, passkey) {
                                    // Refresh or Navigate back? Vue does location.href = '' which reloads.
                                    viewModel.loadReception(id, passkey)
                                }
                            },
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text("並ぶ")
                        }
                    }

                    if (reception.adminNames.isNotEmpty()) {
                        Divider(modifier = Modifier.padding(vertical = 8.dp))
                        Text("現在の待機列", style = MaterialTheme.typography.titleMedium)
                        
                        LazyColumn(verticalArrangement = Arrangement.spacedBy(8.dp)) {
                            items(reception.queues) { q ->
                                Card(modifier = Modifier.fillMaxWidth()) {
                                    Row(
                                        modifier = Modifier.padding(8.dp),
                                        verticalAlignment = Alignment.CenterVertically
                                    ) {
                                        IconButton(onClick = {
                                            viewModel.queueRemove(id, q.queueName) {
                                                viewModel.loadReception(id, passkey)
                                            }
                                        }) {
                                            Icon(Icons.Default.Delete, contentDescription = "削除")
                                        }
                                        Column {
                                            Text("${q.queueName} - ${q.waitingGuest} 名")
                                            Text(
                                                "受付: ${formatTime(q.queuedAt)}",
                                                style = MaterialTheme.typography.bodySmall
                                            )
                                        }
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}

private fun formatTime(isoString: String): String {
    return try {
        val df = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss", Locale.getDefault())
        val date = df.parse(isoString)
        val outDf = SimpleDateFormat("HH:mm:ss", Locale.getDefault())
        outDf.format(date ?: Date())
    } catch (e: Exception) {
        isoString
    }
}

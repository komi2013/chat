package com.chat.android.feature.reception

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import java.time.LocalDateTime
import java.time.format.DateTimeFormatter

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun BookModal(
    timeStart: String,
    timeEnd: String,
    reception: Reception,
    menus: List<Menu>,
    initialMenuID: Int?,
    onDismiss: () -> Unit,
    onSubmit: (String, String, Int, List<String>) -> Unit
) {
    var startTime by remember { mutableStateOf(timeStart) }
    var endTime by remember { mutableStateOf(timeEnd) }
    var selectedMenuID by remember { mutableStateOf(initialMenuID ?: menus.firstOrNull { it.forBookType.isNotEmpty() }?.menuID ?: 0) }
    
    val totalAsks = reception.asks.size + reception.askChoices.size + reception.askMultiChoices.size
    val answers = remember { mutableStateListOf(*Array(totalAsks) { "" }) }

    AlertDialog(
        onDismissRequest = onDismiss,
        title = { Text("予約を追加") },
        text = {
            Column(
                modifier = Modifier
                    .fillMaxWidth()
                    .verticalScroll(rememberScrollState()),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                Text("開始時間")
                TextField(
                    value = startTime,
                    onValueChange = { startTime = it },
                    modifier = Modifier.fillMaxWidth()
                )

                // Questions
                reception.asks.forEachIndexed { index, ask ->
                    Text(ask.question)
                    TextField(
                        value = answers[index],
                        onValueChange = { answers[index] = it },
                        modifier = Modifier.fillMaxWidth()
                    )
                }

                var offset = reception.asks.size
                reception.askChoices.forEachIndexed { index, ask ->
                    Text(ask.question)
                    var expanded by remember { mutableStateOf(false) }
                    ExposedDropdownMenuBox(
                        expanded = expanded,
                        onExpandedChange = { expanded = !expanded }
                    ) {
                        TextField(
                            value = answers[offset + index],
                            onValueChange = {},
                            readOnly = true,
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
                            modifier = Modifier.menuAnchor().fillMaxWidth()
                        )
                        ExposedDropdownMenu(
                            expanded = expanded,
                            onDismissRequest = { expanded = false }
                        ) {
                            ask.choices.drop(1).forEach { choice ->
                                DropdownMenuItem(
                                    text = { Text(choice) },
                                    onClick = {
                                        answers[offset + index] = choice
                                        expanded = false
                                    }
                                )
                            }
                        }
                    }
                }

                offset += reception.askChoices.size
                reception.askMultiChoices.forEachIndexed { index, ask ->
                    Text(ask.question)
                    var expanded by remember { mutableStateOf(false) }
                    ExposedDropdownMenuBox(
                        expanded = expanded,
                        onExpandedChange = { expanded = !expanded }
                    ) {
                        TextField(
                            value = answers[offset + index],
                            onValueChange = {},
                            readOnly = true,
                            trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = expanded) },
                            modifier = Modifier.menuAnchor().fillMaxWidth()
                        )
                        ExposedDropdownMenu(
                            expanded = expanded,
                            onDismissRequest = { expanded = false }
                        ) {
                            ask.choices.drop(1).forEach { choice ->
                                DropdownMenuItem(
                                    text = { Text(choice) },
                                    onClick = {
                                        answers[offset + index] = choice
                                        expanded = false
                                    }
                                )
                            }
                        }
                    }
                }

                Text("メニュー")
                var menuExpanded by remember { mutableStateOf(false) }
                val selectedMenu = menus.find { it.menuID == selectedMenuID }
                ExposedDropdownMenuBox(
                    expanded = menuExpanded,
                    onExpandedChange = { menuExpanded = !menuExpanded }
                ) {
                    TextField(
                        value = selectedMenu?.let { "${it.menuName} - ${it.price}円" } ?: "選択してください",
                        onValueChange = {},
                        readOnly = true,
                        trailingIcon = { ExposedDropdownMenuDefaults.TrailingIcon(expanded = menuExpanded) },
                        modifier = Modifier.menuAnchor().fillMaxWidth()
                    )
                    ExposedDropdownMenu(
                        expanded = menuExpanded,
                        onDismissRequest = { menuExpanded = false }
                    ) {
                        menus.filter { it.forBookType.isNotEmpty() }.forEach { menu ->
                            DropdownMenuItem(
                                text = { Text("${menu.menuName} - ${menu.price}円") },
                                onClick = {
                                    selectedMenuID = menu.menuID
                                    menuExpanded = false
                                }
                            )
                        }
                    }
                }
            }
        },
        confirmButton = {
            Button(onClick = { onSubmit(startTime, endTime, selectedMenuID, answers.toList()) }) {
                Text("投稿")
            }
        },
        dismissButton = {
            TextButton(onClick = onDismiss) {
                Text("閉じる")
            }
        }
    )
}

package com.chat.android.feature.ticket

import androidx.compose.foundation.layout.*
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun TicketEditScreen(
    navController: NavController,
    ticketID: String?,
    viewModel: TicketViewModel = hiltViewModel()
) {
    val tickets by viewModel.tickets.collectAsState()
    val ticket = remember(ticketID, tickets) {
        tickets.find { it.ticketID == ticketID }
    }

    var title by remember { mutableStateOf(ticket?.title ?: "") }
    var description by remember { mutableStateOf(ticket?.description ?: "") }
    var status by remember { mutableStateOf(ticket?.status ?: "open") }
    var priority by remember { mutableStateOf(ticket?.priority ?: "medium") }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (ticketID == null) "New Ticket" else "Edit Ticket") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    if (ticketID != null) {
                        IconButton(onClick = {
                            viewModel.deleteTicket(ticketID)
                            navController.navigateUp()
                        }) {
                            Icon(Icons.Default.Delete, contentDescription = "Delete")
                        }
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
            TextField(
                value = title,
                onValueChange = { title = it },
                label = { Text("Title") },
                modifier = Modifier.fillMaxWidth()
            )
            Spacer(modifier = Modifier.height(8.dp))
            TextField(
                value = description,
                onValueChange = { description = it },
                label = { Text("Description") },
                modifier = Modifier.fillMaxWidth(),
                minLines = 3
            )
            Spacer(modifier = Modifier.height(8.dp))
            
            Text("Status", style = MaterialTheme.typography.labelMedium)
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
                listOf("open", "in_progress", "closed").forEach { s ->
                    FilterChip(
                        selected = status == s,
                        onClick = { status = s },
                        label = { Text(s) }
                    )
                }
            }
            
            Spacer(modifier = Modifier.height(8.dp))
            Text("Priority", style = MaterialTheme.typography.labelMedium)
            Row(modifier = Modifier.fillMaxWidth(), horizontalArrangement = Arrangement.SpaceEvenly) {
                listOf("low", "medium", "high").forEach { p ->
                    FilterChip(
                        selected = priority == p,
                        onClick = { priority = p },
                        label = { Text(p) }
                    )
                }
            }

            Spacer(modifier = Modifier.weight(1f))
            Button(
                onClick = {
                    if (ticketID == null) {
                        viewModel.createTicket(title, description, status, priority)
                    } else {
                        viewModel.updateTicket(ticketID, title, description, status, priority)
                    }
                    navController.navigateUp()
                },
                modifier = Modifier.fillMaxWidth(),
                enabled = title.isNotBlank()
            ) {
                Text("Save")
            }
        }
    }
}

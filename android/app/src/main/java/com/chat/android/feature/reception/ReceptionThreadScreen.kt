package com.chat.android.feature.reception

import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.feature.thread.ThreadScreen
import com.chat.android.navigation.ThreadHeadRoute
import com.chat.android.navigation.ThreadRoute

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun ReceptionThreadScreen(
    navController: NavController,
    channelID: String,
    code: String,
    backID: String? = null,
    viewModel: ReceptionThreadViewModel = hiltViewModel()
) {
    val isFetched by viewModel.isFetched.collectAsState()
    val nickname by viewModel.nickname.collectAsState()
    val errorMessage by viewModel.errorMessage.collectAsState()
    val threadUiState by viewModel.threadUiState.collectAsState()

    LaunchedEffect(channelID, code, backID) {
        viewModel.init(channelID, code, backID)
    }

    if (!isFetched) {
        Box(modifier = Modifier.fillMaxSize(), contentAlignment = Alignment.Center) {
            if (errorMessage != null) {
                Column(horizontalAlignment = Alignment.CenterHorizontally) {
                    Text(errorMessage!!, color = MaterialTheme.colorScheme.error)
                    Spacer(modifier = Modifier.height(16.dp))
                    Button(onClick = { /* Navigate to settings or sign in */ }) {
                        Text("サインインページ")
                    }
                }
            } else {
                CircularProgressIndicator()
            }
        }
        return
    }

    // Vue code used Messages component and ReceptionEditBox.
    // In Android, we can reuse ThreadScreen's content or wrap ThreadScreen.
    // However, ReceptionThread has specific headers like "⬅" back and "📄" copyable toggle.

    var copyable by remember { mutableStateOf(false) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = {
                    Text(
                        threadUiState.title,
                        modifier = Modifier.clickable {
                            navController.navigate(ThreadHeadRoute(parentID = threadUiState.parentID))
                        }
                    )
                },
                navigationIcon = {
                    IconButton(onClick = {
                        val backID = threadUiState.head?.backID.orEmpty()
                        if (backID.isNotBlank()) {
                            // Back to parent thread
                            navController.navigate(ThreadRoute(channelID = channelID, parentID = backID))
                        } else {
                            navController.navigateUp()
                        }
                    }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "Back")
                    }
                },
                actions = {
                    IconButton(onClick = { copyable = !copyable }) {
                        Text("📄", fontSize = 20.sp, color = if (copyable) MaterialTheme.colorScheme.primary else Color.Gray)
                    }
                }
            )
        }
    ) { padding ->
        Box(modifier = Modifier.padding(padding)) {
            // Reusing ThreadScreen logic by delegating to its composable if possible, 
            // but ThreadScreen is a top-level screen. 
            // For now, I'll copy the core part or just use ThreadScreen directly if it fits.
            // But since ReceptionThread needs to manage its own state (fetching nickname first), 
            // I will call ThreadScreen with the resolved parentID.
            
            val resolvedParentID = "@$nickname"
            ThreadScreen(
                navController = navController,
                channelID = channelID,
                parentID = resolvedParentID,
                backID = backID
            )
        }
    }
}

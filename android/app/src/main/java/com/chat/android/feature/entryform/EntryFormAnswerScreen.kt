package com.chat.android.feature.entryform

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EntryFormAnswerScreen(
    navController: NavController,
    id: String,
    viewModel: EntryFormViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()
    val answers = remember { mutableStateMapOf<Int, Any?>() }

    LaunchedEffect(id) {
        viewModel.loadAnswerForm(id)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(state.form?.title ?: "回答入力") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                },
                actions = {
                    Button(
                        onClick = { viewModel.submitAnswers(answers.toMap()) },
                        enabled = !state.saving && state.form != null
                    ) {
                        Text("送信")
                    }
                }
            )
        }
    ) { padding ->
        when {
            state.loading -> {
                Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                    CircularProgressIndicator()
                }
            }
            state.error != null -> {
                Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                    Text(state.error ?: "", color = MaterialTheme.colorScheme.error)
                }
            }
            state.form != null -> {
                LazyColumn(
                    modifier = Modifier.fillMaxSize().padding(padding),
                    contentPadding = PaddingValues(16.dp),
                    verticalArrangement = Arrangement.spacedBy(16.dp)
                ) {
                    items(state.form?.questions.orEmpty(), key = { it.sequence }) { question ->
                        Card(modifier = Modifier.fillMaxWidth()) {
                            Column(modifier = Modifier.padding(16.dp)) {
                                Text("質問 ${question.sequence}: ${getAskText(question)}", style = MaterialTheme.typography.titleMedium)
                                Spacer(Modifier.height(8.dp))
                                
                                when (question) {
                                    is Question.Text -> {
                                        val currentAnswer = (answers[question.sequence] as? String).orEmpty()
                                        OutlinedTextField(
                                            value = currentAnswer,
                                            onValueChange = { answers[question.sequence] = it },
                                            modifier = Modifier.fillMaxWidth(),
                                            label = { Text("回答を入力") }
                                        )
                                    }
                                    is Question.SingleChoice -> {
                                        val currentAnswer = answers[question.sequence] as? String
                                        question.choices.forEach { choice ->
                                            Row(
                                                verticalAlignment = Alignment.CenterVertically,
                                                modifier = Modifier.fillMaxWidth()
                                            ) {
                                                RadioButton(
                                                    selected = currentAnswer == choice,
                                                    onClick = { answers[question.sequence] = choice }
                                                )
                                                Text(choice)
                                            }
                                        }
                                    }
                                    else -> {
                                        Text("サポートされていない質問タイプです", style = MaterialTheme.typography.bodySmall)
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

private fun getAskText(question: Question): String {
    return when (question) {
        is Question.Text -> question.askText
        is Question.SingleChoice -> question.askText
        is Question.DatePicker -> question.askText
        is Question.MultiChoice -> question.askText
    }
}

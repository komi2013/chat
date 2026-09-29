package com.chat.android.feature.entryform

import androidx.compose.foundation.layout.*
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.navigation.EntryFormEditRoute
import com.chat.android.navigation.EntryFormAnswerRoute

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EntryFormListScreen(
    navController: NavController,
    viewModel: EntryFormViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()

    LaunchedEffect(Unit) {
        viewModel.loadEditor(null, null)
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("フォーム一覧") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        },
        floatingActionButton = {
            FloatingActionButton(onClick = { 
                viewModel.createForm()
                navController.navigate(EntryFormEditRoute())
            }) {
                Icon(Icons.Default.Add, contentDescription = "新規作成")
            }
        }
    ) { padding ->
        if (state.loading) {
            Box(Modifier.fillMaxSize().padding(padding), contentAlignment = Alignment.Center) {
                CircularProgressIndicator()
            }
        } else {
            LazyColumn(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                items(state.forms) { storedForm ->
                    val form = storedForm.form
                    Card(
                        modifier = Modifier.fillMaxWidth(),
                        onClick = { navController.navigate(EntryFormEditRoute(id = form.id)) }
                    ) {
                        Row(
                            modifier = Modifier.padding(16.dp),
                            verticalAlignment = Alignment.CenterVertically
                        ) {
                            Column(modifier = Modifier.weight(1f)) {
                                Text(form.title, style = MaterialTheme.typography.titleMedium)
                                Text("ID: ${form.id}", style = MaterialTheme.typography.bodySmall)
                            }
                            Button(onClick = { navController.navigate(EntryFormAnswerRoute(id = form.id)) }) {
                                Text("回答")
                            }
                        }
                    }
                }
            }
        }
    }
}

@OptIn(ExperimentalMaterial3Api::class)
@Composable
fun EntryFormEditScreen(
    navController: NavController,
    id: String? = null,
    formJson: String? = null,
    viewModel: EntryFormViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()

    LaunchedEffect(id, formJson) {
        if (state.form == null) {
            viewModel.loadEditor(id, formJson)
        }
    }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (id == null) "フォーム作成" else "フォーム編集") },
                navigationIcon = {
                    IconButton(onClick = { navController.navigateUp() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                },
                actions = {
                    Button(
                        onClick = { viewModel.saveAndPublish() },
                        enabled = !state.saving
                    ) {
                        Text("保存・送信")
                    }
                }
            )
        }
    ) { padding ->
        state.form?.let { form ->
            LazyColumn(
                modifier = Modifier.fillMaxSize().padding(padding),
                contentPadding = PaddingValues(16.dp),
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                item {
                    OutlinedTextField(
                        value = form.title,
                        onValueChange = viewModel::setTitle,
                        label = { Text("フォームタイトル") },
                        modifier = Modifier.fillMaxWidth()
                    )
                }

                items(form.questions, key = { it.sequence }) { question ->
                    QuestionEditor(
                        question = question,
                        onTextChange = { viewModel.setQuestionText(question.sequence, it) },
                        onDelete = { viewModel.removeQuestion(question.sequence) }
                    )
                }

                item {
                    Row(horizontalArrangement = Arrangement.spacedBy(8.dp)) {
                        Button(onClick = { viewModel.addQuestion(QuestionType.TEXT) }) { Text("+ テキスト") }
                        Button(onClick = { viewModel.addQuestion(QuestionType.SELECT) }) { Text("+ 単一選択") }
                    }
                }
            }
        }
    }
}

@Composable
fun QuestionEditor(
    question: Question,
    onTextChange: (String) -> Unit,
    onDelete: () -> Unit
) {
    Card(modifier = Modifier.fillMaxWidth()) {
        Column(modifier = Modifier.padding(16.dp)) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("質問 ${question.sequence}", modifier = Modifier.weight(1f))
                IconButton(onClick = onDelete) {
                    Icon(Icons.Default.Delete, contentDescription = "削除")
                }
            }
            OutlinedTextField(
                value = when(question) {
                    is Question.Text -> question.askText
                    is Question.SingleChoice -> question.askText
                    is Question.DatePicker -> question.askText
                    is Question.MultiChoice -> question.askText
                },
                onValueChange = onTextChange,
                label = { Text("質問文") },
                modifier = Modifier.fillMaxWidth()
            )
        }
    }
}

package com.chat.android.ui.screen

import android.app.DatePickerDialog
import android.app.TimePickerDialog
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.material.icons.Icons
import androidx.compose.material.icons.filled.Add
import androidx.compose.material.icons.filled.ArrowBack
import androidx.compose.material.icons.filled.Delete
import androidx.compose.material3.Button
import androidx.compose.material3.Card
import androidx.compose.material3.Checkbox
import androidx.compose.material3.DropdownMenu
import androidx.compose.material3.DropdownMenuItem
import androidx.compose.material3.ExperimentalMaterial3Api
import androidx.compose.material3.Icon
import androidx.compose.material3.IconButton
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedButton
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.Scaffold
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.material3.TopAppBar
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.collectAsState
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateMapOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.hilt.navigation.compose.hiltViewModel
import androidx.navigation.NavController
import com.chat.android.feature.entryform.EntryFormViewModel
import com.chat.android.feature.entryform.Question
import com.chat.android.feature.entryform.QuestionType
import com.chat.android.feature.entryform.StoredEntryForm
import java.util.Calendar

@Composable
@OptIn(ExperimentalMaterial3Api::class)
fun EntryFormScreen(
    navController: NavController,
    id: String,
    viewModel: EntryFormViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()
    val answers = remember(id) { mutableStateMapOf<Int, Any>() }

    LaunchedEffect(id) { viewModel.loadAnswerForm(id) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text("フォーム回答") },
                navigationIcon = {
                    IconButton(onClick = { navController.popBackStack() }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        }
    ) { insets ->
        Column(Modifier.fillMaxSize().padding(insets)) {
            DeliveryFields(
                channelId = state.channelId,
                updatedBy = state.updatedBy,
                pushNames = state.pushNames,
                onChannelIdChange = viewModel::setChannelId,
                onUpdatedByChange = viewModel::setUpdatedBy,
                onPushNamesChange = viewModel::setPushNames
            )
            state.error?.let { MessageText(it, isError = true) }
            state.notice?.let { MessageText(it, isError = false) }
            when {
                state.loading -> MessageText("フォームを読み込んでいます", isError = false)
                state.form != null -> {
                    val form = state.form!!
                    LazyColumn(
                        modifier = Modifier.weight(1f).fillMaxWidth(),
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                        contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp)
                    ) {
                        item {
                            Text(form.title, style = MaterialTheme.typography.headlineSmall, fontWeight = FontWeight.SemiBold)
                        }
                        items(form.questions, key = { it.sequence }) { question ->
                            AnswerQuestion(
                                question = question,
                                value = answers[question.sequence],
                                onAnswer = { value -> answers[question.sequence] = value }
                            )
                        }
                    }
                    Button(
                        onClick = { viewModel.submitAnswers(answers.toMap()) },
                        enabled = !state.saving,
                        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp)
                    ) {
                        Text(if (state.saving) "送信中..." else "回答を送信")
                    }
                }
            }
        }
    }
}

@Composable
private fun AnswerQuestion(
    question: Question,
    value: Any?,
    onAnswer: (Any) -> Unit
) {
    Column(verticalArrangement = Arrangement.spacedBy(8.dp)) {
        Text("${question.sequence}. ${question.text()}", style = MaterialTheme.typography.titleMedium)
        when (question) {
            is Question.Text -> OutlinedTextField(
                value = value as? String ?: "",
                onValueChange = onAnswer,
                modifier = Modifier.fillMaxWidth(),
                singleLine = true
            )
            is Question.SingleChoice -> ChoiceAnswer(
                choices = question.choices,
                value = value as? String ?: "",
                onAnswer = onAnswer
            )
            is Question.DatePicker -> DateAnswer(
                dateType = question.dateType,
                value = value as? String ?: "",
                onAnswer = onAnswer
            )
            is Question.MultiChoice -> {
                val selected = (value as? List<*>)?.filterIsInstance<String>().orEmpty()
                question.choices.forEach { choice ->
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Checkbox(
                            checked = choice in selected,
                            onCheckedChange = { checked ->
                                onAnswer(if (checked) selected + choice else selected - choice)
                            }
                        )
                        Text(choice)
                    }
                }
            }
        }
    }
}

@Composable
private fun ChoiceAnswer(choices: List<String>, value: String, onAnswer: (Any) -> Unit) {
    var expanded by remember { mutableStateOf(false) }
    Column {
        OutlinedButton(onClick = { expanded = true }, modifier = Modifier.fillMaxWidth()) {
            Text(value.ifBlank { "選択してください" })
        }
        DropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
            choices.forEach { choice ->
                DropdownMenuItem(
                    text = { Text(choice) },
                    onClick = {
                        onAnswer(choice)
                        expanded = false
                    }
                )
            }
        }
    }
}

@Composable
private fun DateAnswer(dateType: Int, value: String, onAnswer: (Any) -> Unit) {
    val context = LocalContext.current
    val calendar = remember { Calendar.getInstance() }
    val label = when (dateType) {
        2 -> "日付を選択"
        3 -> "時間を選択"
        else -> "日時を選択"
    }
    OutlinedTextField(
        value = value,
        onValueChange = {},
        readOnly = true,
        label = { Text(label) },
        modifier = Modifier.fillMaxWidth().clickable {
            if (dateType == 3) {
                showTimePicker(context, calendar, onAnswer)
            } else {
                DatePickerDialog(
                    context,
                    { _, year, month, day ->
                        calendar.set(year, month, day)
                        val date = "%04d-%02d-%02d".format(year, month + 1, day)
                        if (dateType == 1) {
                            showTimePicker(context, calendar) { time -> onAnswer("$date $time") }
                        } else {
                            onAnswer(date)
                        }
                    },
                    calendar.get(Calendar.YEAR),
                    calendar.get(Calendar.MONTH),
                    calendar.get(Calendar.DAY_OF_MONTH)
                ).show()
            }
        }
    )
}

private fun showTimePicker(
    context: android.content.Context,
    calendar: Calendar,
    onAnswer: (Any) -> Unit
) {
    TimePickerDialog(
        context,
        { _, hour, minute -> onAnswer("%02d:%02d".format(hour, minute)) },
        calendar.get(Calendar.HOUR_OF_DAY),
        calendar.get(Calendar.MINUTE),
        true
    ).show()
}

@Composable
@OptIn(ExperimentalMaterial3Api::class)
fun EntryFormEditScreen(
    navController: NavController,
    id: String? = null,
    formJson: String? = null,
    viewModel: EntryFormViewModel = hiltViewModel()
) {
    val state by viewModel.state.collectAsState()
    var addMenuExpanded by remember { mutableStateOf(false) }

    LaunchedEffect(id, formJson) { viewModel.loadEditor(id, formJson) }

    Scaffold(
        topBar = {
            TopAppBar(
                title = { Text(if (state.form == null) "フォーム一覧" else "フォーム編集") },
                navigationIcon = {
                    IconButton(onClick = {
                        if (state.form != null && id == null && formJson.isNullOrBlank()) {
                            viewModel.loadEditor(null, null)
                        } else {
                            navController.popBackStack()
                        }
                    }) {
                        Icon(Icons.Default.ArrowBack, contentDescription = "戻る")
                    }
                }
            )
        }
    ) { insets ->
        Column(Modifier.fillMaxSize().padding(insets)) {
            DeliveryFields(
                channelId = state.channelId,
                updatedBy = state.updatedBy,
                pushNames = state.pushNames,
                onChannelIdChange = viewModel::setChannelId,
                onUpdatedByChange = viewModel::setUpdatedBy,
                onPushNamesChange = viewModel::setPushNames
            )
            state.error?.let { MessageText(it, isError = true) }
            state.notice?.let { MessageText(it, isError = false) }
            when {
                state.loading -> MessageText("読み込んでいます", isError = false)
                state.form == null -> FormList(
                    forms = state.forms,
                    onCreate = viewModel::createForm,
                    onSelect = { stored -> viewModel.loadEditor(stored.form.id, null) },
                    onView = { stored -> navController.navigate("entryForm/${stored.form.id}") }
                )
                else -> {
                    val form = state.form!!
                    LazyColumn(
                        modifier = Modifier.weight(1f),
                        verticalArrangement = Arrangement.spacedBy(12.dp),
                        contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp)
                    ) {
                        item {
                            Text("ID: ${form.id}", style = MaterialTheme.typography.labelMedium)
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
                                onRemove = { viewModel.removeQuestion(question.sequence) },
                                onAddChoice = { viewModel.addChoice(question.sequence) },
                                onChoiceChange = { index, text -> viewModel.setChoice(question.sequence, index, text) },
                                onRemoveChoice = { index -> viewModel.removeChoice(question.sequence, index) },
                                onDateTypeChange = { viewModel.setDateType(question.sequence, it) },
                                onOptionKeyChange = { viewModel.setOptionKey(question.sequence, it) }
                            )
                        }
                        item {
                            Column {
                                OutlinedButton(onClick = { addMenuExpanded = true }) {
                                    Icon(Icons.Default.Add, contentDescription = null)
                                    Text("質問を追加")
                                }
                                DropdownMenu(
                                    expanded = addMenuExpanded,
                                    onDismissRequest = { addMenuExpanded = false }
                                ) {
                                    QuestionType.entries.forEach { type ->
                                        DropdownMenuItem(
                                            text = { Text(type.label()) },
                                            onClick = {
                                                viewModel.addQuestion(type)
                                                addMenuExpanded = false
                                            }
                                        )
                                    }
                                }
                            }
                        }
                    }
                    Button(
                        onClick = viewModel::saveAndPublish,
                        enabled = !state.saving,
                        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 12.dp)
                    ) {
                        Text(if (state.saving) "保存中..." else "保存して送信")
                    }
                }
            }
        }
    }
}

@Composable
private fun FormList(
    forms: List<StoredEntryForm>,
    onCreate: () -> Unit,
    onSelect: (StoredEntryForm) -> Unit,
    onView: (StoredEntryForm) -> Unit
) {
    LazyColumn(
        modifier = Modifier.fillMaxSize(),
        verticalArrangement = Arrangement.spacedBy(10.dp),
        contentPadding = androidx.compose.foundation.layout.PaddingValues(16.dp)
    ) {
        item {
            Button(onClick = onCreate, modifier = Modifier.fillMaxWidth()) {
                Icon(Icons.Default.Add, contentDescription = null)
                Text("新しいフォーム")
            }
        }
        if (forms.isEmpty()) {
            item { MessageText("保存されたフォームはありません", isError = false) }
        }
        items(forms, key = { it.form.id }) { stored ->
            Card(Modifier.fillMaxWidth().clickable { onSelect(stored) }) {
                Row(
                    modifier = Modifier.padding(16.dp).fillMaxWidth(),
                    horizontalArrangement = Arrangement.SpaceBetween,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    Column(modifier = Modifier.weight(1f)) {
                        Text(stored.form.title.ifBlank { "タイトルなし" }, fontWeight = FontWeight.Medium)
                        Spacer(Modifier.height(4.dp))
                        Text("${stored.form.questions.size} 件の質問 · ${stored.form.id}", style = MaterialTheme.typography.bodySmall)
                    }
                    Button(onClick = { onView(stored) }) {
                        Text("回答画面")
                    }
                }
            }
        }
    }
}

@Composable
private fun QuestionEditor(
    question: Question,
    onTextChange: (String) -> Unit,
    onRemove: () -> Unit,
    onAddChoice: () -> Unit,
    onChoiceChange: (Int, String) -> Unit,
    onRemoveChoice: (Int) -> Unit,
    onDateTypeChange: (Int) -> Unit,
    onOptionKeyChange: (Boolean) -> Unit
) {
    Card(Modifier.fillMaxWidth()) {
        Column(
            modifier = Modifier.padding(14.dp),
            verticalArrangement = Arrangement.spacedBy(8.dp)
        ) {
            Row(verticalAlignment = Alignment.CenterVertically) {
                Text("${question.sequence}. ${question.type.label()}", modifier = Modifier.weight(1f), fontWeight = FontWeight.Medium)
                IconButton(onClick = onRemove) {
                    Icon(Icons.Default.Delete, contentDescription = "質問を削除")
                }
            }
            OutlinedTextField(
                value = question.text(),
                onValueChange = onTextChange,
                label = { Text("質問文") },
                modifier = Modifier.fillMaxWidth()
            )
            when (question) {
                is Question.SingleChoice -> ChoiceEditor(question.choices, onAddChoice, onChoiceChange, onRemoveChoice)
                is Question.MultiChoice -> {
                    ChoiceEditor(question.choices, onAddChoice, onChoiceChange, onRemoveChoice)
                    OptionKeyToggle(question.optionKey, onOptionKeyChange)
                }
                is Question.DatePicker -> {
                    var expanded by remember(question.sequence) { mutableStateOf(false) }
                    Column {
                        OutlinedButton(onClick = { expanded = true }) {
                            Text(when (question.dateType) {
                                2 -> "日付のみ"
                                3 -> "時間のみ"
                                else -> "日時"
                            })
                        }
                        DropdownMenu(expanded = expanded, onDismissRequest = { expanded = false }) {
                            listOf(1 to "日時", 2 to "日付のみ", 3 to "時間のみ").forEach { (type, label) ->
                                DropdownMenuItem(
                                    text = { Text(label) },
                                    onClick = {
                                        onDateTypeChange(type)
                                        expanded = false
                                    }
                                )
                            }
                        }
                    }
                    OptionKeyToggle(question.optionKey, onOptionKeyChange)
                }
                is Question.Text -> Unit
            }
        }
    }
}

@Composable
private fun ChoiceEditor(
    choices: List<String>,
    onAdd: () -> Unit,
    onChange: (Int, String) -> Unit,
    onRemove: (Int) -> Unit
) {
    choices.forEachIndexed { index, choice ->
        Row(verticalAlignment = Alignment.CenterVertically) {
            OutlinedTextField(
                value = choice,
                onValueChange = { onChange(index, it) },
                label = { Text("選択肢 ${index + 1}") },
                modifier = Modifier.weight(1f)
            )
            IconButton(onClick = { onRemove(index) }) {
                Icon(Icons.Default.Delete, contentDescription = "選択肢を削除")
            }
        }
    }
    TextButton(onClick = onAdd) {
        Icon(Icons.Default.Add, contentDescription = null)
        Text("選択肢を追加")
    }
}

@Composable
private fun OptionKeyToggle(value: Boolean, onChange: (Boolean) -> Unit) {
    Row(verticalAlignment = Alignment.CenterVertically) {
        Checkbox(checked = value, onCheckedChange = onChange)
        Text("オプションキーに含める")
    }
}

@Composable
private fun DeliveryFields(
    channelId: String,
    updatedBy: String,
    pushNames: String,
    onChannelIdChange: (String) -> Unit,
    onUpdatedByChange: (String) -> Unit,
    onPushNamesChange: (String) -> Unit
) {
    Column(
        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp),
        verticalArrangement = Arrangement.spacedBy(8.dp)
    ) {
        OutlinedTextField(
            value = channelId,
            onValueChange = onChannelIdChange,
            label = { Text("チャネルID") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )
        OutlinedTextField(
            value = updatedBy,
            onValueChange = onUpdatedByChange,
            label = { Text("送信者の別名") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )
        OutlinedTextField(
            value = pushNames,
            onValueChange = onPushNamesChange,
            label = { Text("通知対象の別名 (カンマ区切り)") },
            modifier = Modifier.fillMaxWidth(),
            singleLine = true
        )
    }
}

@Composable
private fun MessageText(message: String, isError: Boolean) {
    Text(
        text = message,
        color = if (isError) MaterialTheme.colorScheme.error else MaterialTheme.colorScheme.primary,
        modifier = Modifier.fillMaxWidth().padding(horizontal = 16.dp, vertical = 8.dp)
    )
}

private fun Question.text(): String = when (this) {
    is Question.Text -> askText
    is Question.SingleChoice -> askText
    is Question.DatePicker -> askText
    is Question.MultiChoice -> askText
}

private fun QuestionType.label(): String = when (this) {
    QuestionType.TEXT -> "テキスト"
    QuestionType.SELECT -> "単一選択"
    QuestionType.DATE -> "日付・時刻"
    QuestionType.MULTI_SELECT -> "複数選択"
}

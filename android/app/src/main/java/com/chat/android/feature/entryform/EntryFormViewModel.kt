package com.chat.android.feature.entryform

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import dagger.hilt.android.lifecycle.HiltViewModel
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import java.util.UUID
import javax.inject.Inject

data class EntryFormUiState(
    val forms: List<StoredEntryForm> = emptyList(),
    val form: EntryForm? = null,
    val channelId: String = "",
    val updatedBy: String = "",
    val pushNames: String = "",
    val loading: Boolean = false,
    val saving: Boolean = false,
    val error: String? = null,
    val notice: String? = null
)

@HiltViewModel
class EntryFormViewModel @Inject constructor(
    private val repository: EntryFormRepository
) : ViewModel() {
    private val mutableState = MutableStateFlow(
        EntryFormUiState(
            channelId = repository.getChannelId(),
            updatedBy = repository.getUpdatedBy(),
            pushNames = repository.getPushNames()
        )
    )
    val state = mutableState.asStateFlow()

    fun loadAnswerForm(id: String) {
        viewModelScope.launch {
            mutableState.value = mutableState.value.copy(loading = true, error = null)
            val form = repository.getForm(id)
            mutableState.value = mutableState.value.copy(
                form = form,
                loading = false,
                error = if (form == null) "フォームが端末に保存されていません" else null
            )
        }
    }

    fun loadEditor(id: String?, formJson: String?) {
        viewModelScope.launch {
            mutableState.value = mutableState.value.copy(loading = true, error = null)
            when {
                !formJson.isNullOrBlank() -> {
                    val (imported, error) = repository.importForm(formJson)
                    mutableState.value = mutableState.value.copy(
                        form = imported,
                        loading = false,
                        error = error
                    )
                }
                !id.isNullOrBlank() -> {
                    val form = repository.getForm(id)
                    mutableState.value = mutableState.value.copy(
                        form = form,
                        loading = false,
                        error = if (form == null) "フォームが端末に保存されていません" else null
                    )
                }
                else -> {
                    val forms = repository.getAllForms()
                    mutableState.value = mutableState.value.copy(forms = forms, form = null, loading = false)
                }
            }
        }
    }

    fun createForm() {
        val form = EntryForm(
            id = UUID.randomUUID().toString().replace("-", "").take(12),
            title = "",
            questions = listOf(Question.Text(sequence = 1, askText = ""))
        )
        mutableState.value = mutableState.value.copy(form = form, error = null, notice = null)
    }

    fun setTitle(title: String) = updateForm { copy(title = title) }

    fun setChannelId(channelId: String) {
        repository.setChannelId(channelId)
        mutableState.value = mutableState.value.copy(channelId = channelId)
    }

    fun setUpdatedBy(updatedBy: String) {
        repository.setUpdatedBy(updatedBy)
        mutableState.value = mutableState.value.copy(updatedBy = updatedBy)
    }

    fun setPushNames(pushNames: String) {
        repository.setPushNames(pushNames)
        mutableState.value = mutableState.value.copy(pushNames = pushNames)
    }

    fun addQuestion(type: QuestionType) = updateForm {
        val question = when (type) {
            QuestionType.TEXT -> Question.Text(questions.size + 1, "")
            QuestionType.SELECT -> Question.SingleChoice(questions.size + 1, "", listOf("", ""))
            QuestionType.DATE -> Question.DatePicker(questions.size + 1, "", 1, false)
            QuestionType.MULTI_SELECT -> Question.MultiChoice(questions.size + 1, "", listOf("", ""), false)
        }
        copy(questions = EntryFormCodec.normalize(questions + question))
    }

    fun removeQuestion(sequence: Int) = updateForm {
        copy(questions = EntryFormCodec.normalize(questions.filterNot { it.sequence == sequence }))
    }

    fun setQuestionText(sequence: Int, text: String) = updateQuestion(sequence) { question ->
        when (question) {
            is Question.Text -> question.copy(askText = text)
            is Question.SingleChoice -> question.copy(askText = text)
            is Question.DatePicker -> question.copy(askText = text)
            is Question.MultiChoice -> question.copy(askText = text)
        }
    }

    fun setDateType(sequence: Int, dateType: Int) = updateQuestion(sequence) { question ->
        (question as? Question.DatePicker)?.copy(dateType = dateType) ?: question
    }

    fun setOptionKey(sequence: Int, enabled: Boolean) = updateQuestion(sequence) { question ->
        when (question) {
            is Question.DatePicker -> question.copy(optionKey = enabled)
            is Question.MultiChoice -> question.copy(optionKey = enabled)
            else -> question
        }
    }

    fun addChoice(sequence: Int) = updateQuestion(sequence) { question ->
        when (question) {
            is Question.SingleChoice -> question.copy(choices = question.choices + "")
            is Question.MultiChoice -> question.copy(choices = question.choices + "")
            else -> question
        }
    }

    fun setChoice(sequence: Int, index: Int, value: String) = updateQuestion(sequence) { question ->
        when (question) {
            is Question.SingleChoice -> question.copy(choices = question.choices.updated(index, value))
            is Question.MultiChoice -> question.copy(choices = question.choices.updated(index, value))
            else -> question
        }
    }

    fun removeChoice(sequence: Int, index: Int) = updateQuestion(sequence) { question ->
        when (question) {
            is Question.SingleChoice -> question.copy(choices = question.choices.removed(index))
            is Question.MultiChoice -> question.copy(choices = question.choices.removed(index))
            else -> question
        }
    }

    fun saveAndPublish() {
        val form = mutableState.value.form ?: return
        if (form.title.isBlank()) {
            mutableState.value = mutableState.value.copy(error = "フォームタイトルを入力してください")
            return
        }
        viewModelScope.launch {
            mutableState.value = mutableState.value.copy(saving = true, error = null, notice = null)
            val error = repository.saveAndPublish(form)
            val forms = repository.getAllForms()
            mutableState.value = mutableState.value.copy(
                forms = forms,
                saving = false,
                error = error,
                notice = if (error == null) "フォームを保存して送信しました" else null
            )
        }
    }

    fun submitAnswers(answers: Map<Int, Any?>) {
        val form = mutableState.value.form ?: return
        viewModelScope.launch {
            mutableState.value = mutableState.value.copy(saving = true, error = null, notice = null)
            val error = repository.submitAnswers(form, answers)
            mutableState.value = mutableState.value.copy(
                saving = false,
                error = error,
                notice = if (error == null) "回答を送信しました" else null
            )
        }
    }

    private fun updateQuestion(sequence: Int, transform: (Question) -> Question) = updateForm {
        copy(questions = questions.map { question ->
            if (question.sequence == sequence) transform(question) else question
        })
    }

    private fun updateForm(transform: EntryForm.() -> EntryForm) {
        val form = mutableState.value.form ?: return
        mutableState.value = mutableState.value.copy(
            form = transform(form).let { it.copy(questions = EntryFormCodec.normalize(it.questions)) },
            error = null
        )
    }

    private fun List<String>.updated(index: Int, value: String): List<String> =
        mapIndexed { itemIndex, item -> if (itemIndex == index) value else item }

    private fun List<String>.removed(index: Int): List<String> =
        filterIndexed { itemIndex, _ -> itemIndex != index }
}

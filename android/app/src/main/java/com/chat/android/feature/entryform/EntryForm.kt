package com.chat.android.feature.entryform

enum class QuestionType {
    TEXT,
    SELECT,
    DATE,
    MULTI_SELECT
}

sealed class Question(
    open val sequence: Int,
    open val type: QuestionType,
    open val optionKey: Boolean = false
) {
    data class Text(
        override val sequence: Int,
        val askText: String
    ) : Question(sequence, QuestionType.TEXT)

    data class SingleChoice(
        override val sequence: Int,
        val askText: String,
        val choices: List<String>
    ) : Question(sequence, QuestionType.SELECT)

    data class DatePicker(
        override val sequence: Int,
        val askText: String,
        val dateType: Int,
        override val optionKey: Boolean
    ) : Question(sequence, QuestionType.DATE, optionKey)

    data class MultiChoice(
        override val sequence: Int,
        val askText: String,
        val choices: List<String>,
        override val optionKey: Boolean
    ) : Question(sequence, QuestionType.MULTI_SELECT, optionKey)
}

data class EntryForm(
    val id: String,
    val title: String,
    val questions: List<Question>
)

data class Answer(
    val answerID: String,
    val channelID: String,
    val from: String, // "1" for entryForm, "2" for reception
    val askID: String,
    val answerBy: String,
    val answers: String, // JSON string of [{sequence: int, answer: any}, ...]
    val updatedAt: Long = System.currentTimeMillis()
)

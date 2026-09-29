package com.chat.android.feature.entryform

import com.google.gson.JsonArray
import com.google.gson.JsonElement
import com.google.gson.JsonObject
import com.google.gson.JsonParser
import com.google.gson.JsonPrimitive
import java.util.UUID

object EntryFormCodec {
    fun parse(json: String): EntryForm {
        val root = JsonParser().parse(json).asJsonObject
        val id = root.string("id", "entryFormID").ifBlank {
            UUID.randomUUID().toString().replace("-", "").take(12)
        }
        val title = root.string("title", "entryFormTitle")
        val questions = if (root.has("questions") && root["questions"].isJsonArray) {
            parseUnifiedQuestions(root["questions"].asJsonArray)
        } else {
            parseLegacyQuestions(root)
        }
        return EntryForm(id, title, normalize(questions))
    }

    fun toStorageJson(form: EntryForm): String {
        val root = JsonObject().apply {
            addProperty("id", form.id)
            addProperty("title", form.title)
            add("questions", unifiedQuestions(form.questions))
        }
        return root.toString()
    }

    fun toWireJson(form: EntryForm): String {
        val questions = normalize(form.questions)
        val root = JsonObject().apply {
            addProperty("entryFormID", form.id)
            addProperty("entryFormTitle", form.title)
            add("asks", JsonArray())
            add("askChoices", JsonArray())
            add("dates", JsonArray())
            add("askMultiChoices", JsonArray())
        }
        questions.forEach { question ->
            val item = JsonObject().apply {
                addProperty("question", question.askText())
                addProperty("sequence", question.sequence)
            }
            when (question) {
                is Question.Text -> root["asks"].asJsonArray.add(item)
                is Question.SingleChoice -> {
                    item.add("choices", stringArray(question.choices))
                    root["askChoices"].asJsonArray.add(item)
                }
                is Question.DatePicker -> {
                    item.addProperty("dateType", question.dateType)
                    item.addProperty("optionKey", question.optionKey)
                    root["dates"].asJsonArray.add(item)
                }
                is Question.MultiChoice -> {
                    item.add("choices", stringArray(question.choices))
                    item.addProperty("optionKey", question.optionKey)
                    root["askMultiChoices"].asJsonArray.add(item)
                }
            }
        }
        return root.toString()
    }

    fun normalize(questions: List<Question>): List<Question> =
        questions.withIndex()
            .sortedWith(compareBy<IndexedValue<Question>> { it.value.sequence }.thenBy { it.index })
            .mapIndexed { index, item -> item.value.withSequence(index + 1) }

    private fun parseLegacyQuestions(root: JsonObject): List<Question> {
        val questions = mutableListOf<Question>()
        root.array("asks").forEach { item ->
            questions += Question.Text(item.int("sequence"), item.string("question"))
        }
        root.array("askChoices").forEach { item ->
            questions += Question.SingleChoice(
                item.int("sequence"), item.string("question"), item.strings("choices")
            )
        }
        root.array("dates").forEach { item ->
            questions += Question.DatePicker(
                item.int("sequence"), item.string("question"), item.int("dateType", 1),
                item.boolean("optionKey")
            )
        }
        root.array("askMultiChoices").forEach { item ->
            questions += Question.MultiChoice(
                item.int("sequence"), item.string("question"), item.strings("choices"),
                item.boolean("optionKey")
            )
        }
        return questions
    }

    private fun parseUnifiedQuestions(array: JsonArray): List<Question> = array.mapNotNull { element ->
        if (!element.isJsonObject) return@mapNotNull null
        val item = element.asJsonObject
        val sequence = item.int("sequence")
        val text = item.string("askText", "question")
        when (item.string("type").uppercase()) {
            "TEXT" -> Question.Text(sequence, text)
            "SELECT", "SINGLE_CHOICE" -> Question.SingleChoice(sequence, text, item.strings("choices"))
            "DATE", "DATEPICKER" -> Question.DatePicker(
                sequence, text, item.int("dateType", 1), item.boolean("optionKey")
            )
            "MULTI_SELECT", "MULTICHOICE" -> Question.MultiChoice(
                sequence, text, item.strings("choices"), item.boolean("optionKey")
            )
            else -> null
        }
    }

    private fun unifiedQuestions(questions: List<Question>): JsonArray = JsonArray().apply {
        normalize(questions).forEach { question ->
            add(JsonObject().apply {
                addProperty("sequence", question.sequence)
                addProperty("type", question.type.name)
                addProperty("askText", question.askText())
                when (question) {
                    is Question.SingleChoice -> add("choices", stringArray(question.choices))
                    is Question.DatePicker -> {
                        addProperty("dateType", question.dateType)
                        addProperty("optionKey", question.optionKey)
                    }
                    is Question.MultiChoice -> {
                        add("choices", stringArray(question.choices))
                        addProperty("optionKey", question.optionKey)
                    }
                    is Question.Text -> Unit
                }
            })
        }
    }

    private fun Question.askText(): String = when (this) {
        is Question.Text -> askText
        is Question.SingleChoice -> askText
        is Question.DatePicker -> askText
        is Question.MultiChoice -> askText
    }

    private fun Question.withSequence(sequence: Int): Question = when (this) {
        is Question.Text -> copy(sequence = sequence)
        is Question.SingleChoice -> copy(sequence = sequence)
        is Question.DatePicker -> copy(sequence = sequence)
        is Question.MultiChoice -> copy(sequence = sequence)
    }

    private fun JsonObject.array(name: String): List<JsonObject> =
        get(name)?.takeIf(JsonElement::isJsonArray)?.asJsonArray?.mapNotNull {
            it.takeIf(JsonElement::isJsonObject)?.asJsonObject
        }.orEmpty()

    private fun JsonObject.string(vararg names: String): String = names.firstNotNullOfOrNull { name ->
        get(name)?.takeIf { it.isJsonPrimitive && it.asJsonPrimitive.isString }?.asString
    }.orEmpty()

    private fun JsonObject.int(name: String, default: Int = 0): Int =
        get(name)?.takeIf(JsonElement::isJsonPrimitive)?.let { runCatching { it.asInt }.getOrNull() } ?: default

    private fun JsonObject.boolean(name: String): Boolean =
        get(name)?.takeIf(JsonElement::isJsonPrimitive)?.let { runCatching { it.asBoolean }.getOrNull() } ?: false

    private fun JsonObject.strings(name: String): List<String> =
        get(name)?.takeIf(JsonElement::isJsonArray)?.asJsonArray?.mapNotNull {
            it.takeIf(JsonElement::isJsonPrimitive)?.asString
        }.orEmpty()

    private fun stringArray(values: List<String>): JsonArray = JsonArray().apply {
        values.forEach { add(JsonPrimitive(it)) }
    }
}

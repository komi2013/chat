package com.chat.android.feature.entryform

import com.google.gson.JsonParser
import org.junit.Assert.assertEquals
import org.junit.Assert.assertTrue
import org.junit.Test

class EntryFormCodecTest {
    @Test
    fun legacyQuestionsReceiveOneUniqueSequenceAcrossTypes() {
        val form = EntryFormCodec.parse(
            """{"entryFormID":"form-a","entryFormTitle":"Apply","asks":[{"question":"Text","sequence":2}],"askChoices":[{"question":"Select","choices":["A","B"],"sequence":1}],"dates":[{"question":"Date","sequence":1,"dateType":2,"optionKey":true}],"askMultiChoices":[{"question":"Multi","choices":["X","Y"],"sequence":2,"optionKey":true}]}"""
        )

        assertEquals(listOf(1, 2, 3, 4), form.questions.map { it.sequence })
        assertTrue(form.questions[1] is Question.DatePicker)
        assertEquals(2, (form.questions[1] as Question.DatePicker).dateType)
    }

    @Test
    fun wireJsonKeepsDateQuestionsInDateArray() {
        val form = EntryFormCodec.parse(
            """{"entryFormID":"form-b","dates":[{"question":"Visit date","sequence":1,"dateType":3,"optionKey":true}]}"""
        )

        val wire = JsonParser().parse(EntryFormCodec.toWireJson(form)).asJsonObject
        val date = wire.getAsJsonArray("dates").single().asJsonObject

        assertEquals("Visit date", date.get("question").asString)
        assertEquals(1, date.get("sequence").asInt)
        assertEquals(3, date.get("dateType").asInt)
        assertTrue(date.get("optionKey").asBoolean)
        assertEquals(0, wire.getAsJsonArray("asks").size())
    }

    @Test
    fun normalizedFormSurvivesSqliteJsonFormatRoundTrip() {
        val original = EntryFormCodec.parse(
            """{"entryFormID":"form-c","entryFormTitle":"Choices","askMultiChoices":[{"question":"Pick","choices":["A"],"sequence":7,"optionKey":true}]}"""
        )

        val restored = EntryFormCodec.parse(EntryFormCodec.toStorageJson(original))

        assertEquals(original, restored)
    }
}

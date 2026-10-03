package com.chat.android.feature.channel

import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.input.TextFieldValue
import com.chat.android.core.util.MarkdownFormatRepair
import org.junit.Assert.assertEquals
import org.junit.Test

class MarkdownCodecTest {

    @Test
    fun editorTextHidesBoldMarkersAndMapsSelectionToContent() {
        val original = "aa**aa**aa"
        val transformed = MarkdownCodec.editorText(original)

        assertEquals("aaaaaa", transformed.text.text)
        val selectionStart = transformed.offsetMapping.transformedToOriginal(2)
        val selectionEnd = transformed.offsetMapping.transformedToOriginal(4)
        assertEquals(4, selectionStart)
        assertEquals(6, selectionEnd)

        val afterDelete = original.removeRange(selectionStart, selectionEnd)
        val repaired = MarkdownFormatRepair.repair(
            TextFieldValue(afterDelete, TextRange(selectionStart))
        )
        assertEquals("aaaa", repaired.text)
    }

    @Test
    fun editorTextHidesFullWidthBoldMarkersAndMapsSelectionToContent() {
        val original = "aa＊太＊aa・＊太＊aa"
        val transformed = MarkdownCodec.editorText(original)

        assertEquals("aaaaaa", transformed.text.text)
        val selectionStart = transformed.offsetMapping.transformedToOriginal(2)
        val selectionEnd = transformed.offsetMapping.transformedToOriginal(4)
        assertEquals(5, selectionStart)
        assertEquals(7, selectionEnd)

        val afterDelete = original.removeRange(selectionStart, selectionEnd)
        val repaired = MarkdownFormatRepair.repair(
            TextFieldValue(afterDelete, TextRange(selectionStart))
        )
        assertEquals("aaaa", repaired.text)
    }
}

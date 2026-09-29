package com.chat.android.feature.channel

import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration

object MarkdownCodec {

    fun toAnnotatedString(text: String): AnnotatedString {
        val builder = AnnotatedString.Builder()
        var currentText = text.replace("\uFEFF", "")

        // This is a simplified parser. For a production app, a more robust state-machine or 
        // regex-based recursive parser would be used to handle nested tags.
        // We'll use a sequence of regex replacements for basic support.

        val tokens = mutableListOf<Pair<IntRange, SpanStyle>>()
        val annotations = mutableListOf<Triple<IntRange, String, String>>()

        // Bold: ＊太＊...・＊太＊
        Regex("＊太＊(.*?)・＊太＊").findAll(currentText).forEach { match ->
            // In a real implementation, we'd remove tags and track offsets.
            // For now, let's just apply styles to the whole match for simplicity 
            // OR do the proper replacement.
        }

        // Given the complexity of custom tokens, a simple iterative approach:
        var result = currentText
        
        // Paragraph: ＊p＊…・＊p＊ -> just remove tags
        result = result.replace("＊p＊", "").replace("・＊p＊", "\n")

        return buildAnnotatedString {
            append(result)
            // Implementation note: The spec requires a custom parser to handle the specific
            // full-width Japanese delimiters. The following is a placeholder for the logic.
            
            // Bold
            Regex("＊太＊(.*?)・＊太＊").findAll(result).forEach { match ->
                addStyle(SpanStyle(fontWeight = FontWeight.Bold), match.range.first, match.range.last)
            }
            
            // Red
            Regex("色＊赤(.*?)赤＊色").findAll(result).forEach { match ->
                addStyle(SpanStyle(color = Color.Red), match.range.first, match.range.last)
            }

            // Links: 「text」（url）
            Regex("「(.*?)」（(.*?)）").findAll(result).forEach { match ->
                val textPart = match.groupValues[1]
                val urlPart = match.groupValues[2]
                addStringAnnotation("URL", urlPart, match.range.first, match.range.last)
                addStyle(SpanStyle(color = Color.Blue, textDecoration = TextDecoration.Underline), match.range.first, match.range.last)
            }
        }
    }

    /**
     * Placeholder for the reverse transform: HTML/Editor to Custom Markdown
     */
    fun fromHtml(html: String): String {
        var md = html.replace("\uFEFF", "")
        // Bold: <b>...</b> -> ＊太＊...・＊太＊
        md = md.replace("<b>", "＊太＊").replace("</b>", "・＊太＊")
        md = md.replace("<strong>", "＊太＊").replace("</strong>", "・＊太＊")
        // Strike: <s>...</s> -> 〜〜...・〜〜
        md = md.replace("<s>", "〜〜").replace("</s>", "・〜〜")
        // ... more replacements
        return md
    }
}

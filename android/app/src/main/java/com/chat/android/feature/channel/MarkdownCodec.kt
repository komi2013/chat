package com.chat.android.feature.channel

import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.AnnotatedString
import androidx.compose.ui.text.SpanStyle
import androidx.compose.ui.text.buildAnnotatedString
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontStyle
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextDecoration
import androidx.compose.ui.text.withStyle
import androidx.compose.ui.unit.sp

object MarkdownCodec {

    /**
     * カスタム markdown 記号（全角）を解析し、装飾された AnnotatedString を返す。
     * 表示時には記号自体は非表示（または置換）にする。
     */
    fun toAnnotatedString(text: String): AnnotatedString {
        val result = text.replace("\uFEFF", "")
            .replace("＊p＊", "")
            .replace("・＊p＊", "\n")

        return buildAnnotatedString {
            var currentPos = 0
            // 非強欲マッチで各タグを抽出
            val pattern = "(＊太＊.*?・＊太＊|〜〜.*?・〜〜|＜quote＞.*?＜・quote＞|｀｀｀.*?・｀｀｀|色＊赤.*?赤＊色|＠＠.*?・＠＠|「.*?」（.*?）)"
            val regex = Regex(pattern)

            regex.findAll(result).forEach { match ->
                // タグの前のテキストを追加
                append(result.substring(currentPos, match.range.first))

                val tag = match.value
                when {
                    tag.startsWith("＊太＊") -> {
                        val content = tag.substring(3, tag.length - 4)
                        withStyle(SpanStyle(fontWeight = FontWeight.Bold)) { append(content) }
                    }
                    tag.startsWith("〜〜") -> {
                        val content = tag.substring(2, tag.length - 3)
                        withStyle(SpanStyle(textDecoration = TextDecoration.LineThrough)) { append(content) }
                    }
                    tag.startsWith("＜quote＞") -> {
                        val content = tag.substring(7, tag.length - 8)
                        withStyle(SpanStyle(color = Color.Gray, fontStyle = FontStyle.Italic)) { append(content) }
                    }
                    tag.startsWith("｀｀｀") -> {
                        val content = tag.substring(3, tag.length - 4)
                        withStyle(SpanStyle(fontFamily = FontFamily.Monospace, background = Color.LightGray.copy(alpha = 0.3f))) { append(content) }
                    }
                    tag.startsWith("色＊赤") -> {
                        val content = tag.substring(3, tag.length - 3)
                        withStyle(SpanStyle(color = Color.Red)) { append(content) }
                    }
                    tag.startsWith("＠＠") -> {
                        val content = tag.substring(2, tag.length - 3)
                        withStyle(SpanStyle(color = Color(0xFF1976D2), fontWeight = FontWeight.Bold)) { append(content) }
                    }
                    tag.startsWith("「") -> {
                        // 「text」（url）
                        val mid = tag.indexOf("」（")
                        if (mid != -1) {
                            val textPart = tag.substring(1, mid)
                            val urlPart = tag.substring(mid + 2, tag.length - 1)
                            val start = length
                            withStyle(SpanStyle(color = Color.Blue, textDecoration = TextDecoration.Underline)) {
                                append(textPart)
                            }
                            addStringAnnotation("URL", urlPart, start, length)
                        } else {
                            append(tag)
                        }
                    }
                    else -> append(tag)
                }
                currentPos = match.range.last + 1
            }
            // 残りのテキストを追加
            append(result.substring(currentPos))
        }
    }

    /**
     * 文字列内の記号を維持したまま、装飾だけを適用する（エディタ用）。
     * 記号自体は薄い色にして目立たなくする。
     */
    fun styleTags(text: String): AnnotatedString {
        return buildAnnotatedString {
            append(text)
            val tagStyle = SpanStyle(
                color = Color.Gray.copy(alpha = 0.3f),
                fontSize = 10.sp
            )

            // Bold: ＊太＊...・＊太＊
            Regex("＊太＊(.*?)・＊太＊").findAll(text).forEach { match ->
                addStyle(tagStyle, match.range.first, match.range.first + 3)
                addStyle(SpanStyle(fontWeight = FontWeight.Bold), match.range.first + 3, match.range.last - 3)
                addStyle(tagStyle, match.range.last - 3, match.range.last + 1)
            }

            // Strike: 〜〜...・〜〜
            Regex("〜〜(.*?)・〜〜").findAll(text).forEach { match ->
                addStyle(tagStyle, match.range.first, match.range.first + 2)
                addStyle(SpanStyle(textDecoration = TextDecoration.LineThrough), match.range.first + 2, match.range.last - 2)
                addStyle(tagStyle, match.range.last - 2, match.range.last + 1)
            }

            // Quote: ＜quote＞...＜・quote＞
            Regex("＜quote＞(.*?)＜・quote＞").findAll(text).forEach { match ->
                addStyle(tagStyle, match.range.first, match.range.first + 7)
                addStyle(SpanStyle(color = Color.Gray, fontStyle = FontStyle.Italic), match.range.first + 7, match.range.last - 7)
                addStyle(tagStyle, match.range.last - 7, match.range.last + 1)
            }

            // Code: ｀｀｀...・｀｀｀
            Regex("｀｀｀(.*?)・｀｀｀").findAll(text).forEach { match ->
                addStyle(tagStyle, match.range.first, match.range.first + 3)
                addStyle(SpanStyle(fontFamily = FontFamily.Monospace, background = Color.LightGray.copy(alpha = 0.3f)), match.range.first + 3, match.range.last - 3)
                addStyle(tagStyle, match.range.last - 3, match.range.last + 1)
            }

            // Red: 色＊赤...赤＊色
            Regex("色＊赤(.*?)赤＊色").findAll(text).forEach { match ->
                addStyle(tagStyle, match.range.first, match.range.first + 3)
                addStyle(SpanStyle(color = Color.Red), match.range.first + 3, match.range.last - 2)
                addStyle(tagStyle, match.range.last - 2, match.range.last + 1)
            }

            // Mentions: ＠＠名前・＠＠
            Regex("＠＠(.*?)・＠＠").findAll(text).forEach { match ->
                addStyle(tagStyle, match.range.first, match.range.first + 2)
                addStyle(SpanStyle(color = Color(0xFF1976D2), fontWeight = FontWeight.Bold), match.range.first + 2, match.range.last - 2)
                addStyle(tagStyle, match.range.last - 2, match.range.last + 1)
            }

            // Links: 「text」（url）
            Regex("「(.*?)」（(.*?)）").findAll(text).forEach { match ->
                val mid = text.indexOf("」（", match.range.first)
                if (mid != -1) {
                    addStyle(tagStyle, match.range.first, match.range.first + 1)
                    addStyle(SpanStyle(color = Color.Blue, textDecoration = TextDecoration.Underline), match.range.first + 1, mid)
                    addStyle(tagStyle, mid, mid + 2)
                    addStyle(SpanStyle(color = Color.Gray, textDecoration = TextDecoration.Underline, fontSize = 10.sp), mid + 2, match.range.last)
                    addStyle(tagStyle, match.range.last, match.range.last + 1)
                }
            }
        }
    }

    /**
     * Editor の HTML からカスタム Markdown へ変換する（vue/src/components/MarkdownCodec.js 相当）。
     */
    fun fromHtml(html: String): String {
        var md = html.replace("\uFEFF", "")
        md = md.replace("<b>", "＊太＊").replace("</b>", "・＊太＊")
        md = md.replace("<strong>", "＊太＊").replace("</strong>", "・＊太＊")
        md = md.replace("<s>", "〜〜").replace("</s>", "・〜〜")
        md = md.replace("<blockquote>", "＜quote＞").replace("</blockquote>", "＜・quote＞")
        md = md.replace("<pre><code>", "｀｀｀").replace("</code></pre>", "・｀｀｀")
        md = md.replace("<span style=\"color: red;\">", "色＊赤").replace("</span>", "赤＊色")
        return md
    }
}

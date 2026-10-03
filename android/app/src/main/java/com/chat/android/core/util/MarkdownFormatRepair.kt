package com.chat.android.core.util

import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.input.TextFieldValue

/**
 * 入力欄の装飾記号を保守する。
 *
 * 本文は「open + 内容 + close」の記号で装飾するため、削除操作で記号の片方だけが
 * 消えたり 1 文字欠けたりすると、壊れた記号が本文に残ってしまう。
 * 例: 「aa＊太＊aa・＊太＊aa」の一部を削除して「aa太＊aa＊太aa」になる。
 *
 * ここでは「正しく対になった記号」だけを残し、対になっていない記号や 1 文字欠けの
 * 断片は取り除く。片方だけ消したら装飾ごと全部消える（＝本文だけが残る）挙動にする。
 */
object MarkdownFormatRepair {

    /** open と close が異なる記号ペア。 */
    private val PAIRS = listOf(
        "＊太＊" to "・＊太＊",
        "〜〜" to "・〜〜",
        "＜quote＞" to "＜・quote＞",
        "｀｀｀" to "・｀｀｀",
        "色＊赤" to "赤＊色",
        "＠＠" to "・＠＠"
    )

    /** open と close が同じ記号ペア。 */
    private val SAME_TOKENS = listOf("**", "~~", "```")

    /** 長い記号から順に照合する（短い断片が先に一致しないようにする）。 */
    private val ALL_TOKENS: List<String> = run {
        val tokens = mutableSetOf<String>()
        // 元のトークンを追加
        for ((open, close) in PAIRS) {
            tokens.add(open)
            tokens.add(close)
            // すべての断片（1文字以上）を追加して、壊れた際に確実に削除されるようにする
            for (i in 1..open.length) tokens.add(open.substring(0, i))
            for (i in 0 until open.length) tokens.add(open.substring(i))
            for (i in 1..close.length) tokens.add(close.substring(0, i))
            for (i in 0 until close.length) tokens.add(close.substring(i))
        }
        for (s in SAME_TOKENS) {
            tokens.add(s)
            for (i in 1..s.length) tokens.add(s.substring(0, i))
        }
        // 特殊な断片や記号
        tokens.addAll(listOf("太＊", "＊太", "・＊", "＊・", "quote", "＜・", "・＞"))
        
        tokens.filter { it.isNotEmpty() }
            .distinct()
            .sortedByDescending { it.length }
    }

    /** 内容が空のペア。 */
    private val EMPTY_PAIRS = listOf(
        "****", "~~~~", "```\n\n```", "[]()",
        "＊太＊・＊太＊", "〜〜・〜〜", "＜quote＞＜・quote＞", "｀｀｀・｀｀｀", "色＊赤赤＊色", "＠＠・＠＠"
    )

    /** TextFieldValue 版。カーソル・選択位置もできるだけ保つ。 */
    fun repair(value: TextFieldValue): TextFieldValue {
        val original = value.text
        if (original.isEmpty()) return value

        val start = value.selection.start
        val end = value.selection.end

        // 1) 空のペアを取り除く。
        val (t1, s1, e1) = removeEmptyPairs(original, start, end)
        // 2) 対になっていない記号・断片を取り除く。
        val (t2, s2, e2) = removeBrokenMarkers(t1, s1, e1)

        if (t2 == original && s2 == start && e2 == end) return value
        return TextFieldValue(
            t2,
            TextRange(s2.coerceIn(0, t2.length), e2.coerceIn(0, t2.length))
        )
    }

    private fun removeEmptyPairs(text: String, start: Int, end: Int): Triple<String, Int, Int> {
        var result = text
        var s = start.coerceIn(0, text.length)
        var e = end.coerceIn(0, text.length)
        var changed: Boolean
        do {
            changed = false
            for (pair in EMPTY_PAIRS) {
                val index = result.indexOf(pair)
                if (index < 0) continue
                val pairEnd = index + pair.length
                fun adjust(offset: Int): Int = when {
                    offset <= index -> offset
                    offset >= pairEnd -> offset - pair.length
                    else -> index
                }
                s = adjust(s)
                e = adjust(e)
                result = result.removeRange(index, pairEnd)
                changed = true
            }
        } while (changed)
        return Triple(result, s, e)
    }

    private fun removeBrokenMarkers(text: String, start: Int, end: Int): Triple<String, Int, Int> {
        if (text.isEmpty()) return Triple(text, start, end)
        val protected = findProtectedRanges(text)
        val removed = BooleanArray(text.length)
        var i = 0
        while (i < text.length) {
            val token = ALL_TOKENS.firstOrNull { text.startsWith(it, i) }
            if (token != null && !protected[i]) {
                for (k in i until i + token.length) removed[k] = true
                i += token.length
            } else {
                i += token?.length ?: 1
            }
        }
        if (removed.none { it }) return Triple(text, start, end)

        fun mapIndex(offset: Int): Int {
            val upto = offset.coerceIn(0, text.length)
            var count = 0
            for (k in 0 until upto) if (!removed[k]) count++
            return count
        }
        val builder = StringBuilder(text.length)
        for (k in text.indices) if (!removed[k]) builder.append(text[k])
        val newText = builder.toString()
        return Triple(newText, mapIndex(start), mapIndex(end))
    }

    /**
     * 正しく対になった記号が占める文字位置を true にする。
     * ここに含まれない記号は「壊れている」とみなして取り除く。
     */
    private fun findProtectedRanges(text: String): BooleanArray {
        val protected = BooleanArray(text.length)

        // open と close が異なる記号。
        for ((open, close) in PAIRS) {
            val openStack = ArrayList<Int>()
            var i = 0
            while (i < text.length) {
                if (text.startsWith(close, i)) {
                    if (openStack.isNotEmpty()) {
                        val openPos = openStack.removeAt(openStack.lastIndex)
                        markProtected(protected, openPos, open.length)
                        markProtected(protected, i, close.length)
                    }
                    i += close.length
                } else if (text.startsWith(open, i)) {
                    openStack.add(i)
                    i += open.length
                } else {
                    i += 1
                }
            }
        }

        // open と close が同じ記号（2 個で 1 対）。
        for (token in SAME_TOKENS) {
            val positions = ArrayList<Int>()
            var i = 0
            while (i < text.length) {
                if (text.startsWith(token, i)) {
                    positions.add(i)
                    i += token.length
                } else {
                    i += 1
                }
            }
            var k = 0
            while (k + 1 < positions.size) {
                markProtected(protected, positions[k], token.length)
                markProtected(protected, positions[k + 1], token.length)
                k += 2
            }
        }

        return protected
    }

    private fun markProtected(protected: BooleanArray, from: Int, length: Int) {
        for (k in from until (from + length).coerceAtMost(protected.size)) {
            protected[k] = true
        }
    }
}
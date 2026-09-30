package com.chat.android.core.util

import org.junit.Assert.assertEquals
import org.junit.Assert.assertNotNull
import org.junit.Assert.assertTrue
import org.junit.Test

/**
 * QrEncoder の構造検証。
 *
 * 第三者ライブラリを使わずにQRの正しさを確認するため、
 * 「構造が仕様どおりか」と「データビットを読み戻せるか」を検証する。
 */
class QrEncoderTest {

    private fun finderOk(m: Array<BooleanArray>, x0: Int, y0: Int): Boolean {
        // 7x7 の finder: 外周 black, 内周 white, 中心3x3 black
        for (dy in 0..6) {
            for (dx in 0..6) {
                val expected = when {
                    dx == 0 || dx == 6 || dy == 0 || dy == 6 -> true
                    dx in 2..4 && dy in 2..4 -> true
                    else -> false
                }
                if (m[y0 + dy][x0 + dx] != expected) return false
            }
        }
        return true
    }

    @Test
    fun `Encodes simple text into structurally valid QR`() {
        val qr = QrEncoder.encode("HELLO")
        assertNotNull(qr)
        val size = qr!!.size
        // バージョンの取りうる値
        assertTrue("size should be 21..57 but was $size", size in 21..57)
        assertEquals("size must be 4v+17", 0, (size - 17) % 4)

        // finder 3箇所
        assertTrue(finderOk(qr.modules, 0, 0))
        assertTrue(finderOk(qr.modules, size - 7, 0))
        assertTrue(finderOk(qr.modules, 0, size - 7))

        // タイミングパターン（6行目・6列目は白黒交互）
        for (i in 8 until size - 8) {
            assertEquals("timing row at $i", i % 2 == 0, qr.modules[6][i])
            assertEquals("timing col at $i", i % 2 == 0, qr.modules[i][6])
        }

        // ダークモジュール
        assertTrue("dark module must be set", qr.modules[size - 8][8])
    }

    @Test
    fun `Round trips data bits back to original text`() {
        val texts = listOf(
            "https://chat.quigen.info/profile/J/?code=DtG9LRpT39W7Xcd9",
            "https://chat.quigen.info/",
            "a",
            "0123456789"
        )
        for (text in texts) {
            val decoded = QrEncoder.decodeForTest(text)
            assertEquals("round trip failed for $text", text, decoded)
        }
    }

    @Test
    fun `Rejects text that exceeds capacity`() {
        // バージョン10-L のデータ容量(271バイト)を超える
        val tooLong = "x".repeat(400)
        assertEquals(null, QrEncoder.encode(tooLong))
    }
}

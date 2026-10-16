package com.chat.android.core.util

/**
 * 外部ライブラリなしでQRコードを生成する。
 *
 * 対応範囲:
 *  - モード: バイトモード（URLなどのASCII文字列）
 *  - エラーレベル: L
 *  - バージョン: 1〜10
 */
object QrEncoder {

    class QrCode(val size: Int, val modules: Array<BooleanArray>)

    private val EC_PER_BLOCK = intArrayOf(7, 10, 15, 20, 26, 18, 20, 24, 30, 18)
    private val BLOCK_COUNT = intArrayOf(1, 1, 1, 1, 1, 2, 2, 2, 2, 4)
    private val SHORT_BLOCK_DATA = intArrayOf(19, 34, 55, 80, 108, 68, 78, 97, 116, 68)
    private val LONG_BLOCK_COUNT = intArrayOf(0, 0, 0, 0, 0, 0, 0, 0, 0, 2)
    private val LONG_BLOCK_DATA = intArrayOf(0, 0, 0, 0, 0, 0, 0, 0, 0, 69)

    private val ALIGNMENT_CENTERS = arrayOf(
        intArrayOf(),
        intArrayOf(),
        intArrayOf(6, 18),
        intArrayOf(6, 22),
        intArrayOf(6, 26),
        intArrayOf(6, 30),
        intArrayOf(6, 34),
        intArrayOf(6, 22, 38),
        intArrayOf(6, 24, 42),
        intArrayOf(6, 26, 46),
        intArrayOf(6, 28, 50)
    )

    private val EXP = IntArray(256)
    private val LOG = IntArray(256)

    init {
        var x = 1
        for (i in 0 until 255) {
            EXP[i] = x
            LOG[x] = i
            x = x shl 1
            if (x and 0x100 != 0) x = x xor 0x11D
        }
        EXP[255] = EXP[0]
    }

    private fun gfMul(a: Int, b: Int): Int {
        if (a == 0 || b == 0) return 0
        return EXP[(LOG[a] + LOG[b]) % 255]
    }

    private fun totalDataCodewords(v: Int): Int {
        val idx = v - 1
        return (BLOCK_COUNT[idx] - LONG_BLOCK_COUNT[idx]) * SHORT_BLOCK_DATA[idx] +
                LONG_BLOCK_COUNT[idx] * LONG_BLOCK_DATA[idx]
    }

    fun encode(text: String): QrCode? {
        val bytes = text.toByteArray(Charsets.UTF_8)
        val version = selectVersion(bytes.size) ?: return null
        val dataBits = buildDataBits(bytes, version)
        val allCodewords = interleave(dataBits, version)
        val matrix = buildMatrix(version)
        placeData(matrix, allCodewords)
        applyBestMask(matrix)
        return QrCode(matrix.size, matrix.modules)
    }

    private fun selectVersion(byteCount: Int): Int? {
        for (v in 1..10) {
            val countBits = if (v < 10) 8 else 16
            val needed = 4 + countBits + byteCount * 8 + 4
            if (totalDataCodewords(v) * 8 >= needed) return v
        }
        return null
    }

    private fun buildDataBits(bytes: ByteArray, version: Int): IntArray {
        val bits = ArrayList<Boolean>()
        // Mode: Byte (0100)
        bits.add(false); bits.add(true); bits.add(false); bits.add(false)
        // Length
        val countBits = if (version < 10) 8 else 16
        for (i in countBits - 1 downTo 0) bits.add(((bytes.size shr i) and 1) == 1)
        // Data
        for (b in bytes) {
            for (i in 7 downTo 0) bits.add(((b.toInt() shr i) and 1) == 1)
        }
        // Terminator
        val capacity = totalDataCodewords(version) * 8
        repeat(4) { if (bits.size < capacity) bits.add(false) }
        while (bits.size % 8 != 0) bits.add(false)

        val codewords = IntArray(totalDataCodewords(version))
        for (i in 0 until (bits.size / 8)) {
            var v = 0
            for (j in 0 until 8) if (bits[i * 8 + j]) v = v or (1 shl (7 - j))
            codewords[i] = v
        }
        var pad = 0
        for (i in (bits.size / 8) until codewords.size) {
            codewords[i] = if (pad % 2 == 0) 0xEC else 0x11
            pad++
        }
        return codewords
    }

    private fun reedSolomon(data: IntArray, ecCount: Int): IntArray {
        val gen = IntArray(ecCount)
        var g = intArrayOf(1)
        for (i in 0 until ecCount) {
            val nextG = IntArray(g.size + 1)
            val root = EXP[i]
            for (j in g.indices) {
                nextG[j] = nextG[j] xor gfMul(g[j], root)
                nextG[j + 1] = nextG[j + 1] xor g[j]
            }
            g = nextG
        }
        for (i in 0 until ecCount) gen[i] = g[ecCount - 1 - i]

        val remainder = IntArray(ecCount)
        for (value in data) {
            val factor = value xor remainder[0]
            System.arraycopy(remainder, 1, remainder, 0, ecCount - 1)
            remainder[ecCount - 1] = 0
            for (i in 0 until ecCount) {
                remainder[i] = remainder[i] xor gfMul(gen[i], factor)
            }
        }
        return remainder
    }

    private fun interleave(data: IntArray, version: Int): IntArray {
        val idx = version - 1
        val ecCount = EC_PER_BLOCK[idx]
        val numShort = BLOCK_COUNT[idx] - LONG_BLOCK_COUNT[idx]
        val shortLen = SHORT_BLOCK_DATA[idx]
        val longLen = LONG_BLOCK_DATA[idx]

        val blocks = Array(BLOCK_COUNT[idx]) { i ->
            val dataLen = if (i < numShort) shortLen else longLen
            val offset = if (i < numShort) i * shortLen else numShort * shortLen + (i - numShort) * longLen
            val d = data.copyOfRange(offset, offset + dataLen)
            d + reedSolomon(d, ecCount)
        }

        val result = IntArray(totalDataCodewords(version) + BLOCK_COUNT[idx] * ecCount)
        var p = 0
        val maxData = if (LONG_BLOCK_COUNT[idx] > 0) longLen else shortLen
        for (i in 0 until maxData) {
            for (b in blocks) {
                val bDataLen = b.size - ecCount
                if (i < bDataLen) result[p++] = b[i]
            }
        }
        for (i in 0 until ecCount) {
            for (b in blocks) result[p++] = b[b.size - ecCount + i]
        }
        return result
    }

    private class Matrix(val size: Int) {
        val modules = Array(size) { BooleanArray(size) }
        val reserved = Array(size) { BooleanArray(size) }
        fun set(x: Int, y: Int, dark: Boolean, reserve: Boolean = true) {
            if (x in 0 until size && y in 0 until size) {
                modules[y][x] = dark
                if (reserve) reserved[y][x] = true
            }
        }
    }

    private fun buildMatrix(version: Int): Matrix {
        val size = version * 4 + 17
        val m = Matrix(size)
        placeFinder(m, 0, 0)
        placeFinder(m, size - 7, 0)
        placeFinder(m, 0, size - 7)
        for (i in 8 until size - 8) {
            val dark = i % 2 == 0
            m.set(i, 6, dark)
            m.set(6, i, dark)
        }
        val centers = ALIGNMENT_CENTERS[version]
        for (cy in centers) {
            for (cx in centers) {
                if ((cx <= 8 && cy <= 8) || (cx <= 8 && cy >= size - 9) || (cx >= size - 9 && cy <= 8)) continue
                placeAlignment(m, cx, cy)
            }
        }
        m.set(8, size - 8, true)
        reserveFormatInfo(m)
        return m
    }

    private fun placeFinder(m: Matrix, x0: Int, y0: Int) {
        for (dy in -1..7) {
            for (dx in -1..7) {
                val x = x0 + dx; val y = y0 + dy
                if (x in 0 until m.size && y in 0 until m.size) {
                    val dark = (dx in 0..6 && (dy == 0 || dy == 6)) || (dy in 0..6 && (dx == 0 || dx == 6)) || (dx in 2..4 && dy in 2..4)
                    m.set(x, y, dark)
                }
            }
        }
    }

    private fun placeAlignment(m: Matrix, cx: Int, cy: Int) {
        for (dy in -2..2) {
            for (dx in -2..2) {
                m.set(cx + dx, cy + dy, dx == -2 || dx == 2 || dy == -2 || dy == 2 || (dx == 0 && dy == 0))
            }
        }
    }

    private fun reserveFormatInfo(m: Matrix) {
        for (i in 0..14) {
            val (x, y) = FORMAT_TOP_LEFT[i]
            m.set(x, y, false)
        }
        val size = m.size
        for (i in 0..7) m.set(size - 1 - i, 8, false)
        for (i in 0..7) m.set(8, size - 1 - i, false)
    }

    private fun placeData(m: Matrix, codewords: IntArray) {
        val size = m.size
        var bit = 0; var up = true; var col = size - 1
        while (col > 0) {
            if (col == 6) col--
            for (row in 0 until size) {
                val y = if (up) size - 1 - row else row
                for (c in 0..1) {
                    val x = col - c
                    if (!m.reserved[y][x]) {
                        val dark = if (bit / 8 < codewords.size) (codewords[bit / 8] shr (7 - (bit % 8))) and 1 == 1 else false
                        m.set(x, y, dark, false)
                        bit++
                    }
                }
            }
            up = !up; col -= 2
        }
    }

    private fun applyBestMask(m: Matrix) {
        var bestMask = 0; var minPenalty = Int.MAX_VALUE
        val original = Array(m.size) { m.modules[it].copyOf() }
        for (mask in 0..7) {
            for (y in 0 until m.size) for (x in 0 until m.size) if (!m.reserved[y][x]) if (maskBit(mask, x, y)) m.modules[y][x] = !m.modules[y][x]
            writeFormatInfo(m, mask)
            val penalty = penaltyScore(m)
            if (penalty < minPenalty) { minPenalty = penalty; bestMask = mask }
            for (y in 0 until m.size) m.modules[y] = original[y].copyOf()
        }
        for (y in 0 until m.size) for (x in 0 until m.size) if (!m.reserved[y][x]) if (maskBit(bestMask, x, y)) m.modules[y][x] = !m.modules[y][x]
        writeFormatInfo(m, bestMask)
    }

    private fun maskBit(mask: Int, x: Int, y: Int): Boolean = when (mask) {
        0 -> (x + y) % 2 == 0
        1 -> y % 2 == 0
        2 -> x % 3 == 0
        3 -> (x + y) % 3 == 0
        4 -> (y / 2 + x / 3) % 2 == 0
        5 -> ((x * y) % 2) + ((x * y) % 3) == 0
        6 -> (((x * y) % 2) + ((x * y) % 3)) % 2 == 0
        else -> (((x + y) % 2) + ((x * y) % 3)) % 2 == 0
    }

    private fun penaltyScore(m: Matrix): Int {
        val size = m.size; var p = 0
        for (y in 0 until size) {
            var run = 1
            for (x in 1 until size) {
                if (m.modules[y][x] == m.modules[y][x - 1]) run++ else { if (run >= 5) p += 3 + (run - 5); run = 1 }
            }
            if (run >= 5) p += 3 + (run - 5)
        }
        for (x in 0 until size) {
            var run = 1
            for (y in 1 until size) {
                if (m.modules[y][x] == m.modules[y - 1][x]) run++ else { if (run >= 5) p += 3 + (run - 5); run = 1 }
            }
            if (run >= 5) p += 3 + (run - 5)
        }
        for (y in 0 until size - 1) for (x in 0 until size - 1) {
            val c = m.modules[y][x]
            if (c == m.modules[y][x + 1] && c == m.modules[y + 1][x] && c == m.modules[y + 1][x + 1]) p += 3
        }
        val pat = booleanArrayOf(true, false, true, true, true, false, true, false, false, false, false)
        val rev = booleanArrayOf(false, false, false, false, true, false, true, true, true, false, true)
        for (y in 0 until size) for (x in 0..size - 11) {
            var matchP = true; var matchR = true
            for (i in 0..10) { if (m.modules[y][x + i] != pat[i]) matchP = false; if (m.modules[y][x + i] != rev[i]) matchR = false }
            if (matchP || matchR) p += 40
        }
        for (x in 0 until size) for (y in 0..size - 11) {
            var matchP = true; var matchR = true
            for (i in 0..10) { if (m.modules[y + i][x] != pat[i]) matchP = false; if (m.modules[y + i][x] != rev[i]) matchR = false }
            if (matchP || matchR) p += 40
        }
        var dark = 0; for (y in 0 until size) for (x in 0 until size) if (m.modules[y][x]) dark++
        val ratio = dark * 100 / (size * size)
        p += (kotlin.math.abs(ratio - 50) / 5) * 10
        return p
    }

    private fun writeFormatInfo(m: Matrix, mask: Int) {
        val size = m.size
        val data = (0b01 shl 3) or mask
        var rem = data shl 10
        for (i in 0 until 5) if ((rem shr (14 - i)) and 1 == 1) rem = rem xor (0x537 shl (4 - i))
        val bits = ((data shl 10) or rem) xor 0x5412
        for (i in 0..14) {
            val bit = (bits shr i) and 1 == 1
            val (lx, ly) = FORMAT_TOP_LEFT[14 - i]
            m.set(lx, ly, bit, false)
            if (i < 8) {
                m.set(size - 1 - i, 8, bit, false)
            } else {
                m.set(8, size - 15 + i, bit, false)
            }
        }
        m.set(8, size - 8, true, false)
    }

    private val FORMAT_TOP_LEFT = arrayOf(
        intArrayOf(8, 0), intArrayOf(8, 1), intArrayOf(8, 2), intArrayOf(8, 3), intArrayOf(8, 4), intArrayOf(8, 5), intArrayOf(8, 7), intArrayOf(8, 8),
        intArrayOf(7, 8), intArrayOf(5, 8), intArrayOf(4, 8), intArrayOf(3, 8), intArrayOf(2, 8), intArrayOf(1, 8), intArrayOf(0, 8)
    )

    fun decodeForTest(text: String): String? {
        val qr = encode(text) ?: return null
        val size = qr.size; val version = (size - 17) / 4
        var fi = 0
        for (i in 0..14) {
            val (x, y) = FORMAT_TOP_LEFT[14 - i]
            if (qr.modules[y][x]) fi = fi or (1 shl i)
        }
        val mask = ((fi xor 0x5412) shr 10) and 0b111
        val reserve = buildReservedMap(size, version)
        val bits = ArrayList<Boolean>()
        var up = true; var col = size - 1
        while (col > 0) {
            if (col == 6) col--
            for (row in 0 until size) {
                val y = if (up) size - 1 - row else row
                for (c in 0..1) {
                    val x = col - c
                    if (!reserve[y][x]) {
                        var dark = qr.modules[y][x]
                        if (maskBit(mask, x, y)) dark = !dark
                        bits.add(dark)
                    }
                }
            }
            up = !up; col -= 2
        }
        val codewords = IntArray(bits.size / 8)
        for (i in codewords.indices) {
            var v = 0; for (j in 0 until 8) if (bits[i * 8 + j]) v = v or (1 shl (7 - j))
            codewords[i] = v
        }
        val data = deinterleave(codewords, version) ?: return null
        val full = ArrayList<Boolean>(); for (c in data) for (i in 7 downTo 0) full.add(((c shr i) and 1) == 1)
        val mode = (if (full[0]) 8 else 0) or (if (full[1]) 4 else 0) or (if (full[2]) 2 else 0) or (if (full[3]) 1 else 0)
        if (mode != 4) return null
        val countBits = if (version < 10) 8 else 16
        var len = 0; for (i in 0 until countBits) if (full[4 + i]) len = len or (1 shl (countBits - 1 - i))
        val bytes = ByteArray(len)
        for (i in 0 until len) {
            var v = 0; val start = 4 + countBits + i * 8
            if (start + 8 > full.size) break
            for (j in 0 until 8) if (full[start + j]) v = v or (1 shl (7 - j))
            bytes[i] = v.toByte()
        }
        return String(bytes, Charsets.UTF_8)
    }

    private fun deinterleave(codewords: IntArray, version: Int): IntArray? {
        val idx = version - 1
        val ecCount = EC_PER_BLOCK[idx]
        val numShort = BLOCK_COUNT[idx] - LONG_BLOCK_COUNT[idx]
        val shortLen = SHORT_BLOCK_DATA[idx]
        val longLen = LONG_BLOCK_DATA[idx]
        val data = IntArray(totalDataCodewords(version))

        var inP = 0
        val blocksData = Array(BLOCK_COUNT[idx]) { b ->
            IntArray(if (b < numShort) shortLen else longLen)
        }
        val maxDataLen = if (LONG_BLOCK_COUNT[idx] > 0) longLen else shortLen
        for (i in 0 until maxDataLen) {
            for (b in 0 until BLOCK_COUNT[idx]) {
                if (i < blocksData[b].size) blocksData[b][i] = codewords[inP++]
            }
        }
        var outP = 0
        for (b in 0 until BLOCK_COUNT[idx]) {
            for (i in blocksData[b].indices) {
                if (outP < data.size) data[outP++] = blocksData[b][i]
            }
        }
        return data
    }

    private fun buildReservedMap(size: Int, version: Int): Array<BooleanArray> {
        val m = Matrix(size)
        placeFinder(m, 0, 0); placeFinder(m, size - 7, 0); placeFinder(m, 0, size - 7)
        for (i in 8 until size - 8) { m.set(i, 6, false); m.set(6, i, false) }
        val centers = ALIGNMENT_CENTERS[version]
        for (cy in centers) for (cx in centers) {
            if ((cx <= 8 && cy <= 8) || (cx <= 8 && cy >= size - 9) || (cx >= size - 9 && cy <= 8)) continue
            placeAlignment(m, cx, cy)
        }
        m.set(8, size - 8, true)
        for (i in 0..14) { val (x, y) = FORMAT_TOP_LEFT[i]; m.set(x, y, false) }
        for (i in 0..7) m.set(size - 1 - i, 8, false)
        for (i in 0..7) m.set(8, size - 1 - i, false)
        return m.reserved
    }
}

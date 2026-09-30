package com.chat.android.core.util

/**
 * 外部ライブラリなしでQRコードを生成する。
 *
 * Maven Central に到達できない環境でも招待QRを出せるように、
 * 仕様（JIS X0510 相当）のエンコーダを最小限の実装で自前用意している。
 *
 * 対応範囲:
 *  - モード: バイトモード（URLなどのASCII文字列）
 *  - エラーレベル: L（訂正能力と容量のバランスが最も良い設定）
 *  - バージョン: 1〜10
 *  - マスク: 8パターンからペナルティが最小のものを自動選択
 *
 * 生成結果の正しさは [com.chat.android.core.util.QrEncoderTest] で
 * 構造（ファインダパターン・タイミング・フォーマット情報）と
 * データビットの往復読み戻しで検証している。
 */
object QrEncoder {

    /** 生成されたQRのモジュール配列。 [modules][y][x] が true の=black。 */
    class QrCode(val size: Int, val modules: Array<BooleanArray>)

    // ---- エラーレベル L のブロック構成（バージョン1〜10） ----
    // 1ブロックあたりの誤り訂正コードワード数
    private val EC_PER_BLOCK = intArrayOf(7, 10, 15, 20, 26, 18, 20, 24, 30, 18)
    // ブロック数
    private val BLOCK_COUNT = intArrayOf(1, 1, 1, 1, 1, 2, 2, 2, 2, 4)
    // 短いブロックのデータコードワード数
    private val SHORT_BLOCK_DATA = intArrayOf(19, 34, 55, 80, 108, 68, 78, 97, 116, 68)
    // 長いブロックの個数（バージョン10のみ2個）
    private val LONG_BLOCK_COUNT = intArrayOf(0, 0, 0, 0, 0, 0, 0, 0, 0, 2)
    // 長いブロックのデータコードワード数
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

    /** バージョン v のデータコードワード総数。 */
    private fun totalDataCodewords(v: Int): Int {
        val idx = v - 1
        val shortCount = BLOCK_COUNT[idx] - LONG_BLOCK_COUNT[idx]
        return shortCount * SHORT_BLOCK_DATA[idx] + LONG_BLOCK_COUNT[idx] * LONG_BLOCK_DATA[idx]
    }

    /**
     * [text] をQR化する。容量を超える場合は null。
     */
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

    /** バイト数から最小バージョンを決める。 */
    private fun selectVersion(byteCount: Int): Int? {
        for (v in 1..10) {
            val countBits = if (v < 10) 8 else 16
            // モード4bit + 文字数 + データ8bit*n + 終端4bit
            val needed = 4 + countBits + byteCount * 8 + 4
            if (totalDataCodewords(v) * 8 >= needed) return v
        }
        return null
    }

    /** バイト列をデータコードワード列（誤り訂正前）にする。 */
    private fun buildDataBits(bytes: ByteArray, version: Int): IntArray {
        val bits = ArrayList<Boolean>(totalDataCodewords(version) * 8)

        // モード指示子: バイトモード = 0100
        bits.addAll(listOf(false, true, false, false))
        // 文字数指示子: バージョン1〜9は8bit、10以降は16bit
        val countBits = if (version < 10) 8 else 16
        for (i in countBits - 1 downTo 0) {
            bits.add((bytes.size shr i) and 1 == 1)
        }
        // データ本体
        for (b in bytes) {
            for (i in 7 downTo 0) {
                bits.add((b.toInt() shr i) and 1 == 1)
            }
        }
        // 終端（最大4bit）
        val capacity = totalDataCodewords(version) * 8
        repeat(4) { if (bits.size < capacity) bits.add(false) }
        // 8bit 境界まで埋め
        while (bits.size % 8 != 0) bits.add(false)

        // パディング（交互に 0xEC / 0x11）
        // ビット列を MSB ファーストで8bitずつコードワードに詰める。
        val codewords = IntArray(totalDataCodewords(version))
        for (codewordIndex in 0 until totalDataCodewords(version)) {
            var value = 0
            for (bit in 0 until 8) {
                val bitIndex = codewordIndex * 8 + bit
                val isSet = bitIndex < bits.size && bits[bitIndex]
                if (isSet) value = value or (1 shl (7 - bit))
            }
            codewords[codewordIndex] = value
        }
        // 末尾の余ったビット列にパディングバイトを詰める
        var padIndex = 0
        var bi = bits.size / 8
        while (bi < codewords.size) {
            codewords[bi] = if (padIndex % 2 == 0) 0xEC else 0x11
            padIndex++
            bi++
        }
        return codewords
    }

    /**
     * GF(256) 上での剰余。QRの原始多項式は x^8+x^4+x^3+x^2+1 (0x11D)。
     */
    private fun gfMul(a: Int, b: Int): Int {
        var a0 = a
        var b0 = b
        var result = 0
        var i = 7
        while (i >= 0) {
            result = result shl 1
            if ((b0 shr i) and 1 == 1) result = result xor 0x11D
            a0 = a0 shl 1
            if (i != 0) {
                // 桁上がりで 0x100 が出たときは必ず 1 なので常にフィードバックする
                if (a0 shr 8 == 1) a0 = a0 xor 0x11D
            }
            i--
        }
        return result
    }

    /** Reed-Solomon の誤り訂正コードワードを生成する。 */
    private fun reedSolomon(data: IntArray, ecCount: Int): IntArray {
        // 発生多項式 (x-α^0)(x-α^1)...(x-α^(ecCount-1))
        val generator = IntArray(ecCount + 1) { if (it == 0 || it == ecCount) 1 else 0 }
        for (i in 0 until ecCount) {
            val root = gfExp(i)
            for (j in ecCount downTo 1) {
                generator[j] = generator[j - 1]
            }
            generator[0] = 0
            for (j in ecCount downTo 1) {
                generator[j] = generator[j] xor gfMul(generator[j - 1], root)
            }
        }

        val result = IntArray(ecCount)
        for (d in data) {
            val factor = d xor result[ecCount - 1]
            for (j in ecCount - 1 downTo 1) {
                result[j] = result[j - 1] xor gfMul(generator[j], factor)
            }
            result[0] = gfMul(generator[0], factor)
        }
        return result
    }

    // GF(256) の α^i を返す
    private fun gfExp(i: Int): Int {
        var value = 1
        repeat(i) { value = gfMul(value, 2) }
        return value
    }

    /**
     * データブロックに分割し、誤り訂正を計算、数据/誤り訂正を交互に並べる。
     */
    private fun interleave(data: IntArray, version: Int): IntArray {
        val idx = version - 1
        val ecPerBlock = EC_PER_BLOCK[idx]
        val shortCount = BLOCK_COUNT[idx] - LONG_BLOCK_COUNT[idx]
        val blocks = ArrayList<IntArray>()

        var offset = 0
        for (i in 0 until shortCount) {
            val block = data.copyOfRange(offset, offset + SHORT_BLOCK_DATA[idx])
            offset += SHORT_BLOCK_DATA[idx]
            blocks.add(block + reedSolomon(block, ecPerBlock))
        }
        for (i in 0 until LONG_BLOCK_COUNT[idx]) {
            val block = data.copyOfRange(offset, offset + LONG_BLOCK_DATA[idx])
            offset += LONG_BLOCK_DATA[idx]
            blocks.add(block + reedSolomon(block, ecPerBlock))
        }

        val result = IntArray(blocks.size * (blocks[0].size))
        var outIndex = 0
        val maxData = blocks.maxOf { dataCountOf(it.size, ecPerBlock) }
        // データ部を先に、誤り訂正部を後に、各ブロック内で交差させて並べる。
        for (i in 0 until maxData) {
            for (block in blocks) {
                val dataLen = dataCountOf(block.size, ecPerBlock)
                if (i < dataLen) result[outIndex++] = block[i]
            }
        }
        for (i in 0 until ecPerBlock) {
            for (block in blocks) {
                val dataLen = dataCountOf(block.size, ecPerBlock)
                result[outIndex++] = block[dataLen + i]
            }
        }
        return result
    }

    private fun dataCountOf(blockSize: Int, ecPerBlock: Int): Int = blockSize - ecPerBlock

    /** 機能パターン（ファインダ・アライメント・タイミング・形式情報）を配置した空の行列を作る。 */
    private class Matrix(val size: Int) {
        val modules = Array(size) { BooleanArray(size) }
        val reserved = Array(size) { BooleanArray(size) }
        val version: Int

        init {
            version = (size - 17) / 4
        }

        fun set(x: Int, y: Int, dark: Boolean, reserve: Boolean = true) {
            modules[y][x] = dark
            if (reserve) reserved[y][x] = true
        }
    }

    private fun buildMatrix(version: Int): Matrix {
        val size = version * 4 + 17
        val m = Matrix(size)

        // ファインダパターン（3箇所）＋ 区切り
        placeFinder(m, 0, 0)
        placeFinder(m, size - 7, 0)
        placeFinder(m, 0, size - 7)

        // タイミングパターン（形式情報と重ならない範囲だけ）
        for (i in 8 until size - 8) {
            val dark = i % 2 == 0
            m.set(i, 6, dark)
            m.set(6, i, dark)
        }

        // アライメントパターン
        val centers = ALIGNMENT_CENTERS[version]
        for (cy in centers) {
            for (cx in centers) {
                // ファインダと重なる位置は飛ばす
                val nearFinder = (cx <= 8 && cy <= 8) ||
                    (cx <= 8 && cy >= size - 9) ||
                    (cx >= size - 9 && cy <= 8)
                if (nearFinder) continue
                placeAlignment(m, cx, cy)
            }
        }

        // ダークモジュール（形式情報の下）
        m.set(8, size - 8, true)

        // 形式情報領域を予約（実際のビットはマスク確定後に書く）
        reserveFormatInfo(m)

        return m
    }

    private fun placeFinder(m: Matrix, x0: Int, y0: Int) {
        for (dy in -1..7) {
            for (dx in -1..7) {
                val x = x0 + dx
                val y = y0 + dy
                if (x < 0 || y < 0 || x >= m.size || y >= m.size) continue
                val inRing = (dx in 0..6 && (dy == 0 || dy == 6)) ||
                    (dy in 0..6 && (dx == 0 || dx == 6)) ||
                    (dx in 2..4 && dy in 2..4)
                m.set(x, y, inRing)
            }
        }
    }

    private fun placeAlignment(m: Matrix, cx: Int, cy: Int) {
        for (dy in -2..2) {
            for (dx in -2..2) {
                val dark = (dx == -2 || dx == 2 || dy == -2 || dy == 2) ||
                    (dx == 0 && dy == 0)
                m.set(cx + dx, cy + dy, dark)
            }
        }
    }

    private fun reserveFormatInfo(m: Matrix) {
        val size = m.size
        // 左上
        for (i in 0..14) {
            val (x, y) = FORMAT_TOP_LEFT[i]
            m.set(x, y, false)
        }
        // 右上（水平方向）
        for (i in 0..7) m.set(size - 1 - i, 8, false)
        // 左下（垂直方向）
        for (i in 0..7) m.set(8, size - 1 - i, false)
    }

    /** コードワードをジグザグに配置する。 */
    private fun placeData(m: Matrix, codewords: IntArray) {
        val size = m.size
        var bitIndex = 0
        var upward = true

        var col = size - 1
        while (col > 0) {
            if (col == 6) col--  // タイミングパターンの列は飛ばす
            for (step in 0 until size) {
                val y = if (upward) size - 1 - step else step
                for (c in 0..1) {
                    val x = col - c
                    if (m.reserved[y][x]) continue
                    val codewordIndex = bitIndex / 8
                    val bitInByte = 7 - (bitIndex % 8)
                    // 容量を超える残りのビットは 0 のままにする
                    val dark = if (codewordIndex < codewords.size) {
                        (codewords[codewordIndex] shr bitInByte) and 1 == 1
                    } else {
                        false
                    }
                    m.set(x, y, dark, reserve = false)
                    bitIndex++
                }
            }
            upward = !upward
            col -= 2
        }
    }

    /** 8パターンからペナルティが最小のマスクを選び、形式情報を書き込む。 */
    private fun applyBestMask(m: Matrix) {
        var bestMask = 0
        var bestPenalty = Int.MAX_VALUE
        val original = Array(m.size) { m.modules[it].copyOf() }

        for (mask in 0..7) {
            // データ領域にマスクを適用
            for (y in 0 until m.size) {
                for (x in 0 until m.size) {
                    if (m.reserved[y][x]) continue
                    if (maskBit(mask, x, y)) m.modules[y][x] = !m.modules[y][x]
                }
            }
            val penalty = penaltyScore(m)
            if (penalty < bestPenalty) {
                bestPenalty = penalty
                bestMask = mask
            }
            // 元に戻す
            for (y in 0 until m.size) {
                m.modules[y] = original[y].copyOf()
            }
        }

        // 採用したマスクを適用
        for (y in 0 until m.size) {
            for (x in 0 until m.size) {
                if (m.reserved[y][x]) continue
                if (maskBit(bestMask, x, y)) m.modules[y][x] = !m.modules[y][x]
            }
        }
        writeFormatInfo(m, bestMask)
    }

    /** マスクパターンの判定関数。 */
    private fun maskBit(mask: Int, x: Int, y: Int): Boolean = when (mask) {
        0 -> (x + y) % 2 == 0
        1 -> y % 2 == 0
        2 -> x % 3 == 0
        3 -> (x + y) % 3 == 0
        4 -> (x / 3 + y / 2) % 2 == 0
        5 -> (x * y) % 2 + (x * y) % 3 == 0
        6 -> ((x * y) % 2 + (x * y) % 3) % 2 == 0
        else -> ((x + y) % 2 + (x * y) % 3) % 2 == 0
    }

    /** 4つのルールでペナルティを計算する。 */
    private fun penaltyScore(m: Matrix): Int {
        val size = m.size
        var penalty = 0

        // ルール1: 同じ色が5つ以上連続
        for (y in 0 until size) {
            var run = 1
            for (x in 1 until size) {
                if (m.modules[y][x] == m.modules[y][x - 1]) {
                    run++
                } else {
                    if (run >= 5) penalty += 3 + (run - 5)
                    run = 1
                }
            }
            if (run >= 5) penalty += 3 + (run - 5)
        }
        for (x in 0 until size) {
            var run = 1
            for (y in 1 until size) {
                if (m.modules[y][x] == m.modules[y - 1][x]) {
                    run++
                } else {
                    if (run >= 5) penalty += 3 + (run - 5)
                    run = 1
                }
            }
            if (run >= 5) penalty += 3 + (run - 5)
        }

        // ルール2: 2x2の同色ブロック
        for (y in 0 until size - 1) {
            for (x in 0 until size - 1) {
                val c = m.modules[y][x]
                if (c == m.modules[y][x + 1] && c == m.modules[y + 1][x] &&
                    c == m.modules[y + 1][x + 1]
                ) {
                    penalty += 3
                }
            }
        }

        // ルール3: 1:1:3:1:1 に似た finder パターン
        val pattern = booleanArrayOf(true, false, true, true, true, false, true, false, false, false, false)
        val reversed = BooleanArray(pattern.size) { pattern[pattern.size - 1 - it] }
        for (y in 0 until size) {
            for (x in 0 until size - pattern.size) {
                if (matchesAt(m, x, y, pattern) || matchesAt(m, x, y, reversed)) {
                    penalty += 40
                }
            }
        }
        for (x in 0 until size) {
            for (y in 0 until size - pattern.size) {
                if (matchesVertical(m, x, y, pattern) || matchesVertical(m, x, y, reversed)) {
                    penalty += 40
                }
            }
        }

        // ルール4: 黒比率の偏り
        var dark = 0
        for (y in 0 until size) for (x in 0 until size) if (m.modules[y][x]) dark++
        val ratio = dark * 100 / (size * size)
        penalty += kotlin.math.abs(ratio - 50) / 5 * 10

        return penalty
    }

    private fun matchesAt(m: Matrix, x0: Int, y0: Int, p: BooleanArray): Boolean {
        for (i in p.indices) if (m.modules[y0][x0 + i] != p[i]) return false
        return true
    }

    private fun matchesVertical(m: Matrix, x0: Int, y0: Int, p: BooleanArray): Boolean {
        for (i in p.indices) if (m.modules[y0 + i][x0] != p[i]) return false
        return true
    }

    /**
     * 形式情報（エラーレベル L + マスク番号）を書き込む。
     *
     * L = 01, BCH(15,5) + マスクパターン 0x5412。
     * 座標は仕様どおり、左上と左下・右上の2か所へ同じ15ビットを置く。
     */
    private fun writeFormatInfo(m: Matrix, mask: Int) {
        val size = m.size
        val ecBits = 0b01  // L
        val data = (ecBits shl 3) or mask

        // BCH(15,5): data<<10 を発生多項式 0x537 で割った余り。
        // 0x537 の最高次は x^10 なので、bit10 が立ったときに還元する。
        var rem = data
        for (i in 0..9) {
            rem = rem shl 1
            if (rem and 0x400 != 0) rem = rem xor 0x537
        }
        val bits = ((data shl 10) or rem) xor 0x5412

        for (i in 0..14) {
            val bit = (bits shr i) and 1 == 1
            // 左上（bit0 から順に）
            val (lx, ly) = FORMAT_TOP_LEFT[i]
            m.set(lx, ly, bit, reserve = false)
            // 右上（水平方向）: x = size-1-i, y = 8
            if (FORMAT_COPY_TOP[i] >= 0) {
                m.set(size - 1 - i, 8, bit, reserve = false)
            }
            // 左下（垂直方向）: x = 8, y = size-1-i
            if (FORMAT_COPY_BOTTOM[i] >= 0) {
                m.set(8, size - 1 - i, bit, reserve = false)
            }
        }
        // ダークモジュール
        m.set(8, size - 8, true, reserve = false)
    }

    // 形式情報の左上配置（bit index -> x, y）
    private val FORMAT_TOP_LEFT = arrayOf(
        intArrayOf(8, 0), intArrayOf(8, 1), intArrayOf(8, 2), intArrayOf(8, 3),
        intArrayOf(8, 4), intArrayOf(8, 5), intArrayOf(8, 7), intArrayOf(8, 8),
        intArrayOf(7, 8), intArrayOf(5, 8), intArrayOf(4, 8), intArrayOf(3, 8),
        intArrayOf(2, 8), intArrayOf(1, 8), intArrayOf(0, 8)
    )

    // 右上の水平方向のコピー（bit0..7）
    private val FORMAT_COPY_TOP = intArrayOf(0, 1, 2, 3, 4, 5, 6, 7, -1, -1, -1, -1, -1, -1, -1)

    // 左下の垂直方向のコピー（bit0..7）
    private val FORMAT_COPY_BOTTOM = intArrayOf(-1, -1, -1, -1, -1, -1, -1, -1, 0, 1, 2, 3, 4, 5, 6)

    /**
     * テスト専用。[text] をQR化し、逆順に読み戻して元文字列が復元できるか確認する。
     * マスク・形式情報・機能パターンを剥がしてデータビット列を取り出し直す。
     */
    fun decodeForTest(text: String): String? {
        val qr = encode(text) ?: return null
        val size = qr.size
        val version = (size - 17) / 4
        val mask = readMaskFromFormatInfo(qr.modules)

        // 機能パターンを再構築して予約判定を再現する
        val reserve = buildReservedMap(size, version)

        // データビットを読み戻す（placeData と逆順）
        val bits = ArrayList<Boolean>()
        var upward = true
        var col = size - 1
        while (col > 0) {
            if (col == 6) col--
            for (step in 0 until size) {
                val y = if (upward) size - 1 - step else step
                for (c in 0..1) {
                    val x = col - c
                    if (reserve[y][x]) continue
                    var dark = qr.modules[y][x]
                    if (maskBit(mask, x, y)) dark = !dark
                    bits.add(dark)
                }
            }
            upward = !upward
            col -= 2
        }

        // コードワード化
        val codewords = IntArray(bits.size / 8)
        for (i in codewords.indices) {
            var v = 0
            for (b in 0 until 8) {
                if (bits[i * 8 + b]) v = v or (1 shl (7 - b))
            }
            codewords[i] = v
        }

        // ブロックからデータ部だけを取り戻す
        val idx = version - 1
        val shortCount = BLOCK_COUNT[idx] - LONG_BLOCK_COUNT[idx]
        val data = IntArray(totalDataCodewords(version))
        var out = 0
        val maxData = SHORT_BLOCK_DATA[idx]
        for (i in 0 until maxData) {
            for (b in 0 until BLOCK_COUNT[idx]) {
                val blockDataLen =
                    if (b < shortCount) SHORT_BLOCK_DATA[idx] else LONG_BLOCK_DATA[idx]
                if (i >= blockDataLen) continue
                if (out < data.size) data[out++] = codewords[i * BLOCK_COUNT[idx] + b]
            }
        }
        if (out < data.size) return null

        // モード指示子と文字数を読む
        val full = ArrayList<Boolean>(data.size * 8)
        for (c in data) for (i in 7 downTo 0) full.add((c shr i) and 1 == 1)

        val mode = listOf(full[0], full[1], full[2], full[3])
        if (mode != listOf(false, true, false, false)) return null
        val countBits = if (version < 10) 8 else 16
        var length = 0
        for (i in 4 until 4 + countBits) {
            length = (length shl 1) or (if (full[i]) 1 else 0)
        }
        val bytes = ArrayList<Byte>()
        var p = 4 + countBits
        for (i in 0 until length) {
            if (p + 8 > full.size) return null
            var v = 0
            for (b in 0 until 8) v = (v shl 1) or (if (full[p + b]) 1 else 0)
            bytes.add(v.toByte())
            p += 8
        }
        return String(bytes.toByteArray(), Charsets.UTF_8)
    }

    /** 形式情報（左上）からマスク番号を読み戻す。 */
    private fun readMaskFromFormatInfo(m: Array<BooleanArray>): Int {
        var value = 0
        for (i in 0..14) {
            val (x, y) = FORMAT_TOP_LEFT[i]
            if (m[y][x]) value = value or (1 shl i)
        }
        val unmasked = value xor 0x5412
        return (unmasked shr 10) and 0b111
    }

    /** 機能パターン部分の予約マップを再構築する。 */
    private fun buildReservedMap(size: Int, version: Int): Array<BooleanArray> {
        val m = Matrix(size)
        placeFinder(m, 0, 0)
        placeFinder(m, size - 7, 0)
        placeFinder(m, 0, size - 7)
        for (i in 8 until size - 8) {
            m.set(i, 6, i % 2 == 0)
            m.set(6, i, i % 2 == 0)
        }
        val centers = ALIGNMENT_CENTERS[version]
        for (cy in centers) {
            for (cx in centers) {
                val nearFinder = (cx <= 8 && cy <= 8) ||
                    (cx <= 8 && cy >= size - 9) ||
                    (cx >= size - 9 && cy <= 8)
                if (nearFinder) continue
                placeAlignment(m, cx, cy)
            }
        }
        m.set(8, size - 8, true)
        reserveFormatInfo(m)
        return m.reserved
    }
}
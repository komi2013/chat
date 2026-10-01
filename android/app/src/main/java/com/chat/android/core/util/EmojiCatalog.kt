package com.chat.android.core.util

import android.content.Context
import org.json.JSONArray

/**
 * メッセージへの絵文字リアクションと、絵文字ピッカーの一覧
 * （vue/src/my/emoji.js に対応）。
 *
 * サーバーは絵文字を文字列で受け取る。/img/... で始まる値は画像として扱う。
 */
object EmojiCatalog {

    private val PREFS = "chat_prefs"

    /** vue/src/my/emoji.js の emojiRanges と同じ範囲。 */
    private val EMOJI_RANGES = listOf(
        0x1F300 to 0x1F5FF,
        0x1F600 to 0x1F64F,
        0x1F680 to 0x1F6FF,
        0x1F900 to 0x1F9FF,
        0x1FA70 to 0x1FAFF,
        0x2600 to 0x26FF,
        0x2702 to 0x27B0
    )

    /** 初期一覧。vue/src/my/emoji.js の masterEmojis と同じ。 */
    private val DEFAULT_EMOJIS = listOf(
        "🙇", "😁", "🤔", "😂", "🤣", "😱", "😭", "😅", "👍", "👌",
        "/img/arigatou.png", "/img/kakunin.png", "/img/odaijini.png", "/img/soudesune.png",
        "/img/naruhodo.png", "/img/shouchi.png",
        "👎", "👏", "💪", "🤝", "✅", "☑️", "🎉", "💖", "🔥", "🎶",
        "😜", "😋", "😇", "😊", "😎", "🥰", "🤩", "🚫"
    )

    /** 直近使った絵文字を先頭に並べた一覧を読み込む。 */
    fun masterEmojis(context: Context): List<String> {
        val prefs = context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
        val stored = prefs.getString("emojis", null)
        if (stored.isNullOrBlank()) return DEFAULT_EMOJIS
        return runCatching {
            val array = JSONArray(stored)
            (0 until array.length()).mapNotNull { array.optString(it, "").ifBlank { null } }
        }.getOrNull()?.takeIf { it.isNotEmpty() } ?: DEFAULT_EMOJIS
    }

    /**
     * 直近使った絵文字を先頭へ移動して保存する（vue の rotateEmoji と同じ）。
     * 画像は末尾、普通絵文字は先頭35件までに整理する。
     */
    fun rotateEmoji(context: Context, emoji: String) {
        val current = masterEmojis(context).filter { it != emoji }.toMutableList()
        current.add(0, emoji)
        val imageEmojis = current.filter { it.startsWith("/img/") }
        val normalEmojis = current.filter { !it.startsWith("/img/") }.take(35)
        val rotated = normalEmojis + imageEmojis

        context.getSharedPreferences(PREFS, Context.MODE_PRIVATE)
            .edit()
            .putString("emojis", JSONArray(rotated).toString())
            .apply()
    }

    /** /img/... のような画像パスかどうか。 */
    fun isImageEmoji(value: String): Boolean =
        Regex("^/[\\w.-]+(/[\\w.-]+)*$").matches(value)

    fun isEmojiInRange(value: String): Boolean {
        if (value.isEmpty()) return false
        val codePoint = value.codePointAt(0)
        return EMOJI_RANGES.any { codePoint in it.first..it.second }
    }

    /** vue の validateEmoji と同じ。絵文字でないなら1文字まで許可する。 */
    fun validateEmoji(value: String): Boolean {
        if (value.isEmpty()) return false
        return if (!isEmojiInRange(value) && value.length > 1) false else true
    }

    /** 集計結果（vue の calcEmoji 相当）。 */
    data class EmojiCount(
        val emoji: String,
        val count: Int,
        val selected: Boolean
    )

    /**
     * リアクションを絵文字ごとにまとめる（vue の calcEmoji と同じ）。
     * 同じエイリアスが同じ絵文字を2回リアクションしている場合は1件にまとめる。
     */
    fun calcEmojis(emojisJson: String, myAliasName: String): List<EmojiCount> {
        val array = runCatching { JSONArray(emojisJson) }.getOrNull() ?: return emptyList()
        if (array.length() == 0) return emptyList()

        val order = mutableListOf<String>()
        val counts = mutableMapOf<String, Int>()
        val selected = mutableMapOf<String, Boolean>()

        for (i in 0 until array.length()) {
            val obj = array.optJSONObject(i) ?: continue
            val value = obj.optString("emoji", "")
            val aliasName = obj.optString("aliasName", "")
            if (value.isEmpty()) continue
            if (!counts.containsKey(value)) {
                order.add(value)
                counts[value] = 0
                selected[value] = false
            }
            counts[value] = (counts[value] ?: 0) + 1
            if (aliasName == myAliasName) selected[value] = true
        }

        return order.map { EmojiCount(it, counts[it] ?: 0, selected[it] ?: false) }
    }
}
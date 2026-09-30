package com.chat.android.core.util

import kotlin.random.Random

/**
 * アイコン用のランダム絵文字と背景色を生成する（vue/src/my/emoji.js 相当）。
 *
 * サーバーは ",<絵文字>,#RRGGBB" 形式しか受け付けない。
 * common.EmojiImgValid の判定は `^,.{1,2},#[0-9a-fA-F]{6}$` で、
 * Go の `.` は1符文（rune）に一致するため、絵文字は1文字分に収まる必要がある。
 */
object RandomAvatar {

    // 1符文で完結する絵文字のみを使う（結合絵文字は2符文になるため使わない）
    val EMOJI_PRESETS = listOf(
        "🙂", "😀", "😆", "😎", "🥳", "🤔",
        "🐱", "🐶", "🦊", "🐼", "🐨", "🦁",
        "🌸", "🌟", "🍀", "🍎", "⚽", "🎵",
        "🚀", "🔥", "💡", "📌", "✅", "🎉"
    )

    val COLOR_PRESETS = listOf(
        "#E53935", "#8E24AA", "#3949AB", "#00897B",
        "#F9A825", "#FB8C00", "#6D4C41", "#546E7A"
    )

    /** ランダムな ",絵文字,#色" を返す。 */
    fun random(): String {
        val emoji = EMOJI_PRESETS.random()
        val color = COLOR_PRESETS.random()
        return ",$emoji,$color"
    }

    /** ランダムな絵文字を1つ返す。 */
    fun randomEmoji(): String = EMOJI_PRESETS.random()

    /** ランダムな背景色を #RRGGBB で返す。 */
    fun randomColor(): String {
        fun byte() = Random.nextInt(0, 256)
        return "#%02X%02X%02X".format(byte(), byte(), byte())
    }
}

package com.chat.android.core.util

/**
 * messageID の末尾1文字を除く前半が Base62 エンコードされた unix 時刻。
 * vue の base62Decode(messageID.slice(0, -1)) と同じ。
 */
object Base62 {
    private const val CHARS = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

    fun decode(value: String): Long {
        var result = 0L
        for (c in value) {
            val index = CHARS.indexOf(c)
            if (index < 0) continue
            result = result * 62 + index
        }
        return result
    }

    fun encode(num: Long): String {
        if (num <= 0) return "0"
        val sb = StringBuilder()
        var n = num
        while (n > 0) {
            sb.append(CHARS[(n % 62).toInt()])
            n /= 62
        }
        return sb.reverse().toString()
    }
}

/** unix 秒を vue の timeFormat('YYYY/MM/DD hh:mm:ss') と同じ書式にする。 */
fun formatThreadTime(unixSeconds: Long): String {
    val cal = java.util.Calendar.getInstance()
    cal.timeInMillis = unixSeconds * 1000
    return String.format(
        "%04d/%02d/%02d %02d:%02d:%02d",
        cal.get(java.util.Calendar.YEAR),
        cal.get(java.util.Calendar.MONTH) + 1,
        cal.get(java.util.Calendar.DAY_OF_MONTH),
        cal.get(java.util.Calendar.HOUR_OF_DAY),
        cal.get(java.util.Calendar.MINUTE),
        cal.get(java.util.Calendar.SECOND)
    )
}

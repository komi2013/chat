package com.chat.android.core.push.handlers

import android.content.ContentValues
import android.database.sqlite.SQLiteDatabase
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.timestamp.TimestampDbHelper

class TimestampCodeHandler(context: android.content.Context) : PushHandler {
    private val dbHelper = TimestampDbHelper(context)

    override suspend fun handle(pd: PushData) {
        val arr = pd.getContentsAsArray() ?: return
        val db = dbHelper.writableDatabase
        for (i in 0 until arr.length()) {
            val obj = arr.optJSONObject(i) ?: continue
            val code = obj.optString("code")
            val isDel = obj.optBoolean("isDel", false)
            if (isDel) {
                db.delete("timestampCode", "code=?", arrayOf(code))
            } else {
                val values = ContentValues().apply {
                    put("code", code)
                    put("channelID", obj.optString("channelID"))
                    put("adminGroup", obj.optString("adminGroup"))
                    put("planDate", obj.optString("planDate"))
                    put("until", obj.optInt("until"))
                    put("isDel", 0)
                }
                db.insertWithOnConflict("timestampCode", null, values, SQLiteDatabase.CONFLICT_REPLACE)
            }
        }
    }
}

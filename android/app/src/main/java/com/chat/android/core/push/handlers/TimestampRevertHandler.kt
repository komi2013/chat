package com.chat.android.core.push.handlers

import android.content.ContentValues
import android.database.sqlite.SQLiteDatabase
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.timestamp.TimestampDbHelper

class TimestampRevertHandler(context: android.content.Context) : PushHandler {
    private val dbHelper = TimestampDbHelper(context)

    override suspend fun handle(pd: PushData) {
        val arr = pd.getContentsAsArray() ?: return
        val action = arr.optString(1)
        val now = arr.optString(2)
        val aliasName = arr.optString(3)
        val timestampID = now + aliasName
        
        val db = dbHelper.writableDatabase
        if (action == "startWork") {
            db.delete("timestamp", "timestampID=?", arrayOf(timestampID))
        } else if (action == "endWork") {
            val values = ContentValues().apply { putNull("timeOut") }
            db.update("timestamp", values, "timestampID=?", arrayOf(timestampID))
        }
    }
}

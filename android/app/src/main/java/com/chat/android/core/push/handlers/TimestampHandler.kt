package com.chat.android.core.push.handlers

import android.content.ContentValues
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.timestamp.TimestampDbHelper

class TimestampHandler(context: android.content.Context, private val logDb: SQLiteOpenHelper) : PushHandler {
    private val dbHelper = TimestampDbHelper(context)

    override suspend fun handle(pd: PushData) {
        val arr = pd.getContentsAsArray() ?: return
        if (arr.length() < 4) return
        val code = arr.optString(0)
        val action = arr.optString(1)
        val now = arr.optString(2)
        val aliasName = arr.optString(3)
        val channelID = pd.channelID

        val db = dbHelper.writableDatabase
        
        if (action == "startWork") {
            val values = ContentValues().apply {
                put("timestampID", now + aliasName)
                put("channelID", channelID)
                put("aliasName", aliasName)
                put("timeIn", now)
            }
            db.insertWithOnConflict("timestamp", null, values, SQLiteDatabase.CONFLICT_REPLACE)
        } else if (action == "endWork") {
            // Need to get latest timestamp for this user and add timeOut
            val c = db.query("timestamp", arrayOf("timestampID"), "channelID=? AND aliasName=?", arrayOf(channelID, aliasName), null, null, "timeIn DESC", "1")
            if (c.moveToFirst()) {
                val tid = c.getString(0)
                val values = ContentValues().apply { put("timeOut", now) }
                db.update("timestamp", values, "timestampID=?", arrayOf(tid))
            }
            c.close()
        } else if (action == "startBreak" || action == "endBreak") {
            // simplified: get latest, parse breaks json, append/update, save
        }
    }
}

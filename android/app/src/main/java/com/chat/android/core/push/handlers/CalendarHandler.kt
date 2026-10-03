package com.chat.android.core.push.handlers

import android.content.ContentValues
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.calendar.CalendarDbHelper
import com.chat.android.feature.calendar.CalendarItem
import org.json.JSONArray
import org.json.JSONObject

class CalendarHandler(context: android.content.Context, private val logDb: SQLiteOpenHelper) : PushHandler {
    private val calendars = CalendarDbHelper(context)

    override suspend fun handle(pd: PushData) {
        val data = pd.getContentsAsObject() ?: return
        val id = data.optString("calendarID", "")
        val channelID = data.optString("channelID", "")
        val aliasName = data.optString("aliasName", "")
        if (id.isBlank() || channelID != pd.channelID || aliasName != pd.updatedBy) return
        val previous = calendars.get(id)
        if (data.optInt("delete", 0) != 0 || data.optBoolean("delete", false)) {
            calendars.delete(id)
        } else {
            val attendees = data.optJSONArray("aliasNames")?.toString()
                ?: data.optString("aliasNames", "[]").takeIf { it.startsWith("[") } ?: "[]"
            calendars.save(CalendarItem(
                calendarID = id,
                channelID = channelID,
                aliasName = aliasName,
                timeStart = data.optString("timeStart", ""),
                timeEnd = data.optString("timeEnd", ""),
                todo = data.optString("todo", ""),
                aliasNamesJson = attendees,
                repeatOption = data.optString("repeatOption", "none")
            ))
        }
        if (previous != null) {
            val values = ContentValues().apply {
                put("logID", "${pd.title}${pd.channelID}${pd.updatedBy}${pd.pushID}")
                put("pushID", pd.pushID); put("pushTitle", pd.title); put("channelID", pd.channelID)
                put("updatedBy", pd.updatedBy); put("updatedAt", System.currentTimeMillis())
                put("preContents", previous.toJson().toString())
            }
            runCatching { logDb.writableDatabase.insertWithOnConflict("log", null, values, SQLiteDatabase.CONFLICT_REPLACE) }
                .onFailure { Log.w("CalendarHandler", "Could not log previous calendar event", it) }
        }
    }
}

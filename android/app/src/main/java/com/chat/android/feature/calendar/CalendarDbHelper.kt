package com.chat.android.feature.calendar

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import org.json.JSONObject

data class CalendarItem(
    val calendarID: String,
    val channelID: String,
    val aliasName: String,
    val timeStart: String,
    val timeEnd: String,
    val todo: String,
    val aliasNamesJson: String,
    val repeatOption: String = "none"
) {
    val title: String get() = todo.take(10)
    fun toJson(): JSONObject = JSONObject().apply {
        put("calendarID", calendarID); put("channelID", channelID); put("aliasName", aliasName)
        put("timeStart", timeStart); put("timeEnd", timeEnd); put("todo", todo)
        put("aliasNames", org.json.JSONArray(aliasNamesJson)); put("repeatOption", repeatOption)
    }
}

class CalendarDbHelper(context: Context) : SQLiteOpenHelper(context, "calendar.db", null, 1) {
    override fun onCreate(db: SQLiteDatabase) {
        db.execSQL("CREATE TABLE calendar (calendarID TEXT PRIMARY KEY, channelID TEXT, aliasName TEXT, timeStart TEXT, timeEnd TEXT, todo TEXT, aliasNames TEXT, repeatOption TEXT)")
    }
    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) = Unit

    fun all(): List<CalendarItem> = readableDatabase.query("calendar", null, null, null, null, null, "timeStart ASC").use { c ->
        buildList { while (c.moveToNext()) add(CalendarItem(c.getString(0), c.getString(1), c.getString(2), c.getString(3), c.getString(4), c.getString(5), c.getString(6), c.getString(7))) }
    }
    fun get(id: String): CalendarItem? = readableDatabase.query("calendar", null, "calendarID=?", arrayOf(id), null, null, null).use { c ->
        if (!c.moveToFirst()) null else CalendarItem(c.getString(0), c.getString(1), c.getString(2), c.getString(3), c.getString(4), c.getString(5), c.getString(6), c.getString(7))
    }
    fun save(item: CalendarItem) {
        val values = ContentValues().apply {
            put("calendarID", item.calendarID); put("channelID", item.channelID); put("aliasName", item.aliasName)
            put("timeStart", item.timeStart); put("timeEnd", item.timeEnd); put("todo", item.todo)
            put("aliasNames", item.aliasNamesJson); put("repeatOption", item.repeatOption)
        }
        writableDatabase.insertWithOnConflict("calendar", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }
    fun delete(id: String) { writableDatabase.delete("calendar", "calendarID=?", arrayOf(id)) }
    fun deleteByAlias(channelID: String, aliasName: String) { writableDatabase.delete("calendar", "channelID=? AND aliasName=?", arrayOf(channelID, aliasName)) }
}

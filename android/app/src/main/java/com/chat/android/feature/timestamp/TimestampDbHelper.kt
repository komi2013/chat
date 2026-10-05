package com.chat.android.feature.timestamp

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import org.json.JSONArray
import org.json.JSONObject

data class TimestampCodeItem(
    val code: String,
    val channelID: String,
    val planDate: String,
    val until: Int,
    val adminName: String,
    val updatedBy: String,
    val updatedAt: String,
    val approvers: String
)

data class TimestampItem(
    val timestampID: String,
    val channelID: String,
    val aliasName: String,
    val timeIn: String,
    val timeOut: String,
    val breaks: String,
    val stampStatus: Int,
    val ticketID: String,
    val geos: String,
    val ips: String,
    val approveds: String,
    val offday: Int,
    val overwork: Int,
    val offmin: Int,
    val overtime: Int,
    val routine_start: String,
    val routine_end: String,
    val schedules: String
)

class TimestampDbHelper(context: Context) : SQLiteOpenHelper(context, "timestamp.db", null, 2) {
    override fun onCreate(db: SQLiteDatabase) {
        db.execSQL("""
            CREATE TABLE timestamp (
                timestampID TEXT PRIMARY KEY,
                channelID TEXT,
                aliasName TEXT,
                timeIn TEXT,
                timeOut TEXT,
                breaks TEXT,
                stampStatus INTEGER,
                ticketID TEXT,
                geos TEXT,
                ips TEXT,
                approveds TEXT,
                offday INTEGER,
                overwork INTEGER,
                offmin INTEGER,
                overtime INTEGER,
                routine_start TEXT,
                routine_end TEXT,
                schedules TEXT
            )
        """.trimIndent())
        
        db.execSQL("""
            CREATE TABLE timestampCode (
                code TEXT PRIMARY KEY,
                channelID TEXT,
                planDate TEXT,
                until INTEGER,
                adminName TEXT,
                updatedBy TEXT,
                updatedAt TEXT,
                approvers TEXT
            )
        """.trimIndent())
    }

    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) {
        db.execSQL("DROP TABLE IF EXISTS timestamp")
        db.execSQL("DROP TABLE IF EXISTS timestampCode")
        onCreate(db)
    }
}

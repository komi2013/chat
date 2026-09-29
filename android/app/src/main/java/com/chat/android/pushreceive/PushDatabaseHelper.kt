package com.chat.android.pushreceive

import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper

class PushDatabaseHelper(context: Context) : SQLiteOpenHelper(context, DATABASE_NAME, null, DATABASE_VERSION) {

    override fun onCreate(db: SQLiteDatabase) {
        // Table for deduplication
        db.execSQL("""
            CREATE TABLE pushDuplication (
                pushDuplicationID TEXT PRIMARY KEY,
                pushID TEXT,
                pushTitle TEXT,
                channelID TEXT,
                updatedBy TEXT,
                updatedAt INTEGER,
                preContents TEXT
            )
        """.trimIndent())

        // Table for logs
        db.execSQL("""
            CREATE TABLE log (
                logID TEXT PRIMARY KEY,
                pushID TEXT,
                pushTitle TEXT,
                channelID TEXT,
                updatedBy TEXT,
                updatedAt INTEGER,
                preContents TEXT
            )
        """.trimIndent())
    }

    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) {
        // Handle migrations if necessary
    }

    companion object {
        private const val DATABASE_NAME = "push_receive.db"
        private const val DATABASE_VERSION = 1
    }
}

package com.chat.android.core.push

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

        // Table for chunks
        db.execSQL("""
            CREATE TABLE IF NOT EXISTS chunks (
                chunkID TEXT PRIMARY KEY,
                strChunk TEXT,
                chunkPass TEXT,
                chunkIndex INTEGER,
                chunkLength INTEGER
            )
        """.trimIndent())
    }

    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) {
        if (oldVersion < 2) {
            db.execSQL("""
                CREATE TABLE IF NOT EXISTS chunks (
                    chunkID TEXT PRIMARY KEY,
                    strChunk TEXT,
                    chunkPass TEXT,
                    chunkIndex INTEGER,
                    chunkLength INTEGER
                )
            """.trimIndent())
        }
    }

    fun saveChunk(chunkPass: String, index: Int, length: Int, content: String) {
        val chunkID = chunkPass + index
        val values = android.content.ContentValues().apply {
            put("chunkID", chunkID)
            put("strChunk", content)
            put("chunkPass", chunkPass)
            put("chunkIndex", index)
            put("chunkLength", length)
        }
        writableDatabase.insertWithOnConflict("chunks", null, values, android.database.sqlite.SQLiteDatabase.CONFLICT_REPLACE)
    }

    fun getChunks(chunkPass: String): List<Pair<Int, String>> {
        val chunks = mutableListOf<Pair<Int, String>>()
        val cursor = readableDatabase.query(
            "chunks",
            arrayOf("chunkIndex", "strChunk"),
            "chunkPass = ?",
            arrayOf(chunkPass),
            null, null, "chunkIndex ASC"
        )
        cursor.use {
            while (it.moveToNext()) {
                chunks.add(it.getInt(0) to it.getString(1))
            }
        }
        return chunks
    }

    fun deleteChunks(chunkPass: String) {
        writableDatabase.delete("chunks", "chunkPass = ?", arrayOf(chunkPass))
    }

    companion object {
        private const val DATABASE_NAME = "push_receive.db"
        private const val DATABASE_VERSION = 2
    }
}

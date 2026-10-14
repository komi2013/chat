package com.chat.android.feature.entryform

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper

data class StoredEntryForm(
    val form: EntryForm,
    val updatedAt: Long
)

class EntryFormDbHelper(context: Context) : SQLiteOpenHelper(context, DATABASE_NAME, null, DATABASE_VERSION) {
    override fun onCreate(db: SQLiteDatabase) {
        db.execSQL(
            "CREATE TABLE entry_forms (" +
                "id TEXT PRIMARY KEY, " +
                "title TEXT NOT NULL, " +
                "form_json TEXT NOT NULL, " +
                "updated_at INTEGER NOT NULL)"
        )
        db.execSQL(
            "CREATE TABLE IF NOT EXISTS answers (" +
                "answerID TEXT PRIMARY KEY, " +
                "channelID TEXT NOT NULL, " +
                "formType TEXT NOT NULL, " +
                "askID TEXT NOT NULL, " +
                "answerBy TEXT NOT NULL, " +
                "answers_json TEXT NOT NULL, " +
                "updated_at INTEGER NOT NULL)"
        )
    }

    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) {
        if (oldVersion < 2) {
            db.execSQL(
                "CREATE TABLE IF NOT EXISTS answers (" +
                    "answerID TEXT PRIMARY KEY, " +
                    "channelID TEXT NOT NULL, " +
                    "formType TEXT NOT NULL, " +
                    "askID TEXT NOT NULL, " +
                    "answerBy TEXT NOT NULL, " +
                    "answers_json TEXT NOT NULL, " +
                    "updated_at INTEGER NOT NULL)"
            )
        }
    }

    fun saveAnswer(answer: Answer): Long {
        val values = ContentValues().apply {
            put("answerID", answer.answerID)
            put("channelID", answer.channelID)
            put("formType", answer.from)
            put("askID", answer.askID)
            put("answerBy", answer.answerBy)
            put("answers_json", answer.answers)
            put("updated_at", answer.updatedAt)
        }
        return writableDatabase.insertWithOnConflict(
            "answers",
            null,
            values,
            SQLiteDatabase.CONFLICT_REPLACE
        )
    }

    fun saveEntryForm(form: EntryForm): Long {
        val values = ContentValues().apply {
            put("id", form.id)
            put("title", form.title)
            put("form_json", EntryFormCodec.toStorageJson(form))
            put("updated_at", System.currentTimeMillis())
        }
        return writableDatabase.insertWithOnConflict(
            TABLE_NAME,
            null,
            values,
            SQLiteDatabase.CONFLICT_REPLACE
        )
    }

    fun getEntryForm(id: String): EntryForm? {
        val cursor = readableDatabase.query(
            TABLE_NAME,
            arrayOf("form_json"),
            "id = ?",
            arrayOf(id),
            null,
            null,
            null
        )
        return cursor.use {
            if (it.moveToFirst()) EntryFormCodec.parse(it.getString(0)) else null
        }
    }

    fun getAllEntryForms(): List<StoredEntryForm> {
        val forms = mutableListOf<StoredEntryForm>()
        val cursor = readableDatabase.query(
            TABLE_NAME,
            arrayOf("form_json", "updated_at"),
            null,
            null,
            null,
            null,
            "updated_at DESC"
        )
        cursor.use {
            while (it.moveToNext()) {
                forms += StoredEntryForm(
                    form = EntryFormCodec.parse(it.getString(0)),
                    updatedAt = it.getLong(1)
                )
            }
        }
        return forms
    }

    fun deleteEntryForm(id: String): Int = writableDatabase.delete(TABLE_NAME, "id = ?", arrayOf(id))

    companion object {
        private const val DATABASE_NAME = "EntryFormDB.db"
        private const val DATABASE_VERSION = 2
        private const val TABLE_NAME = "entry_forms"
    }
}

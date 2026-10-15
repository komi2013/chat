package com.chat.android.feature.ticket

import android.content.ContentValues
import android.content.Context
import android.database.Cursor
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import com.chat.android.core.network.Ticket

class TicketDbHelper(context: Context) : SQLiteOpenHelper(context, "ticket.db", null, 1) {
    override fun onCreate(db: SQLiteDatabase) {
        db.execSQL("""
            CREATE TABLE ticket (
                ticketID TEXT PRIMARY KEY,
                title TEXT,
                description TEXT,
                status TEXT,
                priority TEXT,
                createdAt TEXT
            )
        """.trimIndent())
    }

    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) {
        db.execSQL("DROP TABLE IF EXISTS ticket")
        onCreate(db)
    }

    fun saveTicket(ticket: Ticket) {
        val db = writableDatabase
        val values = ContentValues().apply {
            put("ticketID", ticket.ticketID)
            put("title", ticket.title)
            put("description", ticket.description)
            put("status", ticket.status)
            put("priority", ticket.priority)
            put("createdAt", ticket.createdAt)
        }
        db.insertWithOnConflict("ticket", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    fun deleteTicket(ticketID: String) {
        writableDatabase.delete("ticket", "ticketID = ?", arrayOf(ticketID))
    }

    fun getTickets(): List<Ticket> {
        val list = mutableListOf<Ticket>()
        val db = readableDatabase
        val cursor = db.query("ticket", null, null, null, null, null, "createdAt DESC")
        cursor.use {
            while (it.moveToNext()) {
                list.add(it.toTicket())
            }
        }
        return list
    }

    fun getTicket(ticketID: String): Ticket? {
        val db = readableDatabase
        val cursor = db.query("ticket", null, "ticketID = ?", arrayOf(ticketID), null, null, null)
        return cursor.use {
            if (it.moveToFirst()) it.toTicket() else null
        }
    }

    private fun Cursor.toTicket(): Ticket {
        return Ticket(
            ticketID = getString(getColumnIndexOrThrow("ticketID")),
            title = getString(getColumnIndexOrThrow("title")),
            description = getString(getColumnIndexOrThrow("description")),
            status = getString(getColumnIndexOrThrow("status")),
            priority = getString(getColumnIndexOrThrow("priority")),
            createdAt = getString(getColumnIndexOrThrow("createdAt"))
        )
    }
}

package com.chat.android.core.push.handlers

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import android.util.Log
import com.chat.android.feature.entryform.EntryFormCodec
import com.chat.android.feature.entryform.EntryFormDbHelper
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler

class EntryFormHandler(
    private val context: Context,
    private val logDbHelper: SQLiteOpenHelper
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        // According to the spec, entryForm contents is an object: entryForm{entryFormID, ...}
        val entryFormJson = pd.getContentsAsObject()
        if (entryFormJson == null) {
            Log.e("EntryFormHandler", "Payload contents is not a JSONObject")
            return
        }

        val entryFormID = entryFormJson.optString("entryFormID", "")
        if (entryFormID.isEmpty()) {
            Log.e("EntryFormHandler", "Missing entryFormID in payload")
            return
        }

        // 1. Save to the specialized EntryForm database
        try {
            val entryForm = EntryFormCodec.parse(entryFormJson.toString())
            EntryFormDbHelper(context).saveEntryForm(entryForm)
        } catch (e: Exception) {
            Log.e("EntryFormHandler", "Failed to save entryForm to EntryFormDbHelper", e)
        }

        val db = logDbHelper.writableDatabase

        // 2. Save to Log table in PushDatabaseHelper (for auditing/history)
        db.beginTransaction()
        try {
            // logID = pd[1]+pd[2]+pd[3]+pd[0] 
            val logID = "${pd.title}${pd.channelID}${pd.updatedBy}${pd.pushID}"
            val logValues = ContentValues().apply {
                put("logID", logID)
                put("pushID", pd.pushID)
                put("pushTitle", pd.title)
                put("channelID", pd.channelID)
                put("updatedBy", pd.updatedBy)
                put("updatedAt", System.currentTimeMillis())
                put("preContents", entryFormJson.toString()) // As per web quirk: pre falls back to incoming
            }
            db.insertWithOnConflict("log", null, logValues, SQLiteDatabase.CONFLICT_REPLACE)

            db.setTransactionSuccessful()
            Log.i("EntryFormHandler", "Successfully logged entryForm push: $entryFormID")
        } catch (e: Exception) {
            Log.e("EntryFormHandler", "Failed to log entryForm push", e)
        } finally {
            db.endTransaction()
        }
    }
}

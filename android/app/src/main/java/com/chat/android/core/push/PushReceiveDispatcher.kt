package com.chat.android.core.push

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import android.util.Log
import com.chat.android.core.push.handlers.EntryFormHandler
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.launch

class PushReceiveDispatcher(
    private val context: Context,
    private val dbHelper: SQLiteOpenHelper
) {
    // Application-scoped Coroutine guarantees background SQLite execution completes
    // even if the FCM service starts shutting down.
    private val applicationScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    // Registry of available handlers
    private val actions: Map<String, PushHandler> = mapOf(
        "entryForm" to EntryFormHandler(context, dbHelper),
        // "thread" to ThreadHandler(context, dbHelper),
        // ... add other handlers here later
    )

    fun receive(notificationData: String, fromPush: Boolean = false) {
        // Parse the wire array
        val pd = try {
            PushData.parse(notificationData, fromPush)
        } catch (e: Exception) {
            Log.e("PushDispatcher", "Failed to parse PushData", e)
            return
        }

        // Toggle Deduplication Logic
        val pushDuplicationID = "${pd.title}${pd.channelID}${pd.updatedBy}${pd.pushID}"
        val db = dbHelper.writableDatabase

        applicationScope.launch {
            try {
                // 1. Check if marker exists
                val cursor = db.query(
                    "pushDuplication",
                    arrayOf("pushDuplicationID"),
                    "pushDuplicationID = ?",
                    arrayOf(pushDuplicationID),
                    null, null, null
                )

                val isDuplicate = cursor.use { it.moveToFirst() }

                if (isDuplicate) {
                    // 2. TOGGLE SEMANTICS: Delete the marker and drop the payload
                    db.delete("pushDuplication", "pushDuplicationID = ?", arrayOf(pushDuplicationID))
                    Log.d("PushDispatcher", "Duplicate dropped, marker deleted for re-sync: $pushDuplicationID")
                    return@launch
                }

                // 3. Cleanup old markers (older than 30 days)
                val thirtyDaysAgo = System.currentTimeMillis() - (30L * 24 * 60 * 60 * 1000)
                db.delete("pushDuplication", "updatedAt < ?", arrayOf(thirtyDaysAgo.toString()))

                // 4. Save new marker
                val values = ContentValues().apply {
                    put("pushDuplicationID", pushDuplicationID)
                    put("pushID", pd.pushID)
                    put("pushTitle", pd.title)
                    put("channelID", pd.channelID)
                    put("updatedBy", pd.updatedBy)
                    put("updatedAt", System.currentTimeMillis())
                    put("preContents", pd.getContentsAsString() ?: "") 
                }
                db.insertWithOnConflict("pushDuplication", null, values, SQLiteDatabase.CONFLICT_REPLACE)

                // 5. Dispatch to Handler
                val action = actions[pd.title]
                if (action != null) {
                    action.handle(pd)
                } else if (pd.title == "pushCheck") {
                    Log.i("PushDispatcher", "Health check ping received for channel: ${pd.channelID}")
                } else {
                    Log.w("PushDispatcher", "Unknown action: ${pd.title}")
                }

            } catch (e: Exception) {
                Log.e("PushDispatcher", "Error processing push for ${pd.title}", e)
            }
        }
    }
}

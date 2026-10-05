package com.chat.android.core.push

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import android.util.Log
import com.chat.android.core.push.handlers.AliasHandler
import com.chat.android.core.push.handlers.BookmarkHandler
import com.chat.android.core.push.handlers.CalendarHandler
import com.chat.android.core.push.handlers.TimestampHandler
import com.chat.android.core.push.handlers.TimestampCodeHandler
import com.chat.android.core.push.handlers.TimestampReportHandler
import com.chat.android.core.push.handlers.TimestampRevertHandler
import com.chat.android.core.push.handlers.ChannelEditHandler
import com.chat.android.core.push.handlers.EmojiHandler
import com.chat.android.core.push.handlers.EntryFormHandler
import com.chat.android.core.push.handlers.GroupHandler
import com.chat.android.core.push.handlers.ThreadHandler
import com.chat.android.core.push.handlers.ThreadHeadHandler
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.SupervisorJob
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.SharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import kotlinx.coroutines.launch
import org.json.JSONArray

class PushReceiveDispatcher(
    private val context: Context,
    private val dbHelper: SQLiteOpenHelper
) {
    // Application-scoped Coroutine guarantees background SQLite execution completes
    // even if the FCM service starts shutting down.
    private val applicationScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    private val _onUpdated = MutableSharedFlow<String>(replay = 0)
    val onUpdated: SharedFlow<String> = _onUpdated.asSharedFlow()

    // Registry of available handlers
    private val actions: Map<String, PushHandler> = mapOf(
        "entryForm" to EntryFormHandler(context, dbHelper),
        "channelEdit" to ChannelEditHandler(context),
        "alias" to AliasHandler(context),
        "group" to GroupHandler(context),
        "thread" to ThreadHandler(context),
        "threadEdit" to ThreadHandler(context),
        "threadHead" to ThreadHeadHandler(context),
        "emoji" to EmojiHandler(context),
        "bookmark" to BookmarkHandler(context),
        "calendar" to CalendarHandler(context, dbHelper),
        "timestamp" to TimestampHandler(context, dbHelper),
        "timestampCode" to TimestampCodeHandler(context),
        "timestampReport" to TimestampReportHandler(context),
        "timestampRevert" to TimestampRevertHandler(context),
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
        // 驥崎､・愛螳壹・ vue/src/pushReceive/pushReceive.js 縺ｨ蜷後§繝医げ繝ｫ譁ｹ蠑・
        //   1蝗樒岼 = 驕ｩ逕ｨ縺励※繝槭・繧ｫ繝ｼ菫晏ｭ・/ 2蝗樒岼 = 繝槭・繧ｫ繝ｼ蜑企勁縺励※遐ｴ譽・ｼ亥・蜷梧悄逕ｨ・・
        val pushDuplicationID = "${pd.title}${pd.channelID}${pd.updatedBy}${pd.pushID}"

        applicationScope.launch {
            try {
                // 1. Check if marker exists
                // FCM 縺ｮ onMessageReceived・医Γ繧､繝ｳ繧ｹ繝ｬ繝・ラ・峨°繧峨ｂ蜻ｼ縺ｰ繧後ｋ縺溘ａ縲・
                // DB 繧ｪ繝ｼ繝励Φ縺ｯ蠢・★ IO 蛛ｴ縺ｮ縺薙・菴咲ｽｮ縺ｧ陦後≧縲・
                val db = dbHelper.writableDatabase
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
                Log.d("PushDispatcher", "Dispatching action: ${pd.title} for channel: ${pd.channelID}")
                if (action != null) {
                    action.handle(pd)
                    _onUpdated.emit(pd.channelID)
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



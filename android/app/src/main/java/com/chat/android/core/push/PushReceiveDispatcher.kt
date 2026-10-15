package com.chat.android.core.push

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import android.util.Log
import com.chat.android.core.push.handlers.AnswerHandler
import com.chat.android.core.push.handlers.ChunkHandler
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
import com.chat.android.core.push.handlers.ReceptionOrderHandler
import com.chat.android.core.push.handlers.StoreSelectHandler
import com.chat.android.core.push.handlers.ShiftStaffHandler
import com.chat.android.core.push.handlers.TopEditHandler
import com.chat.android.core.push.handlers.TicketHandler
import com.chat.android.core.network.SessionManager
import com.chat.android.core.network.ApiService
import com.chat.android.core.repository.UserRepository
import com.chat.android.feature.reception.ReceptionApiService
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
    private val dbHelper: SQLiteOpenHelper,
    private val sessionManager: SessionManager,
    private val apiService: ApiService,
    private val receptionApiService: ReceptionApiService,
    private val userRepository: UserRepository
) {
    // Application-scoped Coroutine guarantees background SQLite execution completes
    // even if the FCM service starts shutting down.
    private val applicationScope = CoroutineScope(SupervisorJob() + Dispatchers.IO)

    private val _onUpdated = MutableSharedFlow<String>(replay = 0)
    val onUpdated: SharedFlow<String> = _onUpdated.asSharedFlow()

    // Registry of available handlers
    private val actions: Map<String, PushHandler> by lazy {
        mapOf(
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
            "receptionOrder" to ReceptionOrderHandler(
                context,
                com.chat.android.feature.reception.ReceptionDbHelper(context),
                dbHelper
            ),
            "shiftStaffEdit" to ShiftStaffHandler(
                context,
                receptionApiService,
                sessionManager
            ),
            "answer" to AnswerHandler(context),
            "storeSelect" to StoreSelectHandler(context, apiService, sessionManager),
            "storeShare" to StoreSelectHandler(context, apiService, sessionManager),
            "chunk" to ChunkHandler(this, dbHelper as PushDatabaseHelper),
            "topEdit" to TopEditHandler(userRepository),
            "ticket" to TicketHandler(context),
        )
    }

    fun receive(notificationData: String, fromPush: Boolean = false) {
        // Parse the wire array
        val pd = try {
            PushData.parse(notificationData, fromPush)
        } catch (e: Exception) {
            Log.e("PushDispatcher", "Failed to parse PushData", e)
            return
        }

        // Toggle Deduplication Logic
        // 重褁E��定�E vue/src/pushReceive/pushReceive.js と同じトグル方弁E
        //   1回目 = 適用してマ�Eカー保孁E/ 2回目 = マ�Eカー削除して破棁E���E同期用�E�E
        val pushDuplicationID = "${pd.title}${pd.channelID}${pd.updatedBy}${pd.pushID}"

        applicationScope.launch {
            try {
                // 1. Check if marker exists
                // FCM の onMessageReceived�E�メインスレチE���E�からも呼ばれるため、E
                // DB オープンは忁E�� IO 側のこ�E位置で行う、E
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



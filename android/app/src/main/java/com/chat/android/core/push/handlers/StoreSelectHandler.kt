package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.network.ApiService
import com.chat.android.core.network.SessionManager
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.calendar.CalendarDbHelper
import com.chat.android.feature.channel.ChannelDbHelper
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import org.json.JSONObject
import java.text.SimpleDateFormat
import java.util.*

/**
 * Handles 'storeSelect' and 'storeShare' push actions (corresponding to vue/src/pushReceive/storeSelect.js and storeShare.js).
 */
class StoreSelectHandler(
    private val context: Context,
    private val apiService: ApiService,
    private val sessionManager: SessionManager
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        when (pd.title) {
            "storeSelect" -> handleStoreSelect(pd)
            "storeShare" -> handleStoreShare(pd)
        }
    }

    private suspend fun handleStoreSelect(pd: PushData) {
        val array = pd.getContentsAsArray() ?: return
        val targetStore = array.optString(0)
        val param = array.optJSONObject(1) ?: return

        if (targetStore == "calendar") {
            val dateStr = param.optString("date")
            if (dateStr.isEmpty()) return

            val sdf = SimpleDateFormat("yyyy-MM-dd", Locale.getDefault())
            val date = try { sdf.parse(dateStr) } catch (e: Exception) { null } ?: return

            val calendar = Calendar.getInstance()
            calendar.time = date
            calendar.add(Calendar.DAY_OF_YEAR, -1)
            val startDate = calendar.time

            calendar.time = date
            calendar.add(Calendar.DAY_OF_YEAR, 13)
            val endDate = calendar.time

            val calendarDb = CalendarDbHelper(context)
            val allEvents = calendarDb.all()

            val calendarData = JSONArray()
            val isoSdf = SimpleDateFormat("yyyy-MM-dd'T'HH:mm:ss", Locale.getDefault())
            isoSdf.timeZone = TimeZone.getDefault()

            allEvents.forEach { event ->
                try {
                    val eventDate = isoSdf.parse(event.timeStart)
                    if (eventDate != null && eventDate >= startDate && eventDate <= endDate) {
                        val obj = JSONObject().apply {
                            put("calendarID", event.calendarID)
                            put("timeStart", event.timeStart)
                            put("timeEnd", event.timeEnd)
                            put("read", 1)
                        }
                        calendarData.put(obj)
                    }
                } catch (e: Exception) {
                    Log.e("StoreSelectHandler", "Failed to parse date for event ${event.calendarID}", e)
                }
            }

            // Send back to the requester via ContentsJustPush
            val channelDb = ChannelDbHelper(context)
            val channel = channelDb.getChannel(pd.channelID) ?: return
            val myname = channel.myname ?: "Unknown"
            val csrf = sessionManager.getCsrf() ?: return

            val contents = JSONArray().apply {
                put("calendar")
                put(0) // save flag
                put(calendarData)
            }

            val requestBody = contents.toString().toRequestBody("application/json".toMediaTypeOrNull())
            val csrfBody = csrf.toRequestBody("text/plain".toMediaTypeOrNull())
            val pushTitleBody = "storeShare".toRequestBody("text/plain".toMediaTypeOrNull())
            val channelIDBody = pd.channelID.toRequestBody("text/plain".toMediaTypeOrNull())
            val updatedByBody = myname.toRequestBody("text/plain".toMediaTypeOrNull())
            val pushNamesBody = JSONArray().put(pd.updatedBy).toString().toRequestBody("application/json".toMediaTypeOrNull())

            try {
                val response = apiService.contentsJustPush(
                    csrf = csrfBody,
                    channelID = channelIDBody,
                    updatedBy = updatedByBody,
                    pushNames = pushNamesBody,
                    contents = requestBody,
                    pushTitle = pushTitleBody
                )
                if (response.isSuccessful) {
                    Log.i("StoreSelectHandler", "Successfully sent storeShare to ${pd.updatedBy}")
                    
                    // Save log like Vue version
                    val pushDb = com.chat.android.core.push.PushDatabaseHelper(context)
                    val logID = pd.title + pd.channelID + pd.updatedBy + pd.contents.toString() + pd.pushID
                    pushDb.saveLog(
                        logID = logID,
                        pushID = pd.pushID,
                        pushTitle = pd.title,
                        channelID = pd.channelID,
                        updatedBy = pd.updatedBy
                    )
                } else {
                    Log.e("StoreSelectHandler", "Failed to send storeShare: ${response.code()}")
                }
            } catch (e: Exception) {
                Log.e("StoreSelectHandler", "Error sending storeShare", e)
            }
        }
    }

    private suspend fun handleStoreShare(pd: PushData) {
        val array = pd.getContentsAsArray() ?: return
        val targetStore = array.optString(0)
        val save = array.optInt(1, 0)
        val contents = array.optJSONArray(2) ?: return

        if (targetStore == "calendar") {
            val calendarDb = CalendarDbHelper(context)
            for (i in 0 until contents.length()) {
                val obj = contents.optJSONObject(i) ?: continue
                // Note: The Vue version doesn't seem to persist storeShare data 
                // unless save == 1. In storeShare.js: if (save == 1) { ... upsertIDB ... }
                // And it updates a calendarsStore (state management) in Vue.
                // For Android, we might want to at least log it or update local DB if save == 1.
                if (save == 1) {
                    val id = obj.optString("calendarID")
                    if (id.isEmpty()) continue
                    
                    // The contents from storeSelect only have calendarID, timeStart, timeEnd.
                    // We might need a full object for save(), but let's see what we can do.
                    val existing = calendarDb.get(id)
                    calendarDb.save(
                        com.chat.android.feature.calendar.CalendarItem(
                            calendarID = id,
                            channelID = pd.channelID,
                            aliasName = obj.optString("aliasName", pd.updatedBy),
                            timeStart = obj.optString("timeStart"),
                            timeEnd = obj.optString("timeEnd"),
                            todo = obj.optString("todo", existing?.todo ?: ""),
                            aliasNamesJson = obj.optJSONArray("aliasNames")?.toString() ?: existing?.aliasNamesJson ?: "[]",
                            repeatOption = obj.optString("repeatOption", existing?.repeatOption ?: "none")
                        )
                    )
                }
            }
            Log.i("StoreSelectHandler", "Processed storeShare for $targetStore, save=$save")
        }
    }
}

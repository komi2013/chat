package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.network.SessionManager
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.reception.ReceptionApiService
import com.chat.android.feature.reception.ReceptionDbHelper
import com.chat.android.feature.reception.ShiftStaff
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext

class ShiftStaffHandler(
    private val context: Context,
    private val apiService: ReceptionApiService,
    private val sessionManager: SessionManager
) : PushHandler {
    private val gson = Gson()
    private val dbHelper = ReceptionDbHelper(context)

    override suspend fun handle(pd: PushData) {
        val contents = pd.getContentsAsString()
        if (contents == null) {
            Log.e("ShiftStaffHandler", "Payload contents is not a string")
            return
        }

        try {
            val listType = object : TypeToken<List<ShiftStaff>>() {}.type
            val staffList: List<ShiftStaff> = gson.fromJson(contents, listType)

            withContext(Dispatchers.IO) {
                staffList.forEach { staff ->
                    if (staff.type == "upsert") {
                        dbHelper.saveShiftStaff(staff)
                    } else if (staff.type == "delete") {
                        dbHelper.deleteShiftStaff(staff)
                    }
                }
            }

            // Acknowledge the push
            sessionManager.serializeAuthenticatedRequest {
                val csrf = sessionManager.getCsrf()
                if (csrf != null) {
                    val response = apiService.pushResponse(csrf, pd.pushID)
                    if (response.isSuccessful) {
                        sessionManager.applyResponseCsrf(csrf, response.body()?.csrf)
                    }
                }
            }

            Log.i("ShiftStaffHandler", "Successfully processed shiftStaffEdit push")
        } catch (e: Exception) {
            Log.e("ShiftStaffHandler", "Failed to process shiftStaffEdit push", e)
        }
    }
}

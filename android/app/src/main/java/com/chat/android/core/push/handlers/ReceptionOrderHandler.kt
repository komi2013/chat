package com.chat.android.core.push.handlers

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.reception.ReceptionDbHelper
import com.chat.android.feature.reception.ReceptionOrder
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken

class ReceptionOrderHandler(
    private val context: Context,
    private val db: ReceptionDbHelper,
    private val logDbHelper: SQLiteOpenHelper
) : PushHandler {
    private val gson = Gson()

    override suspend fun handle(pd: PushData) {
        val contents = pd.getContentsAsString() ?: return
        
        try {
            val type = object : TypeToken<List<ReceptionOrder>>() {}.type
            val orders: List<ReceptionOrder> = gson.fromJson(contents, type)
            
            if (orders.isEmpty()) {
                // If the list is empty, it might mean the orders for this seat were cleared
                // or we need to find which seat this push belongs to.
                // In the web version, a push with an empty list for receptionOrder usually triggers a full refresh
                // or is accompanied by other data. 
                // However, ReceptionOrder delete push usually just says "deleted" or similar.
                // For now, let's just save whatever comes.
            }

            orders.forEach { order ->
                if (order.price == -1) { // Assuming -1 price might mean deletion in some protocols, or check if specific fields are null
                     db.deleteOrder(order.receptionOrderID)
                } else {
                     db.saveOrder(order)
                }
            }
            Log.i("ReceptionOrderHandler", "Saved ${orders.size} orders from push")

            // Log the push for auditing
            val logDb = logDbHelper.writableDatabase
            val logID = "${pd.title}${pd.channelID}${pd.updatedBy}${pd.pushID}"
            val logValues = ContentValues().apply {
                put("logID", logID)
                put("pushID", pd.pushID)
                put("pushTitle", pd.title)
                put("channelID", pd.channelID)
                put("updatedBy", pd.updatedBy)
                put("updatedAt", System.currentTimeMillis())
                put("preContents", contents)
            }
            logDb.insertWithOnConflict("log", null, logValues, SQLiteDatabase.CONFLICT_REPLACE)
        } catch (e: Exception) {
            Log.e("ReceptionOrderHandler", "Failed to parse receptionOrder push", e)
        }
    }
}

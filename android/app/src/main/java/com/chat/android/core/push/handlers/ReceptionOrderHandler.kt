package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.reception.ReceptionDbHelper
import com.chat.android.feature.reception.ReceptionOrder
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken

class ReceptionOrderHandler(private val context: Context) : PushHandler {
    private val gson = Gson()

    override suspend fun handle(pd: PushData) {
        val contents = pd.getContentsAsArray() ?: return
        // In Vue: receptionOrder: (pd) => { ... }
        // The payload seems to be a list of orders based on the context in ReceptionOrder.vue
        // However, usually push data for "receptionOrder" event contains the new order or a list of them.
        // Assuming it's a list of ReceptionOrder objects in the contents array or as a JSON string in pd[4].
        
        val ordersJson = contents.optString(1) // Adjust index based on actual payload structure
        if (ordersJson.isNullOrEmpty()) return

        try {
            val type = object : TypeToken<List<ReceptionOrder>>() {}.type
            val orders: List<ReceptionOrder> = gson.fromJson(ordersJson, type)
            val db = ReceptionDbHelper(context)
            orders.forEach { db.saveOrder(it) }
            Log.i("ReceptionOrderHandler", "Saved ${orders.size} orders from push")
        } catch (e: Exception) {
            Log.e("ReceptionOrderHandler", "Failed to parse receptionOrder push", e)
        }
    }
}

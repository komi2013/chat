package com.chat.android.feature.reception

import android.content.ContentValues
import android.content.Context
import android.database.sqlite.SQLiteDatabase
import android.database.sqlite.SQLiteOpenHelper
import com.google.gson.Gson
import com.google.gson.reflect.TypeToken

class ReceptionDbHelper(context: Context) : SQLiteOpenHelper(context, DATABASE_NAME, null, DATABASE_VERSION) {
    private val gson = Gson()

    override fun onCreate(db: SQLiteDatabase) {
        db.execSQL("""
            CREATE TABLE receptionOrder (
                receptionOrderID TEXT PRIMARY KEY,
                seatName TEXT,
                menuID INTEGER,
                menuName TEXT,
                itemChoices TEXT,
                freeOptions TEXT,
                paidOptions TEXT,
                freeMultiOptions TEXT,
                price INTEGER,
                items TEXT
            )
        """.trimIndent())
    }

    override fun onUpgrade(db: SQLiteDatabase, oldVersion: Int, newVersion: Int) {}

    fun saveOrder(order: ReceptionOrder) {
        val values = ContentValues().apply {
            put("receptionOrderID", order.receptionOrderID)
            put("seatName", order.seatName)
            put("menuID", order.menuID)
            put("menuName", order.menuName)
            put("itemChoices", gson.toJson(order.itemChoices))
            put("freeOptions", gson.toJson(order.freeOptions))
            put("paidOptions", gson.toJson(order.paidOptions))
            put("freeMultiOptions", gson.toJson(order.freeMultiOptions))
            put("price", order.price)
            put("items", gson.toJson(order.items))
        }
        writableDatabase.insertWithOnConflict("receptionOrder", null, values, SQLiteDatabase.CONFLICT_REPLACE)
    }

    fun deleteOrder(orderID: String) {
        writableDatabase.delete("receptionOrder", "receptionOrderID = ?", arrayOf(orderID))
    }

    fun getAllOrders(): List<ReceptionOrder> {
        val orders = mutableListOf<ReceptionOrder>()
        val cursor = readableDatabase.query("receptionOrder", null, null, null, null, null, null)
        cursor.use {
            while (it.moveToNext()) {
                val itemChoicesType = object : TypeToken<Map<String, String>>() {}.type
                val intListType = object : TypeToken<List<Int>>() {}.type
                val paidOptionsType = object : TypeToken<List<ReceptionOrderPaidOption>>() {}.type
                val itemsType = object : TypeToken<List<ReceptionOrderItem>>() {}.type

                orders.add(ReceptionOrder(
                    receptionOrderID = it.getString(it.getColumnIndexOrThrow("receptionOrderID")),
                    seatName = it.getString(it.getColumnIndexOrThrow("seatName")),
                    menuID = it.getInt(it.getColumnIndexOrThrow("menuID")),
                    menuName = it.getString(it.getColumnIndexOrThrow("menuName")),
                    itemChoices = gson.fromJson(it.getString(it.getColumnIndexOrThrow("itemChoices")), itemChoicesType),
                    freeOptions = gson.fromJson(it.getString(it.getColumnIndexOrThrow("freeOptions")), intListType),
                    paidOptions = gson.fromJson(it.getString(it.getColumnIndexOrThrow("paidOptions")), paidOptionsType),
                    freeMultiOptions = gson.fromJson(it.getString(it.getColumnIndexOrThrow("freeMultiOptions")), intListType),
                    price = it.getInt(it.getColumnIndexOrThrow("price")),
                    items = gson.fromJson(it.getString(it.getColumnIndexOrThrow("items")), itemsType)
                ))
            }
        }
        return orders
    }

    fun deleteOrdersForSeat(seatName: String) {
        writableDatabase.delete("receptionOrder", "seatName = ?", arrayOf(seatName))
    }

    companion object {
        private const val DATABASE_NAME = "ReceptionDB.db"
        private const val DATABASE_VERSION = 1
    }
}

package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "calendar")
data class CalendarEntity(
    @PrimaryKey
    val calendarID: String,
    val channelID: String? = null,
    val title: String? = null,
    val description: String? = null,
    val date: String? = null,
    val startTime: String? = null,
    val endTime: String? = null,
    val location: String? = null,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

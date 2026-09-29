package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey
import androidx.room.Index

@Entity(
    tableName = "ticket",
    indices = [
        Index(value = ["ticketID"]),
        Index(value = ["status"])
    ]
)
data class TicketEntity(
    @PrimaryKey
    val ticketID: String,
    val channelID: String? = null,
    val aliasName: String? = null,
    val title: String? = null,
    val description: String? = null,
    val status: String? = null,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

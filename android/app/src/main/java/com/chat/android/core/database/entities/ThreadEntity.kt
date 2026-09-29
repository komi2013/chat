package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey
import androidx.room.Index

@Entity(
    tableName = "thread",
    indices = [
        Index(value = ["messageID"]),
        Index(value = ["parentID"]),
        Index(value = ["channelID"]),
        Index(value = ["channelID", "parentID"])
    ]
)
data class ThreadEntity(
    @PrimaryKey
    val messageID: String,
    val channelID: String,
    val parentID: String? = null,
    val aliasName: String,
    val aliasImg: String? = null,
    val content: String,
    val good: Int = 0,
    val bad: Int = 0,
    val bookmark: Boolean = false,
    val timestamp: String,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

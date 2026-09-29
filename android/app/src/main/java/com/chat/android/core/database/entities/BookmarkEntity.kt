package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "bookmark")
data class BookmarkEntity(
    @PrimaryKey
    val messageID: String,
    val channelID: String? = null,
    val aliasName: String? = null,
    val bookmarked: Boolean = true,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

package com.chat.android.data.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey
import androidx.room.Index

@Entity(
    tableName = "channel",
    indices = [
        Index(value = ["channelID"]),
        Index(value = ["displayStatus"])
    ]
)
data class ChannelEntity(
    @PrimaryKey
    val channelID: String,
    val name: String,
    val description: String? = null,
    val myname: String? = null,
    val myimg: String? = null,
    val displayStatus: String? = null,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

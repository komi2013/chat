package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey
import androidx.room.Index

@Entity(
    tableName = "alias",
    indices = [
        Index(value = ["aliasID"]),
        Index(value = ["channelID"])
    ]
)
data class AliasEntity(
    @PrimaryKey
    val aliasID: String,
    val aliasName: String,
    val channelID: String? = null,
    val aliasImg: String? = null,
    val aliasBio: String? = null,
    val accessRight: String? = null,
    val good: Int = 0,
    val bad: Int = 0,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey
import androidx.room.Index

@Entity(
    tableName = "advertisement",
    indices = [Index(value = ["advertisementID"])]
)
data class AdvertisementEntity(
    @PrimaryKey
    val advertisementID: String,
    val adName: String? = null,
    val adStart: String? = null,
    val adEnd: String? = null,
    val adImg: String? = null,
    val adLink: String? = null,
    val adText: String? = null,
    val adPublic: Boolean = false,
    val channelID: String? = null,
    val createdAt: String? = null,
    val updatedAt: String? = null
)

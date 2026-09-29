package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "user_nickname")
data class UserNicknameEntity(
    @PrimaryKey val nickname: String,
    val nickImg: String? = null,
    val nickBio: String? = null,
    val good: Int = 0,
    val bad: Int = 0,
    val createdAt: String? = null,
    val accessRight: String? = null
)
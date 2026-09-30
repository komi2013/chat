package com.chat.android.core.network.model

data class UserNicknameEntity(
    val nickname: String,
    val nickImg: String? = null,
    val nickBio: String? = null,
    val good: Int = 0,
    val bad: Int = 0,
    val createdAt: String? = null,
    val accessRight: String? = null
)

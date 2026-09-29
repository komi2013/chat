package com.chat.android.core.database.entities

import androidx.room.Entity
import androidx.room.PrimaryKey

@Entity(tableName = "user_profile")
data class UserProfileEntity(
    @PrimaryKey val profileId: String = CURRENT_PROFILE_ID,
    val name: String? = null,
    val mail: String? = null,
    val nickname: String? = null,
    val channelID: String? = null,
    val accessRight: String? = null,
    val admin: Boolean? = null,
    val telephone: String? = null,
    val walletAddress: String? = null,
    val latitude: Double? = null,
    val longitude: Double? = null,
    val nickImg: String? = null,
    val nickBio: String? = null
) {
    companion object {
        const val CURRENT_PROFILE_ID = "current"
    }
}
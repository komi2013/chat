package com.chat.android.navigation

import kotlinx.serialization.Serializable

@Serializable
data object HomeRoute

@Serializable
data object UserRoute

@Serializable
data object EntryFormListRoute

@Serializable
data class EntryFormEditRoute(val id: String? = null, val formJson: String? = null)

@Serializable
data class EntryFormAnswerRoute(val id: String)

@Serializable
data class ChannelRoute(val id: String? = null)

/** /profile/{id}/?code=... に対応するルート。 */
@Serializable
data class ProfileRoute(val id: String? = null, val code: String? = null)

/** グループ編集画面。 */
@Serializable
data class GroupRoute(val id: String? = null)


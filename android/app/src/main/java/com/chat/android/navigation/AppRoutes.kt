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

/**
 * /people/{id}/{name}/ に対応するルート。
 * vue/src/views/People.vue に対応し、name はエイリアス名またはグループ名のどちらでも受け付ける。
 */
@Serializable
data class PeopleRoute(val id: String? = null, val name: String? = null)

/**
 * /thread/{channelID}/{parentID}/ に対応するルート。
 * vue/src/views/Thread.vue に対応し、parentID はスレッドID。
 */
@Serializable
data class ThreadRoute(val channelID: String? = null, val parentID: String? = null)

/** /threadHead/{parentID}/ に対応するルート（スレッド設定）。 */
@Serializable
data class ThreadHeadRoute(val parentID: String? = null)


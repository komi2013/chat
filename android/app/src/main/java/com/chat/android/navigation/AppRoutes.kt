package com.chat.android.navigation

import kotlinx.serialization.Serializable

@Serializable
data object HomeRoute

@Serializable
data object LoginRoute

@Serializable
data object UserRoute

@Serializable
data object SettingsRoute

@Serializable
data object EntryFormListRoute

@Serializable
data class EntryFormEditRoute(val id: String? = null, val formJson: String? = null)

@Serializable
data class EntryFormAnswerRoute(val id: String)

@Serializable
data class CalendarRoute(val date: String? = null)

@Serializable
data class CalendarEditRoute(val id: String? = null, val dateTime: String? = null)

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
data class ThreadRoute(
    val channelID: String? = null,
    val parentID: String? = null,
    val backID: String? = null
)

/** Native SkyWay video call room, shared with the Vue /threadCall/{code}/ route. */
@Serializable
data class ThreadCallRoute(val code: String)

/** /threadHead/{parentID}/ に対応するルート（スレッド設定）。 */
@Serializable
data class ThreadHeadRoute(val parentID: String? = null)


@Serializable
data class TimestampRoute(val adminName: String, val code: String)

@Serializable
data object TimestampCodeRoute

@Serializable
data class TimestampReportRoute(val admin: String, val month: String? = null, val stamper: String? = null)

@Serializable
data class ReceptionRoute(val id: String? = null, val code: String? = null)

@Serializable
data class ReceptionEnterRoute(val id: String, val passkey: String)

@Serializable
data class ReceptionMenuRoute(val id: String, val code: String, val codeType: String)

@Serializable
data class ReceptionBookRoute(val id: String, val date: String? = null, val menuID: Int? = null)

@Serializable
data class ReceptionOrderRoute(val id: String, val code: String)

@Serializable
data class ReceptionShiftRoute(val receptionID: String, val date: String? = null)

@Serializable
data class ReceptionThreadRoute(val channelID: String, val code: String, val backID: String? = null)

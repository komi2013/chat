package com.chat.android.navigation

import kotlinx.serialization.Serializable

@Serializable
data object UserRoute

@Serializable
data object EntryFormListRoute

@Serializable
data class EntryFormEditRoute(val id: String? = null, val formJson: String? = null)

@Serializable
data class EntryFormAnswerRoute(val id: String)

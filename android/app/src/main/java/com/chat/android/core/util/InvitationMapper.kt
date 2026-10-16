package com.chat.android.core.util

import android.net.Uri
import com.chat.android.navigation.ProfileRoute

object InvitationMapper {

    /**
     * Parses a direct App Link or custom scheme URL into a ProfileRoute.
     * Supported formats:
     * - https://chat.quigen.info/invitation/{M}/?code={code}
     * - https://chat.quigen.info/profile/{id}/?code=...
     * - chat://profile/{id}?code=...
     */
    fun parseUri(uri: Uri?): ProfileRoute? {
        if (uri == null) return null

        if (uri.scheme == "chat") {
            if (uri.host != "profile") return null
            val channelID = uri.pathSegments?.firstOrNull()?.takeIf { it.isNotBlank() } ?: return null
            return ProfileRoute(id = channelID, code = uri.getQueryParameter("code"))
        }

        if (uri.scheme != "https" && uri.scheme != "http") return null
        if (uri.host != "chat.quigen.info") return null

        val segments = uri.pathSegments ?: return null
        if (segments.size < 2) return null

        return when (segments[0]) {
            "profile" -> {
                val channelID = segments[1].takeIf { it.isNotBlank() } ?: return null
                ProfileRoute(id = channelID, code = uri.getQueryParameter("code"))
            }
            "invitation" -> {
                val m = segments[1].takeIf { it.isNotBlank() } ?: return null
                val code = uri.getQueryParameter("code").takeIf { !it.isNullOrBlank() } ?: return null
                ProfileRoute(id = m, code = code)
            }
            else -> null
        }
    }

    /**
     * Parses the Google Play Install Referrer string.
     * Expected format: m=shop001&code=abc123
     */
    fun parseReferrer(referrerUrl: String?): ProfileRoute? {
        if (referrerUrl.isNullOrBlank()) return null
        
        val params = try {
            referrerUrl.split("&").associate {
                val parts = it.split("=")
                if (parts.size == 2) parts[0] to parts[1] else parts[0] to ""
            }
        } catch (e: Exception) {
            return null
        }

        val m = params["m"]?.takeIf { it.isNotBlank() } ?: return null
        val code = params["code"]?.takeIf { it.isNotBlank() } ?: return null
        
        return ProfileRoute(id = m, code = code)
    }
}

package com.chat.android.core.util

import com.chat.android.BuildConfig

/**
 * The backend stores uploaded images as root relative paths such as
 * "/data/img/-/nickname.png" (see common/file.go ImgSave: PublicImgPath + "/img/...").
 *
 * The web client binds those values straight into an img src, so the browser resolves them
 * against the current origin. Coil has no origin to resolve against, so every reference the
 * app loads has to be turned into an absolute URL first.
 */
fun String.toAbsoluteImageUrl(): String {
    val reference = trim()
    if (reference.isEmpty() || !reference.startsWith("/")) return reference
    return BuildConfig.BASE_URL.trimEnd('/') + reference
}

package com.chat.android.core.push

interface PushHandler {
    /**
     * Executes the business logic for a specific push title.
     * Runs on an IO Coroutine thread.
     */
    suspend fun handle(pd: PushData)
}

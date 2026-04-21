package com.chat.android.receiver

import android.content.BroadcastReceiver
import android.content.Context
import android.content.Intent
import com.chat.android.service.CallNotificationService

class CallReceiver : BroadcastReceiver() {
    
    override fun onReceive(context: Context, intent: Intent) {
        when (intent.action) {
            "INCOMING_CALL" -> {
                val callerName = intent.getStringExtra("caller_name") ?: "Unknown"
                val roomId = intent.getStringExtra("room_id") ?: ""
                CallNotificationService.startIncomingCall(context, callerName, roomId)
            }
            "END_CALL" -> {
                CallNotificationService.endCall(context)
            }
        }
    }
    
    companion object {
        const val ACTION_INCOMING_CALL = "INCOMING_CALL"
        const val ACTION_END_CALL = "END_CALL"
    }
}
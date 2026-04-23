package com.chat.android.ui.call

import android.content.Intent
import android.os.Bundle
import android.view.WindowManager
import android.widget.Button
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class IncomingCallActivity : AppCompatActivity() {
    
    private lateinit var callerNameText: TextView
    private lateinit var answerButton: Button
    private lateinit var declineButton: Button
    private var callerName: String = ""
    private var roomId: String = ""
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        // Set up full-screen immersive mode for incoming call
        window.addFlags(
            WindowManager.LayoutParams.FLAG_SHOW_WHEN_LOCKED or
            WindowManager.LayoutParams.FLAG_DISMISS_KEYGUARD or
            WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON or
            WindowManager.LayoutParams.FLAG_TURN_SCREEN_ON
        )
        
        val layout = android.widget.LinearLayout(this).apply {
            orientation = android.widget.LinearLayout.VERTICAL
            gravity = android.view.Gravity.CENTER
            setPadding(32, 32, 32, 32)
            setBackgroundColor(android.graphics.Color.DKGRAY)
        }
        
        callerNameText = TextView(this).apply {
            textSize = 24f
            setTextColor(android.graphics.Color.WHITE)
            gravity = android.view.Gravity.CENTER
            setPadding(0, 0, 0, 48)
        }
        
        answerButton = Button(this).apply {
            text = "Answer"
            setBackgroundColor(android.graphics.Color.GREEN)
            setTextColor(android.graphics.Color.WHITE)
            setOnClickListener { answerCall() }
        }
        
        declineButton = Button(this).apply {
            text = "Decline"
            setBackgroundColor(android.graphics.Color.RED)
            setTextColor(android.graphics.Color.WHITE)
            setOnClickListener { declineCall() }
        }
        
        layout.addView(callerNameText)
        layout.addView(answerButton)
        layout.addView(declineButton)
        setContentView(layout)
        
        callerName = intent.getStringExtra("caller_name") ?: "Unknown Caller"
        roomId = intent.getStringExtra("room_id") ?: ""
        callerNameText.text = "Incoming call from\n$callerName"
    }
    
    private fun answerCall() {
        val intent = Intent(this, VideoCallActivity::class.java).apply {
            putExtra("room_id", roomId)
            putExtra("caller_name", callerName)
            flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
        }
        startActivity(intent)
        finish()
    }
    
    private fun declineCall() {
        finish()
    }
}

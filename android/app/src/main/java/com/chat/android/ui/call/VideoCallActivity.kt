package com.chat.android.ui.call

import android.content.Intent
import android.os.Bundle
import android.widget.Button
import android.widget.TextView
import androidx.appcompat.app.AppCompatActivity

class VideoCallActivity : AppCompatActivity() {
    
    private lateinit var endCallButton: Button
    private lateinit var statusText: TextView
    private var roomId: String = ""
    private var callerName: String = ""
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        val layout = android.widget.LinearLayout(this).apply {
            orientation = android.widget.LinearLayout.VERTICAL
            setPadding(20, 50, 20, 50)
            setBackgroundColor(android.graphics.Color.BLACK)
        }
        
        statusText = TextView(this).apply {
            text = "Connecting..."
            textSize = 18f
            setTextColor(android.graphics.Color.WHITE)
            gravity = android.view.Gravity.CENTER
            setPadding(0, 0, 0, 30)
        }
        
        endCallButton = Button(this).apply {
            text = "End Call"
            setBackgroundColor(android.graphics.Color.RED)
            setTextColor(android.graphics.Color.WHITE)
            textSize = 16f
            setOnClickListener { endCall() }
        }
        
        layout.addView(statusText)
        layout.addView(endCallButton)
        setContentView(layout)
        
        roomId = intent.getStringExtra("room_id") ?: ""
        callerName = intent.getStringExtra("caller_name") ?: "Unknown"
        statusText.text = "Call with $callerName"
        
        initializeCall()
    }
    
    private fun initializeCall() {
        statusText.text = "Connected to $callerName"
    }
    
    private fun endCall() {
        finish()
    }
    
    override fun onBackPressed() {
        endCall()
    }
}
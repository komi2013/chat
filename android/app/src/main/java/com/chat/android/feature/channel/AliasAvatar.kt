package com.chat.android.feature.channel

import android.util.Base64
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.shape.CircleShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.unit.dp
import coil.compose.AsyncImage

@Composable
fun AliasAvatar(aliasImg: String, modifier: Modifier = Modifier) {
    if (aliasImg.startsWith(",")) {
        // Format: ",content,colorHex"
        val parts = aliasImg.split(",")
        val content = parts.getOrNull(1) ?: ""
        val colorString = parts.getOrNull(2) ?: "#CCCCCC"
        val color = try {
            Color(android.graphics.Color.parseColor(colorString))
        } catch (_: Exception) {
            Color.Gray
        }
        
        Box(
            modifier = modifier
                .background(androidx.compose.material3.MaterialTheme.colorScheme.surface, CircleShape)
                .border(1.dp, color, CircleShape),
            contentAlignment = Alignment.Center
        ) {
            Text(text = content, color = color)
        }
    } else if (aliasImg.startsWith("data:image")) {
        val bytes = androidx.compose.runtime.remember(aliasImg) {
            runCatching { Base64.decode(aliasImg.substringAfter(','), Base64.DEFAULT) }.getOrNull()
        }
        AsyncImage(model = bytes, contentDescription = "Avatar", modifier = modifier.clip(CircleShape))
    } else {
        // Standard URL image loading using Coil
        AsyncImage(
            model = aliasImg,
            contentDescription = "Avatar",
            modifier = modifier.clip(CircleShape)
        )
    }
}

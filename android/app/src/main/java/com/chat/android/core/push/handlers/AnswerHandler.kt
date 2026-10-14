package com.chat.android.core.push.handlers

import android.content.Context
import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushHandler
import com.chat.android.feature.entryform.Answer
import com.chat.android.feature.entryform.EntryFormDbHelper
import org.json.JSONArray
import java.lang.StringBuilder

class AnswerHandler(
    private val context: Context
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        val contents = pd.getContentsAsArray()
        if (contents == null) {
            Log.e("AnswerHandler", "Payload contents is not a JSONArray")
            return
        }

        val channelID = pd.channelID
        val updatedBy = pd.updatedBy
        
        // pd[4][0] = from (1=entryForm, 2=reception)
        val from = contents.optString(0, "")
        // pd[4][1] = ID
        val id = contents.optString(1, "")
        // pd[4][2] = answers
        val answersJson = contents.opt(2)?.toString() ?: "[]"
        // pd[4][3] = optionKeys
        val optionKeys = contents.optJSONArray(3)
        
        val optionKeyStr = StringBuilder()
        if (optionKeys != null) {
            for (i in 0 until optionKeys.length()) {
                optionKeyStr.append(optionKeys.optString(i))
            }
        }
        
        val answerID = channelID + from + updatedBy + id + optionKeyStr.toString()
        
        val answer = Answer(
            answerID = answerID,
            channelID = channelID,
            from = from,
            askID = id,
            answerBy = updatedBy,
            answers = answersJson,
            updatedAt = System.currentTimeMillis()
        )
        
        try {
            EntryFormDbHelper(context).saveAnswer(answer)
            Log.i("AnswerHandler", "Successfully saved answer: $answerID")
        } catch (e: Exception) {
            Log.e("AnswerHandler", "Failed to save answer", e)
        }
    }
}

package com.chat.android.core.push.handlers

import android.util.Log
import com.chat.android.core.push.PushData
import com.chat.android.core.push.PushDatabaseHelper
import com.chat.android.core.push.PushHandler
import com.chat.android.core.push.PushReceiveDispatcher
import org.json.JSONArray

class ChunkHandler(
    private val dispatcher: PushReceiveDispatcher,
    private val dbHelper: PushDatabaseHelper
) : PushHandler {

    override suspend fun handle(pd: PushData) {
        // Wire format from chunk.js:
        // pd[2] = strChunk
        // pd[3] = chunkPass
        // pd[4] = chunkIndex
        // pd[5] = chunkLength

        val strChunk = pd.rawJson.optString(2, "")
        val chunkPass = pd.rawJson.optString(3, "")
        val chunkIndex = pd.rawJson.optInt(4, -1)
        val chunkLength = pd.rawJson.optInt(5, -1)

        if (chunkPass.isEmpty() || chunkIndex == -1 || chunkLength == -1) {
            Log.e("ChunkHandler", "Invalid chunk data: pass=$chunkPass, index=$chunkIndex, length=$chunkLength")
            return
        }

        // 1. Save current chunk
        dbHelper.saveChunk(chunkPass, chunkIndex, chunkLength, strChunk)

        // 2. Check if all chunks are present
        val chunks = dbHelper.getChunks(chunkPass)
        if (chunks.size == chunkLength) {
            try {
                // 3. Combine chunks
                val combinedString = chunks.joinToString("") { it.second }
                
                // 4. Parse combined string and re-dispatch
                // The JS version does JSON.parse and then data.unshift("")
                // which effectively turns the payload back into the [pushID, title, ...] array
                // where index 0 (pushID) is empty.
                val combinedArray = JSONArray(combinedString)
                
                // Create a new JSONArray with an empty string at index 0 as per JS: data.unshift("")
                val finalArray = JSONArray()
                finalArray.put("") // Empty pushID placeholder
                for (i in 0 until combinedArray.length()) {
                    finalArray.put(combinedArray.get(i))
                }

                Log.i("ChunkHandler", "Reassembling $chunkLength chunks for pass $chunkPass")
                
                // Re-dispatch the reassembled payload
                dispatcher.receive(finalArray.toString(), fromPush = pd.directPush)
                
                // 5. Cleanup
                dbHelper.deleteChunks(chunkPass)
            } catch (e: Exception) {
                Log.e("ChunkHandler", "Failed to reassemble chunks for pass $chunkPass", e)
            }
        } else {
            Log.d("ChunkHandler", "Stored chunk $chunkIndex/$chunkLength for pass $chunkPass")
        }
    }
}

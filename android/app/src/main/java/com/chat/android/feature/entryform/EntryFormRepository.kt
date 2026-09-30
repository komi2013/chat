package com.chat.android.feature.entryform

import android.content.Context
import com.chat.android.core.data.SessionManager
import com.chat.android.core.network.ApiService
import com.google.gson.Gson
import com.google.gson.JsonArray
import com.google.gson.JsonObject
import dagger.hilt.android.qualifiers.ApplicationContext
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.RequestBody.Companion.toRequestBody
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class EntryFormRepository @Inject constructor(
    @ApplicationContext context: Context,
    private val apiService: ApiService,
    private val entryFormApiService: EntryFormApiService,
    private val sessionManager: SessionManager
) {
    private val database = EntryFormDbHelper(context)
    private val preferences = context.getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
    private val gson = Gson()

    fun getChannelId(): String = preferences.getString("channelID", "").orEmpty()

    fun setChannelId(channelId: String) {
        preferences.edit().putString("channelID", channelId).apply()
    }

    fun getUpdatedBy(): String = preferences.getString("entryFormUpdatedBy", null)
        ?: sessionManager.getNickname().orEmpty()

    fun setUpdatedBy(updatedBy: String) {
        preferences.edit().putString("entryFormUpdatedBy", updatedBy).apply()
    }

    fun getPushNames(): String = preferences.getString("entryFormPushNames", "").orEmpty()

    fun setPushNames(pushNames: String) {
        preferences.edit().putString("entryFormPushNames", pushNames).apply()
    }

    suspend fun getAllForms(): List<StoredEntryForm> = withContext(Dispatchers.IO) {
        database.getAllEntryForms()
    }

    suspend fun getForm(id: String): EntryForm? = withContext(Dispatchers.IO) {
        database.getEntryForm(id)
    }

    suspend fun saveLocal(form: EntryForm): String? = withContext(Dispatchers.IO) {
        if (form.id.isBlank()) return@withContext "フォームIDが空です"
        if (database.saveEntryForm(form) < 0) "フォームを保存できませんでした" else null
    }

    suspend fun saveAndPublish(form: EntryForm): String? {
        val localError = saveLocal(form)
        if (localError != null) return localError
        return sendPush(EntryFormCodec.toWireJson(form), "entryForm")
    }

    suspend fun submitAnswers(form: EntryForm, answers: Map<Int, Any?>): String? {
        val answerData = JsonArray()
        val optionKeys = JsonArray()
        EntryFormCodec.normalize(form.questions).forEach { question ->
            val answer = answers[question.sequence]
            answerData.add(JsonObject().apply {
                addProperty("sequence", question.sequence)
                add("answer", gson.toJsonTree(answer))
            })
            if (question.optionKey) appendOptionKeys(optionKeys, answer)
        }
        val contents = JsonArray().apply {
            add("1")
            add(form.id)
            add(answerData)
            add(optionKeys)
        }
        return sendPush(contents.toString(), "answer")
    }

    suspend fun importForm(json: String): Pair<EntryForm?, String?> = withContext(Dispatchers.IO) {
        runCatching {
            EntryFormCodec.parse(json).also { form ->
                val seededForm = if (form.questions.isEmpty()) {
                    form.copy(questions = defaultQuestions())
                } else form
                database.saveEntryForm(seededForm)
            }
        }.fold({ it to null }, { null to (it.message ?: "フォームJSONを読み込めませんでした") })
    }

    private suspend fun sendPush(contents: String, pushTitle: String): String? {
        val csrf = (sessionManager.getCsrf() ?: preferences.getString("csrf", null))
            ?.takeIf(String::isNotBlank)
            ?: return "サインインが必要です"
        val channelID = preferences.getString("channelID", null)
            ?.takeIf(String::isNotBlank)
            ?: return "送信先チャネルが設定されていません"
        val updatedBy = (preferences.getString("entryFormUpdatedBy", null)
            ?: sessionManager.getNickname() ?: preferences.getString("nickname", null))
            ?.takeIf(String::isNotBlank)
            ?: return "送信者名が設定されていません"
        val pushNames = getPushNames().split(',').map(String::trim).filter(String::isNotBlank)
            .ifEmpty { listOf(updatedBy) }

        return try {
            val response = entryFormApiService.sendEntryFormPush(
                pushNames = gson.toJson(pushNames).part(),
                channelID = channelID.part(),
                updatedBy = updatedBy.part(),
                csrf = csrf.part(),
                contents = contents.part(),
                pushTitle = pushTitle.part()
            )
            if (!response.isSuccessful) return "送信に失敗しました (${response.code()})"
            val body = response.body() ?: return "サーバーから空の応答が返されました"
            // ContentsPush はセッション確認に成功していればエラー応答でも CSRF を回転させて
            // 返す（common/response.go の WriteResponseWithSession）。エラーで早期 return
            // する前に保存しないと、回転後の値を取りこぼして以降の API が失敗し続ける。
            sessionManager.applyResponseCsrf(csrf, body.csrf)?.let { rotatedCsrf ->
                preferences.edit().putString("csrf", rotatedCsrf).apply()
            }
            if (!body.error.isNullOrBlank()) return body.error
            if (body.csrf.isNullOrBlank()) return "CSRFトークンを更新できませんでした"
            null
        } catch (error: Exception) {
            error.message ?: "送信に失敗しました"
        }
    }

    private fun appendOptionKeys(target: JsonArray, value: Any?) {
        when (value) {
            is String -> target.add(value)
            is Iterable<*> -> value.forEach { item -> item?.toString()?.let(target::add) }
            is Array<*> -> value.forEach { item -> item?.toString()?.let(target::add) }
            null -> Unit
            else -> target.add(value.toString())
        }
    }

    private fun defaultQuestions() = listOf(
        Question.SingleChoice(1, "", emptyList()),
        Question.MultiChoice(2, "", emptyList(), false),
        Question.Text(3, ""),
        Question.DatePicker(4, "", 1, false)
    )

    private fun String.part() = toRequestBody("text/plain".toMediaType())
}

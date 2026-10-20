package com.chat.android.feature.ticket

import com.chat.android.core.network.ApiService
import com.chat.android.core.network.SessionManager
import com.chat.android.core.network.Ticket
import com.chat.android.core.push.PushReceiveDispatcher
import kotlinx.coroutines.flow.Flow
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.asSharedFlow
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.RequestBody
import okhttp3.RequestBody.Companion.toRequestBody
import org.json.JSONArray
import javax.inject.Inject
import javax.inject.Singleton

@Singleton
class TicketRepository @Inject constructor(
    private val apiService: ApiService,
    private val ticketApiService: TicketApiService,
    private val dbHelper: TicketDbHelper,
    private val sessionManager: SessionManager,
    private val pushDispatcher: PushReceiveDispatcher
) {
    private val _dbUpdateFlow = MutableSharedFlow<Unit>(replay = 0)
    val dbUpdateFlow = _dbUpdateFlow.asSharedFlow()

    suspend fun fetchTickets(): Result<List<Ticket>> = sessionManager.serializeAuthenticatedRequest {
        val csrf = sessionManager.getCsrf() ?: ""
        return@serializeAuthenticatedRequest try {
            val response = apiService.getTickets(csrf)
            if (response.isSuccessful) {
                val tickets = response.body() ?: emptyList()
                tickets.forEach { dbHelper.saveTicket(it) }
                _dbUpdateFlow.emit(Unit)
                Result.success(tickets)
            } else {
                Result.failure(Exception("Failed to fetch tickets: ${response.code()}"))
            }
        } catch (e: Exception) {
            Result.failure(e)
        }
    }

    fun getTickets(): List<Ticket> = dbHelper.getTickets()

    fun getTicket(ticketID: String): Ticket? = dbHelper.getTicket(ticketID)

    private fun toPart(value: String): RequestBody = value.toRequestBody("text/plain".toMediaTypeOrNull())

    suspend fun createTicket(
        channelID: String,
        updatedBy: String,
        title: String,
        description: String,
        status: String,
        priority: String
    ): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = ticketApiService.ticketPost(
            csrf = toPart(csrf),
            channelID = toPart(channelID),
            updatedBy = toPart(updatedBy),
            pushNames = toPart("[]"),
            title = toPart(title),
            description = toPart(description),
            status = toPart(status),
            priority = toPart(priority)
        )
        return@serializeAuthenticatedRequest handleResponse(response, csrf)
    }

    suspend fun updateTicket(
        channelID: String,
        updatedBy: String,
        ticketID: String,
        title: String?,
        description: String?,
        status: String?,
        priority: String?
    ): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = ticketApiService.ticketEdit(
            csrf = toPart(csrf),
            channelID = toPart(channelID),
            updatedBy = toPart(updatedBy),
            pushNames = toPart("[]"),
            ticketID = toPart(ticketID),
            title = title?.let { toPart(it) },
            description = description?.let { toPart(it) },
            status = status?.let { toPart(it) },
            priority = priority?.let { toPart(it) }
        )
        return@serializeAuthenticatedRequest handleResponse(response, csrf)
    }

    suspend fun deleteTicket(
        channelID: String,
        updatedBy: String,
        ticketID: String
    ): Result<Unit> = sessionManager.serializeAuthenticatedRequest {
        val csrf = sessionManager.getCsrf() ?: ""
        val response = ticketApiService.ticketDelete(
            csrf = toPart(csrf),
            channelID = toPart(channelID),
            updatedBy = toPart(updatedBy),
            pushNames = toPart("[]"),
            ticketID = toPart(ticketID)
        )
        return@serializeAuthenticatedRequest handleResponse(response, csrf)
    }

    private suspend fun handleResponse(response: retrofit2.Response<com.chat.android.core.network.ApiResponse<Any>>, sentCsrf: String): Result<Unit> {
        if (!response.isSuccessful) {
            return Result.failure(Exception("Network error: ${response.code()}"))
        }
        val body = response.body() ?: return Result.failure(Exception("Empty response"))
        sessionManager.applyResponseCsrf(sentCsrf, body.csrf)
        if (body.success == false) {
            return Result.failure(Exception(body.error ?: "Unknown error"))
        }
        body.pushContents?.forEach { pushDispatcher.receive(it) }
        _dbUpdateFlow.emit(Unit)
        return Result.success(Unit)
    }
}

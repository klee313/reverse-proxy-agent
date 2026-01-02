package com.rpa.android.core

import android.content.Context
import com.rpa.android.data.LogStore
import kotlinx.coroutines.flow.MutableSharedFlow
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asSharedFlow
import java.io.File
import java.time.Instant
import java.time.ZoneId
import java.time.format.DateTimeFormatter

object ServiceEvents {
    private const val MAX_LOGS = 200
    private val formatter = DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss").withZone(ZoneId.systemDefault())

    private val _status = MutableStateFlow(ServiceStatus())
    private val _logs = MutableStateFlow<List<LogLine>>(emptyList())
    private val _events = MutableSharedFlow<LogLine>(extraBufferCapacity = 50)

    val status: StateFlow<ServiceStatus> = _status
    val logs: StateFlow<List<LogLine>> = _logs
    val events = _events.asSharedFlow()

    fun init(context: android.content.Context) {
        LogStore.init(context)
        val lines = LogStore.loadRecent(MAX_LOGS).mapNotNull { parseLine(it) }
        _logs.value = lines
    }

    fun updateStatus(status: ServiceStatus) {
        _status.value = status
    }

    fun log(level: String, message: String) {
        val line = LogLine(formatter.format(Instant.now()), level, message)
        _events.tryEmit(line)
        val next = (_logs.value + line).takeLast(MAX_LOGS)
        _logs.value = next
        LogStore.append("${line.timestamp}|${line.level}|${line.message}")
    }

    fun loadLogPage(offsetFromEnd: Int, pageSize: Int): List<LogLine> {
        return LogStore.loadPage(offsetFromEnd, pageSize).mapNotNull { parseLine(it) }
    }

    fun exportLogs(context: Context): File? {
        return LogStore.exportToCache(context)
    }

    private fun parseLine(raw: String): LogLine? {
        val parts = raw.split("|", limit = 3)
        if (parts.size < 3) return null
        return LogLine(parts[0], parts[1], parts[2])
    }
}

package com.rpa.android

import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import java.time.Instant

object MetricsStore {
    private val _metrics = MutableStateFlow(MetricsSnapshot())
    val metrics: StateFlow<MetricsSnapshot> = _metrics

    fun onStartAttempt() {
        update { it.copy(startAttemptTotal = it.startAttemptTotal + 1) }
    }

    fun onStartSuccess() {
        update { it.copy(startSuccessTotal = it.startSuccessTotal + 1) }
    }

    fun onStartFailure() {
        update { it.copy(startFailureTotal = it.startFailureTotal + 1) }
    }

    fun onExitSuccess() {
        update { it.copy(exitSuccessTotal = it.exitSuccessTotal + 1) }
    }

    fun onExitFailure() {
        update { it.copy(exitFailureTotal = it.exitFailureTotal + 1) }
    }

    fun onRestart() {
        update { it.copy(restartTotal = it.restartTotal + 1) }
    }

    fun onRestartScheduled(delayMs: Long) {
        update { it.copy(backoffMs = delayMs) }
    }

    fun onState(state: ServiceState) {
        update { it.copy(state = state) }
    }

    fun onUptimeStart() {
        update { it.copy(uptimeStart = Instant.now().epochSecond) }
    }

    fun onLastSuccess() {
        update { it.copy(lastSuccessUnix = Instant.now().epochSecond) }
    }

    fun onLastExit(reason: String) {
        update { it.copy(lastExit = reason) }
    }

    fun onLastClass(value: String) {
        update { it.copy(lastClass = value) }
    }

    fun onLastTrigger(reason: String) {
        update { it.copy(lastTrigger = reason) }
    }

    private fun update(transform: (MetricsSnapshot) -> MetricsSnapshot) {
        _metrics.value = transform(_metrics.value)
    }
}

data class MetricsSnapshot(
    val state: ServiceState = ServiceState.STOPPED,
    val restartTotal: Int = 0,
    val startAttemptTotal: Int = 0,
    val startSuccessTotal: Int = 0,
    val startFailureTotal: Int = 0,
    val exitSuccessTotal: Int = 0,
    val exitFailureTotal: Int = 0,
    val lastSuccessUnix: Long? = null,
    val lastExit: String = "-",
    val lastClass: String = "-",
    val lastTrigger: String = "-",
    val backoffMs: Long? = null,
    val uptimeStart: Long? = null
) {
    fun toItems(): List<MetricItem> {
        return buildList {
            add(MetricItem("rpa_client_state", state.ordinal.toString()))
            add(MetricItem("rpa_client_restart_total", restartTotal.toString()))
            add(MetricItem("rpa_client_start_attempt_total", startAttemptTotal.toString()))
            add(MetricItem("rpa_client_start_success_total", startSuccessTotal.toString()))
            add(MetricItem("rpa_client_start_failure_total", startFailureTotal.toString()))
            add(MetricItem("rpa_client_exit_success_total", exitSuccessTotal.toString()))
            add(MetricItem("rpa_client_exit_failure_total", exitFailureTotal.toString()))
            add(MetricItem("rpa_client_last_exit", lastExit))
            add(MetricItem("rpa_client_last_class", lastClass))
            add(MetricItem("rpa_client_last_trigger", lastTrigger))
            lastSuccessUnix?.let { add(MetricItem("rpa_client_last_success_unix", it.toString())) }
            backoffMs?.let { add(MetricItem("rpa_client_backoff_ms", it.toString())) }
            uptimeSeconds()?.let { add(MetricItem("rpa_client_uptime_sec", it.toString())) }
        }
    }
    fun uptimeSeconds(now: Long = Instant.now().epochSecond): Long? {
        val start = uptimeStart ?: return null
        return now - start
    }
}

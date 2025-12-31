package com.rpa.android.core

data class LogLine(
    val timestamp: String,
    val level: String,
    val message: String,
)

data class MetricItem(
    val key: String,
    val value: String,
)

data class DoctorItem(
    val title: String,
    val status: String,
    val detail: String,
)

data class ServiceStatus(
    val state: ServiceState = ServiceState.STOPPED,
    val detail: String = "-",
)

enum class ServiceState(val label: String) {
    STOPPED("Stopped"),
    CONNECTING("Connecting"),
    RUNNING("Running"),
}

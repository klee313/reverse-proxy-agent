package com.rpa.android.data

import android.content.Context
import org.yaml.snakeyaml.Yaml
import java.io.File

object ConfigStore {
    private const val CONFIG_FILE = "rpa.yaml"

    fun loadText(context: Context): String {
        val file = File(context.filesDir, CONFIG_FILE)
        if (!file.exists()) {
            return defaultTemplate()
        }
        return file.readText()
    }

    fun saveText(context: Context, text: String) {
        val file = File(context.filesDir, CONFIG_FILE)
        file.writeText(text)
    }

    fun parse(text: String): RpaConfig {
        val yaml = Yaml()
        val data = yaml.load<Any>(text)
        if (data !is Map<*, *>) {
            throw IllegalArgumentException("invalid config format")
        }
        val sshMap = data["ssh"] as? Map<*, *> ?: emptyMap<String, Any>()
        val clientMap = data["client"] as? Map<*, *> ?: emptyMap<String, Any>()

        val user = (sshMap["user"] as? String)?.trim().orEmpty()
        val host = (sshMap["host"] as? String)?.trim().orEmpty()
        val port = (sshMap["port"] as? Int) ?: (sshMap["port"] as? String)?.toIntOrNull() ?: 22

        if (user.isBlank()) {
            throw IllegalArgumentException("ssh.user is required")
        }
        if (host.isBlank()) {
            throw IllegalArgumentException("ssh.host is required")
        }
        if (port <= 0) {
            throw IllegalArgumentException("ssh.port must be > 0")
        }

        val restartMap = clientMap["restart"] as? Map<*, *> ?: emptyMap<String, Any>()
        val minDelayMs = (restartMap["min_delay_ms"] as? Int)
            ?: (restartMap["min_delay_ms"] as? String)?.toIntOrNull()
            ?: 2000
        val maxDelayMs = (restartMap["max_delay_ms"] as? Int)
            ?: (restartMap["max_delay_ms"] as? String)?.toIntOrNull()
            ?: 30000
        val factor = (restartMap["factor"] as? Double)
            ?: (restartMap["factor"] as? Int)?.toDouble()
            ?: (restartMap["factor"] as? String)?.toDoubleOrNull()
            ?: 2.0
        val jitter = (restartMap["jitter"] as? Double)
            ?: (restartMap["jitter"] as? Int)?.toDouble()
            ?: (restartMap["jitter"] as? String)?.toDoubleOrNull()
            ?: 0.2
        val debounceMs = (restartMap["debounce_ms"] as? Int)
            ?: (restartMap["debounce_ms"] as? String)?.toIntOrNull()
            ?: 2000
        val periodicRestartSec = (clientMap["periodic_restart_sec"] as? Int)
            ?: (clientMap["periodic_restart_sec"] as? String)?.toIntOrNull()
            ?: 3600
        val sleepCheckSec = (clientMap["sleep_check_sec"] as? Int)
            ?: (clientMap["sleep_check_sec"] as? String)?.toIntOrNull()
            ?: 5
        val sleepGapSec = (clientMap["sleep_gap_sec"] as? Int)
            ?: (clientMap["sleep_gap_sec"] as? String)?.toIntOrNull()
            ?: 30
        val networkPollSec = (clientMap["network_poll_sec"] as? Int)
            ?: (clientMap["network_poll_sec"] as? String)?.toIntOrNull()
            ?: 5

        val forwards = when (val raw = clientMap["local_forwards"]) {
            is List<*> -> raw.mapNotNull { it?.toString()?.trim() }.filter { it.isNotBlank() }
            is String -> listOf(raw.trim()).filter { it.isNotBlank() }
            else -> emptyList()
        }
        if (forwards.isEmpty()) {
            throw IllegalArgumentException("client.local_forwards is required")
        }

        return RpaConfig(
            ssh = SshConfig(user = user, host = host, port = port),
            client = ClientConfig(
                localForwards = forwards,
                restart = RestartConfig(
                    minDelayMs = minDelayMs,
                    maxDelayMs = maxDelayMs,
                    factor = factor,
                    jitter = jitter,
                    debounceMs = debounceMs
                ),
                periodicRestartSec = periodicRestartSec,
                sleepCheckSec = sleepCheckSec,
                sleepGapSec = sleepGapSec,
                networkPollSec = networkPollSec
            )
        )
    }

    private fun defaultTemplate(): String {
        return """
client:
  restart:
    min_delay_ms: 2000
    max_delay_ms: 30000
    factor: 2.0
    jitter: 0.2
    debounce_ms: 2000
  periodic_restart_sec: 3600
  sleep_check_sec: 5
  sleep_gap_sec: 30
  network_poll_sec: 5
  local_forwards:
    - "127.0.0.1:15432:127.0.0.1:5432"
ssh:
  user: "ubuntu"
  host: "example.com"
  port: 22
""".trimIndent()
    }
}

data class RpaConfig(
    val ssh: SshConfig,
    val client: ClientConfig
)

data class SshConfig(
    val user: String,
    val host: String,
    val port: Int
)

data class ClientConfig(
    val localForwards: List<String>,
    val restart: RestartConfig,
    val periodicRestartSec: Int,
    val sleepCheckSec: Int,
    val sleepGapSec: Int,
    val networkPollSec: Int
)

data class RestartConfig(
    val minDelayMs: Int,
    val maxDelayMs: Int,
    val factor: Double,
    val jitter: Double,
    val debounceMs: Int
)

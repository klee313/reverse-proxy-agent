package com.rpa.android

import android.content.Context
import java.io.File
import java.net.InetSocketAddress
import java.net.Socket

object DoctorChecks {
    fun run(context: Context): List<DoctorItem> {
        val configText = ConfigStore.loadText(context)
        val configResult = runCatching { ConfigStore.parse(configText) }
        val items = mutableListOf<DoctorItem>()

        if (configResult.isFailure) {
            items.add(
                DoctorItem(
                    "Config valid",
                    "ERROR",
                    configResult.exceptionOrNull()?.message ?: "Invalid config"
                )
            )
            return items
        }

        val config = configResult.getOrThrow()
        items.add(DoctorItem("Config valid", "OK", "rpa.yaml parsed"))

        val keyExists = File(context.filesDir, "keys/rpa_ed25519").exists()
        items.add(
            DoctorItem(
                "SSH key",
                if (keyExists) "OK" else "ERROR",
                if (keyExists) "private key present" else "missing private key"
            )
        )

        val forwardChecks = config.client.localForwards.map { spec ->
            val result = runCatching { canBindLocalForward(spec) }
            DoctorItem(
                "Forward ${spec}",
                if (result.getOrDefault(false)) "OK" else "ERROR",
                if (result.getOrDefault(false)) "port available" else (result.exceptionOrNull()?.message ?: "port in use")
            )
        }
        items.addAll(forwardChecks)

        val hostReachable = runCatching { canConnect(config.ssh.host, config.ssh.port, 2000) }.getOrDefault(false)
        items.add(
            DoctorItem(
                "SSH host reachable",
                if (hostReachable) "OK" else "WARN",
                if (hostReachable) "tcp connect ok" else "unable to connect"
            )
        )

        val knownHostsFile = File(context.filesDir, "known_hosts")
        items.add(
            DoctorItem(
                "Known hosts",
                if (knownHostsFile.exists()) "OK" else "WARN",
                if (knownHostsFile.exists()) "known_hosts present" else "no known_hosts yet"
            )
        )

        val lastClass = MetricsStore.metrics.value.lastClass
        if (lastClass != "-" && lastClass != "clean") {
            items.add(
                DoctorItem(
                    "Last failure class",
                    "WARN",
                    lastClass
                )
            )
        }

        return items
    }

    private fun canBindLocalForward(spec: String): Boolean {
        val parts = spec.split(":")
        if (parts.size != 4) {
            throw IllegalArgumentException("invalid forward spec")
        }
        val localHost = parts[0].ifBlank { "127.0.0.1" }
        val localPort = parts[1].toInt()
        val socket = java.net.ServerSocket()
        socket.reuseAddress = true
        socket.bind(InetSocketAddress(localHost, localPort))
        socket.close()
        return true
    }

    private fun canConnect(host: String, port: Int, timeoutMs: Int): Boolean {
        Socket().use { socket ->
            socket.connect(InetSocketAddress(host, port), timeoutMs)
            return true
        }
    }
}

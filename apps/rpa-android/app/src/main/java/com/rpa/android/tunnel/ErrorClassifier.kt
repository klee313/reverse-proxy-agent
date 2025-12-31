package com.rpa.android.tunnel

object ErrorClassifier {
    fun classify(message: String?): String {
        val text = message?.lowercase().orEmpty()
        return when {
            text.contains("auth") || text.contains("permission denied") -> "auth"
            text.contains("host key") || text.contains("known_hosts") -> "hostkey"
            text.contains("unknown host") || text.contains("name or service") || text.contains("unresolved") -> "dns"
            text.contains("refused") -> "refused"
            text.contains("timed out") || text.contains("timeout") -> "timeout"
            text.contains("network") || text.contains("route") -> "network"
            else -> "unknown"
        }
    }
}

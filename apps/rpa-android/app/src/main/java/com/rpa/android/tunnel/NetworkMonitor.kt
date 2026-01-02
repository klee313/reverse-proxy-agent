package com.rpa.android.tunnel

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import com.rpa.android.core.ServiceEvents

enum class NetworkEventType {
    AVAILABLE,
    LOST,
    CHANGED,
    DEGRADED,
}

data class NetworkEvent(
    val type: NetworkEventType,
    val message: String
)

class NetworkMonitor(
    context: Context,
    private val onEvent: (NetworkEvent) -> Unit
) {
    private val connectivityManager =
        context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager
    private var lastValidated: Boolean? = null

    private val callback = object : ConnectivityManager.NetworkCallback() {
        override fun onAvailable(network: Network) {
            onEvent(NetworkEvent(NetworkEventType.AVAILABLE, "network available"))
        }

        override fun onLost(network: Network) {
            onEvent(NetworkEvent(NetworkEventType.LOST, "network lost"))
        }

        override fun onCapabilitiesChanged(network: Network, networkCapabilities: NetworkCapabilities) {
            val validated = networkCapabilities.hasCapability(NetworkCapabilities.NET_CAPABILITY_VALIDATED)
            val last = lastValidated
            lastValidated = validated
            if (last == true && !validated) {
                onEvent(NetworkEvent(NetworkEventType.DEGRADED, "network degraded (validated=false)"))
                return
            }
            onEvent(NetworkEvent(NetworkEventType.CHANGED, "network changed"))
        }
    }

    fun register() {
        runCatching { connectivityManager.registerDefaultNetworkCallback(callback) }
            .onFailure { ServiceEvents.log("ERROR", "network callback failed: ${it.message}") }
    }

    fun unregister() {
        runCatching { connectivityManager.unregisterNetworkCallback(callback) }
    }
}

package com.rpa.android.tunnel

import android.content.Context
import android.net.ConnectivityManager
import android.net.Network
import android.net.NetworkCapabilities
import com.rpa.android.core.ServiceEvents

class NetworkMonitor(
    context: Context,
    private val onEvent: (String) -> Unit
) {
    private val connectivityManager =
        context.getSystemService(Context.CONNECTIVITY_SERVICE) as ConnectivityManager

    private val callback = object : ConnectivityManager.NetworkCallback() {
        override fun onAvailable(network: Network) {
            onEvent("network available")
        }

        override fun onLost(network: Network) {
            onEvent("network lost")
        }

        override fun onCapabilitiesChanged(network: Network, networkCapabilities: NetworkCapabilities) {
            onEvent("network changed")
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

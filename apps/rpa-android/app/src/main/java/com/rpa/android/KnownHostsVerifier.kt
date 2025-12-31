package com.rpa.android

import net.schmizz.sshj.common.KeyType
import net.schmizz.sshj.transport.verification.OpenSSHKnownHosts
import java.io.File
import java.security.PublicKey

class AcceptNewKnownHosts(private val file: File, private val onWarning: (String) -> Unit) : OpenSSHKnownHosts(file) {
    override fun hostKeyUnverifiableAction(hostname: String, key: PublicKey): Boolean {
        return try {
            val type = KeyType.fromKey(key)
            val entry = HostEntry(null, hostname, type, key, "rpa-android")
            entries().add(entry)
            write(entry)
            true
        } catch (e: Exception) {
            onWarning("failed to write known_hosts: ${e.message}")
            false
        }
    }

    override fun hostKeyChangedAction(hostname: String, key: PublicKey): Boolean {
        onWarning("host key changed for $hostname")
        return false
    }
}

package com.rpa.android

import org.bouncycastle.jce.provider.BouncyCastleProvider
import java.security.Security

object CryptoProvider {
    @Volatile
    private var installed = false

    fun install() {
        if (installed) {
            return
        }
        installed = true
        runCatching {
            val provider = BouncyCastleProvider()
            val name = provider.name
            if (Security.getProvider(name) != null) {
                Security.removeProvider(name)
            }
            Security.insertProviderAt(provider, 1)
        }
    }
}

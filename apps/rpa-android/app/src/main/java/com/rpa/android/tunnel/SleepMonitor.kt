package com.rpa.android.tunnel

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.launch

class SleepMonitor(
    private val scope: CoroutineScope,
    private val intervalMs: Long,
    private val gapMs: Long,
    private val onEvent: (String) -> Unit
) {
    private var job: Job? = null

    fun start() {
        if (intervalMs <= 0) {
            return
        }
        job?.cancel()
        job = scope.launch {
            var last = System.currentTimeMillis()
            while (true) {
                delay(intervalMs)
                val now = System.currentTimeMillis()
                if (now - last > gapMs) {
                    onEvent("wake")
                }
                last = now
            }
        }
    }

    fun stop() {
        job?.cancel()
        job = null
    }
}

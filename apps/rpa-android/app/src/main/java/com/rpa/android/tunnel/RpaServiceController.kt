package com.rpa.android.tunnel

import android.content.Context
import android.content.Intent
import androidx.core.content.ContextCompat

object RpaServiceController {
    fun start(context: Context) {
        val intent = Intent(context, RpaService::class.java)
        ContextCompat.startForegroundService(context, intent)
    }

    fun stop(context: Context) {
        val intent = Intent(context, RpaService::class.java).apply {
            action = RpaService.ACTION_STOP
        }
        ContextCompat.startForegroundService(context, intent)
    }
}

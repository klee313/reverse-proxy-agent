package com.rpa.android.tunnel

import android.app.PendingIntent
import android.content.Context
import android.content.Intent

object PendingIntentHelper {
    fun service(context: Context, requestCode: Int, intent: Intent): PendingIntent {
        return PendingIntent.getService(
            context,
            requestCode,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )
    }
}

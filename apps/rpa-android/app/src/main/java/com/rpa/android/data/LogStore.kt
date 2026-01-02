package com.rpa.android.data

import android.content.Context
import java.io.File

object LogStore {
    private const val FILE_NAME = "rpa.log"
    private var file: File? = null

    fun init(context: Context) {
        file = File(context.filesDir, FILE_NAME)
    }

    fun append(line: String) {
        file?.appendText(line + "\n")
    }

    fun loadRecent(maxLines: Int): List<String> {
        val target = file ?: return emptyList()
        if (!target.exists()) {
            return emptyList()
        }
        val lines = target.readLines()
        return lines.takeLast(maxLines)
    }

    fun loadPage(offsetFromEnd: Int, pageSize: Int): List<String> {
        val target = file ?: return emptyList()
        if (!target.exists()) {
            return emptyList()
        }
        val lines = target.readLines()
        if (lines.isEmpty()) {
            return emptyList()
        }
        val safeOffset = offsetFromEnd.coerceAtLeast(0)
        val endIndex = (lines.size - safeOffset).coerceAtLeast(0)
        val startIndex = (endIndex - pageSize).coerceAtLeast(0)
        if (startIndex >= endIndex) {
            return emptyList()
        }
        return lines.subList(startIndex, endIndex)
    }

    fun exportToCache(context: Context): File? {
        val source = file ?: File(context.filesDir, FILE_NAME)
        if (!source.exists()) {
            return null
        }
        val exportFile = File(context.cacheDir, "rpa-log-${System.currentTimeMillis()}.log")
        source.copyTo(exportFile, overwrite = true)
        return exportFile
    }
}

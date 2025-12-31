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
}

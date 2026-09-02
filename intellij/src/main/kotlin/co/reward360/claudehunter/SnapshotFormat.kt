package co.reward360.claudehunter

import kotlin.math.min
import kotlin.math.roundToInt

// Formatting helpers for status-widget rendering.
object SnapshotFormat {
    fun tokensCompact(tokens: Long): String = when {
        tokens >= 1_000_000 -> "%.1fM".format(tokens / 1_000_000.0)
        tokens >= 1_000 -> "${(tokens / 1_000.0).roundToInt()}k"
        else -> tokens.toString()
    }

    fun percentBar(percent: Double, widthCells: Int = 10): String {
        val clamped = min(100.0, maxOf(0.0, percent))
        val filled = ((clamped / 100.0) * widthCells).roundToInt()
        return "▓".repeat(filled) + "░".repeat(widthCells - filled)
    }

    fun costUSD(dollars: Double): String = when {
        dollars < 0.01 -> "$0"
        else -> "$" + "%.2f".format(dollars)
    }
}

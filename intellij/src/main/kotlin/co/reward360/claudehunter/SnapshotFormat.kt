package co.reward360.claudehunter

import kotlin.math.roundToInt

// Formatting helpers for status-widget rendering.
object SnapshotFormat {
    fun tokensCompact(tokens: Long): String = when {
        tokens >= 1_000_000 -> "%.1fM".format(tokens / 1_000_000.0)
        tokens >= 1_000 -> "${(tokens / 1_000.0).roundToInt()}k"
        else -> tokens.toString()
    }

    fun costUSD(dollars: Double): String = when {
        dollars < 0.01 -> "$0"
        else -> "$" + "%.2f".format(dollars)
    }
}

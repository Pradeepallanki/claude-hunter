package co.reward360.claudehunter

import com.google.gson.Gson
import com.google.gson.JsonSyntaxException
import com.intellij.openapi.diagnostic.Logger

// SnapshotParser decodes a single NDJSON line from the claude-hunter core.
// Uses the Gson build bundled with the IntelliJ Platform.
object SnapshotParser {
    private val logger = Logger.getInstance(SnapshotParser::class.java)
    private val gson = Gson()

    private data class RawSnapshot(
        val type: String? = null,
        val model: String? = null,
        val window5h: RawWindow? = null,
    )

    private data class RawWindow(
        val inputTokens: Long = 0,
        val outputTokens: Long = 0,
        val cacheCreateTokens: Long = 0,
        val cacheReadTokens: Long = 0,
        val effectiveTokens: Long = 0,
        val costUSD: Double = 0.0,
        val burnRatePerMinute: Double = 0.0,
        val percentOfCeilingEstimate: Double = 0.0,
        val perModel: List<RawModelBreakdown>? = null,
    )

    private data class RawModelBreakdown(
        val model: String = "",
        val inputTokens: Long = 0,
        val outputTokens: Long = 0,
        val cacheCreateTokens: Long = 0,
        val cacheReadTokens: Long = 0,
        val effectiveTokens: Long = 0,
        val costUSD: Double = 0.0,
    )

    fun parseLine(rawLine: String): Snapshot? {
        return try {
            val decoded = gson.fromJson(rawLine, RawSnapshot::class.java) ?: return null
            if (decoded.type != "snapshot" || decoded.window5h == null) return null
            val rawWindow = decoded.window5h
            Snapshot(
                model = decoded.model?.takeIf { it.isNotEmpty() },
                window = WindowSummary(
                    inputTokens = rawWindow.inputTokens,
                    outputTokens = rawWindow.outputTokens,
                    cacheCreateTokens = rawWindow.cacheCreateTokens,
                    cacheReadTokens = rawWindow.cacheReadTokens,
                    effectiveTokens = rawWindow.effectiveTokens,
                    costUSD = rawWindow.costUSD,
                    burnRatePerMinute = rawWindow.burnRatePerMinute,
                    percentOfCeilingEstimate = rawWindow.percentOfCeilingEstimate,
                    perModel = rawWindow.perModel.orEmpty().map { rawEntry ->
                        ModelBreakdown(
                            model = rawEntry.model,
                            inputTokens = rawEntry.inputTokens,
                            outputTokens = rawEntry.outputTokens,
                            cacheCreateTokens = rawEntry.cacheCreateTokens,
                            cacheReadTokens = rawEntry.cacheReadTokens,
                            effectiveTokens = rawEntry.effectiveTokens,
                            costUSD = rawEntry.costUSD,
                        )
                    },
                ),
            )
        } catch (parseError: JsonSyntaxException) {
            logger.warn("bad snapshot line: ${parseError.message}")
            null
        }
    }
}

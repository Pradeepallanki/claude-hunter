package co.reward360.claudehunter

// ModelBreakdown is one row of the per-model attribution table.
data class ModelBreakdown(
    val model: String,
    val inputTokens: Long,
    val outputTokens: Long,
    val cacheCreateTokens: Long,
    val cacheReadTokens: Long,
    val effectiveTokens: Long,
    val costUSD: Double,
)

// WindowSummary mirrors the window5h object emitted by the claude-hunter core.
data class WindowSummary(
    val inputTokens: Long,
    val outputTokens: Long,
    val cacheCreateTokens: Long,
    val cacheReadTokens: Long,
    val effectiveTokens: Long,
    val costUSD: Double,
    val burnRatePerMinute: Double,
    val percentOfCeilingEstimate: Double,
    val perModel: List<ModelBreakdown>,
)

data class SessionActivity(
    val sessionId: String,
    val model: String,
    val contextTokens: Long,
    val contextWindowSize: Long,
    val totalTokens: Long,
    val sidechainTurns: Long,
    val lastActiveAt: String,
)

data class Snapshot(
    val model: String?,
    val window: WindowSummary,
    val agents: List<SessionActivity> = emptyList(),
)

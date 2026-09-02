package co.reward360.claudehunter

// TooltipRenderer turns a Snapshot into the HTML string an IntelliJ status
// bar tooltip needs (plain \n is ignored by Swing tooltips).
object TooltipRenderer {
    fun render(snapshot: Snapshot): String {
        val window = snapshot.window
        return buildString {
            append("<html><body style='padding:2px 6px;'>")
            append("<b>Claude Hunter</b> &middot; ")
            append(escapeHtml(snapshot.model ?: "unknown model"))
            append("<br><br>")

            append("<b>5h window</b> &mdash; ")
            append(SnapshotFormat.tokensCompact(window.effectiveTokens))
            append(" effective (")
            append("%.1f".format(window.percentOfCeilingEstimate))
            append("% of ceiling)")
            append("<br><br>")

            append("<table cellpadding='1' cellspacing='0'>")
            appendRow("Input", SnapshotFormat.tokensCompact(window.inputTokens))
            appendRow("Output", SnapshotFormat.tokensCompact(window.outputTokens))
            appendRow("Cache write", SnapshotFormat.tokensCompact(window.cacheCreateTokens))
            appendRow(
                "Cache read",
                "${SnapshotFormat.tokensCompact(window.cacheReadTokens)} " +
                    "(&times;0.1 &rarr; ${SnapshotFormat.tokensCompact((window.cacheReadTokens * 0.1).toLong())})",
            )
            appendRow("Cost", SnapshotFormat.costUSD(window.costUSD))
            appendRow("Burn rate", "${window.burnRatePerMinute.toInt()} tok/min")
            append("</table>")

            if (window.perModel.isNotEmpty()) {
                append("<br><b>By model</b><br>")
                append("<table cellpadding='1' cellspacing='0'>")
                append("<tr><td style='padding-right:12px; color:#888;'>Model</td>")
                append("<td style='padding-right:12px; color:#888;'>Tokens</td>")
                append("<td style='color:#888;'>Cost</td></tr>")
                for (modelEntry in window.perModel) {
                    append("<tr><td style='padding-right:12px;'>")
                    append(escapeHtml(modelEntry.model))
                    append("</td><td style='padding-right:12px;'>")
                    append(SnapshotFormat.tokensCompact(modelEntry.effectiveTokens))
                    append("</td><td>")
                    append(SnapshotFormat.costUSD(modelEntry.costUSD))
                    append("</td></tr>")
                }
                append("</table>")
            }

            append("</body></html>")
        }
    }

    private fun StringBuilder.appendRow(label: String, value: String) {
        append("<tr><td style='padding-right:12px; color:#888;'>")
        append(label)
        append("</td><td>")
        append(value)
        append("</td></tr>")
    }

    private fun escapeHtml(raw: String): String =
        raw.replace("&", "&amp;")
            .replace("<", "&lt;")
            .replace(">", "&gt;")
}

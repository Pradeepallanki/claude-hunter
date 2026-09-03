package co.reward360.claudehunter

import com.intellij.openapi.ui.popup.JBPopup
import com.intellij.openapi.ui.popup.JBPopupFactory
import com.intellij.ui.awt.RelativePoint
import java.awt.BorderLayout
import java.awt.Dimension
import java.awt.FlowLayout
import java.awt.event.MouseEvent
import javax.swing.BorderFactory
import javax.swing.JButton
import javax.swing.JEditorPane
import javax.swing.JPanel
import javax.swing.JScrollPane
import kotlin.math.max
import kotlin.math.min

// DetailsPopup shows the current snapshot in a floating popup with a Close button.
object DetailsPopup {
    fun show(snapshot: Snapshot, mouseEvent: MouseEvent) {
        val htmlPane = JEditorPane("text/html", buildHtml(snapshot)).apply {
            isEditable = false
            border = BorderFactory.createEmptyBorder(8, 12, 8, 12)
        }
        val scrollPane = JScrollPane(htmlPane).apply {
            preferredSize = Dimension(380, 320)
            border = BorderFactory.createEmptyBorder()
        }

        var popupHolder: JBPopup? = null
        val closeButton = JButton("Close").apply {
            addActionListener { popupHolder?.cancel() }
        }
        val buttonBar = JPanel(FlowLayout(FlowLayout.RIGHT)).apply {
            add(closeButton)
        }

        val container = JPanel(BorderLayout()).apply {
            add(scrollPane, BorderLayout.CENTER)
            add(buttonBar, BorderLayout.SOUTH)
        }

        val popup = JBPopupFactory.getInstance()
            .createComponentPopupBuilder(container, closeButton)
            .setTitle("Claude Hunter")
            .setMovable(true)
            .setResizable(true)
            .setFocusable(true)
            .setRequestFocus(true)
            .setCancelOnClickOutside(false)
            .createPopup()
        popupHolder = popup
        popup.show(RelativePoint(mouseEvent))
    }

    private fun buildHtml(snapshot: Snapshot): String {
        val window = snapshot.window
        val percentUsed = window.percentOfCeilingEstimate
        val percentRemaining = min(100.0, max(0.0, 100.0 - percentUsed))
        val weightedCacheRead = (window.cacheReadTokens * 0.1).toLong()

        return buildString {
            append("<html><body style='font-family:sans-serif;'>")
            append("<h3 style='margin:0 0 4px 0;'>Claude Hunter</h3>")
            append("<div style='color:#888; margin-bottom:10px;'>")
            append(escapeHtml(snapshot.model ?: "unknown model"))
            append("</div>")

            append("<div style='color:#2ea043; font-size:14pt; margin-bottom:12px;'>")
            append("🔋 ${percentRemaining.toInt()}% remaining ")
            append("<span style='color:#888; font-size:9pt;'>(")
            append("%.1f".format(percentUsed))
            append("% of 5h ceiling used)</span></div>")

            append("<table cellpadding='2' cellspacing='0'>")
            appendRow("Input", SnapshotFormat.tokensCompact(window.inputTokens))
            appendRow("Output", SnapshotFormat.tokensCompact(window.outputTokens))
            appendRow("Cache write", SnapshotFormat.tokensCompact(window.cacheCreateTokens))
            appendRow(
                "Cache read",
                "${SnapshotFormat.tokensCompact(window.cacheReadTokens)} " +
                    "(&times;0.1 &rarr; ${SnapshotFormat.tokensCompact(weightedCacheRead)})",
            )
            appendRow("Effective", SnapshotFormat.tokensCompact(window.effectiveTokens))
            appendRow("Cost", SnapshotFormat.costUSD(window.costUSD))
            appendRow("Burn rate", "${window.burnRatePerMinute.toInt()} tok/min")
            append("</table>")

            append("<br><b>Active agents")
            if (snapshot.agents.isNotEmpty()) append(" (${snapshot.agents.size})")
            append("</b><br>")
            if (snapshot.agents.isEmpty()) {
                append("<div style='color:#888;'>No sessions active in the last 5 minutes.</div>")
            } else {
                append("<table cellpadding='2' cellspacing='0'>")
                append("<tr>")
                append("<td style='padding-right:12px; color:#888;'>Session</td>")
                append("<td style='padding-right:12px; color:#888;'>Model</td>")
                append("<td style='padding-right:12px; color:#888;'>Context</td>")
                append("<td style='padding-right:12px; color:#888;'>5h tokens</td>")
                append("<td style='color:#888;'>Last active</td>")
                append("</tr>")
                for (agent in snapshot.agents) {
                    val contextPct = if (agent.contextWindowSize > 0) {
                        (agent.contextTokens * 100 / agent.contextWindowSize).toInt()
                    } else 0
                    val contextLabel = if (agent.contextWindowSize > 0) {
                        "${SnapshotFormat.tokensCompact(agent.contextTokens)} / " +
                            "${SnapshotFormat.tokensCompact(agent.contextWindowSize)} (${contextPct}%)"
                    } else {
                        SnapshotFormat.tokensCompact(agent.contextTokens)
                    }
                    val badge = if (agent.sidechainTurns > 0) {
                        " <span style='background:#444; color:#ddd; padding:0 4px; border-radius:6px; font-size:8pt;'>+${agent.sidechainTurns} subagent</span>"
                    } else ""
                    append("<tr>")
                    append("<td style='padding-right:12px;'><tt>")
                    append(escapeHtml(agent.sessionId.take(8)))
                    append("</tt>")
                    append(badge)
                    append("</td>")
                    append("<td style='padding-right:12px;'>")
                    append(escapeHtml(shortenModel(agent.model)))
                    append("</td>")
                    append("<td style='padding-right:12px;'>")
                    append(contextLabel)
                    append("</td>")
                    append("<td style='padding-right:12px;'>")
                    append(SnapshotFormat.tokensCompact(agent.totalTokens))
                    append("</td>")
                    append("<td>")
                    append(relativeTime(agent.lastActiveAt))
                    append("</td>")
                    append("</tr>")
                }
                append("</table>")
            }

            if (window.perModel.isNotEmpty()) {
                append("<br><b>By model</b><br>")
                append("<table cellpadding='2' cellspacing='0'>")
                append("<tr><td style='padding-right:12px; color:#888;'>Model</td>")
                append("<td style='padding-right:12px; color:#888;'>Tokens</td>")
                append("<td style='color:#888;'>Cost</td></tr>")
                for (entry in window.perModel) {
                    append("<tr><td style='padding-right:12px;'>")
                    append(escapeHtml(entry.model))
                    append("</td><td style='padding-right:12px;'>")
                    append(SnapshotFormat.tokensCompact(entry.effectiveTokens))
                    append("</td><td>")
                    append(SnapshotFormat.costUSD(entry.costUSD))
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
        raw.replace("&", "&amp;").replace("<", "&lt;").replace(">", "&gt;")

    private fun shortenModel(model: String): String {
        if (model.isEmpty()) return "unknown"
        return model.removePrefix("claude-").replace(Regex("-\\d{8}$"), "")
    }

    private fun relativeTime(iso: String): String {
        return try {
            val then = java.time.Instant.parse(iso).toEpochMilli()
            val seconds = kotlin.math.max(0L, (System.currentTimeMillis() - then) / 1000L)
            when {
                seconds < 60 -> "${seconds}s ago"
                seconds < 3600 -> "${seconds / 60}m ago"
                else -> "${seconds / 3600}h ago"
            }
        } catch (_: Exception) {
            "—"
        }
    }
}

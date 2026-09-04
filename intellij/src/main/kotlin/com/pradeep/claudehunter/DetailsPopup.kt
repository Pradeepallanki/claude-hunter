package com.pradeep.claudehunter

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
import javax.swing.JTabbedPane
import kotlin.math.max
import kotlin.math.min

// DetailsPopup shows the current snapshot in a tabbed floating popup.
object DetailsPopup {
    fun show(snapshot: Snapshot, mouseEvent: MouseEvent) {
        val tabbedPane = JTabbedPane().apply {
            addTab("Overview", tabPanel(overviewHtml(snapshot)))
            addTab("Agents", tabPanel(agentsHtml(snapshot.agents)))
            addTab("Projects", tabPanel(projectsHtml(snapshot.projects)))
            addTab("Breakdown", tabPanel(breakdownHtml(snapshot)))
            preferredSize = Dimension(460, 380)
        }

        var popupHolder: JBPopup? = null
        val closeButton = JButton("Close").apply {
            addActionListener { popupHolder?.cancel() }
        }
        val buttonBar = JPanel(FlowLayout(FlowLayout.RIGHT)).apply {
            add(closeButton)
        }
        val container = JPanel(BorderLayout()).apply {
            add(tabbedPane, BorderLayout.CENTER)
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

    private fun tabPanel(bodyHtml: String): JScrollPane {
        val pane = JEditorPane(
            "text/html",
            "<html><body style='font-family:sans-serif; padding:0 4px;'>$bodyHtml</body></html>",
        ).apply {
            isEditable = false
            border = BorderFactory.createEmptyBorder(8, 12, 8, 12)
        }
        return JScrollPane(pane).apply {
            border = BorderFactory.createEmptyBorder()
        }
    }

    private fun overviewHtml(snapshot: Snapshot): String {
        val window = snapshot.window
        val percentUsed = window.percentOfCeilingEstimate
        val percentRemaining = min(100.0, max(0.0, 100.0 - percentUsed))
        val batteryColor = when {
            percentRemaining <= 10 -> "#c14545"
            percentRemaining <= 20 -> "#c1a545"
            else -> "#2ea043"
        }
        val etaText = formatEta(window.secondsToLimit)

        val contextWarning = snapshot.agents.firstOrNull {
            it.contextWindowSize > 0 &&
                it.contextTokens.toDouble() / it.contextWindowSize >= 0.8
        }
        val warningBanner = if (contextWarning != null) {
            val pct = (contextWarning.contextTokens * 100 / contextWarning.contextWindowSize).toInt()
            val bg = if (pct >= 95) "#5a2020" else "#5a4820"
            "<div style='background:$bg; padding:8px 12px; border-radius:4px; margin-bottom:12px;'>" +
                "⚠ Session <tt>${escapeHtml(contextWarning.sessionId.take(8))}</tt> context at $pct% of " +
                "${SnapshotFormat.tokensCompact(contextWarning.contextWindowSize)} — consider <tt>/compact</tt>." +
                "</div>"
        } else ""

        return buildString {
            append("<h3 style='margin:0 0 4px 0;'>Claude Hunter</h3>")
            append("<div style='color:#888; margin-bottom:14px;'>")
            append(escapeHtml(snapshot.model ?: "unknown model"))
            append("</div>")
            append("<div style='color:$batteryColor; font-size:16pt; margin-bottom:6px;'>")
            append("🔋 ${percentRemaining.toInt()}% remaining")
            append("</div>")
            append("<div style='color:#888; margin-bottom:14px;'>")
            append(etaText)
            append(" · ")
            append("%.1f".format(percentUsed))
            append("% of 5h ceiling used")
            append("</div>")
            append(warningBanner)
            append("<table cellpadding='2' cellspacing='0'>")
            appendRow("Effective tokens", SnapshotFormat.tokensCompact(window.effectiveTokens))
            appendRow("Cost", SnapshotFormat.costUSD(window.costUSD))
            appendRow("Burn rate", "${window.burnRatePerMinute.toInt()} tok/min")
            append("</table>")
        }
    }

    private fun agentsHtml(agents: List<SessionActivity>): String {
        if (agents.isEmpty()) {
            return "<h3 style='margin:0 0 6px 0;'>Active agents</h3>" +
                "<div style='color:#888;'>No sessions active in the last 5 minutes.</div>"
        }
        return buildString {
            append("<h3 style='margin:0 0 6px 0;'>Active agents (${agents.size})</h3>")
            append("<table cellpadding='2' cellspacing='0'>")
            append("<tr>")
            append("<td style='padding-right:12px; color:#888;'>Session</td>")
            append("<td style='padding-right:12px; color:#888;'>Model</td>")
            append("<td style='padding-right:12px; color:#888;'>Context</td>")
            append("<td style='padding-right:12px; color:#888;'>5h tokens</td>")
            append("<td style='color:#888;'>Last active</td>")
            append("</tr>")
            for (agent in agents) {
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
    }

    private fun projectsHtml(projects: List<ProjectActivity>): String {
        if (projects.isEmpty()) {
            return "<h3 style='margin:0 0 6px 0;'>Projects</h3>" +
                "<div style='color:#888;'>No project activity in the current window.</div>"
        }
        return buildString {
            append("<h3 style='margin:0 0 6px 0;'>Projects (${projects.size})</h3>")
            append("<table cellpadding='2' cellspacing='0'>")
            append("<tr>")
            append("<td style='padding-right:12px; color:#888;'>Project</td>")
            append("<td style='padding-right:12px; color:#888;'>Sessions</td>")
            append("<td style='padding-right:12px; color:#888;'>Tokens</td>")
            append("<td style='padding-right:12px; color:#888;'>Cost</td>")
            append("<td style='color:#888;'>Last active</td>")
            append("</tr>")
            for (project in projects) {
                append("<tr>")
                append("<td style='padding-right:12px;'>")
                append(escapeHtml(project.project))
                append("</td>")
                append("<td style='padding-right:12px;'>")
                append(project.sessions.toString())
                append("</td>")
                append("<td style='padding-right:12px;'>")
                append(SnapshotFormat.tokensCompact(project.totalTokens))
                append("</td>")
                append("<td style='padding-right:12px;'>")
                append(SnapshotFormat.costUSD(project.costUSD))
                append("</td>")
                append("<td>")
                append(relativeTime(project.lastActiveAt))
                append("</td>")
                append("</tr>")
            }
            append("</table>")
        }
    }

    private fun breakdownHtml(snapshot: Snapshot): String {
        val window = snapshot.window
        val weightedCacheRead = (window.cacheReadTokens * 0.1).toLong()
        return buildString {
            append("<h3 style='margin:0 0 6px 0;'>Tokens</h3>")
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
            append("</table>")

            if (window.perModel.isNotEmpty()) {
                append("<h3 style='margin:14px 0 6px 0;'>By model</h3>")
                append("<table cellpadding='2' cellspacing='0'>")
                append("<tr>")
                append("<td style='padding-right:12px; color:#888;'>Model</td>")
                append("<td style='padding-right:12px; color:#888;'>Tokens</td>")
                append("<td style='color:#888;'>Cost</td>")
                append("</tr>")
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
        }
    }

    private fun StringBuilder.appendRow(label: String, value: String) {
        append("<tr><td style='padding-right:12px; color:#888;'>")
        append(label)
        append("</td><td>")
        append(value)
        append("</td></tr>")
    }

    private fun formatEta(secondsToLimit: Long): String {
        if (secondsToLimit < 0) return "ETA to limit: — (no burn detected)"
        if (secondsToLimit == 0L) return "ETA to limit: limit reached"
        val minutes = (secondsToLimit + 30) / 60
        if (minutes < 60) return "≈${minutes}m to limit at current pace"
        val hours = minutes / 60
        val remainderMinutes = minutes % 60
        return "≈${hours}h ${remainderMinutes}m to limit at current pace"
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

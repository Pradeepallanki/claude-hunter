package com.pradeep.claudehunter

import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.StatusBar
import com.intellij.openapi.wm.StatusBarWidget
import com.intellij.util.Consumer
import java.awt.event.MouseEvent
import kotlin.math.max
import kotlin.math.min

// StatusWidget renders the most recent Snapshot as a status-bar text item.
class StatusWidget(private val hostProject: Project) : StatusBarWidget, StatusBarWidget.TextPresentation {
    @Volatile private var latestSnapshot: Snapshot? = null
    private var attachedStatusBar: StatusBar? = null
    private var hunterProcess: HunterProcess? = null

    override fun ID(): String = "ClaudeHunterStatusWidget"

    override fun install(statusBar: StatusBar) {
        attachedStatusBar = statusBar

        val binaryFile = BinaryLocator.locate()
            ?: run {
                latestSnapshot = null
                statusBar.updateWidget(ID())
                return
            }
        hunterProcess = HunterProcess(binaryFile) { snapshot ->
            latestSnapshot = snapshot
            attachedStatusBar?.updateWidget(ID())
        }.also { it.start() }
    }

    override fun dispose() {
        hunterProcess?.stop()
        hunterProcess = null
        attachedStatusBar = null
    }

    override fun getPresentation(): StatusBarWidget.WidgetPresentation = this

    override fun getText(): String {
        val snapshot = latestSnapshot ?: return "Claude ⏳"
        val window = snapshot.window
        val percentRemaining = min(100.0, max(0.0, 100.0 - window.percentOfCeilingEstimate))
        return shortenModelName(snapshot.model) + " " +
            "🔋 ${percentRemaining.toInt()}% · " +
            SnapshotFormat.tokensCompact(window.effectiveTokens) + "/5h · " +
            SnapshotFormat.costUSD(window.costUSD)
    }

    private fun shortenModelName(model: String?): String {
        if (model.isNullOrEmpty()) return "Claude"
        return model.removePrefix("claude-").replace(Regex("-\\d{8}$"), "")
    }

    override fun getTooltipText(): String {
        val snapshot = latestSnapshot ?: return "Claude Hunter starting…"
        return TooltipRenderer.render(snapshot)
    }

    override fun getAlignment(): Float = 0f

    override fun getClickConsumer(): Consumer<MouseEvent> = Consumer { mouseEvent ->
        val snapshot = latestSnapshot ?: return@Consumer
        DetailsPopup.show(snapshot, mouseEvent)
    }
}

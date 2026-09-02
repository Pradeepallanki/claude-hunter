package co.reward360.claudehunter

import com.intellij.openapi.project.Project
import com.intellij.openapi.wm.StatusBarWidget
import com.intellij.openapi.wm.StatusBarWidgetFactory

class StatusWidgetFactory : StatusBarWidgetFactory {
    override fun getId(): String = "ClaudeHunterStatusWidget"
    override fun getDisplayName(): String = "Claude Hunter"
    override fun isAvailable(project: Project): Boolean = true
    override fun createWidget(project: Project): StatusBarWidget = StatusWidget(project)
    override fun disposeWidget(widget: StatusBarWidget) { widget.dispose() }
    override fun canBeEnabledOn(statusBar: com.intellij.openapi.wm.StatusBar): Boolean = true
}

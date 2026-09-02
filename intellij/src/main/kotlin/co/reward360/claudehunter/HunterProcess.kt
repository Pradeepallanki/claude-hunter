package co.reward360.claudehunter

import com.intellij.openapi.diagnostic.Logger
import java.io.BufferedReader
import java.io.File
import java.io.InputStreamReader
import kotlin.concurrent.thread

// HunterProcess spawns the claude-hunter binary and delivers each parsed
// snapshot to a listener callback. It is single-use — call stop() to end.
class HunterProcess(
    private val binaryFile: File,
    private val onSnapshot: (Snapshot) -> Unit,
) {
    private val logger = Logger.getInstance(HunterProcess::class.java)
    private var runningProcess: Process? = null

    fun start() {
        val processBuilder = ProcessBuilder(binaryFile.absolutePath)
            .redirectErrorStream(false)
        runningProcess = processBuilder.start()
        val startedProcess = runningProcess ?: return

        thread(isDaemon = true, name = "claude-hunter-stdout") {
            BufferedReader(InputStreamReader(startedProcess.inputStream)).use { stdoutReader ->
                stdoutReader.lineSequence().forEach { rawLine ->
                    val decoded = SnapshotParser.parseLine(rawLine)
                    if (decoded != null) onSnapshot(decoded)
                }
            }
        }
        thread(isDaemon = true, name = "claude-hunter-stderr") {
            BufferedReader(InputStreamReader(startedProcess.errorStream)).use { stderrReader ->
                stderrReader.lineSequence().forEach { logger.warn("[claude-hunter] $it") }
            }
        }
    }

    fun stop() {
        runningProcess?.destroy()
        runningProcess = null
    }
}

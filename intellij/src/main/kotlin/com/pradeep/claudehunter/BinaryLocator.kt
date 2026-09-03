package com.pradeep.claudehunter

import com.intellij.openapi.diagnostic.Logger
import java.io.File
import java.nio.file.Files
import java.nio.file.Path
import java.nio.file.StandardCopyOption

// BinaryLocator resolves the claude-hunter binary path. Preference order:
//   1. An explicit path from the -Dclaude-hunter.binary-path JVM property.
//   2. The binary bundled under the JAR resource
//      /bin/<platform>-<arch>/claude-hunter, extracted to a temp cache and
//      marked executable.
object BinaryLocator {
    private val logger = Logger.getInstance(BinaryLocator::class.java)

    fun locate(): File? {
        val configuredPath = System.getProperty("claude-hunter.binary-path").orEmpty()
        if (configuredPath.isNotEmpty()) {
            val configuredFile = File(configuredPath)
            if (configuredFile.exists()) return configuredFile
        }
        return extractBundledBinary()
    }

    private fun extractBundledBinary(): File? {
        val resourcePath = "/bin/${platformFolder()}-${archFolder()}/${binaryFileName()}"
        val resourceStream = javaClass.getResourceAsStream(resourcePath) ?: run {
            logger.warn("no bundled binary at $resourcePath")
            return null
        }
        return resourceStream.use { stream ->
            val cacheDirectory = cacheDirectoryForBinary()
            Files.createDirectories(cacheDirectory)
            val extractedFile = cacheDirectory.resolve(binaryFileName())
            Files.copy(stream, extractedFile, StandardCopyOption.REPLACE_EXISTING)
            extractedFile.toFile().apply { setExecutable(true) }
        }
    }

    private fun cacheDirectoryForBinary(): Path {
        val tempRoot = System.getProperty("java.io.tmpdir")
        return Path.of(tempRoot, "claude-hunter", pluginVersionTag())
    }

    private fun pluginVersionTag(): String = "v0.1.0"

    private fun platformFolder(): String {
        val osName = System.getProperty("os.name").lowercase()
        return when {
            osName.contains("mac") -> "darwin"
            osName.contains("linux") -> "linux"
            osName.contains("windows") -> "win"
            else -> "unknown"
        }
    }

    private fun archFolder(): String {
        val arch = System.getProperty("os.arch").lowercase()
        return when {
            arch.contains("aarch64") || arch.contains("arm64") -> "arm64"
            arch.contains("64") -> "x64"
            else -> arch
        }
    }

    private fun binaryFileName(): String {
        val osName = System.getProperty("os.name").lowercase()
        return if (osName.contains("windows")) "claude-hunter.exe" else "claude-hunter"
    }
}

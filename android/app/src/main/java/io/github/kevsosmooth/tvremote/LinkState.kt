package io.github.kevsosmooth.tvremote

import android.os.Handler
import android.os.Looper

/** What the screen shows about the link. Changed by [LinkService], read on the main thread. */
object LinkState {
    enum class Phase { STOPPED, STARTING, WAITING, HELPER_JOINING, HELPER_CONNECTED }

    @Volatile var phase = Phase.STOPPED
        private set
    @Volatile var code: String? = null
        private set
    @Volatile var lastLine: String = ""
        private set

    private val main = Handler(Looper.getMainLooper())
    private val listeners = mutableSetOf<() -> Unit>()

    fun listen(listener: () -> Unit) { listeners += listener }
    fun unlisten(listener: () -> Unit) { listeners -= listener }

    fun update(phase: Phase = this.phase, code: String? = this.code, lastLine: String = this.lastLine) {
        this.phase = phase
        this.code = code
        this.lastLine = lastLine
        main.post { listeners.toList().forEach { it() } }
    }

    /** Maps one tvlink log line to the phase it means, or null if it changes nothing. */
    fun phaseFor(line: String): Phase? = when {
        "tvlink ready" in line -> Phase.WAITING
        "asking to connect" in line -> Phase.HELPER_JOINING
        line.endsWith(": connected") -> Phase.HELPER_CONNECTED
        line.endsWith(": disconnected") || line.endsWith(": failed") || line.endsWith(": closed") -> Phase.WAITING
        else -> null
    }
}

package io.github.kevsosmooth.tvremote

import android.app.Activity
import android.os.Bundle
import android.widget.Button
import android.widget.TextView
import java.net.InetSocketAddress
import java.net.Socket

/** One screen: is ADB on, the pairing code to share, and Start/Stop. */
class MainActivity : Activity() {
    private lateinit var status: TextView
    private lateinit var code: TextView
    private lateinit var adb: TextView
    private lateinit var detail: TextView
    private lateinit var toggle: Button
    private lateinit var newCode: Button
    private val render: () -> Unit = { show() }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        setContentView(R.layout.activity_main)
        status = findViewById(R.id.status)
        code = findViewById(R.id.code)
        adb = findViewById(R.id.adb)
        detail = findViewById(R.id.detail)
        toggle = findViewById(R.id.toggle)
        newCode = findViewById(R.id.new_code)
        toggle.setOnClickListener {
            if (LinkState.phase == LinkState.Phase.STOPPED) LinkService.start(this) else LinkService.stop(this)
        }
        newCode.setOnClickListener { LinkService.newCode(this) }
        toggle.requestFocus()
    }

    override fun onStart() {
        super.onStart()
        LinkState.listen(render)
        show()
        checkAdb()
    }

    override fun onStop() {
        LinkState.unlisten(render)
        super.onStop()
    }

    private fun show() {
        val phase = LinkState.phase
        status.setText(
            when (phase) {
                LinkState.Phase.STOPPED -> R.string.status_stopped
                LinkState.Phase.STARTING -> R.string.status_starting
                LinkState.Phase.WAITING -> R.string.status_waiting
                LinkState.Phase.HELPER_JOINING -> R.string.status_joining
                LinkState.Phase.HELPER_CONNECTED -> R.string.status_connected
            },
        )
        code.text = LinkState.code?.let(::grouped) ?: getString(R.string.code_none)
        detail.text = LinkState.lastLine
        toggle.setText(if (phase == LinkState.Phase.STOPPED) R.string.start else R.string.stop)
        newCode.isEnabled = phase != LinkState.Phase.STOPPED
    }

    /** tvlink can only help if the TV's adb answers on 127.0.0.1:5555. */
    private fun checkAdb() {
        adb.setText(R.string.adb_checking)
        Thread {
            val on = runCatching {
                Socket().use { it.connect(InetSocketAddress("127.0.0.1", 5555), 1000) }
            }.isSuccess
            runOnUiThread { adb.setText(if (on) R.string.adb_on else R.string.adb_off) }
        }.start()
    }

    /** "abcdefgh..." -> "abcd efgh ...", easier to read aloud. Spaces are ignored by tvctl. */
    private fun grouped(raw: String) = raw.chunked(4).joinToString(" ")
}

package io.github.kevsosmooth.tvremote

import android.app.Notification
import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.app.Service
import android.content.Context
import android.content.Intent
import android.content.pm.ServiceInfo
import android.os.Build
import android.os.IBinder
import java.io.File

/**
 * Runs tvlink (bundled as libtvlink.so) while the TV owner keeps remote help on.
 * tvlink forwards a helper holding the pairing code to the TV's own adb.
 */
class LinkService : Service() {
    @Volatile private var process: Process? = null

    override fun onBind(intent: Intent?): IBinder? = null

    override fun onStartCommand(intent: Intent?, flags: Int, startId: Int): Int {
        when (intent?.action) {
            ACTION_STOP -> {
                stopLink()
                stopSelf()
            }
            ACTION_NEW_CODE -> {
                stopLink()
                codeFile().delete()
                startLink()
            }
            else -> startLink()
        }
        return START_NOT_STICKY
    }

    override fun onDestroy() {
        stopLink()
        super.onDestroy()
    }

    private fun startLink() {
        if (process != null) return
        goForeground()
        LinkState.update(phase = LinkState.Phase.STARTING, code = null, lastLine = "")
        val binary = File(applicationInfo.nativeLibraryDir, "libtvlink.so")
        val started = try {
            ProcessBuilder(
                binary.path,
                "--code-file", codeFile().path,
                "--target", ADB_TARGET,
                "--android-sdk", Build.VERSION.SDK_INT.toString(),
            )
                .redirectErrorStream(true)
                .start()
        } catch (e: Exception) {
            LinkState.update(phase = LinkState.Phase.STOPPED, lastLine = "Could not start tvlink: ${e.message}")
            stopSelf()
            return
        }
        process = started
        Thread({ readOutput(started) }, "tvlink-output").start()
    }

    private fun readOutput(p: Process) {
        p.inputStream.bufferedReader().forEachLine { line ->
            if (process !== p) return@forEachLine
            val phase = LinkState.phaseFor(line)
            val code = if (phase == LinkState.Phase.WAITING && LinkState.code == null) readCode() else LinkState.code
            LinkState.update(phase = phase ?: LinkState.phase, code = code, lastLine = line)
        }
        val exit = p.waitFor()
        if (process === p) {
            process = null
            LinkState.update(phase = LinkState.Phase.STOPPED, lastLine = "tvlink stopped (exit $exit). ${LinkState.lastLine}")
            stopSelf()
        }
    }

    private fun stopLink() {
        val p = process ?: return
        process = null
        p.destroy()
        LinkState.update(phase = LinkState.Phase.STOPPED, code = null, lastLine = "Stopped")
    }

    private fun readCode(): String? = runCatching { codeFile().readText().trim() }.getOrNull()

    private fun codeFile() = File(filesDir, "tvlink.code")

    private fun goForeground() {
        val nm = getSystemService(NotificationManager::class.java)
        nm.createNotificationChannel(
            NotificationChannel(CHANNEL, getString(R.string.app_name), NotificationManager.IMPORTANCE_LOW),
        )
        val open = PendingIntent.getActivity(
            this, 0, Intent(this, MainActivity::class.java), PendingIntent.FLAG_IMMUTABLE,
        )
        val notification = Notification.Builder(this, CHANNEL)
            .setSmallIcon(R.drawable.icon)
            .setContentTitle(getString(R.string.notification_title))
            .setContentIntent(open)
            .setOngoing(true)
            .build()
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.UPSIDE_DOWN_CAKE) {
            startForeground(NOTIFICATION_ID, notification, ServiceInfo.FOREGROUND_SERVICE_TYPE_SPECIAL_USE)
        } else {
            startForeground(NOTIFICATION_ID, notification)
        }
    }

    companion object {
        private const val ADB_TARGET = "127.0.0.1:5555"
        private const val CHANNEL = "link"
        private const val NOTIFICATION_ID = 1
        private const val ACTION_STOP = "stop"
        private const val ACTION_NEW_CODE = "new-code"

        fun start(context: Context) = send(context, null)
        fun stop(context: Context) = send(context, ACTION_STOP)
        fun newCode(context: Context) = send(context, ACTION_NEW_CODE)

        private fun send(context: Context, action: String?) {
            val intent = Intent(context, LinkService::class.java).setAction(action)
            if (action == ACTION_STOP) context.startService(intent) else context.startForegroundService(intent)
        }
    }
}

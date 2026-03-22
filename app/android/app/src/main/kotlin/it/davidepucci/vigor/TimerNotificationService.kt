package it.davidepucci.vigor

import android.app.*
import android.content.Context
import android.content.Intent
import android.os.Build
import androidx.core.app.NotificationCompat
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class TimerNotificationService {
    companion object {
        private const val CHANNEL_ID = "vigor_timer_channel"
        private const val NOTIFICATION_ID = 1001
        private const val CHANNEL_NAME = "Training Timer"

        fun createNotificationChannel(context: Context) {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
                val channel = NotificationChannel(
                    CHANNEL_ID,
                    CHANNEL_NAME,
                    NotificationManager.IMPORTANCE_HIGH
                ).apply {
                    description = "Live training timer updates"
                    setShowBadge(false)
                }

                val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
                manager.createNotificationChannel(channel)
            }
        }

        fun showLiveTimerNotification(
            context: Context,
            trainingName: String,
            totalElapsedSeconds: Int,
            contentText: String,
            intervalRemainingSeconds: Int,
            isDurationBased: Boolean,
            progress: Double,
            stopLabel: String,
            completeLabel: String,
            isPaused: Boolean = false
        ) {
            if (Build.VERSION.SDK_INT < Build.VERSION_CODES.UPSIDE_DOWN_CAKE) return

            val intent = Intent(context, MainActivity::class.java).apply {
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            val pendingIntent = PendingIntent.getActivity(
                context,
                0,
                intent,
                PendingIntent.FLAG_IMMUTABLE
            )

            // stop action
            val stopIntent = Intent(context, MainActivity::class.java).apply {
                action = "TIMER_STOP"
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            val stopPendingIntent = PendingIntent.getActivity(
                context,
                1,
                stopIntent,
                PendingIntent.FLAG_IMMUTABLE
            )

            // complete action
            val completeIntent = Intent(context, MainActivity::class.java).apply {
                action = "TIMER_COMPLETE"
                flags = Intent.FLAG_ACTIVITY_NEW_TASK or Intent.FLAG_ACTIVITY_CLEAR_TOP
            }
            val completePendingIntent = PendingIntent.getActivity(
                context,
                2,
                completeIntent,
                PendingIntent.FLAG_IMMUTABLE
            )

            // format total elapsed time for chip
            val totalMinutes = totalElapsedSeconds / 60
            val totalSeconds = totalElapsedSeconds % 60
            val totalElapsedTimeText = String.format("%02d:%02d", totalMinutes, totalSeconds)

            // format interval remaining time for content
            val remainingMinutes = intervalRemainingSeconds / 60
            val remainingSeconds = intervalRemainingSeconds % 60
            val remainingTimeText = String.format("%02d:%02d", remainingMinutes, remainingSeconds)

            // build content text with countdown if duration-based
            val fullContentText = if (isDurationBased) {
                "$contentText · $remainingTimeText"
            } else {
                contentText
            }

            val builder = NotificationCompat.Builder(context, CHANNEL_ID)
                .setContentTitle(trainingName)
                .setContentText(fullContentText)
                .setSmallIcon(android.R.drawable.ic_media_play)
                .setContentIntent(pendingIntent)
                .setOngoing(true)
                .setCategory(NotificationCompat.CATEGORY_WORKOUT)
                .setPriority(NotificationCompat.PRIORITY_HIGH)
                .setOnlyAlertOnce(true)
                .setAutoCancel(false)
                .addAction(android.R.drawable.ic_delete, stopLabel, stopPendingIntent)
                .addAction(android.R.drawable.ic_menu_save, completeLabel, completePendingIntent)

            // android 16+ uses live updates with chip showing total elapsed time
            if (Build.VERSION.SDK_INT >= 36) {
                builder.extras.putBoolean("android.extraRequestPromotedOngoing", true)
                builder.setShortCriticalText(totalElapsedTimeText)
                if (isPaused) {
                    // freeze chronometer: disable auto-tick so it stops advancing
                    builder.setUsesChronometer(false)
                    builder.setWhen(System.currentTimeMillis() - (totalElapsedSeconds * 1000L))
                } else {
                    builder.setWhen(System.currentTimeMillis() - (totalElapsedSeconds * 1000L))
                    builder.setUsesChronometer(true)
                    builder.setChronometerCountDown(false)
                }
            }

            val notification = builder.build()

            val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
            manager.notify(NOTIFICATION_ID, notification)
        }

        fun cancelNotification(context: Context) {
            val manager = context.getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager
            manager.cancel(NOTIFICATION_ID)
        }
    }
}

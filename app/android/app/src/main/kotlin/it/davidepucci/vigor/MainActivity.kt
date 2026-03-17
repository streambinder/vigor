package it.davidepucci.vigor

import android.os.Bundle
import io.flutter.embedding.android.FlutterFragmentActivity
import io.flutter.embedding.engine.FlutterEngine
import io.flutter.plugin.common.MethodChannel

class MainActivity : FlutterFragmentActivity() {
    private val channelName = "it.davidepucci.vigor/timer_notification"
    private var methodChannel: MethodChannel? = null

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        TimerNotificationService.createNotificationChannel(this)
        handleIntent(intent)
    }

    override fun onNewIntent(intent: android.content.Intent) {
        super.onNewIntent(intent)
        handleIntent(intent)
    }

    private fun handleIntent(intent: android.content.Intent?) {
        when (intent?.action) {
            "TIMER_STOP" -> {
                methodChannel?.invokeMethod("onTimerStop", null)
            }
            "TIMER_COMPLETE" -> {
                methodChannel?.invokeMethod("onTimerComplete", null)
            }
        }
    }

    override fun configureFlutterEngine(flutterEngine: FlutterEngine) {
        super.configureFlutterEngine(flutterEngine)

        methodChannel = MethodChannel(flutterEngine.dartExecutor.binaryMessenger, channelName)
        methodChannel?.setMethodCallHandler { call, result ->
            when (call.method) {
                "updateTimerNotification" -> {
                    val trainingName = call.argument<String>("trainingName") ?: "Training"
                    val totalElapsedSeconds = call.argument<Int>("totalElapsedSeconds") ?: 0
                    val contentText = call.argument<String>("contentText") ?: "Workout"
                    val intervalRemainingSeconds = call.argument<Int>("intervalRemainingSeconds") ?: 0
                    val isDurationBased = call.argument<Boolean>("isDurationBased") ?: true
                    val progress = call.argument<Double>("progress") ?: 0.0
                    val stopLabel = call.argument<String>("stopLabel") ?: "Stop"
                    val completeLabel = call.argument<String>("completeLabel") ?: "Complete"

                    TimerNotificationService.showLiveTimerNotification(
                        this,
                        trainingName,
                        totalElapsedSeconds,
                        contentText,
                        intervalRemainingSeconds,
                        isDurationBased,
                        progress,
                        stopLabel,
                        completeLabel
                    )
                    result.success(null)
                }
                "stopTimer" -> {
                    // notify flutter to stop timer
                    methodChannel?.invokeMethod("onTimerStop", null)
                    result.success(null)
                }
                "completeTimer" -> {
                    // notify flutter to complete timer
                    methodChannel?.invokeMethod("onTimerComplete", null)
                    result.success(null)
                }
                "cancelTimerNotification" -> {
                    TimerNotificationService.cancelNotification(this)
                    result.success(null)
                }
                else -> result.notImplemented()
            }
        }
    }

    override fun onDestroy() {
        methodChannel?.setMethodCallHandler(null)
        super.onDestroy()
    }
}

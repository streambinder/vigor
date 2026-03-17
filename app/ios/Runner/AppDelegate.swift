import Flutter
import UIKit

@main
@objc class AppDelegate: FlutterAppDelegate {
  override func application(
    _ application: UIApplication,
    didFinishLaunchingWithOptions launchOptions: [UIApplication.LaunchOptionsKey: Any]?
  ) -> Bool {
    GeneratedPluginRegistrant.register(with: self)

    // setup method channel for live activities
    if let controller = window?.rootViewController as? FlutterViewController {
      setupTimerNotificationChannel(controller: controller)
    }

    return super.application(application, didFinishLaunchingWithOptions: launchOptions)
  }

  private func setupTimerNotificationChannel(controller: FlutterViewController) {
    let channel = FlutterMethodChannel(
      name: "it.davidepucci.vigor/timer_notification",
      binaryMessenger: controller.binaryMessenger
    )

    channel.setMethodCallHandler { [weak self] (call, result) in
      guard #available(iOS 16.2, *) else {
        result(FlutterMethodNotImplemented)
        return
      }

      switch call.method {
      case "updateTimerNotification":
        guard let args = call.arguments as? [String: Any],
              let exerciseName = args["exerciseName"] as? String,
              let remainingSeconds = args["remainingSeconds"] as? Int,
              let progress = args["progress"] as? Double else {
          result(FlutterError(code: "INVALID_ARGS", message: "missing arguments", details: nil))
          return
        }

        let trainingName = args["trainingName"] as? String ?? "Training"

        if TimerLiveActivityManager.shared.currentActivity == nil {
          TimerLiveActivityManager.shared.startActivity(
            trainingName: trainingName,
            exerciseName: exerciseName,
            remainingSeconds: remainingSeconds,
            progress: progress
          )
        } else {
          TimerLiveActivityManager.shared.updateActivity(
            exerciseName: exerciseName,
            remainingSeconds: remainingSeconds,
            progress: progress
          )
        }
        result(nil)

      case "cancelTimerNotification":
        TimerLiveActivityManager.shared.stopActivity()
        result(nil)

      default:
        result(FlutterMethodNotImplemented)
      }
    }
  }
}

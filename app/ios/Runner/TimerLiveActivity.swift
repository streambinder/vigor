import ActivityKit
import Foundation

@available(iOS 16.2, *)
struct TimerActivityAttributes: ActivityAttributes {
    public struct ContentState: Codable, Hashable {
        var exerciseName: String
        var remainingSeconds: Int
        var progress: Double
    }

    var trainingName: String
}

@available(iOS 16.2, *)
class TimerLiveActivityManager {
    static let shared = TimerLiveActivityManager()

    var currentActivity: Activity<TimerActivityAttributes>?

    private init() {}

    func startActivity(trainingName: String, exerciseName: String, remainingSeconds: Int, progress: Double) {
        // end existing activity if any
        stopActivity()

        let attributes = TimerActivityAttributes(trainingName: trainingName)
        let contentState = TimerActivityAttributes.ContentState(
            exerciseName: exerciseName,
            remainingSeconds: remainingSeconds,
            progress: progress
        )

        do {
            let activity = try Activity<TimerActivityAttributes>.request(
                attributes: attributes,
                content: .init(state: contentState, staleDate: nil),
                pushType: nil
            )
            currentActivity = activity
        } catch {
            print("failed to start live activity: \(error)")
        }
    }

    func updateActivity(exerciseName: String, remainingSeconds: Int, progress: Double) {
        guard let activity = currentActivity else { return }

        let contentState = TimerActivityAttributes.ContentState(
            exerciseName: exerciseName,
            remainingSeconds: remainingSeconds,
            progress: progress
        )

        Task {
            await activity.update(using: contentState)
        }
    }

    func stopActivity() {
        guard let activity = currentActivity else { return }

        Task {
            await activity.end(dismissalPolicy: .immediate)
            currentActivity = nil
        }
    }
}

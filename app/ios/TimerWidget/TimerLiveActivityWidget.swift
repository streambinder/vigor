import ActivityKit
import WidgetKit
import SwiftUI

@available(iOS 16.2, *)
struct TimerLiveActivityWidget: Widget {
    var body: some WidgetConfiguration {
        ActivityConfiguration(for: TimerActivityAttributes.self) { context in
            // lock screen / banner ui
            LockScreenTimerView(context: context)
        } dynamicIsland: { context in
            DynamicIsland {
                // expanded
                DynamicIslandExpandedRegion(.leading) {
                    Text(context.state.exerciseName)
                        .font(.headline)
                }
                DynamicIslandExpandedRegion(.trailing) {
                    Text(formatTime(context.state.remainingSeconds))
                        .font(.title2)
                        .monospacedDigit()
                }
                DynamicIslandExpandedRegion(.bottom) {
                    ProgressView(value: context.state.progress)
                        .tint(.orange)
                }
            } compactLeading: {
                Image(systemName: "figure.strengthtraining.traditional")
            } compactTrailing: {
                Text(formatTime(context.state.remainingSeconds))
                    .font(.caption2)
                    .monospacedDigit()
            } minimal: {
                Image(systemName: "timer")
            }
        }
    }

    private func formatTime(_ seconds: Int) -> String {
        let mins = seconds / 60
        let secs = seconds % 60
        return String(format: "%02d:%02d", mins, secs)
    }
}

@available(iOS 16.2, *)
struct LockScreenTimerView: View {
    let context: ActivityViewContext<TimerActivityAttributes>

    var body: some View {
        VStack(alignment: .leading, spacing: 8) {
            HStack {
                Image(systemName: "figure.strengthtraining.traditional")
                    .foregroundColor(.orange)
                Text(context.attributes.trainingName)
                    .font(.headline)
                Spacer()
                Text(formatTime(context.state.remainingSeconds))
                    .font(.title2)
                    .monospacedDigit()
            }

            ProgressView(value: context.state.progress)
                .tint(.orange)

            Text(context.state.exerciseName)
                .font(.subheadline)
                .foregroundColor(.secondary)
        }
        .padding()
    }

    private func formatTime(_ seconds: Int) -> String {
        let mins = seconds / 60
        let secs = seconds % 60
        return String(format: "%02d:%02d", mins, secs)
    }
}

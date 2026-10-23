import 'activity_actual.dart';
import 'training_feedback.dart';

extension TrainingFeedbackExt on TrainingFeedback {
  bool get hasData => quality != null || qualityReason.isNotEmpty || message.isNotEmpty || activityFeedback.isNotEmpty;

  /// per-activity performed values recorded with the feedback,
  /// keyed by activity id
  Map<String, ActivityActual> get actualsByActivityId => {
        for (final entry in activityActuals.entries)
          if (entry.value is Map)
            entry.key: ActivityActual.fromJson(
              Map<String, dynamic>.from(entry.value as Map),
            ),
      };
}

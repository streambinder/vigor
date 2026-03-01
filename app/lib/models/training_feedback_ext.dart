import 'training_feedback.dart';

extension TrainingFeedbackExt on TrainingFeedback {
  bool get hasData => quality != null || qualityReason.isNotEmpty || message.isNotEmpty || activityFeedback.isNotEmpty;
}

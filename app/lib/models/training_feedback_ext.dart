import 'training_feedback.dart';

extension TrainingFeedbackExt on TrainingFeedback {
  // returns true if any feedback data is present
  bool get hasData => quality != null || qualityReason.isNotEmpty || message.isNotEmpty;
}

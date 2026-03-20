import 'dart:async';
import 'package:flutter/widgets.dart';

sealed class AppEvent {}

class TrainingListChanged extends AppEvent {}

class TrainingCompleted extends AppEvent {
  final String trainingId;
  TrainingCompleted(this.trainingId);
}

class GymListChanged extends AppEvent {}

class HealthSyncCompleted extends AppEvent {}

class ProfileUpdated extends AppEvent {}

class FeedbackSubmitted extends AppEvent {
  final String trainingId;
  FeedbackSubmitted(this.trainingId);
}

class FlowSessionListChanged extends AppEvent {}

/// mixin for screens that need to react to app events beyond ValueNotifier changes
mixin AppEventSubscriber<T extends StatefulWidget> on State<T> {
  StreamSubscription<AppEvent>? _eventSub;

  void subscribeToEvents(Stream<AppEvent> events, void Function(AppEvent) handler) {
    _eventSub = events.listen(handler);
  }

  @override
  void dispose() {
    _eventSub?.cancel();
    super.dispose();
  }
}

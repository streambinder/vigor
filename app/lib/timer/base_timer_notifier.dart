import 'package:flutter/material.dart';
import 'timer_controller.dart';
import 'timer_mode.dart';

/// Minimal abstract base for timer notifiers consumed by InlineTimerSection.
/// Both WorkoutTimerNotifier and FlowTimerNotifier extend this.
abstract class BaseTimerNotifier extends ChangeNotifier {
  TimerController? get controller;
  bool get workoutCompleted;
  bool get isSubmitting;
  int get totalElapsedSeconds;
  TimerMode? get currentMode;
  double get progress;

  Color phaseColor(Brightness brightness);
  String formatDuration(int seconds);
  void handleTap(TapUpDetails details, double sectionWidth);
  void stopWhistle();
}

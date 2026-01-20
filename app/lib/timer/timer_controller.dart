import 'package:flutter/foundation.dart';
import 'training_interval.dart';

/// Abstract base controller for all timer modes
/// Each mode implements its own timing and progression logic
abstract class TimerController extends ChangeNotifier {
  /// Current countdown/countup value in seconds
  int get remainingSeconds;

  /// Whether the timer is currently paused
  bool get isPaused;

  /// Whether the workout segment is complete
  bool get isCompleted;

  /// The currently active training interval (null if none)
  TrainingInterval? get currentInterval;

  /// Whether training has started (past initial countdown)
  bool get hasStarted;

  /// Whether backward navigation is available
  bool get canGoBack;

  /// Upcoming intervals for preview (may be empty for some modes)
  List<TrainingInterval> get upcomingIntervals;

  /// Current round number (for AMRAP/ForTime modes, 0 for interval)
  int get currentRound => 0;

  /// Total rounds (for ForTime mode, 0 for others)
  int get totalRounds => 0;

  /// Total/global time remaining or elapsed (for AMRAP/ForTime)
  int get globalSeconds => 0;

  /// Start the initial countdown before training
  void startCountdown();

  /// Start the actual training (after countdown)
  void startTraining();

  /// Pause the timer
  void pause();

  /// Resume the timer
  void resume();

  /// Toggle pause state
  void togglePause() {
    if (isPaused) {
      resume();
    } else {
      pause();
    }
  }

  /// Skip forward to next interval/activity
  void skipForward();

  /// Skip backward to previous interval/activity
  void skipBackward();

  /// Handle user tap action (for user-paced modes: emom, amrap, forTime)
  /// In interval mode this does nothing
  void onUserAction() {}

  /// Clean up timer resources
  @override
  void dispose();
}

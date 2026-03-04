import 'dart:async';
import 'package:flutter/foundation.dart';
import 'training_interval.dart';

/// Abstract base controller for all timer modes
/// Each mode implements its own timing and progression logic
abstract class TimerController extends ChangeNotifier {
  /// Set to true when interval changes due to user skip (not timer)
  /// Screens should check this to avoid playing jingle on skip
  bool wasSkipped = false;

  /// Independent elapsed timer — runs continuously, unaffected by interval transitions
  Timer? _elapsedTimer;
  int _totalElapsedSeconds = 0;
  bool _elapsedRunning = false;
  int get elapsedSeconds => _totalElapsedSeconds;

  /// wall-clock timestamp of last tick — used to compensate timer drift
  /// when the app returns from background
  DateTime? _lastTickTime;

  /// Start the elapsed clock (call once when training begins)
  @protected
  void startElapsedTimer() {
    if (_elapsedRunning) return;
    _elapsedRunning = true;
    _lastTickTime = DateTime.now();
    _elapsedTimer = Timer.periodic(const Duration(seconds: 1), (_) {
      if (!isPaused) {
        _totalElapsedSeconds++;
        _lastTickTime = DateTime.now();
        notifyListeners();
      }
    });
  }

  /// Stop the elapsed clock (call on completion)
  @protected
  void stopElapsedTimer() {
    _elapsedTimer?.cancel();
    _elapsedTimer = null;
    _elapsedRunning = false;
  }

  /// Compensate for time elapsed while the app was backgrounded.
  /// Returns the number of seconds that passed in background.
  /// Subclasses override [onBackgroundDrift] to fast-forward their timers.
  int compensateBackgroundDrift() {
    if (_lastTickTime == null || isPaused || !_elapsedRunning) return 0;
    final now = DateTime.now();
    final drift = now.difference(_lastTickTime!).inSeconds;
    // only compensate if >1s drift (normal tick jitter is ~0)
    if (drift <= 1) return 0;
    _totalElapsedSeconds += drift;
    _lastTickTime = now;
    onBackgroundDrift(drift);
    return drift;
  }

  /// Override to fast-forward mode-specific timers by [driftSeconds].
  /// Called when the app returns from background with accumulated drift.
  @protected
  void onBackgroundDrift(int driftSeconds) {}

  /// Set to true when countdown reaches 3, 2, or 1 seconds
  /// Screens should play countdown jingle when this is true
  bool shouldPlayCountdownJingle = false;

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
  @mustCallSuper
  void dispose() {
    stopElapsedTimer();
    super.dispose();
  }
}

import 'dart:async';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity.dart';
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import 'timer_controller.dart';
import 'training_interval.dart';

/// Controller for EMOM (Every Minute On the Minute) timer mode
///
/// Behavior:
/// - Each minute: user cycles through all activities in block
/// - User taps/skips to advance to next activity
/// - After completing last activity, rest for remainder of minute
/// - At minute boundary, automatically resets to first activity
/// - Total minutes = block.repeats
class EmomController extends TimerController {
  final Training training;
  final Routine routine;
  final Block block;

  Timer? _timer;
  int _secondsInMinute = 60;
  int _currentMinute = 1;
  int _totalMinutes = 0;
  int _currentActivityIndex = 0;
  bool _isResting = false;  // completed all activities, resting until minute ends
  bool _isPaused = false;
  bool _isCompleted = false;
  bool _hasStarted = false;
  final List<_EmomHistoryEntry> _history = [];

  EmomController({
    required this.training,
    required this.routine,
    required this.block,
  }) {
    _totalMinutes = block.repeats > 0 ? block.repeats : 1;
  }

  @override
  int get remainingSeconds => _secondsInMinute;

  @override
  bool get isPaused => _isPaused;

  @override
  bool get isCompleted => _isCompleted;

  @override
  bool get hasStarted => _hasStarted;

  @override
  bool get canGoBack => _history.isNotEmpty;

  @override
  int get currentRound => _currentMinute;

  @override
  int get totalRounds => _totalMinutes;

  /// Whether user completed all activities and is resting until minute ends
  bool get isResting => _isResting;

  /// Current activity index within block (0-based)
  int get activityIndex => _currentActivityIndex;

  /// Total activities in block
  int get totalActivities => block.activities.length;

  /// Whether current activity is the last in the block
  bool get isLastActivity => _currentActivityIndex >= block.activities.length - 1;

  Activity? get _currentActivity =>
      _currentActivityIndex < block.activities.length
          ? block.activities[_currentActivityIndex]
          : null;

  @override
  TrainingInterval? get currentInterval {
    if (_isResting) {
      return TrainingInterval(
        type: IntervalType.rest,
        duration: _secondsInMinute,
        routineName: routine.type,
        activityNumber: _currentMinute,
        totalActivities: _totalMinutes,
        blockNumber: 1,
        totalBlocks: 1,
        routineNumber: 1,
        totalRoutines: 1,
      );
    }

    final activity = _currentActivity;
    if (activity == null) return null;

    return TrainingInterval(
      type: IntervalType.work,
      duration: _secondsInMinute,
      routineName: routine.type,
      activityName: activity.displayName,
      activity: activity,
      exercise: _parseExercise(activity.detail),
      activityNumber: _currentActivityIndex + 1,
      totalActivities: block.activities.length,
      blockNumber: _currentMinute,
      totalBlocks: _totalMinutes,
      routineNumber: 1,
      totalRoutines: 1,
    );
  }

  @override
  List<TrainingInterval> get upcomingIntervals {
    final upcoming = <TrainingInterval>[];

    // show remaining activities in current minute
    for (int i = _currentActivityIndex + 1; i < block.activities.length && upcoming.length < 5; i++) {
      final activity = block.activities[i];
      upcoming.add(TrainingInterval(
        type: IntervalType.work,
        duration: 0,
        routineName: routine.type,
        activityName: activity.displayName,
        activity: activity,
        exercise: _parseExercise(activity.detail),
        activityNumber: i + 1,
        totalActivities: block.activities.length,
        blockNumber: _currentMinute,
        totalBlocks: _totalMinutes,
        routineNumber: 1,
        totalRoutines: 1,
      ));
    }

    // show activities from next minutes
    for (int nextMinute = _currentMinute + 1;
        nextMinute <= _totalMinutes && upcoming.length < 5;
        nextMinute++) {
      for (int i = 0; i < block.activities.length && upcoming.length < 5; i++) {
        final activity = block.activities[i];
        upcoming.add(TrainingInterval(
          type: IntervalType.work,
          duration: 0,
          routineName: routine.type,
          activityName: activity.displayName,
          activity: activity,
          exercise: _parseExercise(activity.detail),
          activityNumber: i + 1,
          totalActivities: block.activities.length,
          blockNumber: nextMinute,
          totalBlocks: _totalMinutes,
          routineNumber: 1,
          totalRoutines: 1,
        ));
      }
    }

    return upcoming;
  }

  @override
  void startCountdown() {
    _secondsInMinute = 5;
    _startTimer();
  }

  @override
  void startTraining() {
    if (_hasStarted) return;
    _timer?.cancel();
    _hasStarted = true;
    _startMinute();
    notifyListeners();
  }

  @override
  void pause() {
    _isPaused = true;
    notifyListeners();
  }

  @override
  void resume() {
    _isPaused = false;
    notifyListeners();
  }

  @override
  void skipForward() {
    if (_isCompleted) return;
    _advanceActivity();
  }

  @override
  void skipBackward() {
    if (_history.isEmpty) return;
    _restoreHistory();
    notifyListeners();
  }

  @override
  void onUserAction() {
    // same as skip forward - advance to next activity
    skipForward();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _advanceActivity() {
    if (_isResting) return;  // can't advance during rest, wait for minute boundary

    _saveHistory();
    _currentActivityIndex++;

    if (_currentActivityIndex >= block.activities.length) {
      // completed all activities in minute, rest until minute ends
      _isResting = true;
    }
    notifyListeners();
  }

  void _startTimer() {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_isPaused) return;

      if (_secondsInMinute > 0) {
        _secondsInMinute--;
        notifyListeners();
      } else {
        timer.cancel();
        if (!_hasStarted) {
          startTraining();
        } else {
          _onMinuteBoundary();
        }
      }
    });
  }

  void _startMinute() {
    _secondsInMinute = 60;
    _currentActivityIndex = 0;
    _isResting = false;
    notifyListeners();
    _startTimer();
  }

  void _onMinuteBoundary() {
    _currentMinute++;
    if (_currentMinute > _totalMinutes) {
      _completeTraining();
    } else {
      _startMinute();
    }
  }

  void _completeTraining() {
    _timer?.cancel();
    _isCompleted = true;
    notifyListeners();
  }

  void _saveHistory() {
    _history.add(_EmomHistoryEntry(
      minute: _currentMinute,
      activityIndex: _currentActivityIndex,
      secondsRemaining: _secondsInMinute,
      isResting: _isResting,
    ));
  }

  void _restoreHistory() {
    final entry = _history.removeLast();
    _currentMinute = entry.minute;
    _currentActivityIndex = entry.activityIndex;
    _secondsInMinute = entry.secondsRemaining;
    _isResting = entry.isResting;
    _isCompleted = false;
  }

  static Exercise? _parseExercise(Map<String, dynamic> detail) {
    if (detail.isEmpty) return null;
    try {
      return Exercise.fromJson(detail);
    } catch (_) {
      return null;
    }
  }
}

class _EmomHistoryEntry {
  final int minute;
  final int activityIndex;
  final int secondsRemaining;
  final bool isResting;

  _EmomHistoryEntry({
    required this.minute,
    required this.activityIndex,
    required this.secondsRemaining,
    required this.isResting,
  });
}

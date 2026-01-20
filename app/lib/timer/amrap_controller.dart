import 'dart:async';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import 'timer_controller.dart';
import 'training_interval.dart';

/// Controller for AMRAP (As Many Rounds As Possible) timer mode
///
/// Behavior:
/// - Global countdown timer for total workout duration
/// - User taps to advance through activities
/// - When block completes, round counter increments and activities restart
/// - Continues until global timer expires
class AmrapController extends TimerController {
  final Training training;
  final Routine routine;
  final Block block;
  final int totalSeconds;  // global workout duration

  Timer? _timer;
  int _globalSecondsRemaining = 0;
  int _currentActivityIndex = 0;
  int _roundCount = 1;  // 1-based
  bool _isPaused = false;
  bool _isCompleted = false;
  bool _hasStarted = false;
  final List<_AmrapHistoryEntry> _history = [];

  AmrapController({
    required this.training,
    required this.routine,
    required this.block,
    required this.totalSeconds,
  });

  @override
  int get remainingSeconds => _globalSecondsRemaining;

  @override
  int get globalSeconds => _globalSecondsRemaining;

  @override
  bool get isPaused => _isPaused;

  @override
  bool get isCompleted => _isCompleted;

  @override
  bool get hasStarted => _hasStarted;

  @override
  bool get canGoBack => _history.isNotEmpty;

  @override
  int get currentRound => _roundCount;

  /// Activities completed in current round (0-based)
  int get activitiesInRound => _currentActivityIndex;

  /// Total activities per round
  int get activitiesPerRound => block.activities.length;

  @override
  TrainingInterval? get currentInterval {
    if (_currentActivityIndex >= block.activities.length) return null;
    final activity = block.activities[_currentActivityIndex];
    return TrainingInterval(
      type: IntervalType.work,
      duration: 0,  // no duration in AMRAP - user paced
      routineName: routine.type,
      activityName: activity.displayName,
      activity: activity,
      exercise: _parseExercise(activity.detail),
      activityNumber: _currentActivityIndex + 1,
      totalActivities: block.activities.length,
      blockNumber: 1,
      totalBlocks: 1,
      routineNumber: 1,
      totalRoutines: 1,
    );
  }

  @override
  List<TrainingInterval> get upcomingIntervals {
    final upcoming = <TrainingInterval>[];
    // show remaining activities in current round + first few of next round
    for (int i = 1; i <= 5; i++) {
      final idx = (_currentActivityIndex + i) % block.activities.length;
      final activity = block.activities[idx];
      upcoming.add(TrainingInterval(
        type: IntervalType.work,
        duration: 0,
        routineName: routine.type,
        activityName: activity.displayName,
        activity: activity,
        exercise: _parseExercise(activity.detail),
        activityNumber: idx + 1,
        totalActivities: block.activities.length,
        blockNumber: 1,
        totalBlocks: 1,
        routineNumber: 1,
        totalRoutines: 1,
      ));
    }
    return upcoming;
  }

  @override
  void startCountdown() {
    _globalSecondsRemaining = 5;  // initial countdown
    _startTimer(isCountdown: true);
  }

  @override
  void startTraining() {
    if (_hasStarted) return;
    _timer?.cancel();
    _hasStarted = true;
    _globalSecondsRemaining = totalSeconds;
    _startTimer(isCountdown: false);
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
    onUserAction();
  }

  @override
  void skipBackward() {
    if (_history.isEmpty) return;
    _restoreHistory();
    notifyListeners();
  }

  @override
  void onUserAction() {
    if (!_hasStarted || _isCompleted) return;

    _saveHistory();
    _currentActivityIndex++;

    if (_currentActivityIndex >= block.activities.length) {
      // round complete
      _roundCount++;
      _currentActivityIndex = 0;
    }
    notifyListeners();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _startTimer({required bool isCountdown}) {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_isPaused) return;

      if (_globalSecondsRemaining > 0) {
        _globalSecondsRemaining--;
        notifyListeners();
      } else {
        timer.cancel();
        if (isCountdown) {
          startTraining();
        } else {
          _completeTraining();
        }
      }
    });
  }

  void _completeTraining() {
    _timer?.cancel();
    _isCompleted = true;
    notifyListeners();
  }

  void _saveHistory() {
    _history.add(_AmrapHistoryEntry(
      roundCount: _roundCount,
      activityIndex: _currentActivityIndex,
    ));
  }

  void _restoreHistory() {
    final entry = _history.removeLast();
    _roundCount = entry.roundCount;
    _currentActivityIndex = entry.activityIndex;
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

class _AmrapHistoryEntry {
  final int roundCount;
  final int activityIndex;

  _AmrapHistoryEntry({
    required this.roundCount,
    required this.activityIndex,
  });
}

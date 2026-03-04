import 'dart:async';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import 'timer_controller.dart';
import 'training_interval.dart';

/// Controller for ForTime timer mode
///
/// Behavior:
/// - Elapsed time counter (counts UP from 0)
/// - User taps to advance through activities
/// - Fixed number of rounds to complete (block.repeats)
/// - Records completion time when all rounds done
class ForTimeController extends TimerController {
  final Training training;
  final Routine routine;
  final Block block;

  Timer? _timer;
  int _elapsedSeconds = 0;
  int _currentActivityIndex = 0;
  int _currentRound = 1;  // 1-based
  int _targetRounds = 0;
  bool _isPaused = false;
  bool _isCompleted = false;
  bool _hasStarted = false;
  int _countdownSeconds = 5;
  int? _completionTime;  // recorded when finished
  final List<_ForTimeHistoryEntry> _history = [];

  ForTimeController({
    required this.training,
    required this.routine,
    required this.block,
  }) {
    _targetRounds = block.repeats > 0 ? block.repeats : 1;
  }

  @override
  int get remainingSeconds => _hasStarted ? _elapsedSeconds : _countdownSeconds;

  @override
  int get globalSeconds => _elapsedSeconds;

  @override
  bool get isPaused => _isPaused;

  @override
  bool get isCompleted => _isCompleted;

  @override
  bool get hasStarted => _hasStarted;

  @override
  bool get canGoBack => _history.isNotEmpty;

  @override
  int get currentRound => _currentRound;

  @override
  int get totalRounds => _targetRounds;

  /// Completion time in seconds (null if not yet complete)
  int? get completionTime => _completionTime;

  @override
  TrainingInterval? get currentInterval {
    if (_currentActivityIndex >= block.activities.length) return null;
    final activity = block.activities[_currentActivityIndex];
    return TrainingInterval(
      type: IntervalType.work,
      duration: 0,  // no duration - user paced
      routineName: routine.type,
      activityName: activity.displayName,
      activity: activity,
      exercise: _parseExercise(activity.detail),
      activityNumber: _currentActivityIndex + 1,
      totalActivities: block.activities.length,
      blockNumber: _currentRound,
      totalBlocks: _targetRounds,
      routineNumber: 1,
      totalRoutines: 1,
    );
  }

  @override
  List<TrainingInterval> get upcomingIntervals {
    final upcoming = <TrainingInterval>[];
    int actIdx = _currentActivityIndex + 1;
    int round = _currentRound;

    for (int i = 0; i < 5; i++) {
      if (actIdx >= block.activities.length) {
        round++;
        actIdx = 0;
      }
      if (round > _targetRounds) break;

      final activity = block.activities[actIdx];
      upcoming.add(TrainingInterval(
        type: IntervalType.work,
        duration: 0,
        routineName: routine.type,
        activityName: activity.displayName,
        activity: activity,
        exercise: _parseExercise(activity.detail),
        activityNumber: actIdx + 1,
        totalActivities: block.activities.length,
        blockNumber: round,
        totalBlocks: _targetRounds,
        routineNumber: 1,
        totalRoutines: 1,
      ));
      actIdx++;
    }
    return upcoming;
  }

  @override
  void startCountdown() {
    _countdownSeconds = 5;
    _startCountdownTimer();
  }

  @override
  void startTraining() {
    if (_hasStarted) return;
    _timer?.cancel();
    _hasStarted = true;
    _elapsedSeconds = 0;
    startElapsedTimer();
    _startWorkTimer();
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
    wasSkipped = true;
    onUserAction();
  }

  @override
  void skipBackward() {
    if (_history.isEmpty) return;
    wasSkipped = true;
    _restoreHistory();
    notifyListeners();
  }

  @override
  void onUserAction() {
    if (!_hasStarted || _isCompleted) return;

    wasSkipped = true;
    _saveHistory();
    _currentActivityIndex++;

    if (_currentActivityIndex >= block.activities.length) {
      _currentRound++;
      _currentActivityIndex = 0;

      if (_currentRound > _targetRounds) {
        _completeTraining();
        return;
      }
    }
    notifyListeners();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  void onBackgroundDrift(int driftSeconds) {
    if (!_hasStarted || _isCompleted || _isPaused) return;
    // ForTime counts up — just add the missed seconds
    _elapsedSeconds += driftSeconds;
    notifyListeners();
  }

  void _startCountdownTimer() {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_isPaused) return;

      if (_countdownSeconds > 0) {
        _countdownSeconds--;
        notifyListeners();
      } else {
        timer.cancel();
        startTraining();
      }
    });
  }

  void _startWorkTimer() {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_isPaused) return;
      _elapsedSeconds++;
      notifyListeners();
    });
  }

  void _completeTraining() {
    _timer?.cancel();
    stopElapsedTimer();
    _completionTime = _elapsedSeconds;
    _isCompleted = true;
    notifyListeners();
  }

  void _saveHistory() {
    _history.add(_ForTimeHistoryEntry(
      round: _currentRound,
      activityIndex: _currentActivityIndex,
      elapsedSeconds: _elapsedSeconds,
    ));
  }

  void _restoreHistory() {
    final entry = _history.removeLast();
    _currentRound = entry.round;
    _currentActivityIndex = entry.activityIndex;
    _elapsedSeconds = entry.elapsedSeconds;
    _isCompleted = false;
    _completionTime = null;
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

class _ForTimeHistoryEntry {
  final int round;
  final int activityIndex;
  final int elapsedSeconds;

  _ForTimeHistoryEntry({
    required this.round,
    required this.activityIndex,
    required this.elapsedSeconds,
  });
}

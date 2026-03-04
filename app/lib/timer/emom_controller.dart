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
/// - Each minute: user cycles through all activities in current block
/// - User taps/skips to advance to next activity
/// - After completing last activity, rest for remainder of minute
/// - At minute boundary, automatically resets to first activity
/// - When all minutes in a block are exhausted, optional rest then next block
/// - Total minutes per block = block.repeats
class EmomController extends TimerController {
  final Training training;
  final Routine routine;
  final List<Block> blocks;

  Timer? _timer;
  int _secondsInMinute = 60;
  int _currentMinuteInBlock = 1;  // 1-based minute within current block
  int _currentBlockIndex = 0;
  int _currentActivityIndex = 0;
  bool _isResting = false;  // completed all activities, resting until minute ends
  bool _isBlockRest = false; // resting between blocks
  bool _isPaused = false;
  bool _isCompleted = false;
  bool _hasStarted = false;
  final List<_EmomHistoryEntry> _history = [];

  EmomController({
    required this.training,
    required this.routine,
    required this.blocks,
  });

  Block get _currentBlock => blocks[_currentBlockIndex];

  int get _currentBlockRepeats =>
      _currentBlock.repeats > 0 ? _currentBlock.repeats : 1;

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
  int get currentRound => _currentMinuteInBlock;

  @override
  int get totalRounds => _currentBlockRepeats;

  /// Whether user completed all activities and is resting until minute ends
  bool get isResting => _isResting || _isBlockRest;

  /// Whether resting between blocks (inter-block rest)
  bool get isBlockRest => _isBlockRest;

  /// Current activity index within block (0-based)
  int get activityIndex => _currentActivityIndex;

  /// Total activities in current block
  int get totalActivities => _currentBlock.activities.length;

  /// Whether current activity is the last in the block
  bool get isLastActivity => _currentActivityIndex >= _currentBlock.activities.length - 1;

  /// Current block index (0-based)
  int get currentBlockIndex => _currentBlockIndex;

  /// Total number of blocks
  int get totalBlocks => blocks.length;

  Activity? get _currentActivity =>
      _currentActivityIndex < _currentBlock.activities.length
          ? _currentBlock.activities[_currentActivityIndex]
          : null;

  @override
  TrainingInterval? get currentInterval {
    if (_isBlockRest) {
      return TrainingInterval(
        type: IntervalType.rest,
        duration: _secondsInMinute,
        routineName: routine.type,
        activityNumber: _currentBlockIndex + 1,
        totalActivities: blocks.length,
        blockNumber: _currentBlockIndex + 1,
        totalBlocks: blocks.length,
        routineNumber: 1,
        totalRoutines: 1,
      );
    }

    if (_isResting) {
      return TrainingInterval(
        type: IntervalType.rest,
        duration: _secondsInMinute,
        routineName: routine.type,
        activityNumber: _currentMinuteInBlock,
        totalActivities: _currentBlockRepeats,
        blockNumber: _currentBlockIndex + 1,
        totalBlocks: blocks.length,
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
      totalActivities: _currentBlock.activities.length,
      blockNumber: _currentMinuteInBlock,
      totalBlocks: _currentBlockRepeats,
      routineNumber: 1,
      totalRoutines: 1,
    );
  }

  @override
  List<TrainingInterval> get upcomingIntervals {
    final upcoming = <TrainingInterval>[];
    if (_isBlockRest) return upcoming;

    // remaining activities in current minute
    for (int i = _currentActivityIndex + 1;
        i < _currentBlock.activities.length && upcoming.length < 5;
        i++) {
      final activity = _currentBlock.activities[i];
      upcoming.add(TrainingInterval(
        type: IntervalType.work,
        duration: 0,
        routineName: routine.type,
        activityName: activity.displayName,
        activity: activity,
        exercise: _parseExercise(activity.detail),
        activityNumber: i + 1,
        totalActivities: _currentBlock.activities.length,
        blockNumber: _currentMinuteInBlock,
        totalBlocks: _currentBlockRepeats,
        routineNumber: 1,
        totalRoutines: 1,
      ));
    }

    // remaining minutes in current block
    for (int nextMinute = _currentMinuteInBlock + 1;
        nextMinute <= _currentBlockRepeats && upcoming.length < 5;
        nextMinute++) {
      for (int i = 0; i < _currentBlock.activities.length && upcoming.length < 5; i++) {
        final activity = _currentBlock.activities[i];
        upcoming.add(TrainingInterval(
          type: IntervalType.work,
          duration: 0,
          routineName: routine.type,
          activityName: activity.displayName,
          activity: activity,
          exercise: _parseExercise(activity.detail),
          activityNumber: i + 1,
          totalActivities: _currentBlock.activities.length,
          blockNumber: nextMinute,
          totalBlocks: _currentBlockRepeats,
          routineNumber: 1,
          totalRoutines: 1,
        ));
      }
    }

    // activities from subsequent blocks
    for (int bi = _currentBlockIndex + 1; bi < blocks.length && upcoming.length < 5; bi++) {
      final block = blocks[bi];
      final blockRepeats = block.repeats > 0 ? block.repeats : 1;
      for (int minute = 1; minute <= blockRepeats && upcoming.length < 5; minute++) {
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
            blockNumber: minute,
            totalBlocks: blockRepeats,
            routineNumber: 1,
            totalRoutines: 1,
          ));
        }
      }
    }

    return upcoming;
  }

  @override
  void startCountdown() {
    _secondsInMinute = 3;
    _startTimer();
  }

  @override
  void startTraining() {
    if (_hasStarted) return;
    _timer?.cancel();
    _hasStarted = true;
    startElapsedTimer();
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
    wasSkipped = true;

    if (_isBlockRest) {
      _saveHistory();
      _advanceToNextBlock();
      return;
    }

    if (_isResting) {
      // skip the within-minute rest, jump to next minute boundary
      _saveHistory();
      _timer?.cancel();
      _onMinuteBoundary();
      return;
    }

    _advanceActivity();
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
    wasSkipped = true;
    skipForward();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  @override
  void onBackgroundDrift(int driftSeconds) {
    if (!_hasStarted || _isCompleted || _isPaused) return;
    int remaining = driftSeconds;
    while (remaining > 0 && !_isCompleted) {
      if (_secondsInMinute > remaining) {
        _secondsInMinute -= remaining;
        remaining = 0;
      } else {
        remaining -= _secondsInMinute;
        _secondsInMinute = 0;
        _timer?.cancel();
        if (_isBlockRest) {
          _currentBlockIndex++;
          if (_currentBlockIndex >= blocks.length) {
            _completeTraining();
            return;
          }
          _currentMinuteInBlock = 1;
          _isBlockRest = false;
          _isResting = false;
          _currentActivityIndex = 0;
          _secondsInMinute = 60;
        } else {
          _currentMinuteInBlock++;
          if (_currentMinuteInBlock > _currentBlockRepeats) {
            if (_currentBlockIndex < blocks.length - 1 && _currentBlock.rest > 0) {
              _isBlockRest = true;
              _isResting = false;
              _secondsInMinute = _currentBlock.rest;
            } else if (_currentBlockIndex < blocks.length - 1) {
              _currentBlockIndex++;
              _currentMinuteInBlock = 1;
              _isBlockRest = false;
              _isResting = false;
              _currentActivityIndex = 0;
              _secondsInMinute = 60;
            } else {
              _completeTraining();
              return;
            }
          } else {
            _isResting = false;
            _currentActivityIndex = 0;
            _secondsInMinute = 60;
          }
        }
      }
    }
    shouldPlayCountdownJingle = _secondsInMinute == 3;
    _startTimer();
    notifyListeners();
  }

  void _advanceActivity() {
    if (_isResting) return;  // can't advance during rest, wait for minute boundary

    _saveHistory();
    _currentActivityIndex++;

    if (_currentActivityIndex >= _currentBlock.activities.length) {
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
        shouldPlayCountdownJingle = _hasStarted && _secondsInMinute == 3;
        notifyListeners();
      } else {
        timer.cancel();
        if (!_hasStarted) {
          startTraining();
        } else if (_isBlockRest) {
          _advanceToNextBlock();
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
    _isBlockRest = false;
    notifyListeners();
    _startTimer();
  }

  void _onMinuteBoundary() {
    _currentMinuteInBlock++;
    if (_currentMinuteInBlock > _currentBlockRepeats) {
      // block finished — check for rest between blocks
      if (_currentBlockIndex < blocks.length - 1 && _currentBlock.rest > 0) {
        _startBlockRest();
      } else {
        _tryAdvanceBlock();
      }
    } else {
      _startMinute();
    }
  }

  void _startBlockRest() {
    _isBlockRest = true;
    _isResting = false;
    _secondsInMinute = _currentBlock.rest;
    notifyListeners();
    _startTimer();
  }

  void _tryAdvanceBlock() {
    if (_currentBlockIndex < blocks.length - 1) {
      _advanceToNextBlock();
    } else {
      _completeTraining();
    }
  }

  void _advanceToNextBlock() {
    _currentBlockIndex++;
    if (_currentBlockIndex >= blocks.length) {
      _completeTraining();
      return;
    }
    _currentMinuteInBlock = 1;
    _isBlockRest = false;
    _startMinute();
  }

  void _completeTraining() {
    _timer?.cancel();
    stopElapsedTimer();
    _isCompleted = true;
    notifyListeners();
  }

  void _saveHistory() {
    _history.add(_EmomHistoryEntry(
      blockIndex: _currentBlockIndex,
      minute: _currentMinuteInBlock,
      activityIndex: _currentActivityIndex,
      secondsRemaining: _secondsInMinute,
      isResting: _isResting,
      isBlockRest: _isBlockRest,
    ));
  }

  void _restoreHistory() {
    final entry = _history.removeLast();
    _currentBlockIndex = entry.blockIndex;
    _currentMinuteInBlock = entry.minute;
    _currentActivityIndex = entry.activityIndex;
    _secondsInMinute = entry.secondsRemaining;
    _isResting = entry.isResting;
    _isBlockRest = entry.isBlockRest;
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
  final int blockIndex;
  final int minute;
  final int activityIndex;
  final int secondsRemaining;
  final bool isResting;
  final bool isBlockRest;

  _EmomHistoryEntry({
    required this.blockIndex,
    required this.minute,
    required this.activityIndex,
    required this.secondsRemaining,
    required this.isResting,
    required this.isBlockRest,
  });
}

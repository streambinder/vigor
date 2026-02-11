import 'dart:async';
import '../models/training.dart';
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import 'timer_controller.dart';
import 'training_interval.dart';

/// Controller for interval-based timer mode
/// Linear progression through pre-built work/rest intervals
/// Used for: strength, circuit, hiit, mobility, endurance
class IntervalController extends TimerController {
  final Training training;
  final List<TrainingInterval> _intervals;
  final List<int> _history = [];

  Timer? _timer;
  int _currentIntervalIndex = 0;
  int _remainingSeconds = 0;
  bool _isPaused = false;
  bool _isCompleted = false;
  bool _hasStarted = false;

  IntervalController({required this.training})
      : _intervals = _buildIntervals(training);

  @override
  int get remainingSeconds => _remainingSeconds;

  @override
  bool get isPaused => _isPaused;

  @override
  bool get isCompleted => _isCompleted;

  @override
  bool get hasStarted => _hasStarted;

  @override
  bool get canGoBack => _history.isNotEmpty;

  @override
  TrainingInterval? get currentInterval =>
      _currentIntervalIndex < _intervals.length ? _intervals[_currentIntervalIndex] : null;

  @override
  List<TrainingInterval> get upcomingIntervals =>
      _currentIntervalIndex + 1 < _intervals.length
          ? _intervals.sublist(_currentIntervalIndex + 1)
          : [];

  @override
  void startCountdown() {
    _remainingSeconds = 5;
    _startTimer();
  }

  @override
  void startTraining() {
    if (_hasStarted) return;
    _timer?.cancel();
    _hasStarted = true;
    _startCurrentInterval();
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
    _timer?.cancel();
    _history.add(_currentIntervalIndex);
    _currentIntervalIndex++;

    if (_currentIntervalIndex >= _intervals.length) {
      _completeTraining();
    } else {
      _startCurrentInterval();
    }
  }

  @override
  void skipBackward() {
    if (_history.isEmpty) return;
    wasSkipped = true;
    _timer?.cancel();
    _currentIntervalIndex = _history.removeLast();
    _isCompleted = false;
    _startCurrentInterval();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  void _startTimer() {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_isPaused) return;

      if (_remainingSeconds > 0) {
        _remainingSeconds--;
        // play countdown jingle at 3, 2, 1 seconds (only during training, not initial countdown)
        shouldPlayCountdownJingle = _hasStarted && _remainingSeconds >= 1 && _remainingSeconds <= 3;
        notifyListeners();
      } else {
        timer.cancel();
        if (!_hasStarted) {
          startTraining();
        } else {
          _onIntervalCompleted();
        }
      }
    });
  }

  void _startCurrentInterval() {
    if (_currentIntervalIndex >= _intervals.length) {
      _completeTraining();
      return;
    }
    _remainingSeconds = _intervals[_currentIntervalIndex].duration;
    notifyListeners();
    _startTimer();
  }

  void _onIntervalCompleted() {
    // timer-driven transition, not user skip - preserve wasSkipped = false
    _timer?.cancel();
    _history.add(_currentIntervalIndex);
    _currentIntervalIndex++;

    if (_currentIntervalIndex >= _intervals.length) {
      _completeTraining();
    } else {
      _startCurrentInterval();
    }
  }

  void _completeTraining() {
    _timer?.cancel();
    _isCompleted = true;
    notifyListeners();
  }

  /// Build flat interval list from training structure
  static List<TrainingInterval> _buildIntervals(Training training) {
    final intervals = <TrainingInterval>[];
    int activityCounter = 0;
    int totalActivities = 0;

    for (final routine in training.routines) {
      for (final block in routine.blocks) {
        totalActivities += block.activities.length * block.repeats;
      }
    }

    void addRestInterval(TrainingInterval restInterval) {
      if (intervals.isNotEmpty && intervals.last.type == IntervalType.rest) {
        if (restInterval.duration > intervals.last.duration) {
          intervals.removeLast();
          intervals.add(restInterval);
        }
      } else {
        intervals.add(restInterval);
      }
    }

    int routineIndex = 0;
    for (final routine in training.routines) {
      routineIndex++;
      int blockIndex = 0;

      for (final block in routine.blocks) {
        blockIndex++;

        for (int repeat = 1; repeat <= block.repeats; repeat++) {
          for (int actIdx = 0; actIdx < block.activities.length; actIdx++) {
            final activity = block.activities[actIdx];
            final isLastActivityInRepeat = actIdx == block.activities.length - 1;
            activityCounter++;
            final exercise = _parseExercise(activity.detail);

            final workDuration = activity.duration > 0
                ? activity.duration
                : (activity.reps > 0 ? activity.reps * 4 : 30);
            intervals.add(TrainingInterval(
              type: IntervalType.work,
              duration: workDuration,
              routineName: routine.type,
              activityName: activity.displayName,
              activity: activity,
              exercise: exercise,
              activityNumber: activityCounter,
              totalActivities: totalActivities,
              blockNumber: blockIndex,
              totalBlocks: routine.blocks.length,
              routineNumber: routineIndex,
              totalRoutines: training.routines.length,
            ));

            final isLastRoutine = routineIndex == training.routines.length;
            final isLastBlock = blockIndex == routine.blocks.length;
            final isLastRepeat = repeat == block.repeats;
            final isLastActivityOfTraining =
                isLastRoutine && isLastBlock && isLastRepeat && isLastActivityInRepeat;

            if (isLastActivityInRepeat && block.rest > 0 && !isLastActivityOfTraining) {
              addRestInterval(TrainingInterval(
                type: IntervalType.rest,
                duration: block.rest,
                routineName: routine.type,
                activityNumber: activityCounter,
                totalActivities: totalActivities,
                blockNumber: blockIndex,
                totalBlocks: routine.blocks.length,
                routineNumber: routineIndex,
                totalRoutines: training.routines.length,
              ));
            } else if (!isLastActivityInRepeat && activity.rest > 0) {
              addRestInterval(TrainingInterval(
                type: IntervalType.rest,
                duration: activity.rest,
                routineName: routine.type,
                activityNumber: activityCounter,
                totalActivities: totalActivities,
                blockNumber: blockIndex,
                totalBlocks: routine.blocks.length,
                routineNumber: routineIndex,
                totalRoutines: training.routines.length,
              ));
            }
          }
        }
      }

      final isLastRoutine = routineIndex == training.routines.length;
      if (routine.rest > 0 && !isLastRoutine) {
        addRestInterval(TrainingInterval(
          type: IntervalType.rest,
          duration: routine.rest,
          routineName: routine.type,
          activityNumber: activityCounter,
          totalActivities: totalActivities,
          blockNumber: routine.blocks.length,
          totalBlocks: routine.blocks.length,
          routineNumber: routineIndex,
          totalRoutines: training.routines.length,
        ));
      }
    }

    return intervals;
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

import 'package:flutter/material.dart';
import 'package:wakelock_plus/wakelock_plus.dart';
import '../design/tokens.dart';
import '../models/training.dart';
import '../services/audio_service.dart';
import '../services/preferences_service.dart';
import '../services/service_locator.dart';
import 'timer_controller.dart';
import 'timer_mode.dart';
import 'training_interval.dart';
import 'routine_segment.dart';
import 'interval_controller.dart';
import 'emom_controller.dart';
import 'amrap_controller.dart';
import 'for_time_controller.dart';

/// Timer orchestration ChangeNotifier. Manages segment lifecycle, audio,
/// wakelock, background drift, and controller transitions.
/// Used by InlineTimerSection in TrainingDetailsScreen.
class WorkoutTimerNotifier extends ChangeNotifier with WidgetsBindingObserver {
  final Training training;
  final PreferencesService prefs;
  final ServiceLocator serviceLocator;

  late final List<RoutineSegment> _segments;
  int _currentSegmentIndex = 0;
  TimerController? _controller;
  bool _workoutCompleted = false;
  bool _isSubmitting = false;
  String? _previousIntervalKey;
  final AudioService _audioService = AudioService();
  String? _methodologyStats;
  int _accumulatedElapsedSeconds = 0;

  WorkoutTimerNotifier({
    required this.training,
    required this.prefs,
    required this.serviceLocator,
  });

  // --- public getters ---

  TimerController? get controller => _controller;
  bool get workoutCompleted => _workoutCompleted;
  bool get isSubmitting => _isSubmitting;
  String? get methodologyStats => _methodologyStats;
  int get accumulatedElapsedSeconds => _accumulatedElapsedSeconds;

  int get totalElapsedSeconds => _workoutCompleted
      ? _accumulatedElapsedSeconds
      : _accumulatedElapsedSeconds + (_controller?.elapsedSeconds ?? 0);

  TimerMode? get currentMode =>
      _currentSegmentIndex < _segments.length
          ? _segments[_currentSegmentIndex].mode
          : null;

  /// current exercise/activity name for compact bar display
  String? get currentExerciseName => _controller?.currentInterval?.activityName;

  /// remaining seconds on the current interval/timer
  int get currentRemainingSeconds => _controller?.remainingSeconds ?? 0;

  /// overall training progress as 0.0-1.0 fraction based on exercises done vs total
  double get progress {
    if (_workoutCompleted) return 1.0;
    final total = _totalActivities;
    if (total <= 0) return 0.0;
    final current = _controller?.currentInterval?.activityNumber ?? 0;
    // activityNumber is 1-based; (current - 1) = completed, current = in progress
    return ((current - 1) / total).clamp(0.0, 1.0);
  }

  int get _totalActivities {
    int count = 0;
    for (final routine in training.routines) {
      for (final block in routine.blocks) {
        count += block.activities.length * block.repeats;
      }
    }
    return count;
  }

  Color phaseColor(Brightness brightness) {
    final controller = _controller;
    if (controller == null || !controller.hasStarted) return VigorColors.persimmon;
    final interval = controller.currentInterval;
    if (interval == null) return VigorColors.gold;
    return interval.type == IntervalType.rest ? VigorColors.indigo : VigorColors.persimmon;
  }

  // --- lifecycle ---

  void initialize() {
    WidgetsBinding.instance.addObserver(this);
    WakelockPlus.enable();
    _segments = RoutineSegment.buildSegments(
      training.routines,
      training.methodology,
    );
    _initializeAudio();
    _startCurrentSegment();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    WakelockPlus.disable();
    _controller?.removeListener(_onControllerUpdate);
    _controller?.dispose();
    super.dispose();
  }

  // --- segment management ---

  void _startCurrentSegment() {
    if (_currentSegmentIndex >= _segments.length) {
      _completeWorkout();
      return;
    }
    _controller?.removeListener(_onControllerUpdate);
    _controller?.dispose();
    _controller = _createController(_segments[_currentSegmentIndex]);
    _controller!.addListener(_onControllerUpdate);
    _controller!.startCountdown();
  }

  TimerController _createController(RoutineSegment segment) {
    switch (segment.mode) {
      case TimerMode.interval:
        return IntervalController(training: training);

      case TimerMode.emom:
        if (segment.blocks.isEmpty) return IntervalController(training: training);
        return EmomController(
          training: training,
          routine: segment.routine,
          blocks: segment.blocks,
        );

      case TimerMode.amrap:
        final block = segment.blocks.isNotEmpty ? segment.blocks.first : null;
        if (block == null) return IntervalController(training: training);
        return AmrapController(
          training: training,
          routine: segment.routine,
          block: block,
          totalSeconds: training.duration > 0 ? training.duration : 15 * 60,
        );

      case TimerMode.forTime:
        final block = segment.blocks.isNotEmpty ? segment.blocks.first : null;
        if (block == null) return IntervalController(training: training);
        return ForTimeController(
          training: training,
          routine: segment.routine,
          block: block,
        );
    }
  }

  void _onControllerUpdate() {
    if (_controller?.isCompleted == true) {
      _accumulatedElapsedSeconds += _controller?.elapsedSeconds ?? 0;
      _captureMethodologyStats();
      _advanceToNextSegment();
      return;
    }

    if (_controller?.shouldPlayCountdownJingle == true) {
      _playWhistleIfEnabled();
      _controller?.shouldPlayCountdownJingle = false;
    }

    // detect interval transition — reset skip flag
    final currentKey = _intervalKey(_controller?.currentInterval);
    if (_controller?.hasStarted == true &&
        currentKey != null &&
        _previousIntervalKey != null &&
        currentKey != _previousIntervalKey) {
      _controller?.wasSkipped = false;
    }
    _previousIntervalKey = currentKey;

    notifyListeners();
  }

  void _captureMethodologyStats() {
    if (_currentSegmentIndex >= _segments.length) return;
    final segment = _segments[_currentSegmentIndex];
    final controller = _controller;
    if (controller == null) return;
    if (segment.routine.type.toLowerCase() != 'work') return;

    switch (segment.mode) {
      case TimerMode.amrap:
        _methodologyStats =
            'Rounds completed: ${(controller as AmrapController).currentRound - 1}. ';
        break;
      case TimerMode.forTime:
        _methodologyStats =
            'Time to complete: ${(controller as ForTimeController).globalSeconds} seconds. ';
        break;
      default:
        break;
    }
  }

  void _advanceToNextSegment() {
    _currentSegmentIndex++;
    if (_currentSegmentIndex >= _segments.length) {
      _completeWorkout();
    } else {
      _startCurrentSegment();
    }
  }

  void _completeWorkout() {
    _workoutCompleted = true;
    notifyListeners();
  }

  // --- actions ---

  void handleTap(TapUpDetails details, double sectionWidth) {
    final controller = _controller;
    if (controller == null) return;

    _audioService.stopWhistle();

    final segment = _currentSegmentIndex < _segments.length
        ? _segments[_currentSegmentIndex]
        : null;

    if (segment != null && segment.mode.isUserPaced) {
      controller.onUserAction();
    } else {
      if (details.globalPosition.dx < sectionWidth / 2) {
        controller.skipBackward();
      } else {
        controller.skipForward();
      }
    }
  }

  void stopWhistle() => _audioService.stopWhistle();

  /// mark submitting state for UI feedback
  set isSubmitting(bool value) {
    _isSubmitting = value;
    notifyListeners();
  }

  /// capture stats when exiting early (before segment completes naturally)
  void captureEarlyExitStats() {
    _captureMethodologyStats();
    _accumulatedElapsedSeconds += _controller?.elapsedSeconds ?? 0;
  }

  // --- audio ---

  Future<void> _initializeAudio() async {
    await _audioService.initialize();
    await _audioService.setDuckOtherAudio(prefs.duckOtherAudio);
  }

  void _playWhistleIfEnabled() {
    if (prefs.intervalJingle) _audioService.playWhistle();
  }

  // --- background drift ---

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _audioService.reactivate();
      _controller?.compensateBackgroundDrift();
      if (_controller?.shouldPlayCountdownJingle == true) {
        _playWhistleIfEnabled();
        _controller?.shouldPlayCountdownJingle = false;
      }
    }
  }

  // --- utils ---

  String formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    final secs = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  String? _intervalKey(TrainingInterval? interval) {
    if (interval == null) return null;
    return '${interval.type}:${interval.activityNumber}:${interval.blockNumber}:${interval.routineNumber}';
  }
}

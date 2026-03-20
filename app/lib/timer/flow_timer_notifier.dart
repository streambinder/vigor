import 'package:flutter/material.dart';
import 'package:wakelock_plus/wakelock_plus.dart';

import '../design/tokens.dart';
import '../models/flow_pose.dart';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity.dart';
import '../models/training_prompt.dart';
import '../models/llm_step.dart';
import '../models/llm_prompt.dart';
import 'base_timer_notifier.dart';
import 'interval_controller.dart';
import 'timer_controller.dart';
import 'timer_mode.dart';
import 'training_interval.dart';

/// Thin notifier that wires FlowPose list into IntervalController + InlineTimerSection.
/// Mirrors WorkoutTimerNotifier's public surface so InlineTimerSection works unchanged.
class FlowTimerNotifier extends BaseTimerNotifier with WidgetsBindingObserver {
  final List<FlowPose> poses;

  late final _FlowIntervalController _controller;
  bool _workoutCompleted = false;
  bool _isSubmitting = false;

  FlowTimerNotifier({required this.poses}) {
    _controller = _FlowIntervalController(poses: poses);
    _controller.addListener(_onControllerUpdate);
  }

  // --- surface InlineTimerSection reads ---

  @override
  TimerController? get controller => _workoutCompleted ? null : _controller;
  @override
  bool get workoutCompleted => _workoutCompleted;
  @override
  bool get isSubmitting => _isSubmitting;
  String? get methodologyStats => null;
  int get accumulatedElapsedSeconds => _controller.elapsedSeconds;
  @override
  int get totalElapsedSeconds => _controller.elapsedSeconds;
  @override
  TimerMode? get currentMode => TimerMode.interval;
  String? get currentExerciseName => _controller.currentInterval?.activityName;
  int get currentRemainingSeconds => _controller.remainingSeconds;

  @override
  double get progress {
    if (_workoutCompleted) return 1.0;
    final total = poses.length;
    if (total == 0) return 0.0;
    return ((_controller.currentInterval?.activityNumber ?? 1) - 1) / total;
  }

  @override
  Color phaseColor(Brightness brightness) {
    if (_workoutCompleted) return VigorColors.gold;
    final interval = _controller.currentInterval;
    if (interval == null) return brightness == Brightness.dark ? VigorColors.byakurokuLight : VigorColors.byakuroku;
    return interval.type == IntervalType.rest ? VigorColors.indigo : (brightness == Brightness.dark ? VigorColors.byakurokuLight : VigorColors.byakuroku);
  }

  @override
  String formatDuration(int seconds) {
    final m = seconds ~/ 60;
    final s = seconds % 60;
    return '${m.toString().padLeft(2, '0')}:${s.toString().padLeft(2, '0')}';
  }

  set isSubmitting(bool value) {
    _isSubmitting = value;
    notifyListeners();
  }

  // callbacks wired by screen
  Function()? onNotificationStop;
  Function()? onNotificationComplete;

  @override
  void stopWhistle() {}

  @override
  void handleTap(TapUpDetails details, double sectionWidth) {
    if (details.globalPosition.dx < sectionWidth / 2) {
      _controller.skipBackward();
    } else {
      _controller.skipForward();
    }
  }

  void captureEarlyExitStats() {}

  void initialize() {
    WidgetsBinding.instance.addObserver(this);
    WakelockPlus.enable();
    _controller.startCountdown();
  }

  @override
  void dispose() {
    WidgetsBinding.instance.removeObserver(this);
    WakelockPlus.disable();
    _controller.removeListener(_onControllerUpdate);
    _controller.dispose();
    super.dispose();
  }

  @override
  void didChangeAppLifecycleState(AppLifecycleState state) {
    if (state == AppLifecycleState.resumed) {
      _controller.compensateBackgroundDrift();
    }
  }

  void _onControllerUpdate() {
    if (_controller.isCompleted && !_workoutCompleted) {
      _workoutCompleted = true;
    }
    notifyListeners();
  }
}

/// IntervalController subclass built from FlowPose — no Training/Routine/Block hierarchy needed
/// at runtime; the synthetic Training is only used by _buildIntervals to produce intervals.
class _FlowIntervalController extends IntervalController {
  _FlowIntervalController({required List<FlowPose> poses})
      : super(training: _syntheticTraining(poses));

  static final _emptyPrompt = TrainingPrompt(
    reasoning: LLMStep(model: '', prompt: LLMPrompt(system: '', user: '')),
    structuring: LLMStep(model: '', prompt: LLMPrompt(system: '', user: '')),
  );

  // minimal Training shell — only `routines` matters for interval building
  static Training _syntheticTraining(List<FlowPose> poses) => Training(
        id: '',
        name: '',
        description: '',
        methodology: 'interval',
        duration: 0,
        equipment: [],
        goals: [],
        muscles: [],
        request: '',
        references: [],
        factIndices: [],
        routines: [
          Routine(
            id: '',
            trainingId: '',
            type: 'flow',
            rest: 0,
            blocks: [
              Block(
                id: '',
                routineId: '',
                repeats: 1,
                rest: 0,
                activities: poses
                    .map((p) => Activity(
                          id: p.exerciseId,
                          blockId: '',
                          exerciseId: p.exerciseId,
                          name: p.name,
                          duration: p.duration,
                          reps: 0,
                          weightKg: 0,
                          modifiers: [],
                          rest: p.rest,
                          detail: p.detail,
                        ))
                    .toList(),
              ),
            ],
          ),
        ],
        prompt: _emptyPrompt,
        completedAt: null,
        completedIn: null,
        hasHealthSession: false,
        createdAt: DateTime.now(),
        userId: '',
        parentId: null,
        gymId: null,
        gym: null,
      );
}

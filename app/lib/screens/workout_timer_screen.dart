import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:wakelock_plus/wakelock_plus.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/training.dart';
import '../services/audio_service.dart';
import '../services/preferences_service.dart';
import '../timer/timer_controller.dart';
import '../timer/timer_mode.dart';
import '../timer/training_interval.dart';
import '../timer/routine_segment.dart';
import '../timer/interval_controller.dart';
import '../timer/emom_controller.dart';
import '../timer/amrap_controller.dart';
import '../timer/for_time_controller.dart';
import '../utils/feedback_modal.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/timer/completion_screen.dart';
import '../widgets/timer/countdown_overlay.dart';
import '../widgets/timer/counters_row.dart';
import '../widgets/timer/exercise_display.dart';
import '../widgets/timer/rest_display.dart';
import '../widgets/timer/timer_controls.dart';
import '../widgets/timer/upcoming_list.dart';
import '../widgets/timer/remaining_exercises_list.dart';
import '../widgets/timer/emom_display.dart';
import '../widgets/timer/amrap_display.dart';
import '../widgets/timer/for_time_display.dart';
import '../services/service_locator.dart';

/// Unified workout timer screen supporting multiple timer modes
/// Handles hybrid mode switching between warmup/work/cooldown routines
class WorkoutTimerScreen extends StatefulWidget {
  final Training training;

  const WorkoutTimerScreen({super.key, required this.training});

  @override
  State<WorkoutTimerScreen> createState() => _WorkoutTimerScreenState();
}

class _WorkoutTimerScreenState extends State<WorkoutTimerScreen> {
  late final List<RoutineSegment> _segments;
  int _currentSegmentIndex = 0;
  TimerController? _controller;
  bool _workoutCompleted = false;
  String? _previousIntervalKey;
  final AudioService _audioService = AudioService();

  // methodology-specific stats captured on work segment completion
  String? _methodologyStats;

  // accumulated elapsed seconds across all segments
  int _accumulatedElapsedSeconds = 0;

  @override
  void initState() {
    super.initState();
    WakelockPlus.enable();
    _segments = RoutineSegment.buildSegments(
      widget.training.routines,
      widget.training.methodology,
    );
    _audioService.initialize();
    _startCurrentSegment();
  }

  @override
  void dispose() {
    WakelockPlus.disable();
    _controller?.removeListener(_onControllerUpdate);
    _controller?.dispose();
    super.dispose();
  }

  void _startCurrentSegment() {
    if (_currentSegmentIndex >= _segments.length) {
      _completeWorkout();
      return;
    }

    final segment = _segments[_currentSegmentIndex];
    _controller?.removeListener(_onControllerUpdate);
    _controller?.dispose();

    _controller = _createController(segment);
    _controller!.addListener(_onControllerUpdate);
    _controller!.startCountdown();
  }

  TimerController _createController(RoutineSegment segment) {
    switch (segment.mode) {
      case TimerMode.interval:
        // for interval mode, we use the full training but only process this routine
        return IntervalController(training: widget.training);

      case TimerMode.emom:
        if (segment.blocks.isEmpty) return IntervalController(training: widget.training);
        return EmomController(
          training: widget.training,
          routine: segment.routine,
          blocks: segment.blocks,
        );

      case TimerMode.amrap:
        final block = segment.blocks.isNotEmpty ? segment.blocks.first : null;
        if (block == null) return IntervalController(training: widget.training);
        // use training duration or default to 15 minutes
        final totalSeconds = widget.training.duration > 0
            ? widget.training.duration
            : 15 * 60;
        return AmrapController(
          training: widget.training,
          routine: segment.routine,
          block: block,
          totalSeconds: totalSeconds,
        );

      case TimerMode.forTime:
        final block = segment.blocks.isNotEmpty ? segment.blocks.first : null;
        if (block == null) return IntervalController(training: widget.training);
        return ForTimeController(
          training: widget.training,
          routine: segment.routine,
          block: block,
        );
    }
  }

  void _onControllerUpdate() {
    if (!mounted) return;

    // check if current segment is complete
    if (_controller?.isCompleted == true) {
      _accumulatedElapsedSeconds += _controller?.elapsedSeconds ?? 0;
      _captureMethodologyStats();
      _playJingleIfEnabled();
      _advanceToNextSegment();
      return;
    }

    // play countdown jingle for last 3 seconds of interval
    if (_controller?.shouldPlayCountdownJingle == true) {
      _playJingleIfEnabled();
      _controller?.shouldPlayCountdownJingle = false;
    }

    // detect interval transition by comparing stable keys
    final currentInterval = _controller?.currentInterval;
    final currentKey = _intervalKey(currentInterval);
    if (_controller?.hasStarted == true &&
        currentKey != null &&
        _previousIntervalKey != null &&
        currentKey != _previousIntervalKey) {
      // only play jingle on natural timer transitions, not skips
      if (_controller?.wasSkipped != true) {
        _playJingleIfEnabled();
      }
      _controller?.wasSkipped = false;
    }
    _previousIntervalKey = currentKey;

    setState(() {});
  }

  void _captureMethodologyStats() {
    if (_currentSegmentIndex >= _segments.length) return;
    final segment = _segments[_currentSegmentIndex];
    final controller = _controller;
    if (controller == null) return;

    // only capture stats for work routine with AMRAP/ForTime
    if (segment.routine.type.toLowerCase() != 'work') return;

    switch (segment.mode) {
      case TimerMode.amrap:
        final amrapController = controller as AmrapController;
        final completedRounds = amrapController.currentRound - 1;
        _methodologyStats = 'Rounds completed: $completedRounds. ';
        break;
      case TimerMode.forTime:
        final forTimeController = controller as ForTimeController;
        _methodologyStats = 'Time to complete: ${forTimeController.globalSeconds} seconds. ';
        break;
      default:
        break;
    }
  }

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    final secs = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  String? _intervalKey(TrainingInterval? interval) {
    if (interval == null) return null;
    return '${interval.type}:${interval.activityNumber}:${interval.blockNumber}:${interval.routineNumber}';
  }

  void _playJingleIfEnabled() {
    if (context.read<PreferencesService>().intervalJingle) {
      _audioService.playJingle();
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
    setState(() {
      _workoutCompleted = true;
    });
  }

  void _handleTap(TapUpDetails details) {
    final controller = _controller;
    if (controller == null) return;

    final segment = _currentSegmentIndex < _segments.length
        ? _segments[_currentSegmentIndex]
        : null;

    if (segment != null && segment.mode.isUserPaced) {
      // for user-paced modes, tap anywhere triggers action
      controller.onUserAction();
    } else {
      // for interval mode, left/right tap navigation
      final screenWidth = MediaQuery.of(context).size.width;
      if (details.globalPosition.dx < screenWidth / 2) {
        controller.skipBackward();
      } else {
        controller.skipForward();
      }
    }
  }

  Future<void> _showFeedbackAndComplete() async {
    final result = await FeedbackModal.show(
      context,
      widget.training,
      messagePrefix: _methodologyStats,
      elapsedSeconds: _accumulatedElapsedSeconds,
    );
    if (result == null) return;
    await _markTrainingComplete(result);
    if (mounted) {
      Navigator.of(context).pop(true);
    }
  }

  Future<void> _markTrainingComplete(FeedbackResult result) async {
    final response = await context.read<ServiceLocator>().trainingService.completeTraining(
      widget.training.id,
      feedback: result.feedback,
      activityFeedback: result.activityFeedback,
      completedIn: result.completedIn,
    );
    if (!response.isSuccess && mounted) {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).failedToMarkComplete,
        rawError: response.error,
      );
    }
  }

  Color get _phaseColor {
    final controller = _controller;
    if (controller == null || !controller.hasStarted) return VigorColors.persimmon;
    final interval = controller.currentInterval;
    if (interval == null) return VigorColors.gold;
    return interval.type == IntervalType.rest ? VigorColors.indigo : VigorColors.persimmon;
  }

  TimerMode? get _currentMode {
    if (_currentSegmentIndex >= _segments.length) return null;
    return _segments[_currentSegmentIndex].mode;
  }

  @override
  Widget build(BuildContext context) {
    final controller = _controller;

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: Text(widget.training.name),
        leading: AdaptiveIconButton(
          icon: const Icon(Icons.close),
          onPressed: _showExitDialog,
        ),
      ),
      body: _workoutCompleted
          ? CompletionScreen(
              trainingName: widget.training.name,
              onDone: _showFeedbackAndComplete,
            )
          : controller == null
              ? const Center(child: CircularProgressIndicator())
              : !controller.hasStarted
                  ? CountdownOverlay(
                      remainingSeconds: controller.remainingSeconds,
                      onTap: controller.startTraining,
                    )
                  : _buildTimerScreen(controller),
    );
  }

  Widget _buildTimerScreen(TimerController controller) {
    final interval = controller.currentInterval;
    if (interval == null) return const SizedBox.shrink();

    final isDark = Theme.of(context).brightness == Brightness.dark;
    final mode = _currentMode ?? TimerMode.interval;

    return GestureDetector(
      onTapUp: _handleTap,
      child: Container(
        color: isDark ? VigorColors.darkBackground : VigorColors.lightBackground,
        child: SafeArea(
          child: Column(
            children: [
              Expanded(
                child: LayoutBuilder(
                  builder: (context, constraints) {
                    final mainAreaHeight = constraints.maxHeight;
                    return SingleChildScrollView(
                      padding: VigorSpacing.paddingLg,
                      child: Column(
                        children: [
                          // main content area - fills viewport and centers content
                          SizedBox(
                            height: mainAreaHeight - VigorSpacing.lg * 2,
                            child: Center(
                              child: _buildModeMainContent(controller, interval, mode),
                            ),
                          ),
                          // exercises list below - visible on scroll
                          _buildExercisesList(controller, mode),
                        ],
                      ),
                    );
                  },
                ),
              ),
              Padding(
                padding: VigorSpacing.paddingLg,
                child: _buildControls(controller, mode),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildModeMainContent(
    TimerController controller,
    TrainingInterval interval,
    TimerMode mode,
  ) {
    switch (mode) {
      case TimerMode.emom:
        final emomController = controller as EmomController;
        return EmomDisplay(
          interval: interval,
          secondsRemaining: controller.remainingSeconds,
          currentMinute: emomController.currentRound,
          totalMinutes: emomController.totalRounds,
          activityIndex: emomController.activityIndex,
          totalActivities: emomController.totalActivities,
          isResting: emomController.isResting,
          currentBlock: emomController.currentBlockIndex + 1,
          totalBlocks: emomController.totalBlocks,
        );

      case TimerMode.amrap:
        final amrapController = controller as AmrapController;
        return AmrapDisplay(
          interval: interval,
          globalSecondsRemaining: amrapController.globalSeconds,
          currentRound: amrapController.currentRound,
          activitiesInRound: amrapController.activitiesInRound,
          activitiesPerRound: amrapController.activitiesPerRound,
        );

      case TimerMode.forTime:
        final forTimeController = controller as ForTimeController;
        return ForTimeDisplay(
          interval: interval,
          elapsedSeconds: forTimeController.globalSeconds,
          currentRound: forTimeController.currentRound,
          totalRounds: forTimeController.totalRounds,
        );

      case TimerMode.interval:
        return _buildIntervalMainContent(controller, interval);
    }
  }

  Widget _buildIntervalMainContent(TimerController controller, TrainingInterval interval) {
    final isRest = interval.type == IntervalType.rest;
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        if (!isRest) _buildActivityName(interval),
        if (!isRest) const SizedBox(height: VigorSpacing.md),
        if (!isRest) CountersRow(interval: interval),
        if (!isRest) const SizedBox(height: VigorSpacing.lg),
        isRest
            ? RestDisplay(remainingSeconds: controller.remainingSeconds)
            : ExerciseDisplay(
                interval: interval,
                remainingSeconds: controller.remainingSeconds,
                phaseColor: _phaseColor,
              ),
      ],
    );
  }

  Widget _buildExercisesList(TimerController controller, TimerMode mode) {
    final intervals = controller.upcomingIntervals;
    if (intervals.isEmpty) return const SizedBox.shrink();

    if (mode == TimerMode.interval) {
      return Padding(
        padding: const EdgeInsets.only(top: VigorSpacing.xl),
        child: UpcomingList(intervals: intervals),
      );
    }

    int? currentRound;
    int? totalRounds;

    switch (mode) {
      case TimerMode.emom:
        final emomController = controller as EmomController;
        currentRound = emomController.currentRound;
        totalRounds = emomController.totalRounds;
        break;
      case TimerMode.amrap:
        currentRound = (controller as AmrapController).currentRound;
        break;
      case TimerMode.forTime:
        final forTimeController = controller as ForTimeController;
        currentRound = forTimeController.currentRound;
        totalRounds = forTimeController.totalRounds;
        break;
      case TimerMode.interval:
        break;
    }

    return Padding(
      padding: const EdgeInsets.only(top: VigorSpacing.xl),
      child: RemainingExercisesList(
        intervals: intervals,
        mode: mode,
        currentRound: currentRound ?? 1,
        totalRounds: totalRounds,
      ),
    );
  }

  Widget _buildControls(TimerController controller, TimerMode mode) {
    // all modes get full controls - skip forward advances activity
    return TimerControls(
      isPaused: controller.isPaused,
      canGoBack: controller.canGoBack,
      phaseColor: _phaseColor,
      onPauseToggle: controller.togglePause,
      onSkipForward: controller.skipForward,
      onSkipBackward: controller.skipBackward,
    );
  }

  Widget _buildActivityName(TrainingInterval interval) {
    return Text(
      interval.activityName?.toUpperCase() ?? '',
      style: VigorTypography.title.copyWith(
        color: VigorColors.textPrimary(context),
      ),
      textAlign: TextAlign.center,
    );
  }

  void _showExitDialog() {
    final l10n = AppLocalizations.of(context);
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) => AlertDialog(
        title: Text(l10n.exitTraining),
        content: Text(l10n.whatWouldYouLikeToDo),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(dialogContext).pop();
              Navigator.of(context).pop();
            },
            child: Text(l10n.exit),
          ),
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: Text(l10n.continueTraining),
          ),
          TextButton(
            onPressed: () async {
              Navigator.of(dialogContext).pop();
              // capture stats if exiting early from work segment
              _captureMethodologyStats();
              _accumulatedElapsedSeconds += _controller?.elapsedSeconds ?? 0;
              final result = await FeedbackModal.show(
                context,
                widget.training,
                messagePrefix: _methodologyStats,
                elapsedSeconds: _accumulatedElapsedSeconds,
              );
              if (result == null) return;
              await _markTrainingComplete(result);
              if (mounted) {
                Navigator.of(context).pop(true);
              }
            },
            child: Text(l10n.markAsComplete),
          ),
        ],
      ),
    );
  }
}

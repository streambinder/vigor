import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/training.dart';
import '../models/activity.dart';
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../utils/exercise_modal.dart';
import '../utils/feedback_modal.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/cached_exercise_image.dart';
import '../services/service_locator.dart';

class TabataTimerScreen extends StatefulWidget {
  final Training training;

  const TabataTimerScreen({super.key, required this.training});

  @override
  State<TabataTimerScreen> createState() => _TabataTimerScreenState();
}

enum IntervalType { work, rest }

class TrainingInterval {
  final IntervalType type;
  final int duration;
  final String routineName;
  final String? activityName;
  final Activity? activity;
  final Exercise? exercise;
  final int activityNumber;
  final int totalActivities;
  final int blockNumber;
  final int totalBlocks;
  final int routineNumber;
  final int totalRoutines;

  TrainingInterval({
    required this.type,
    required this.duration,
    required this.routineName,
    this.activityName,
    this.activity,
    this.exercise,
    required this.activityNumber,
    required this.totalActivities,
    required this.blockNumber,
    required this.totalBlocks,
    required this.routineNumber,
    required this.totalRoutines,
  });
}

class _TabataTimerScreenState extends State<TabataTimerScreen> {
  Timer? _timer;
  late List<TrainingInterval> _intervals;
  int _currentIntervalIndex = 0;
  int _remainingSeconds = 0;
  bool _isPaused = false;
  bool _isCompleted = false;
  bool _hasStarted = false;
  final List<int> _history = [];

  @override
  void initState() {
    super.initState();
    _intervals = _buildIntervals();
    _startInitialCountdown();
  }

  void _startInitialCountdown() {
    setState(() {
      _remainingSeconds = 5;
    });
    _startTimer();
  }

  @override
  void dispose() {
    _timer?.cancel();
    super.dispose();
  }

  List<TrainingInterval> _buildIntervals() {
    final intervals = <TrainingInterval>[];
    int activityCounter = 0;
    int totalActivities = 0;

    for (final routine in widget.training.routines) {
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
    for (final routine in widget.training.routines) {
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

            final workDuration = activity.duration > 0 ? activity.duration : 60;
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
              totalRoutines: widget.training.routines.length,
            ));

            final isLastRoutine = routineIndex == widget.training.routines.length;
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
                totalRoutines: widget.training.routines.length,
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
                totalRoutines: widget.training.routines.length,
              ));
            }
          }
        }
      }

      final isLastRoutine = routineIndex == widget.training.routines.length;
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
          totalRoutines: widget.training.routines.length,
        ));
      }
    }

    return intervals;
  }

  Exercise? _parseExercise(Map<String, dynamic> detail) {
    if (detail.isEmpty) return null;
    try {
      return Exercise.fromJson(detail);
    } catch (_) {
      return null;
    }
  }

  void _startCurrentInterval() {
    if (_currentIntervalIndex >= _intervals.length) {
      _completeTraining();
      return;
    }

    setState(() {
      _remainingSeconds = _intervals[_currentIntervalIndex].duration;
    });
    _startTimer();
  }

  void _startTimer() {
    _timer?.cancel();
    _timer = Timer.periodic(const Duration(seconds: 1), (timer) {
      if (_isPaused) return;

      setState(() {
        if (_remainingSeconds > 0) {
          _remainingSeconds--;
        } else {
          timer.cancel();
          if (!_hasStarted) {
            _startTraining();
          } else {
            _onIntervalCompleted();
          }
        }
      });
    });
  }

  void _startTraining() {
    if (_hasStarted) return;

    _timer?.cancel();
    setState(() {
      _hasStarted = true;
    });
    _startCurrentInterval();
  }

  void _onIntervalCompleted() {
    _skipForward();
  }

  void _skipForward() {
    if (_isCompleted) return;

    _timer?.cancel();
    _history.add(_currentIntervalIndex);

    setState(() {
      _currentIntervalIndex++;
    });

    if (_currentIntervalIndex >= _intervals.length) {
      _completeTraining();
    } else {
      _startCurrentInterval();
    }
  }

  void _skipBackward() {
    if (_history.isEmpty) return;

    _timer?.cancel();

    setState(() {
      _currentIntervalIndex = _history.removeLast();
      _isCompleted = false;
    });

    _startCurrentInterval();
  }

  void _handleTap(TapUpDetails details) {
    final screenWidth = MediaQuery.of(context).size.width;
    if (details.globalPosition.dx < screenWidth / 2) {
      _skipBackward();
    } else {
      _skipForward();
    }
  }

  void _completeTraining() {
    _timer?.cancel();
    setState(() {
      _isCompleted = true;
    });
  }

  Future<void> _showFeedbackAndComplete() async {
    final result = await FeedbackModal.show(context, widget.training);
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
    );
    if (!response.isSuccess && mounted) {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).failedToMarkComplete,
        rawError: response.error,
      );
    }
  }

  void _togglePause() {
    setState(() {
      _isPaused = !_isPaused;
    });
  }

  String _formatTime(int seconds) {
    final minutes = seconds ~/ 60;
    final remainingSeconds = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${remainingSeconds.toString().padLeft(2, '0')}';
  }

  // phase color: persimmon for work, indigo for rest
  Color get _phaseColor {
    if (!_hasStarted) return VigorColors.persimmon;
    if (_currentIntervalIndex >= _intervals.length) return VigorColors.gold;
    return _intervals[_currentIntervalIndex].type == IntervalType.rest
        ? VigorColors.indigo
        : VigorColors.persimmon;
  }

  @override
  Widget build(BuildContext context) {
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: Text(widget.training.name),
        leading: AdaptiveIconButton(
          icon: const Icon(Icons.close),
          onPressed: _showExitDialog,
        ),
      ),
      body: _isCompleted
          ? _buildCompletedScreen()
          : !_hasStarted
              ? _buildStartScreen()
              : _buildTimerScreen(),
    );
  }

  Widget _buildStartScreen() {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: _startTraining,
      child: Container(
        color: isDark ? VigorColors.darkBackground : VigorColors.lightBackground,
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.campaign,
                size: 120,
                color: VigorColors.stone,
              ),
              const SizedBox(height: VigorSpacing.lg),
              Text(
                l10n.tapToStart,
                style: VigorTypography.headline.copyWith(
                  color: VigorColors.textPrimary(context),
                ),
              ),
              const SizedBox(height: VigorSpacing.xxl),
              Text(
                _formatTime(_remainingSeconds),
                style: VigorTypography.dataDisplay.copyWith(
                  color: VigorColors.persimmon.withValues(alpha: 0.5),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCompletedScreen() {
    final l10n = AppLocalizations.of(context);
    return Center(
      child: Padding(
        padding: VigorSpacing.paddingLg,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // kintsugi gold for completion celebration
            Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: VigorColors.gold.withValues(alpha: 0.15),
              ),
              child: const Icon(
                Icons.check_circle,
                size: 80,
                color: VigorColors.gold,
              ),
            ),
            const SizedBox(height: VigorSpacing.xl),
            Text(
              l10n.trainingCompleted,
              style: VigorTypography.title.copyWith(
                color: VigorColors.textPrimary(context),
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: VigorSpacing.md),
            Text(
              l10n.greatJobCompleting(widget.training.name),
              style: VigorTypography.body.copyWith(
                color: VigorColors.textSecondary(context),
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: VigorSpacing.xxl),
            AdaptiveButton(
              onPressed: _showFeedbackAndComplete,
              child: Text(l10n.done),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTimerScreen() {
    if (_currentIntervalIndex >= _intervals.length) {
      return const SizedBox.shrink();
    }

    final interval = _intervals[_currentIntervalIndex];
    final isRest = interval.type == IntervalType.rest;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return GestureDetector(
      onTapUp: _handleTap,
      child: Container(
        color: isDark ? VigorColors.darkBackground : VigorColors.lightBackground,
        child: SafeArea(
          child: Column(
            children: [
              Expanded(
                child: SingleChildScrollView(
                  padding: VigorSpacing.paddingLg,
                  child: Column(
                    children: [
                      if (!isRest) _buildActivityName(interval),
                      if (!isRest) const SizedBox(height: VigorSpacing.md),
                      if (!isRest) _buildCountersRow(interval),
                      if (!isRest) const SizedBox(height: VigorSpacing.lg),
                      isRest ? _buildRestDisplay() : _buildActivityDisplay(interval),
                      const SizedBox(height: VigorSpacing.xl),
                      _buildUpcomingList(),
                    ],
                  ),
                ),
              ),
              Padding(
                padding: VigorSpacing.paddingLg,
                child: _buildControls(),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildUpcomingList() {
    final l10n = AppLocalizations.of(context);
    final remaining = _intervals.sublist(_currentIntervalIndex + 1);
    if (remaining.isEmpty) return const SizedBox.shrink();

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.upcoming,
          style: VigorTypography.label.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        const SizedBox(height: VigorSpacing.sm),
        ...remaining.take(10).map((interval) => _buildUpcomingItem(interval)),
        if (remaining.length > 10)
          Padding(
            padding: const EdgeInsets.only(top: VigorSpacing.sm),
            child: Text(
              '+${remaining.length - 10} more',
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textMuted(context),
              ),
            ),
          ),
      ],
    );
  }

  Widget _buildUpcomingItem(TrainingInterval interval) {
    final l10n = AppLocalizations.of(context);
    final isRest = interval.type == IntervalType.rest;
    final name = isRest ? l10n.rest : (interval.activityName ?? '');

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs + 2),
      child: Row(
        children: [
          if (!isRest &&
              interval.exercise != null &&
              CachedExerciseImage.isValidUrl(interval.exercise!.reference)) ...[
            GestureDetector(
              onTap: () => ExerciseModal.show(context, interval.exercise!),
              child: Container(
                width: 32,
                height: 32,
                decoration: const BoxDecoration(
                  shape: BoxShape.circle,
                ),
                child: CachedExerciseImage(
                  imageUrl: interval.exercise!.reference,
                  width: 32,
                  height: 32,
                  isCircular: true,
                ),
              ),
            ),
          ] else ...[
            Container(
              width: 32,
              height: 32,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: isRest
                    ? VigorColors.indigo.withValues(alpha: 0.1)
                    : VigorColors.stone.withValues(alpha: 0.1),
              ),
              child: Icon(
                isRest ? Icons.local_drink : Icons.fitness_center,
                size: 32 * 0.4,
                color: isRest ? VigorColors.indigo : VigorColors.stone.withValues(alpha: 0.5),
              ),
            ),
          ],
          const SizedBox(width: VigorSpacing.sm),
          Expanded(
            child: Text(
              name,
              style: VigorTypography.body.copyWith(
                fontWeight: isRest ? FontWeight.normal : FontWeight.w500,
                color: isRest ? VigorColors.textMuted(context) : VigorColors.textPrimary(context),
              ),
            ),
          ),
          // duration in mono font
          Text(
            '${interval.duration}s',
            style: VigorTypography.data.copyWith(
              color: VigorColors.textMuted(context),
            ),
          ),
        ],
      ),
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

  Widget _buildCountersRow(TrainingInterval interval) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs + 2),
      decoration: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.glassDecoration(isDark: isDark)
          : BoxDecoration(
              color: isDark ? VigorColors.darkSurfaceElevated : VigorColors.lightSurface,
              borderRadius: VigorRadius.radiusMd,
            ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.fitness_center, size: 16, color: VigorColors.stone),
          const SizedBox(width: 4),
          Text(
            '${interval.activityNumber}/${interval.totalActivities}',
            style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context)),
          ),
          const SizedBox(width: VigorSpacing.sm),
          const Icon(Icons.view_module, size: 16, color: VigorColors.stone),
          const SizedBox(width: 4),
          Text(
            l10n.blockCounter(interval.blockNumber, interval.totalBlocks),
            style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context)),
          ),
          const SizedBox(width: VigorSpacing.sm),
          const Icon(Icons.list, size: 16, color: VigorColors.stone),
          const SizedBox(width: 4),
          Text(
            l10n.routineCounter(interval.routineNumber, interval.totalRoutines),
            style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context)),
          ),
          const SizedBox(width: 4),
          Text(
            interval.routineName,
            style: VigorTypography.label.copyWith(color: VigorColors.textSecondary(context)),
          ),
        ],
      ),
    );
  }

  Widget _buildRestDisplay() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const Icon(
            Icons.local_drink,
            size: 64,
            color: VigorColors.indigo,
          ),
          const SizedBox(height: VigorSpacing.lg),
          Text(
            '$_remainingSeconds',
            style: VigorTypography.dataDisplay.copyWith(
              color: VigorColors.indigo,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildActivityDisplay(TrainingInterval interval) {
    final activity = interval.activity;
    final exercise = interval.exercise;
    final screenWidth = MediaQuery.of(context).size.width;
    final imageSize = screenWidth * 0.5;
    final hasTimer = activity != null && activity.duration > 0;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // circular exercise image - simple border, no shadows (kanso)
          if (exercise != null && CachedExerciseImage.isValidUrl(exercise.reference)) ...[
            GestureDetector(
              onTap: () => ExerciseModal.show(context, exercise),
              child: Container(
                width: imageSize,
                height: imageSize,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  border: Border.all(
                    color: VigorColors.persimmon,
                    width: 3,
                  ),
                ),
                child: ClipOval(
                  child: Image.network(
                    CachedExerciseImage.proxyUrl(exercise.reference),
                    width: imageSize,
                    height: imageSize,
                    fit: BoxFit.cover,
                    loadingBuilder: (context, child, loadingProgress) {
                      if (loadingProgress == null) return child;
                      return Container(
                        width: imageSize,
                        height: imageSize,
                        color: VigorColors.stone.withValues(alpha: 0.1),
                        child: Icon(
                          Icons.fitness_center,
                          size: imageSize * 0.4,
                          color: VigorColors.stone.withValues(alpha: 0.5),
                        ),
                      );
                    },
                    errorBuilder: (_, _, _) => Container(
                      width: imageSize,
                      height: imageSize,
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      child: Icon(
                        Icons.fitness_center,
                        size: imageSize * 0.4,
                        color: VigorColors.stone.withValues(alpha: 0.5),
                      ),
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(height: VigorSpacing.lg),
          ],
          // reps/kg labels - use data typography
          if (activity != null && (activity.reps > 0 || activity.weightKg > 0)) ...[
            Container(
              padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
              decoration: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.glassDecoration(isDark: isDark)
                  : BoxDecoration(
                      color: isDark ? VigorColors.darkSurfaceElevated : VigorColors.lightSurface,
                      borderRadius: VigorRadius.radiusMd,
                    ),
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  if (activity.reps > 0) ...[
                    const Icon(Icons.fitness_center, size: 20, color: VigorColors.stone),
                    const SizedBox(width: 4),
                    Text(
                      '${activity.reps} reps',
                      style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context)),
                    ),
                  ],
                  if (activity.reps > 0 && activity.weightKg > 0)
                    const SizedBox(width: VigorSpacing.md),
                  if (activity.weightKg > 0) ...[
                    const Icon(Icons.scale, size: 20, color: VigorColors.stone),
                    const SizedBox(width: 4),
                    Text(
                      '${activity.weightKg} kg',
                      style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context)),
                    ),
                  ],
                ],
              ),
            ),
            const SizedBox(height: VigorSpacing.md),
          ],
          // timer display - dataDisplay typography, persimmon for work
          if (hasTimer)
            Text(
              _formatTime(_remainingSeconds),
              style: VigorTypography.dataDisplay.copyWith(
                color: VigorColors.persimmon,
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildControls() {
    final isRest = _currentIntervalIndex < _intervals.length &&
        _intervals[_currentIntervalIndex].type == IntervalType.rest;

    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        // back button - stone when disabled, phase color when enabled
        FloatingActionButton(
          heroTag: 'back',
          onPressed: _history.isNotEmpty ? _skipBackward : null,
          backgroundColor: _history.isNotEmpty ? _phaseColor.withValues(alpha: 0.7) : VigorColors.stone,
          child: const Icon(
            Icons.skip_previous,
            color: Colors.white,
            size: 28,
          ),
        ),
        const SizedBox(width: VigorSpacing.lg),
        // pause/play button - phase-appropriate color
        FloatingActionButton.large(
          heroTag: 'pause',
          onPressed: _togglePause,
          backgroundColor: _phaseColor,
          child: Icon(
            _isPaused ? Icons.play_arrow : Icons.pause,
            color: Colors.white,
            size: 36,
          ),
        ),
        const SizedBox(width: VigorSpacing.lg),
        // forward button
        FloatingActionButton(
          heroTag: 'forward',
          onPressed: _skipForward,
          backgroundColor: _phaseColor.withValues(alpha: 0.7),
          child: const Icon(
            Icons.skip_next,
            color: Colors.white,
            size: 28,
          ),
        ),
      ],
    );
  }

  void _showExitDialog() {
    final l10n = AppLocalizations.of(context);
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (context) => AlertDialog(
        title: Text(l10n.exitTraining),
        content: Text(l10n.whatWouldYouLikeToDo),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).pop();
              Navigator.of(context).pop();
            },
            child: Text(l10n.exit),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(l10n.continueTraining),
          ),
          TextButton(
            onPressed: () async {
              Navigator.of(context).pop();
              final result = await FeedbackModal.show(this.context, widget.training);
              if (result == null) return;
              await _markTrainingComplete(result);
              if (mounted) {
                Navigator.of(this.context).pop(true);
              }
            },
            child: Text(l10n.markAsComplete),
          ),
        ],
      ),
    );
  }
}

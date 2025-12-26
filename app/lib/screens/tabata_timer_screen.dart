import 'dart:async';
import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../config/api_config.dart';
import '../generated/app_localizations.dart';
import '../models/training.dart';
import '../models/activity.dart';
import '../models/exercise.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../utils/exercise_modal.dart';
import '../utils/feedback_modal.dart';
import '../widgets/adaptive/adaptive.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';

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
  final int activityNumber; // Overall activity number (1-based)
  final int totalActivities;
  final int blockNumber; // Block number within routine (1-based)
  final int totalBlocks;
  final int routineNumber; // Routine number (1-based)
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
  List<int> _history = [];
  TrainingService? _trainingService;

  @override
  void initState() {
    super.initState();
    _intervals = _buildIntervals();
    // Start countdown immediately
    _startInitialCountdown();
    // Initialize training service
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final storage = context.read<SecureStorageService>();
      _trainingService = TrainingService(storageService: storage);
    });
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

    // First pass: count total activities
    for (final routine in widget.training.routines) {
      for (final block in routine.blocks) {
        totalActivities += block.activities.length * block.repeats;
      }
    }

    // helper function to add a rest interval, merging with previous if needed
    void addRestInterval(TrainingInterval restInterval) {
      if (intervals.isNotEmpty && intervals.last.type == IntervalType.rest) {
        // Last interval is already a rest, keep only the longer one
        if (restInterval.duration > intervals.last.duration) {
          intervals.removeLast();
          intervals.add(restInterval);
        }
        // Otherwise keep the existing longer rest
      } else {
        // No previous rest, add normally
        intervals.add(restInterval);
      }
    }

    int routineIndex = 0;
    for (final routine in widget.training.routines) {
      routineIndex++;
      int blockIndex = 0;

      for (final block in routine.blocks) {
        blockIndex++;

        // Repeat the block
        for (int repeat = 1; repeat <= block.repeats; repeat++) {
          // Go through each activity in the block
          for (int actIdx = 0; actIdx < block.activities.length; actIdx++) {
            final activity = block.activities[actIdx];
            final isLastActivityInRepeat = actIdx == block.activities.length - 1;
            activityCounter++;
            final exercise = _parseExercise(activity.detail);

            // Add work interval
            final workDuration = activity.duration > 0 ? activity.duration : 60;
            intervals.add(TrainingInterval(
              type: IntervalType.work,
              duration: workDuration,
              routineName: routine.type,
              activityName: activity.name,
              activity: activity,
              exercise: exercise,
              activityNumber: activityCounter,
              totalActivities: totalActivities,
              blockNumber: blockIndex,
              totalBlocks: routine.blocks.length,
              routineNumber: routineIndex,
              totalRoutines: widget.training.routines.length,
            ));

            // After last activity in a repeat, use block rest; otherwise use activity rest
            // skip block rest if this is the very last block of the entire training
            final isLastRoutine = routineIndex == widget.training.routines.length;
            final isLastBlock = blockIndex == routine.blocks.length;
            final isLastRepeat = repeat == block.repeats;
            final isLastTrainingBlock = isLastRoutine && isLastBlock && isLastRepeat;

            if (isLastActivityInRepeat && block.rest > 0 && !isLastTrainingBlock) {
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

      // Add routine rest after all blocks
      if (routine.rest > 0) {
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
    try {
      if (detail.isEmpty) return null;
      return Exercise.fromJson(detail);
    } catch (e) {
      return null;
    }
  }

  bool _isValidImageUrl(String? url) {
    if (url == null || url.isEmpty) return false;
    final uri = Uri.tryParse(url);
    if (uri == null) return false;
    return uri.hasScheme && (uri.scheme == 'http' || uri.scheme == 'https');
  }

  // proxy external image URLs through backend to avoid CORS issues on web
  String _proxyImageUrl(String url) {
    return '${ApiConfig.baseUrl}/proxy/image?url=${Uri.encodeComponent(url)}';
  }

  void _startCurrentInterval() {
    if (_currentIntervalIndex >= _intervals.length) {
      _completeTraining();
      return;
    }

    final interval = _intervals[_currentIntervalIndex];
    setState(() {
      _remainingSeconds = interval.duration;
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
            // initial countdown expired, start training
            _startTraining();
          } else {
            _onIntervalCompleted();
          }
        }
      });
    });
  }

  void _startTraining() {
    if (_hasStarted) return; // already started

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

    // Save current position
    _history.add(_currentIntervalIndex);

    // Move to next interval
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

    // Restore previous position
    setState(() {
      _currentIntervalIndex = _history.removeLast();
      _isCompleted = false;
    });

    _startCurrentInterval();
  }

  void _handleTap(TapUpDetails details) {
    final screenWidth = MediaQuery.of(context).size.width;
    final tapPosition = details.globalPosition.dx;

    if (tapPosition < screenWidth / 2) {
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
    if (result == null) {
      // user cancelled, stay on completed screen
      return;
    }
    await _markTrainingComplete(result);
    if (mounted) {
      Navigator.of(context).pop(true);
    }
  }

  Future<void> _markTrainingComplete(FeedbackResult result) async {
    if (_trainingService == null) return;

    final response = await _trainingService!.completeTraining(
      widget.training.id,
      feedback: result.feedback,
      activityFeedback: result.activityFeedback,
    );
    if (response.isSuccess && mounted) {
      // Training marked as complete successfully
    } else if (mounted) {
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

  @override
  Widget build(BuildContext context) {
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: Text(widget.training.name),
        leading: AdaptiveIconButton(
          icon: const Icon(Icons.close),
          onPressed: () {
            _showExitDialog();
          },
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
    return GestureDetector(
      onTap: _startTraining,
      child: Container(
        color: PlatformHelper.useLiquidGlass
            ? const Color(0xFFF5F7FA)
            : Theme.of(context).scaffoldBackgroundColor,
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.campaign,
                size: 120,
                color: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.primaryColor
                    : Theme.of(context).colorScheme.primary,
              ),
              const SizedBox(height: 24),
              Text(
                l10n.tapToStart,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.headlineStyle
                    : Theme.of(context).textTheme.headlineMedium,
              ),
              const SizedBox(height: 48),
              Text(
                _formatTime(_remainingSeconds),
                style: TextStyle(
                  fontSize: 80,
                  fontWeight: FontWeight.bold,
                  color: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.primaryColor.withOpacity(0.5)
                      : Theme.of(context).colorScheme.primary.withOpacity(0.5),
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
        padding: const EdgeInsets.all(24.0),
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.successColor.withOpacity(0.2)
                    : Theme.of(context).colorScheme.primaryContainer,
              ),
              child: Icon(
                Icons.check_circle,
                size: 80,
                color: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.successColor
                    : Theme.of(context).colorScheme.primary,
              ),
            ),
            const SizedBox(height: 32),
            Text(
              l10n.trainingCompleted,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.titleStyle
                  : Theme.of(context).textTheme.headlineMedium,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 16),
            Text(
              l10n.greatJobCompleting(widget.training.name),
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.bodyStyle
                  : Theme.of(context).textTheme.bodyLarge,
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 48),
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

    return GestureDetector(
      onTapUp: _handleTap,
      child: Container(
        color: PlatformHelper.useLiquidGlass
            ? const Color(0xFFF5F7FA)
            : Theme.of(context).scaffoldBackgroundColor,
        child: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              children: [
                // Activity name
                if (!isRest) _buildActivityName(interval),
                if (!isRest) const SizedBox(height: 16),

                // Counters row (activity/block/routine)
                if (!isRest) _buildCountersRow(interval),
                if (!isRest) const SizedBox(height: 24),

                // Main timer/activity display
                Expanded(
                  child: isRest
                      ? _buildRestDisplay()
                      : _buildActivityDisplay(interval),
                ),

                const SizedBox(height: 16),

                // Controls
                _buildControls(),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildActivityName(TrainingInterval interval) {
    return Text(
      interval.activityName?.toUpperCase() ?? '',
      style: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.titleStyle
          : Theme.of(context).textTheme.headlineMedium,
      textAlign: TextAlign.center,
    );
  }

  Widget _buildCountersRow(TrainingInterval interval) {
    final l10n = AppLocalizations.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.glassDecoration()
          : BoxDecoration(
              color: Theme.of(context).colorScheme.surfaceVariant,
              borderRadius: BorderRadius.circular(12),
            ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          const Icon(Icons.fitness_center, size: 16),
          const SizedBox(width: 4),
          Text(
            '${interval.activityNumber}/${interval.totalActivities}',
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.bodyStyle.copyWith(fontSize: 14)
                : Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(width: 12),
          const Icon(Icons.view_module, size: 16),
          const SizedBox(width: 4),
          Text(
            l10n.blockCounter(interval.blockNumber, interval.totalBlocks),
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.bodyStyle.copyWith(fontSize: 14)
                : Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(width: 12),
          const Icon(Icons.list, size: 16),
          const SizedBox(width: 4),
          Text(
            l10n.routineCounter(interval.routineNumber, interval.totalRoutines),
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.bodyStyle.copyWith(fontSize: 14)
                : Theme.of(context).textTheme.bodyMedium,
          ),
          const SizedBox(width: 4),
          Text(
            interval.routineName,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.bodyStyle.copyWith(fontSize: 14)
                : Theme.of(context).textTheme.bodyMedium,
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
          Icon(
            Icons.local_drink,
            size: 80,
            color: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.primaryColor
                : Theme.of(context).colorScheme.primary,
          ),
          const SizedBox(height: 24),
          Text(
            '$_remainingSeconds',
            style: TextStyle(
              fontSize: 120,
              fontWeight: FontWeight.bold,
              color: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.primaryColor
                  : Theme.of(context).colorScheme.primary,
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

    return SingleChildScrollView(
      child: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // Circular exercise image (if available)
            if (exercise != null && _isValidImageUrl(exercise.reference)) ...[
              GestureDetector(
                onTap: () => ExerciseModal.show(context, exercise),
                child: Container(
                  width: imageSize,
                  height: imageSize,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.primaryColor
                          : Theme.of(context).colorScheme.primary,
                      width: 3,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: Colors.black.withOpacity(0.2),
                        blurRadius: 10,
                        offset: const Offset(0, 4),
                      ),
                    ],
                  ),
                  child: ClipOval(
                    child: Image.network(
                      _proxyImageUrl(exercise.reference),
                      fit: BoxFit.cover,
                      errorBuilder: (context, error, stackTrace) {
                        return Container(
                          color: Colors.grey[300],
                          child: const Icon(
                            Icons.fitness_center,
                            size: 40,
                            color: Colors.grey,
                          ),
                        );
                      },
                    ),
                  ),
                ),
              ),
              const SizedBox(height: 24),
            ],
            // Reps/kg labels between image and timer
            if (activity != null &&
                (activity.reps > 0 || activity.weightKg > 0)) ...[
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
                decoration: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.glassDecoration()
                    : BoxDecoration(
                        color: Theme.of(context).colorScheme.surfaceVariant,
                        borderRadius: BorderRadius.circular(12),
                      ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (activity.reps > 0) ...[
                      const Icon(Icons.fitness_center, size: 20),
                      const SizedBox(width: 4),
                      Text(
                        '${activity.reps} reps',
                        style: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.bodyStyle
                            : Theme.of(context).textTheme.bodyLarge,
                      ),
                    ],
                    if (activity.reps > 0 && activity.weightKg > 0)
                      const SizedBox(width: 16),
                    if (activity.weightKg > 0) ...[
                      const Icon(Icons.scale, size: 20),
                      const SizedBox(width: 4),
                      Text(
                        '${activity.weightKg} kg',
                        style: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.bodyStyle
                            : Theme.of(context).textTheme.bodyLarge,
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: 16),
            ],
            // Large timer display (only if activity has duration)
            if (hasTimer) ...[
              Text(
                _formatTime(_remainingSeconds),
                style: TextStyle(
                  fontSize: 80,
                  fontWeight: FontWeight.bold,
                  color: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.primaryColor
                      : Theme.of(context).colorScheme.primary,
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildControls() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        // Back button
        FloatingActionButton(
          heroTag: 'back',
          onPressed: _history.isNotEmpty ? _skipBackward : null,
          backgroundColor: _history.isNotEmpty
              ? (PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.primaryColor.withOpacity(0.7)
                  : Theme.of(context).colorScheme.primary.withOpacity(0.7))
              : Colors.grey[400],
          child: const Icon(
            Icons.skip_previous,
            color: Colors.white,
            size: 28,
          ),
        ),
        const SizedBox(width: 24),
        // Pause/Play button
        FloatingActionButton.large(
          heroTag: 'pause',
          onPressed: _togglePause,
          backgroundColor: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.primaryColor
              : Theme.of(context).colorScheme.primary,
          child: Icon(
            _isPaused ? Icons.play_arrow : Icons.pause,
            color: Colors.white,
            size: 36,
          ),
        ),
        const SizedBox(width: 24),
        // Forward button
        FloatingActionButton(
          heroTag: 'forward',
          onPressed: _skipForward,
          backgroundColor: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.primaryColor.withOpacity(0.7)
              : Theme.of(context).colorScheme.primary.withOpacity(0.7),
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
      builder: (context) => AlertDialog(
        title: Text(l10n.exitTraining),
        content: Text(l10n.whatWouldYouLikeToDo),
        actions: [
          TextButton(
            onPressed: () {
              Navigator.of(context).pop(); // Close dialog
              Navigator.of(context).pop(); // exit training
            },
            child: Text(l10n.exit),
          ),
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(l10n.continueTraining),
          ),
          TextButton(
            onPressed: () async {
              Navigator.of(context).pop(); // Close exit dialog first
              final result = await FeedbackModal.show(this.context, widget.training);
              if (result == null) return; // user cancelled
              await _markTrainingComplete(result);
              if (mounted) {
                Navigator.of(this.context).pop(true); // Return true to indicate completion
              }
            },
            child: Text(l10n.markAsComplete),
          ),
        ],
      ),
    );
  }
}

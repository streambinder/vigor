import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/training.dart';
import '../services/audio_service.dart';
import '../services/preferences_service.dart';
import '../timer/interval_controller.dart';
import '../timer/timer_controller.dart';
import '../timer/training_interval.dart';
import '../utils/feedback_modal.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/timer/completion_screen.dart';
import '../widgets/timer/countdown_overlay.dart';
import '../widgets/timer/counters_row.dart';
import '../widgets/timer/exercise_display.dart';
import '../widgets/timer/rest_display.dart';
import '../widgets/timer/timer_controls.dart';
import '../widgets/timer/upcoming_list.dart';
import '../services/service_locator.dart';

class TabataTimerScreen extends StatefulWidget {
  final Training training;

  const TabataTimerScreen({super.key, required this.training});

  @override
  State<TabataTimerScreen> createState() => _TabataTimerScreenState();
}

class _TabataTimerScreenState extends State<TabataTimerScreen> {
  late final TimerController _controller;
  String? _previousIntervalKey;
  final AudioService _audioService = AudioService();

  @override
  void initState() {
    super.initState();
    _controller = IntervalController(training: widget.training);
    _controller.addListener(_onControllerUpdate);
    _audioService.initialize();
    _controller.startCountdown();
  }

  void _onControllerUpdate() {
    if (!mounted) return;

    // detect interval transition by comparing stable keys
    final currentInterval = _controller.currentInterval;
    final currentKey = _intervalKey(currentInterval);
    if (_controller.hasStarted &&
        currentKey != null &&
        _previousIntervalKey != null &&
        currentKey != _previousIntervalKey) {
      _playJingleIfEnabled();
    }
    _previousIntervalKey = currentKey;

    setState(() {});
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

  @override
  void dispose() {
    _controller.removeListener(_onControllerUpdate);
    _controller.dispose();
    super.dispose();
  }

  void _handleTap(TapUpDetails details) {
    final screenWidth = MediaQuery.of(context).size.width;
    if (details.globalPosition.dx < screenWidth / 2) {
      _controller.skipBackward();
    } else {
      _controller.skipForward();
    }
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

  Color get _phaseColor {
    if (!_controller.hasStarted) return VigorColors.persimmon;
    final interval = _controller.currentInterval;
    if (interval == null) return VigorColors.gold;
    return interval.type == IntervalType.rest ? VigorColors.indigo : VigorColors.persimmon;
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
      body: _controller.isCompleted
          ? CompletionScreen(
              trainingName: widget.training.name,
              onDone: _showFeedbackAndComplete,
            )
          : !_controller.hasStarted
              ? CountdownOverlay(
                  remainingSeconds: _controller.remainingSeconds,
                  onTap: _controller.startTraining,
                )
              : _buildTimerScreen(),
    );
  }

  Widget _buildTimerScreen() {
    final interval = _controller.currentInterval;
    if (interval == null) return const SizedBox.shrink();

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
                      if (!isRest) CountersRow(interval: interval),
                      if (!isRest) const SizedBox(height: VigorSpacing.lg),
                      isRest
                          ? RestDisplay(remainingSeconds: _controller.remainingSeconds)
                          : ExerciseDisplay(
                              interval: interval,
                              remainingSeconds: _controller.remainingSeconds,
                              phaseColor: _phaseColor,
                            ),
                      const SizedBox(height: VigorSpacing.xl),
                      UpcomingList(intervals: _controller.upcomingIntervals),
                    ],
                  ),
                ),
              ),
              Padding(
                padding: VigorSpacing.paddingLg,
                child: TimerControls(
                  isPaused: _controller.isPaused,
                  canGoBack: _controller.canGoBack,
                  phaseColor: _phaseColor,
                  onPauseToggle: _controller.togglePause,
                  onSkipForward: _controller.skipForward,
                  onSkipBackward: _controller.skipBackward,
                ),
              ),
            ],
          ),
        ),
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
              final result = await FeedbackModal.show(context, widget.training);
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

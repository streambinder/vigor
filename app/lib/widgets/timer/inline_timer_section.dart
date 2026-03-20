import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../../models/activity_ext.dart';
import '../../timer/base_timer_notifier.dart';
import '../../timer/amrap_controller.dart';
import '../../timer/emom_controller.dart';
import '../../timer/for_time_controller.dart';
import '../../timer/timer_controller.dart';
import '../../timer/timer_mode.dart';
import '../../timer/training_interval.dart';
import '../../utils/knowledge_labels.dart';
import '../cached_exercise_image.dart';

/// Inline timer section for embedding in TrainingDetailsScreen.
/// Bordered card with progress bar, background exercise image, and attached control buttons.
class InlineTimerSection extends StatelessWidget {
  final BaseTimerNotifier notifier;
  final VoidCallback? onDone;
  final IconData fallbackIcon;

  const InlineTimerSection({
    super.key,
    required this.notifier,
    this.onDone,
    this.fallbackIcon = Icons.fitness_center,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return ListenableBuilder(
      listenable: notifier,
      builder: (context, _) => Container(
        decoration: BoxDecoration(
          color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
          borderRadius: VigorRadius.radiusLg,
        ),
        clipBehavior: Clip.antiAlias,
        child: ConstrainedBox(
          constraints: const BoxConstraints(minHeight: 280),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // "Timer" section header with overall elapsed timer
              Container(
                width: double.infinity,
                padding: VigorSpacing.paddingMd,
                decoration: BoxDecoration(
                  color: VigorColors.stone.withValues(alpha: 0.05),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Text(
                      AppLocalizations.of(context).timer.toUpperCase(),
                      style: VigorTypography.headline.copyWith(color: VigorColors.stone, fontSize: 14),
                    ),
                    const SizedBox(width: VigorSpacing.sm),
                    Text(
                      notifier.formatDuration(notifier.totalElapsedSeconds),
                      style: VigorTypography.data.copyWith(color: VigorColors.stone, fontSize: 12),
                    ),
                  ],
                ),
              ),
              // progress bar right after section title
              TweenAnimationBuilder<Color?>(
                tween: ColorTween(end: notifier.phaseColor(Theme.of(context).brightness)),
                duration: VigorAnimation.medium,
                builder: (context, color, _) => LinearProgressIndicator(
                  value: notifier.progress,
                  minHeight: 5,
                  backgroundColor: VigorColors.stone.withValues(alpha: 0.15),
                  valueColor: AlwaysStoppedAnimation(color),
                ),
              ),
              _buildContent(context),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildContent(BuildContext context) {
    final controller = notifier.controller;

    if (notifier.workoutCompleted) return _buildCompletionPhase(context);
    if (controller == null) {
      return const SizedBox(height: 160, child: Center(child: CircularProgressIndicator()));
    }
    if (!controller.hasStarted) return _buildCountdownPhase(context, controller);
    return _buildActivePhase(context, controller);
  }

  Widget _buildCountdownPhase(BuildContext context, TimerController controller) {
    // flat surface background, same height as active phase (220 image + ~54 buttons)
    return SizedBox(
      height: 274,
      child: Center(
        child: Text(
          '${controller.remainingSeconds}',
          style: VigorTypography.dataDisplay.copyWith(color: VigorColors.stone),
        ),
      ),
    );
  }

  Widget _buildActivePhase(BuildContext context, TimerController controller) {
    final interval = controller.currentInterval;
    if (interval == null) return const SizedBox.shrink();

    final mode = notifier.currentMode ?? TimerMode.interval;
    final brightness = Theme.of(context).brightness;
    final imageUrl = interval.exercise?.reference;
    final hasImage = imageUrl != null && CachedExerciseImage.isValidUrl(imageUrl);
    final isRest = interval.type == IntervalType.rest;
    final l10n = AppLocalizations.of(context);

    return GestureDetector(
      behavior: HitTestBehavior.opaque,
      onTapUp: (details) {
        notifier.handleTap(details, context.size?.width ?? MediaQuery.of(context).size.width);
      },
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          _buildMainDisplay(context, controller, interval, mode, hasImage, imageUrl, isRest, l10n),
          _buildControlButtons(context, controller, mode, brightness),
        ],
      ),
    );
  }

  /// dot-separated row of activity details (no duration — shown in overlay)
  Widget _buildDataRow(BuildContext context, TimerController controller, TrainingInterval interval, bool isRest) {
    if (isRest) return const SizedBox.shrink();
    final parts = <String>[];
    if (interval.activity != null) {
      final activity = interval.activity!;
      final l10n = AppLocalizations.of(context);
      if (activity.reps > 0) parts.add('${activity.reps} reps');
      if (activity.weightKg > 0) parts.add('${activity.weightKgDisplay} kg');
      if (activity.modifiers.isNotEmpty) {
        parts.addAll(activity.modifiers.map((m) => KnowledgeLabels.modifierLabel(m, l10n)));
      }
    }
    if (parts.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.only(bottom: VigorSpacing.xs),
      child: Text(
        parts.join(' · '),
        style: VigorTypography.data.copyWith(color: Colors.white70, fontSize: 16),
        textAlign: TextAlign.center,
      ),
    );
  }

  /// exercise name or "REST" label anchored right above control buttons
  Widget _buildLabel(TrainingInterval interval, bool isRest, AppLocalizations l10n) {
    final label = isRest
        ? l10n.rest.toUpperCase()
        : (interval.activityName?.toUpperCase() ?? '');
    return Text(
      label,
      style: VigorTypography.headline.copyWith(color: Colors.white),
      textAlign: TextAlign.center,
      maxLines: 2,
      overflow: TextOverflow.ellipsis,
    );
  }

  Widget _buildMainDisplay(
    BuildContext context,
    TimerController controller,
    TrainingInterval interval,
    TimerMode mode,
    bool hasImage,
    String? imageUrl,
    bool isRest,
    AppLocalizations l10n,
  ) {
    // always use 220px fixed height — image or fallback background
    final exerciseFallbackIcon = isRest ? Icons.local_drink : fallbackIcon;
    final background = hasImage && !isRest
        ? Image.network(
            CachedExerciseImage.proxyUrl(imageUrl!),
            fit: BoxFit.cover,
            errorBuilder: (_, _, _) => Container(
              color: VigorColors.stone.withValues(alpha: 0.1),
              child: Center(child: Icon(exerciseFallbackIcon, size: 120, color: VigorColors.stone.withValues(alpha: 0.3))),
            ),
          )
        : Container(
            color: VigorColors.stone.withValues(alpha: 0.1),
            child: Center(child: Icon(exerciseFallbackIcon, size: 120, color: VigorColors.stone.withValues(alpha: 0.3))),
          );

    return Stack(
      children: [
        SizedBox(width: double.infinity, height: 220, child: background),
        // gradient overlay for readability
        Container(
          width: double.infinity,
          height: 220,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topCenter,
              end: Alignment.bottomCenter,
              colors: [
                Colors.black.withValues(alpha: 0.3),
                Colors.black.withValues(alpha: 0.7),
              ],
            ),
          ),
        ),
        // content on top
        SizedBox(
          width: double.infinity,
          height: 220,
          child: Padding(
            padding: VigorSpacing.paddingLg,
            child: Column(
              children: [
                Expanded(child: Center(child: _buildOverlayContent(context, controller, interval, mode))),
                // data row: countdown · reps · kg · modifiers
                _buildDataRow(context, controller, interval, isRest),
                // exercise/rest label right above controls
                _buildLabel(interval, isRest, l10n),
              ],
            ),
          ),
        ),
      ],
    );
  }

  /// overlay content on the background image — mode-specific info only
  /// (exercise name, countdown, reps/kg are shown in the strip below)
  Widget _buildOverlayContent(
    BuildContext context,
    TimerController controller,
    TrainingInterval interval,
    TimerMode mode,
  ) {
    switch (mode) {
      case TimerMode.emom:
        final emom = controller as EmomController;
        if (emom.isResting || emom.isBlockRest) {
          return Text(_formatTime(controller.remainingSeconds), style: VigorTypography.dataDisplay.copyWith(color: Colors.white));
        }
        return Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('MIN ${emom.currentRound}/${emom.totalRounds}', style: VigorTypography.label.copyWith(color: Colors.white70)),
            const SizedBox(height: VigorSpacing.md),
            Text(_formatTime(controller.remainingSeconds), style: VigorTypography.dataDisplay.copyWith(color: Colors.white)),
          ],
        );
      case TimerMode.amrap:
        final amrap = controller as AmrapController;
        return Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('ROUND ${amrap.currentRound} (${amrap.activitiesInRound}/${amrap.activitiesPerRound})', style: VigorTypography.label.copyWith(color: Colors.white70)),
            const SizedBox(height: VigorSpacing.md),
            Text(_formatTime(amrap.globalSeconds), style: VigorTypography.dataDisplay.copyWith(color: Colors.white)),
          ],
        );
      case TimerMode.forTime:
        final forTime = controller as ForTimeController;
        return Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text('ROUND ${forTime.currentRound}/${forTime.totalRounds}', style: VigorTypography.label.copyWith(color: Colors.white70)),
            const SizedBox(height: VigorSpacing.md),
            Text(_formatTime(forTime.globalSeconds), style: VigorTypography.dataDisplay.copyWith(color: Colors.white)),
          ],
        );
      case TimerMode.interval:
        final isRest = interval.type == IntervalType.rest;
        if (isRest) {
          return Text(_formatTime(controller.remainingSeconds), style: VigorTypography.dataDisplay.copyWith(color: Colors.white));
        }
        if (interval.activity != null && interval.activity!.duration > 0) {
          return Text(_formatTime(controller.remainingSeconds), style: VigorTypography.dataDisplay.copyWith(color: Colors.white));
        }
        return const SizedBox.shrink();
    }
  }

  Widget _buildCompletionPhase(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        SizedBox(
          height: 220,
          child: Center(
            child: Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: VigorColors.gold.withValues(alpha: 0.15),
              ),
              child: const Icon(Icons.check_circle, size: 80, color: VigorColors.gold),
            ),
          ),
        ),
        // edge-to-edge button matching control buttons area
        GestureDetector(
          onTap: notifier.isSubmitting ? null : onDone,
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.symmetric(vertical: VigorSpacing.md),
            color: notifier.isSubmitting ? VigorColors.stone.withValues(alpha: 0.2) : VigorColors.gold,
            child: Center(
              child: Text(
                l10n.complete,
                style: VigorTypography.label.copyWith(color: notifier.isSubmitting ? VigorColors.stone : Colors.white, fontWeight: FontWeight.w600),
              ),
            ),
          ),
        ),
      ],
    );
  }

  /// attached button row matching the _buildPrimaryActions style from TrainingDetailsScreen
  Widget _buildControlButtons(BuildContext context, TimerController controller, TimerMode mode, Brightness brightness) {
    final phaseColor = notifier.phaseColor(brightness);
    return Row(
      children: [
        Expanded(child: _controlButton(
          icon: Icons.skip_previous,
          color: controller.canGoBack ? phaseColor.withValues(alpha: 0.7) : VigorColors.stone.withValues(alpha: 0.2),
          onPressed: controller.canGoBack ? () {
            notifier.stopWhistle();
            controller.skipBackward();
          } : null,
        )),
        Expanded(child: _controlButton(
          icon: controller.isPaused ? Icons.play_arrow : Icons.pause,
          color: phaseColor,
          onPressed: controller.togglePause,
        )),
        Expanded(child: _controlButton(
          icon: Icons.skip_next,
          color: phaseColor.withValues(alpha: 0.7),
          onPressed: () {
            notifier.stopWhistle();
            controller.skipForward();
          },
        )),
      ],
    );
  }

  Widget _controlButton({required IconData icon, required Color color, VoidCallback? onPressed}) {
    final isDisabled = onPressed == null;
    final targetColor = isDisabled ? VigorColors.stone.withValues(alpha: 0.2) : color;
    return GestureDetector(
      onTap: onPressed,
      child: AnimatedContainer(
        duration: VigorAnimation.medium,
        curve: VigorAnimation.defaultCurve,
        padding: const EdgeInsets.symmetric(vertical: VigorSpacing.md),
        color: targetColor,
        child: Center(
          child: Icon(icon, size: 22, color: isDisabled ? VigorColors.stone : Colors.white),
        ),
      ),
    );
  }

  String _formatTime(int seconds) {
    final minutes = seconds ~/ 60;
    final secs = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }
}

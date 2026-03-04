import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../../models/activity_ext.dart';
import '../../models/exercise.dart';
import '../../timer/timer_mode.dart';
import '../../timer/training_interval.dart';
import '../../utils/exercise_modal.dart';
import '../cached_exercise_image.dart';

/// Displays remaining exercises in a methodology-aware format
/// - EMOM: groups by minute with minute headers
/// - AMRAP: continuous loop indicator with round markers
/// - ForTime: shows progress toward completion with round countdown
/// - Interval: simple sequential list (uses UpcomingList instead)
class RemainingExercisesList extends StatelessWidget {
  final List<TrainingInterval> intervals;
  final TimerMode mode;
  final int currentRound;
  final int? totalRounds;

  const RemainingExercisesList({
    super.key,
    required this.intervals,
    required this.mode,
    this.currentRound = 1,
    this.totalRounds,
  });

  @override
  Widget build(BuildContext context) {
    if (intervals.isEmpty) return const SizedBox.shrink();

    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          _getHeaderText(l10n),
          style: VigorTypography.label.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        const SizedBox(height: VigorSpacing.sm),
        ..._buildItems(context),
      ],
    );
  }

  String _getHeaderText(AppLocalizations l10n) {
    switch (mode) {
      case TimerMode.emom:
        return l10n.upcoming;
      case TimerMode.amrap:
        return l10n.upcoming;
      case TimerMode.forTime:
        final remaining = totalRounds != null ? totalRounds! - currentRound + 1 : 0;
        return remaining > 1 ? '$remaining rounds left' : l10n.upcoming;
      case TimerMode.interval:
        return l10n.upcoming;
    }
  }

  List<Widget> _buildItems(BuildContext context) {
    switch (mode) {
      case TimerMode.emom:
        return _buildEmomItems(context);
      case TimerMode.amrap:
        return _buildAmrapItems(context);
      case TimerMode.forTime:
        return _buildForTimeItems(context);
      case TimerMode.interval:
        return _buildSimpleItems(context);
    }
  }

  // emom: group by blockNumber (minute)
  List<Widget> _buildEmomItems(BuildContext context) {
    final items = <Widget>[];
    int? lastMinute;

    for (final interval in intervals) {
      final minute = interval.blockNumber;
      if (minute != lastMinute) {
        items.add(_MinuteHeader(minute: minute, totalMinutes: interval.totalBlocks));
        lastMinute = minute;
      }
      items.add(_ExerciseItem(interval: interval, showCounter: false));
    }
    return items;
  }

  // amrap: show activities with round markers when crossing round boundary
  List<Widget> _buildAmrapItems(BuildContext context) {
    final items = <Widget>[];
    int currentIndex = 0;
    int projectedRound = currentRound;

    for (final interval in intervals) {
      // detect round boundary based on activity number resetting
      if (currentIndex > 0 && interval.activityNumber == 1) {
        projectedRound++;
        items.add(_RoundMarker(round: projectedRound));
      }
      items.add(_ExerciseItem(interval: interval, showCounter: true));
      currentIndex++;
    }
    return items;
  }

  // forTime: show remaining with round countdown
  List<Widget> _buildForTimeItems(BuildContext context) {
    final items = <Widget>[];
    int? lastRound;

    for (final interval in intervals) {
      final round = interval.blockNumber;
      if (round != lastRound && totalRounds != null) {
        items.add(_RoundHeader(round: round, totalRounds: totalRounds!));
        lastRound = round;
      }
      items.add(_ExerciseItem(interval: interval, showCounter: false));
    }
    return items;
  }

  List<Widget> _buildSimpleItems(BuildContext context) {
    return intervals.map((i) => _ExerciseItem(interval: i, showCounter: false)).toList();
  }
}

class _MinuteHeader extends StatelessWidget {
  final int minute;
  final int totalMinutes;

  const _MinuteHeader({required this.minute, required this.totalMinutes});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: VigorSpacing.sm, bottom: VigorSpacing.xs),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: 2),
        decoration: BoxDecoration(
          color: VigorColors.persimmon.withValues(alpha: 0.1),
          borderRadius: VigorRadius.radiusSm,
        ),
        child: Text(
          'Minute $minute / $totalMinutes',
          style: VigorTypography.caption.copyWith(
            color: VigorColors.persimmon,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}

class _RoundMarker extends StatelessWidget {
  final int round;

  const _RoundMarker({required this.round});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: VigorSpacing.sm, bottom: VigorSpacing.xs),
      child: Row(
        children: [
          Expanded(child: Divider(color: VigorColors.gold.withValues(alpha: 0.3))),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm),
            child: Text(
              'Round $round',
              style: VigorTypography.caption.copyWith(
                color: VigorColors.gold,
                fontWeight: FontWeight.w600,
              ),
            ),
          ),
          Expanded(child: Divider(color: VigorColors.gold.withValues(alpha: 0.3))),
        ],
      ),
    );
  }
}

class _RoundHeader extends StatelessWidget {
  final int round;
  final int totalRounds;

  const _RoundHeader({required this.round, required this.totalRounds});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.only(top: VigorSpacing.sm, bottom: VigorSpacing.xs),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: 2),
        decoration: BoxDecoration(
          color: VigorColors.persimmon.withValues(alpha: 0.1),
          borderRadius: VigorRadius.radiusSm,
        ),
        child: Text(
          'Round $round / $totalRounds',
          style: VigorTypography.caption.copyWith(
            color: VigorColors.persimmon,
            fontWeight: FontWeight.w600,
          ),
        ),
      ),
    );
  }
}

class _ExerciseItem extends StatelessWidget {
  final TrainingInterval interval;
  final bool showCounter;

  const _ExerciseItem({required this.interval, this.showCounter = false});

  @override
  Widget build(BuildContext context) {
    final isRest = interval.type == IntervalType.rest;
    final name = isRest ? 'Rest' : (interval.activityName ?? '');
    final activity = interval.activity;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs + 2),
      child: Row(
        children: [
          _buildThumbnail(context, interval.exercise, isRest),
          const SizedBox(width: VigorSpacing.sm),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  name,
                  style: VigorTypography.body.copyWith(
                    fontWeight: isRest ? FontWeight.normal : FontWeight.w500,
                    color: isRest ? VigorColors.textMuted(context) : VigorColors.textPrimary(context),
                  ),
                ),
                if (activity != null && (activity.reps > 0 || activity.modifiers.isNotEmpty))
                  Text(
                    [
                      if (activity.reps > 0) '${activity.reps} reps${activity.weightKg > 0 ? ' • ${activity.weightKgDisplay} kg' : ''}',
                      if (activity.modifiers.isNotEmpty) activity.modifiers.join(' · '),
                    ].join(' • '),
                    style: VigorTypography.caption.copyWith(color: VigorColors.textMuted(context)),
                  ),
              ],
            ),
          ),
          if (showCounter)
            Container(
              padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.xs, vertical: 2),
              decoration: BoxDecoration(
                color: VigorColors.stone.withValues(alpha: 0.1),
                borderRadius: VigorRadius.radiusSm,
              ),
              child: Text(
                '${interval.activityNumber}/${interval.totalActivities}',
                style: VigorTypography.caption.copyWith(color: VigorColors.textMuted(context)),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildThumbnail(BuildContext context, Exercise? exercise, bool isRest) {
    if (!isRest && exercise != null && CachedExerciseImage.isValidUrl(exercise.reference)) {
      return GestureDetector(
        onTap: () => ExerciseModal.show(context, exercise),
        child: Container(
          width: 32,
          height: 32,
          decoration: const BoxDecoration(shape: BoxShape.circle),
          child: CachedExerciseImage(
            imageUrl: exercise.reference,
            width: 32,
            height: 32,
            isCircular: true,
          ),
        ),
      );
    }
    return Container(
      width: 32,
      height: 32,
      decoration: BoxDecoration(
        shape: BoxShape.circle,
        color: isRest ? VigorColors.indigo.withValues(alpha: 0.1) : VigorColors.stone.withValues(alpha: 0.1),
      ),
      child: Icon(
        isRest ? Icons.local_drink : Icons.fitness_center,
        size: 32 * 0.4,
        color: isRest ? VigorColors.indigo : VigorColors.stone.withValues(alpha: 0.5),
      ),
    );
  }
}

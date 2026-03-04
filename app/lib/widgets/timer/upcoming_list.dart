import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../../models/activity.dart';
import '../../models/activity_ext.dart';
import '../../models/exercise.dart';
import '../../timer/training_interval.dart';
import '../../utils/exercise_modal.dart';
import '../cached_exercise_image.dart';

class UpcomingList extends StatelessWidget {
  final List<TrainingInterval> intervals;
  final int maxItems;

  const UpcomingList({
    super.key,
    required this.intervals,
    this.maxItems = 10,
  });

  @override
  Widget build(BuildContext context) {
    if (intervals.isEmpty) return const SizedBox.shrink();

    final l10n = AppLocalizations.of(context);
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
        ...intervals.take(maxItems).map((interval) => _UpcomingItem(interval: interval)),
        if (intervals.length > maxItems)
          Padding(
            padding: const EdgeInsets.only(top: VigorSpacing.sm),
            child: Text(
              '+${intervals.length - maxItems} more',
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textMuted(context),
              ),
            ),
          ),
      ],
    );
  }
}

class _UpcomingItem extends StatelessWidget {
  final TrainingInterval interval;

  const _UpcomingItem({required this.interval});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isRest = interval.type == IntervalType.rest;
    final name = isRest ? l10n.rest : (interval.activityName ?? '');

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
                if (!isRest && interval.activity != null && _hasDetails(interval.activity!)) ...[
                  const SizedBox(height: VigorSpacing.xs),
                  _buildActivityTags(interval.activity!),
                ],
              ],
            ),
          ),
        ],
      ),
    );
  }

  bool _hasDetails(Activity activity) =>
      activity.reps > 0 || activity.weightKg > 0 || activity.duration > 0 || activity.modifiers.isNotEmpty;

  Widget _buildActivityTags(Activity activity) {
    return Wrap(
      spacing: VigorSpacing.xs,
      runSpacing: VigorSpacing.xs,
      children: [
        if (activity.reps > 0) _buildTag(Icons.repeat, '${activity.reps}'),
        if (activity.weightKg > 0) _buildTag(Icons.fitness_center, '${activity.weightKgDisplay}kg'),
        if (activity.duration > 0) _buildTag(Icons.timer, '${activity.duration}s'),
        if (activity.modifiers.isNotEmpty) _buildTag(Icons.tune, activity.modifiers.join(' · ')),
      ],
    );
  }

  Widget _buildTag(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusSm,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 11, color: VigorColors.stone),
          const SizedBox(width: VigorSpacing.xs),
          Text(label, style: VigorTypography.data.copyWith(fontSize: 11, fontWeight: FontWeight.w600, color: VigorColors.stone)),
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
        color: isRest
            ? VigorColors.indigo.withValues(alpha: 0.1)
            : VigorColors.stone.withValues(alpha: 0.1),
      ),
      child: Icon(
        isRest ? Icons.local_drink : Icons.fitness_center,
        size: 32 * 0.4,
        color: isRest ? VigorColors.indigo : VigorColors.stone.withValues(alpha: 0.5),
      ),
    );
  }
}

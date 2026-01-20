import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../../models/activity.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../timer/training_interval.dart';
import '../../utils/exercise_modal.dart';
import '../../utils/platform_helper.dart';
import '../cached_exercise_image.dart';

/// Display widget for EMOM timer mode
/// Shows current minute, activity progress within minute, and time remaining
class EmomDisplay extends StatelessWidget {
  final TrainingInterval interval;
  final int secondsRemaining;
  final int currentMinute;
  final int totalMinutes;
  final int activityIndex;
  final int totalActivities;
  final bool isResting;

  const EmomDisplay({
    super.key,
    required this.interval,
    required this.secondsRemaining,
    required this.currentMinute,
    required this.totalMinutes,
    required this.activityIndex,
    required this.totalActivities,
    required this.isResting,
  });

  String _formatTime(int seconds) {
    final minutes = seconds ~/ 60;
    final secs = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final activity = interval.activity;
    final exercise = interval.exercise;
    final screenWidth = MediaQuery.of(context).size.width;
    final imageSize = screenWidth * 0.4;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.lg, vertical: VigorSpacing.sm),
              decoration: BoxDecoration(
                color: VigorColors.persimmon.withValues(alpha: 0.15),
                borderRadius: VigorRadius.radiusMd,
              ),
              child: Text(
                'MINUTE $currentMinute / $totalMinutes',
                style: VigorTypography.label.copyWith(color: VigorColors.persimmon, fontWeight: FontWeight.bold),
              ),
            ),
            if (!isResting && totalActivities > 1) ...[
              const SizedBox(width: VigorSpacing.sm),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
                decoration: BoxDecoration(
                  color: VigorColors.gold.withValues(alpha: 0.15),
                  borderRadius: VigorRadius.radiusMd,
                ),
                child: Text(
                  '${activityIndex + 1}/$totalActivities',
                  style: VigorTypography.label.copyWith(color: VigorColors.gold, fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ],
        ),
        const SizedBox(height: VigorSpacing.lg),

        if (isResting) ...[
          const Icon(Icons.local_drink, size: 80, color: VigorColors.indigo),
          const SizedBox(height: VigorSpacing.md),
          Text(l10n.rest.toUpperCase(), style: VigorTypography.title.copyWith(color: VigorColors.indigo)),
        ] else ...[
          if (exercise != null && CachedExerciseImage.isValidUrl(exercise.reference)) ...[
            GestureDetector(
              onTap: () => ExerciseModal.show(context, exercise),
              child: Container(
                width: imageSize,
                height: imageSize,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  border: Border.all(color: VigorColors.persimmon, width: 3),
                ),
                child: ClipOval(
                  child: Image.network(
                    CachedExerciseImage.proxyUrl(exercise.reference),
                    width: imageSize,
                    height: imageSize,
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) => _buildPlaceholder(imageSize),
                  ),
                ),
              ),
            ),
            const SizedBox(height: VigorSpacing.md),
          ],
          Text(
            interval.activityName?.toUpperCase() ?? '',
            style: VigorTypography.title.copyWith(color: VigorColors.textPrimary(context)),
            textAlign: TextAlign.center,
          ),
          const SizedBox(height: VigorSpacing.sm),
          if (activity != null && activity.reps > 0) _buildRepsChip(context, activity, isDark),
        ],

        const SizedBox(height: VigorSpacing.lg),
        Text(
          _formatTime(secondsRemaining),
          style: VigorTypography.dataDisplay.copyWith(color: isResting ? VigorColors.indigo : VigorColors.persimmon),
        ),
      ],
    );
  }

  Widget _buildPlaceholder(double size) {
    return Container(
      width: size,
      height: size,
      color: VigorColors.stone.withValues(alpha: 0.1),
      child: Icon(Icons.fitness_center, size: size * 0.4, color: VigorColors.stone.withValues(alpha: 0.5)),
    );
  }

  Widget _buildRepsChip(BuildContext context, Activity activity, bool isDark) {
    return Container(
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
          const Icon(Icons.fitness_center, size: 20, color: VigorColors.stone),
          const SizedBox(width: 4),
          Text('${activity.reps} reps', style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context))),
          if (activity.weightKg > 0) ...[
            const SizedBox(width: VigorSpacing.md),
            const Icon(Icons.scale, size: 20, color: VigorColors.stone),
            const SizedBox(width: 4),
            Text('${activity.weightKg} kg', style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context))),
          ],
        ],
      ),
    );
  }
}

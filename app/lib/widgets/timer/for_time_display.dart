import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../models/activity.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../timer/training_interval.dart';
import '../../utils/exercise_modal.dart';
import '../../utils/platform_helper.dart';
import '../cached_exercise_image.dart';

/// Display widget for ForTime timer mode
/// Shows elapsed time (counting up), round progress, and current activity
class ForTimeDisplay extends StatelessWidget {
  final TrainingInterval interval;
  final int elapsedSeconds;
  final int currentRound;
  final int totalRounds;

  const ForTimeDisplay({
    super.key,
    required this.interval,
    required this.elapsedSeconds,
    required this.currentRound,
    required this.totalRounds,
  });

  String _formatTime(int seconds) {
    final minutes = seconds ~/ 60;
    final secs = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final activity = interval.activity;
    final exercise = interval.exercise;
    final screenWidth = MediaQuery.of(context).size.width;
    final imageSize = screenWidth * 0.4;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(_formatTime(elapsedSeconds), style: VigorTypography.dataDisplay.copyWith(color: VigorColors.gold)),
        const SizedBox(height: VigorSpacing.md),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.lg, vertical: VigorSpacing.sm),
          decoration: BoxDecoration(
            color: VigorColors.persimmon.withValues(alpha: 0.15),
            borderRadius: VigorRadius.radiusMd,
          ),
          child: Text(
            'ROUND $currentRound / $totalRounds',
            style: VigorTypography.title.copyWith(color: VigorColors.persimmon, fontWeight: FontWeight.bold),
          ),
        ),
        const SizedBox(height: VigorSpacing.lg),
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

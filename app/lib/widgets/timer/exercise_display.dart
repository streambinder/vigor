import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../models/activity.dart';
import '../../models/exercise.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../timer/training_interval.dart';
import '../../utils/exercise_modal.dart';
import '../../utils/platform_helper.dart';
import '../cached_exercise_image.dart';

class ExerciseDisplay extends StatelessWidget {
  final TrainingInterval interval;
  final int remainingSeconds;
  final Color phaseColor;

  const ExerciseDisplay({
    super.key,
    required this.interval,
    required this.remainingSeconds,
    required this.phaseColor,
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
    final imageSize = screenWidth * 0.5;
    final hasTimer = activity != null && activity.duration > 0;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          if (exercise != null && CachedExerciseImage.isValidUrl(exercise.reference)) ...[
            _buildExerciseImage(context, exercise, imageSize),
            const SizedBox(height: VigorSpacing.lg),
          ],
          if (activity != null && (activity.reps > 0 || activity.weightKg > 0)) ...[
            _buildRepsWeightChip(context, activity, isDark),
            const SizedBox(height: VigorSpacing.md),
          ],
          if (hasTimer)
            Text(
              _formatTime(remainingSeconds),
              style: VigorTypography.dataDisplay.copyWith(color: phaseColor),
            ),
        ],
      ),
    );
  }

  Widget _buildExerciseImage(BuildContext context, Exercise exercise, double size) {
    return GestureDetector(
      onTap: () => ExerciseModal.show(context, exercise),
      child: Container(
        width: size,
        height: size,
        decoration: BoxDecoration(
          shape: BoxShape.circle,
          border: Border.all(color: phaseColor, width: 3),
        ),
        child: ClipOval(
          child: Image.network(
            CachedExerciseImage.proxyUrl(exercise.reference),
            width: size,
            height: size,
            fit: BoxFit.cover,
            loadingBuilder: (context, child, loadingProgress) {
              if (loadingProgress == null) return child;
              return _buildPlaceholder(size);
            },
            errorBuilder: (_, __, ___) => _buildPlaceholder(size),
          ),
        ),
      ),
    );
  }

  Widget _buildPlaceholder(double size) {
    return Container(
      width: size,
      height: size,
      color: VigorColors.stone.withValues(alpha: 0.1),
      child: Icon(
        Icons.fitness_center,
        size: size * 0.4,
        color: VigorColors.stone.withValues(alpha: 0.5),
      ),
    );
  }

  Widget _buildRepsWeightChip(BuildContext context, Activity activity, bool isDark) {
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: VigorSpacing.md,
        vertical: VigorSpacing.sm,
      ),
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
          if (activity.reps > 0 && activity.weightKg > 0) const SizedBox(width: VigorSpacing.md),
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
    );
  }
}

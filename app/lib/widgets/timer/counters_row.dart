import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../utils/platform_helper.dart';
import '../../timer/training_interval.dart';

class CountersRow extends StatelessWidget {
  final TrainingInterval interval;

  const CountersRow({super.key, required this.interval});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Container(
      padding: const EdgeInsets.symmetric(
        horizontal: VigorSpacing.sm,
        vertical: VigorSpacing.xs + 2,
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
}

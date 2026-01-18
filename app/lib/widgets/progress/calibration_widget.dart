import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../../models/family_progress.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../utils/platform_helper.dart';

/// Shows overall calibration as an expandable bar.
/// When collapsed: single bar with averaged calibration across all families.
/// When expanded: per-family calibration bars.
class CalibrationWidget extends StatefulWidget {
  final Map<String, FamilyProgress> families;

  const CalibrationWidget({super.key, required this.families});

  @override
  State<CalibrationWidget> createState() => _CalibrationWidgetState();
}

class _CalibrationWidgetState extends State<CalibrationWidget> {
  bool _expanded = false;

  // human-readable family names
  static const _familyLabels = {
    'horizontal_push': 'Push',
    'horizontal_pull': 'Pull',
    'vertical_push': 'Overhead',
    'vertical_pull': 'Pull-up',
    'squat': 'Squat',
    'hinge': 'Hinge',
    'core': 'Core',
    'carry': 'Carry',
    'balance': 'Balance',
    'cardio': 'Cardio',
    'mobility': 'Mobility',
  };

  double get _overallCalibration {
    if (widget.families.isEmpty) return 0;
    final sum = widget.families.values.fold(0.0, (acc, fp) => acc + fp.calibration);
    return (sum / widget.families.length).clamp(0.0, 100.0);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final calibration = _overallCalibration;

    return GestureDetector(
      onTap: () => setState(() => _expanded = !_expanded),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.all(16),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration()
            : BoxDecoration(
                color: Theme.of(context).cardColor,
                borderRadius: BorderRadius.circular(12),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black.withValues(alpha: 0.05),
                    blurRadius: 10,
                  ),
                ],
              ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // header row
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Row(
                  children: [
                    Icon(
                      Icons.tune,
                      size: 18,
                      color: _calibrationColor(calibration),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      l10n.calibration,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.bodyStyle
                          : Theme.of(context).textTheme.bodyMedium,
                    ),
                  ],
                ),
                Row(
                  children: [
                    Text(
                      '${calibration.toInt()}%',
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.bodyStyle.copyWith(fontWeight: FontWeight.w600)
                          : Theme.of(context).textTheme.bodyMedium?.copyWith(
                                fontWeight: FontWeight.w600,
                              ),
                    ),
                    const SizedBox(width: 4),
                    Icon(
                      _expanded ? Icons.expand_less : Icons.expand_more,
                      size: 20,
                      color: VigorColors.stone,
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 4),

            // description text - always visible
            Text(
              l10n.calibrationDescription,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.captionStyle.copyWith(
                      color: LiquidGlassTheme.captionStyle.color?.withValues(alpha: 0.7),
                    )
                  : Theme.of(context).textTheme.bodySmall?.copyWith(
                        color: VigorColors.stone,
                      ),
            ),
            const SizedBox(height: 12),

            // main calibration bar
            _buildProgressBar(calibration),

            // expanded per-family view
            if (_expanded) ...[
              const SizedBox(height: 16),
              const Divider(),
              const SizedBox(height: 12),
              ..._buildFamilyBars(),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildProgressBar(double value) {
    final trackColor = PlatformHelper.useLiquidGlass
        ? Colors.black.withValues(alpha: 0.5)
        : Theme.of(context).colorScheme.surfaceContainerHighest;
    final fillColor = _calibrationColor(value);

    return Stack(
      children: [
        Container(
          height: 8,
          decoration: BoxDecoration(
            color: trackColor,
            borderRadius: BorderRadius.circular(4),
          ),
        ),
        FractionallySizedBox(
          widthFactor: value / 100,
          child: Container(
            height: 8,
            decoration: BoxDecoration(
              color: fillColor,
              borderRadius: BorderRadius.circular(4),
            ),
          ),
        ),
      ],
    );
  }

  List<Widget> _buildFamilyBars() {
    final sorted = widget.families.entries.toList()
      ..sort((a, b) => b.value.calibration.compareTo(a.value.calibration));

    return sorted.map((entry) {
      final label = _familyLabels[entry.key] ?? _formatFamilyName(entry.key);
      final cal = entry.value.calibration.clamp(0.0, 100.0);

      return Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Row(
          children: [
            SizedBox(
              width: 80,
              child: Text(
                label,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.captionStyle
                    : Theme.of(context).textTheme.bodySmall,
              ),
            ),
            Expanded(
              child: _buildProgressBar(cal),
            ),
            const SizedBox(width: 8),
            SizedBox(
              width: 36,
              child: Text(
                '${cal.toInt()}%',
                textAlign: TextAlign.right,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.captionStyle
                    : Theme.of(context).textTheme.bodySmall,
              ),
            ),
          ],
        ),
      );
    }).toList();
  }

  String _formatFamilyName(String family) {
    return family
        .split('_')
        .map((w) => w.isNotEmpty ? '${w[0].toUpperCase()}${w.substring(1)}' : w)
        .join(' ');
  }

  Color _calibrationColor(double calibration) {
    // neutral accent color for calibration - not a warning indicator
    return PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.primaryColor
        : Theme.of(context).colorScheme.primary;
  }
}

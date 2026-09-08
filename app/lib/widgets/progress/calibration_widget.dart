import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../../models/muscle_progress.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../utils/knowledge_labels.dart';
import '../../utils/platform_helper.dart';

/// Shows calibration progress as calibrated muscles out of total (x/7).
/// When collapsed: single bar with the calibrated fraction.
/// When expanded: per-muscle calibration bars.
class CalibrationWidget extends StatefulWidget {
  final Map<String, MuscleProgress> muscles;

  const CalibrationWidget({super.key, required this.muscles});

  @override
  State<CalibrationWidget> createState() => _CalibrationWidgetState();
}

class _CalibrationWidgetState extends State<CalibrationWidget> {
  bool _expanded = false;

  int get _calibratedCount =>
      widget.muscles.values.where((mp) => mp.calibration >= 100.0).length;

  double get _overallCalibration {
    if (widget.muscles.isEmpty) return 0;
    return (_calibratedCount / widget.muscles.length).clamp(0.0, 1.0);
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
                      '$_calibratedCount/${widget.muscles.length}',
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

            // expanded per-muscle view
            if (_expanded) ...[
              const SizedBox(height: 16),
              const Divider(),
              const SizedBox(height: 12),
              ..._buildMuscleBars(),
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
          widthFactor: value,
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

  List<Widget> _buildMuscleBars() {
    final sorted = KnowledgeLabels.muscleDisplayOrder
        .where((m) => widget.muscles.containsKey(m))
        .map((m) => MapEntry(m, widget.muscles[m]!))
        .toList();
    for (final entry in widget.muscles.entries) {
      if (!KnowledgeLabels.muscleDisplayOrder.contains(entry.key)) {
        sorted.add(entry);
      }
    }

    final l10n = AppLocalizations.of(context);
    return sorted.map((entry) {
      final label = KnowledgeLabels.muscleLabel(entry.key, l10n);
      final cal = (entry.value.calibration / 100.0).clamp(0.0, 1.0);
      final calibrated = entry.value.calibration >= 100.0;

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
              child: calibrated
                  ? const Icon(Icons.check_circle, size: 18, color: Colors.green)
                  : Text(
                      '${entry.value.calibration.toInt()}%',
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

  Color _calibrationColor(double calibration) {
    // neutral accent color for calibration - not a warning indicator
    return PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.primaryColor
        : Theme.of(context).colorScheme.primary;
  }
}

import 'package:flutter/material.dart';
import '../../generated/app_localizations.dart';
import '../../models/muscle_progress.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../utils/knowledge_labels.dart';
import '../../utils/platform_helper.dart';

/// Displays proficiency progress bars for each muscle group.
class MuscleProgressWidget extends StatelessWidget {
  final Map<String, MuscleProgress> muscles;

  const MuscleProgressWidget({super.key, required this.muscles});

  @override
  Widget build(BuildContext context) {
    final sortedMuscles = KnowledgeLabels.muscleDisplayOrder
        .where((m) => muscles.containsKey(m))
        .map((m) => MapEntry(m, muscles[m]!))
        .toList();

    // add any muscles not in the predefined order
    for (final entry in muscles.entries) {
      if (!KnowledgeLabels.muscleDisplayOrder.contains(entry.key)) {
        sortedMuscles.add(entry);
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: sortedMuscles.map((entry) {
        return _buildMuscleRow(context, entry.key, entry.value);
      }).toList(),
    );
  }

  Widget _buildMuscleRow(BuildContext context, String muscle, MuscleProgress progress) {
    final label = KnowledgeLabels.muscleLabel(muscle, AppLocalizations.of(context));
    final proficiency = progress.proficiency.clamp(0.0, 100.0);

    final primaryColor = PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.primaryColor
        : Theme.of(context).colorScheme.primary;

    final trackColor = PlatformHelper.useLiquidGlass
        ? Colors.black.withValues(alpha: 0.5)
        : Theme.of(context).colorScheme.surfaceContainerHighest;

    final textStyle = PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.bodyStyle.copyWith(fontSize: 13)
        : Theme.of(context).textTheme.bodySmall;

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          SizedBox(
            width: 80,
            child: Text(label, style: textStyle),
          ),
          Expanded(
            child: Stack(
              children: [
                Container(
                  height: 8,
                  decoration: BoxDecoration(
                    color: trackColor,
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                FractionallySizedBox(
                  widthFactor: proficiency / 100,
                  child: Container(
                    height: 8,
                    decoration: BoxDecoration(
                      color: primaryColor,
                      borderRadius: BorderRadius.circular(4),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 36,
            child: Text(
              proficiency > 0 ? '${proficiency.toInt()}%' : '–',
              textAlign: TextAlign.right,
              style: textStyle?.copyWith(fontWeight: FontWeight.w600),
            ),
          ),
        ],
      ),
    );
  }

}

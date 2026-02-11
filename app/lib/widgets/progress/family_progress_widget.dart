import 'package:flutter/material.dart';
import '../../generated/app_localizations.dart';
import '../../models/family_progress.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../utils/knowledge_labels.dart';
import '../../utils/platform_helper.dart';

/// Displays proficiency progress bars for each movement family.
class FamilyProgressWidget extends StatelessWidget {
  final Map<String, FamilyProgress> families;

  const FamilyProgressWidget({super.key, required this.families});

  @override
  Widget build(BuildContext context) {
    final sortedFamilies = KnowledgeLabels.familyDisplayOrder
        .where((f) => families.containsKey(f))
        .map((f) => MapEntry(f, families[f]!))
        .toList();

    // add any families not in the predefined order
    for (final entry in families.entries) {
      if (!KnowledgeLabels.familyDisplayOrder.contains(entry.key)) {
        sortedFamilies.add(entry);
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: sortedFamilies.map((entry) {
        return _buildFamilyRow(context, entry.key, entry.value);
      }).toList(),
    );
  }

  Widget _buildFamilyRow(BuildContext context, String family, FamilyProgress progress) {
    final label = KnowledgeLabels.familyLabel(family, AppLocalizations.of(context));
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

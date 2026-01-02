import 'package:flutter/material.dart';
import '../../models/family_progress.dart';
import '../../theme/liquid_glass_theme.dart';
import '../../utils/platform_helper.dart';

/// Displays proficiency progress bars for each movement family.
class FamilyProgressWidget extends StatelessWidget {
  final Map<String, FamilyProgress> families;

  const FamilyProgressWidget({super.key, required this.families});

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

  // display order
  static const _familyOrder = [
    'horizontal_push',
    'horizontal_pull',
    'vertical_push',
    'vertical_pull',
    'squat',
    'hinge',
    'core',
    'cardio',
    'mobility',
    'balance',
    'carry',
  ];

  @override
  Widget build(BuildContext context) {
    final sortedFamilies = _familyOrder
        .where((f) => families.containsKey(f))
        .map((f) => MapEntry(f, families[f]!))
        .toList();

    // add any families not in the predefined order
    for (final entry in families.entries) {
      if (!_familyOrder.contains(entry.key)) {
        sortedFamilies.add(entry);
      }
    }

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: sortedFamilies.map((entry) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 12),
          child: _buildFamilyRow(context, entry.key, entry.value),
        );
      }).toList(),
    );
  }

  Widget _buildFamilyRow(BuildContext context, String family, FamilyProgress progress) {
    final label = _familyLabels[family] ?? _formatFamilyName(family);
    final proficiency = progress.proficiency.clamp(0.0, 100.0);

    final primaryColor = PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.primaryColor
        : Theme.of(context).colorScheme.primary;

    final trackColor = PlatformHelper.useLiquidGlass
        ? Colors.black.withOpacity(0.5)
        : Theme.of(context).colorScheme.surfaceContainerHighest;

    final textStyle = PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.bodyStyle.copyWith(fontSize: 13)
        : Theme.of(context).textTheme.bodySmall;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(label, style: textStyle),
            Text(
              '${proficiency.toInt()}%',
              style: textStyle?.copyWith(fontWeight: FontWeight.w600),
            ),
          ],
        ),
        const SizedBox(height: 4),
        Stack(
          children: [
            Container(
              height: 6,
              decoration: BoxDecoration(
                color: trackColor,
                borderRadius: BorderRadius.circular(3),
              ),
            ),
            FractionallySizedBox(
              widthFactor: proficiency / 100,
              child: Container(
                height: 6,
                decoration: BoxDecoration(
                  color: primaryColor,
                  borderRadius: BorderRadius.circular(3),
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  String _formatFamilyName(String family) {
    return family
        .split('_')
        .map((w) => w.isNotEmpty ? '${w[0].toUpperCase()}${w.substring(1)}' : w)
        .join(' ');
  }
}

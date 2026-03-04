import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/gym.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/equipment_selector.dart';

class GymFormDialog extends StatefulWidget {
  final Gym? gym;

  const GymFormDialog({super.key, this.gym});

  @override
  State<GymFormDialog> createState() => _GymFormDialogState();
}

class _GymFormDialogState extends State<GymFormDialog> {
  late TextEditingController _nameController;
  List<String> _equipment = [];
  Set<String> _weightedModifiers = {};
  final Map<String, List<double>> _modifierVariants = {};

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.gym?.name ?? '');
    _equipment = List.from(widget.gym?.equipment ?? []);
    // parse existing modifier variants from gym
    final existingVariants = widget.gym?.modifierVariants;
    if (existingVariants != null) {
      for (final entry in existingVariants.entries) {
        final weights = (entry.value as List<dynamic>?)
            ?.map((v) => (v as num).toDouble())
            .toList();
        if (weights != null && weights.isNotEmpty) {
          _modifierVariants[entry.key] = weights;
        }
      }
    }
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
  }

  /// weighted modifiers that are currently selected as equipment
  List<String> get _selectedWeightedModifiers =>
      _equipment.where((e) => _weightedModifiers.contains(e)).toList();

  void _addWeight(String modifierId, double weight) {
    setState(() {
      final weights = _modifierVariants[modifierId] ?? [];
      if (!weights.contains(weight)) {
        _modifierVariants[modifierId] = [...weights, weight]..sort();
      }
    });
  }

  void _removeWeight(String modifierId, double weight) {
    setState(() {
      final weights = _modifierVariants[modifierId] ?? [];
      weights.remove(weight);
      if (weights.isEmpty) {
        _modifierVariants.remove(modifierId);
      } else {
        _modifierVariants[modifierId] = weights;
      }
    });
  }

  void _showAddWeightDialog(String modifierId) {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text('$modifierId — add weight (kg)'),
        content: TextField(
          controller: controller,
          keyboardType: const TextInputType.numberWithOptions(decimal: true),
          inputFormatters: [FilteringTextInputFormatter.allow(RegExp(r'[\d.]'))],
          autofocus: true,
          decoration: const InputDecoration(hintText: 'e.g. 10'),
          onSubmitted: (value) {
            final weight = double.tryParse(value);
            if (weight != null && weight > 0) {
              _addWeight(modifierId, weight);
              Navigator.of(context).pop();
            }
          },
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(AppLocalizations.of(context).cancel),
          ),
          TextButton(
            onPressed: () {
              final weight = double.tryParse(controller.text);
              if (weight != null && weight > 0) {
                _addWeight(modifierId, weight);
                Navigator.of(context).pop();
              }
            },
            child: Text(AppLocalizations.of(context).add),
          ),
        ],
      ),
    );
  }

  String _formatWeight(double w) =>
      w == w.truncateToDouble() ? '${w.toInt()}kg' : '${w.toStringAsFixed(1)}kg';

  void _submit() {
    final name = _nameController.text.trim();
    if (name.isEmpty) {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).pleaseEnterGymName,
      );
      return;
    }

    // only include variants for modifiers that are still selected
    final selectedModSet = _equipment.toSet();
    final cleanedVariants = Map<String, List<double>>.fromEntries(
      _modifierVariants.entries.where((e) => selectedModSet.contains(e.key)),
    );

    Navigator.of(context).pop({
      'name': name,
      'equipment': _equipment,
      'modifier_variants': cleanedVariants.isEmpty ? null : cleanedVariants,
    });
  }

  @override
  Widget build(BuildContext context) {
    final isEditing = widget.gym != null;
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor = VigorColors.textPrimary(context);
    final weightedSelected = _selectedWeightedModifiers;

    return Dialog(
      shape: const RoundedRectangleBorder(
        borderRadius: VigorRadius.modal,
      ),
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        constraints: const BoxConstraints(maxWidth: 500),
        padding: VigorSpacing.paddingLg,
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: VigorRadius.lg,
                isDark: isDark,
              )
            : null,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    isEditing ? l10n.editGym : l10n.addGym,
                    style: VigorTypography.headline.copyWith(color: textColor),
                  ),
                  IconButton(
                    icon: Icon(Icons.close, color: VigorColors.textSecondary(context)),
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const SizedBox(height: VigorSpacing.lg),
              AdaptiveTextField(
                controller: _nameController,
                labelText: l10n.gymName,
                placeholder: l10n.gymNamePlaceholder,
              ),
              const SizedBox(height: VigorSpacing.lg),
              Text(
                l10n.equipment,
                style: VigorTypography.label.copyWith(
                  color: textColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: VigorSpacing.sm),
              EquipmentSelector(
                selected: _equipment,
                onChanged: (updated) => setState(() => _equipment = updated),
                onWeightedModifiersLoaded: (weighted) =>
                    setState(() => _weightedModifiers = weighted),
              ),
              // modifier variant configuration for selected weighted modifiers
              if (weightedSelected.isNotEmpty) ...[
                const SizedBox(height: VigorSpacing.lg),
                Text(
                  l10n.availableWeights,
                  style: VigorTypography.label.copyWith(
                    color: textColor,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(height: VigorSpacing.xs),
                Text(
                  l10n.availableWeightsHint,
                  style: VigorTypography.caption.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
                const SizedBox(height: VigorSpacing.sm),
                ...weightedSelected.map((modId) {
                  final weights = _modifierVariants[modId] ?? [];
                  return Padding(
                    padding: const EdgeInsets.only(bottom: VigorSpacing.sm),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(modId, style: VigorTypography.body.copyWith(color: textColor)),
                        const SizedBox(height: VigorSpacing.xs),
                        Wrap(
                          spacing: VigorSpacing.xs,
                          runSpacing: VigorSpacing.xs,
                          children: [
                            ...weights.map((w) => Chip(
                              label: Text(_formatWeight(w), style: VigorTypography.caption.copyWith(
                                color: VigorColors.textPrimary(context),
                              )),
                              onDeleted: () => _removeWeight(modId, w),
                              deleteIconColor: VigorColors.textSecondary(context),
                              visualDensity: VisualDensity.compact,
                              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            )),
                            ActionChip(
                              label: Icon(Icons.add, size: 16, color: VigorColors.indigoAdaptive(context)),
                              onPressed: () => _showAddWeightDialog(modId),
                              visualDensity: VisualDensity.compact,
                              materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                            ),
                          ],
                        ),
                      ],
                    ),
                  );
                }),
              ],
              const SizedBox(height: VigorSpacing.lg),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  AdaptiveTextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: Text(l10n.cancel),
                  ),
                  const SizedBox(width: VigorSpacing.sm),
                  AdaptiveButton(
                    onPressed: _submit,
                    child: Text(isEditing ? l10n.update : l10n.add),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

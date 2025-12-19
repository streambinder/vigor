import 'package:flutter/material.dart';
import '../generated/app_localizations.dart';
import '../models/gym.dart';
import '../services/gym_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

class GymFormDialog extends StatefulWidget {
  final Gym? gym;

  const GymFormDialog({super.key, this.gym});

  @override
  State<GymFormDialog> createState() => _GymFormDialogState();
}

class _GymFormDialogState extends State<GymFormDialog> {
  late TextEditingController _nameController;
  final GymService _gymService = GymService();

  List<String> _equipment = [];
  final TextEditingController _equipmentInputController = TextEditingController();
  bool _loadingEquipment = false;

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.gym?.name ?? '');
    _equipment = List.from(widget.gym?.equipment ?? []);
  }

  @override
  void dispose() {
    _nameController.dispose();
    _equipmentInputController.dispose();
    super.dispose();
  }

  void _addEquipment() {
    final equipment = _equipmentInputController.text.trim();
    if (equipment.isNotEmpty && !_equipment.contains(equipment)) {
      setState(() {
        _equipment.add(equipment);
        _equipmentInputController.clear();
      });
    }
  }

  void _removeEquipment(String equipment) {
    setState(() {
      _equipment.remove(equipment);
    });
  }

  Future<void> _addAllEquipment() async {
    setState(() => _loadingEquipment = true);
    final response = await _gymService.getAvailableEquipment();
    setState(() => _loadingEquipment = false);

    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      setState(() {
        _equipment.addAll(response.data!);
      });
    } else {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).failedToLoadEquipment,
      );
    }
  }

  void _submit() {
    final name = _nameController.text.trim();
    if (name.isEmpty) {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).pleaseEnterGymName,
      );
      return;
    }

    Navigator.of(context).pop({
      'name': name,
      'equipment': _equipment,
    });
  }

  @override
  Widget build(BuildContext context) {
    final isEditing = widget.gym != null;
    final l10n = AppLocalizations.of(context);

    return Dialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
      ),
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        constraints: const BoxConstraints(maxWidth: 500),
        padding: const EdgeInsets.all(24),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 20,
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
                    style: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.headlineStyle
                        : const TextStyle(
                            fontSize: 20,
                            fontWeight: FontWeight.bold,
                          ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              AdaptiveTextField(
                controller: _nameController,
                labelText: l10n.gymName,
                placeholder: l10n.gymNamePlaceholder,
              ),
              const SizedBox(height: 24),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    l10n.equipment,
                    style: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                        : const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                  ),
                  TextButton(
                    onPressed: _loadingEquipment ? null : _addAllEquipment,
                    child: _loadingEquipment
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Text(l10n.addAllEquipment),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: AdaptiveTextField(
                      controller: _equipmentInputController,
                      labelText: l10n.addEquipment,
                      placeholder: l10n.equipmentPlaceholder,
                      onSubmitted: (_) => _addEquipment(),
                    ),
                  ),
                  const SizedBox(width: 8),
                  AdaptiveButton(
                    onPressed: _addEquipment,
                    child: const Icon(Icons.add),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              if (_equipment.isEmpty)
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  child: Center(
                    child: Text(
                      l10n.noEquipmentAddedYet,
                      style: TextStyle(
                        color: Colors.grey[600],
                        fontStyle: FontStyle.italic,
                      ),
                    ),
                  ),
                )
              else
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: _equipment.map((equipment) {
                    return Chip(
                      label: Text(equipment),
                      deleteIcon: const Icon(Icons.close, size: 18),
                      onDeleted: () => _removeEquipment(equipment),
                      backgroundColor: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                          : Theme.of(context).colorScheme.primaryContainer,
                    );
                  }).toList(),
                ),
              const SizedBox(height: 24),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: Text(l10n.cancel),
                  ),
                  const SizedBox(width: 8),
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

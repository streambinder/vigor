import 'package:flutter/material.dart';
import '../generated/app_localizations.dart';
import '../models/gym.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/equipment_selector.dart';
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
  List<String> _equipment = [];

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.gym?.name ?? '');
    _equipment = List.from(widget.gym?.equipment ?? []);
  }

  @override
  void dispose() {
    _nameController.dispose();
    super.dispose();
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
              Text(
                l10n.equipment,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                    : const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
              ),
              const SizedBox(height: 8),
              EquipmentSelector(
                selected: _equipment,
                onChanged: (updated) => setState(() => _equipment = updated),
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

import 'package:flutter/material.dart';
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
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final textColor = VigorColors.textPrimary(context);

    return Dialog(
      shape: RoundedRectangleBorder(
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
              SizedBox(height: VigorSpacing.lg),
              AdaptiveTextField(
                controller: _nameController,
                labelText: l10n.gymName,
                placeholder: l10n.gymNamePlaceholder,
              ),
              SizedBox(height: VigorSpacing.lg),
              Text(
                l10n.equipment,
                style: VigorTypography.label.copyWith(
                  color: textColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
              SizedBox(height: VigorSpacing.sm),
              EquipmentSelector(
                selected: _equipment,
                onChanged: (updated) => setState(() => _equipment = updated),
              ),
              SizedBox(height: VigorSpacing.lg),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  AdaptiveTextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: Text(l10n.cancel),
                  ),
                  SizedBox(width: VigorSpacing.sm),
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

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../generated/app_localizations.dart';
import '../models/gym.dart';
import '../models/training.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';
import '../services/preferences_service.dart';
import '../services/user_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/equipment_selector.dart';
import '../widgets/user_select_dialog.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

enum EquipmentMode { bodyweight, gym, custom }

class TrainingGenerationModal extends StatefulWidget {
  final List<Gym> gyms;
  final Function(Training)? onSuccess;

  const TrainingGenerationModal({
    super.key,
    required this.gyms,
    this.onSuccess,
  });

  @override
  State<TrainingGenerationModal> createState() => _TrainingGenerationModalState();
}

class _TrainingGenerationModalState extends State<TrainingGenerationModal> {
  final _formKey = GlobalKey<FormState>();
  final _durationController = TextEditingController(text: '60');
  final _promptController = TextEditingController();

  EquipmentMode _equipmentMode = EquipmentMode.bodyweight;
  Gym? _selectedGym;
  bool _isGenerating = false;
  bool _includeWarmupCooldown = true;
  final List<UserInfo> _partners = [];
  List<String> _equipment = [];

  @override
  void initState() {
    super.initState();
    final prefs = context.read<PreferencesService>();
    final defaultId = prefs.defaultGymId;
    // default to gym mode if user has a default gym set, otherwise bodyweight
    if (defaultId != null && widget.gyms.any((g) => g.id == defaultId)) {
      _equipmentMode = EquipmentMode.gym;
      _selectedGym = widget.gyms.firstWhere((g) => g.id == defaultId);
    } else if (widget.gyms.isNotEmpty) {
      _selectedGym = widget.gyms.first;
    }
  }

  @override
  void dispose() {
    _durationController.dispose();
    _promptController.dispose();
    super.dispose();
  }

  Future<void> _addPartner() async {
    final l10n = AppLocalizations.of(context);
    final user = await showUserSelectDialog(
      context: context,
      title: l10n.addPartner,
    );
    if (user != null && !_partners.any((p) => p.id == user.id)) {
      setState(() {
        _partners.add(user);
      });
    }
  }

  void _removePartner(UserInfo partner) {
    setState(() {
      _partners.remove(partner);
    });
  }

  Widget _buildEquipmentSection() {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.equipment,
          style: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.captionStyle
              : Theme.of(context).textTheme.labelMedium,
        ),
        const SizedBox(height: 8),
        // ternary segmented button
        SizedBox(
          width: double.infinity,
          child: SegmentedButton<EquipmentMode>(
            segments: [
              ButtonSegment(
                value: EquipmentMode.bodyweight,
                label: Text(l10n.bodyweight),
                icon: const Icon(Icons.accessibility_new, size: 16),
              ),
              ButtonSegment(
                value: EquipmentMode.gym,
                label: Text(l10n.gym),
                icon: const Icon(Icons.fitness_center, size: 16),
              ),
              ButtonSegment(
                value: EquipmentMode.custom,
                label: Text(l10n.custom),
                icon: const Icon(Icons.build, size: 16),
              ),
            ],
            selected: {_equipmentMode},
            onSelectionChanged: (Set<EquipmentMode> selected) {
              setState(() {
                _equipmentMode = selected.first;
              });
            },
            showSelectedIcon: false,
          ),
        ),
        const SizedBox(height: 12),
        // mode-specific content
        _buildEquipmentModeContent(),
      ],
    );
  }

  Widget _buildEquipmentModeContent() {
    final l10n = AppLocalizations.of(context);
    switch (_equipmentMode) {
      case EquipmentMode.bodyweight:
        return Text(
          l10n.noEquipmentBodyweightOnly,
          style: TextStyle(
            color: Colors.grey[600],
            fontStyle: FontStyle.italic,
            fontSize: 12,
          ),
        );
      case EquipmentMode.gym:
        return _buildGymSelector();
      case EquipmentMode.custom:
        return _buildCustomEquipmentInput();
    }
  }

  Widget _buildGymSelector() {
    final l10n = AppLocalizations.of(context);
    if (widget.gyms.isEmpty) {
      return Text(
        l10n.noGymsDefinedCreateOne,
        style: TextStyle(
          color: Colors.grey[600],
          fontStyle: FontStyle.italic,
          fontSize: 12,
        ),
      );
    }
    return Container(
      decoration: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.glassDecoration(
              borderRadius: 12,
              border: Border.all(
                color: Colors.white.withOpacity(0.2),
                width: 1.5,
              ),
            )
          : BoxDecoration(
              border: Border.all(
                color: Theme.of(context).colorScheme.outline,
              ),
              borderRadius: BorderRadius.circular(12),
            ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<Gym>(
          value: _selectedGym,
          isExpanded: true,
          hint: Text(
            l10n.selectAGym,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.bodyStyle.copyWith(
                    color: Colors.white.withOpacity(0.5),
                  )
                : TextStyle(color: Colors.grey[600]),
          ),
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
          borderRadius: BorderRadius.circular(12),
          items: widget.gyms.map((gym) {
            return DropdownMenuItem<Gym>(
              value: gym,
              child: Text(
                gym.name,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.bodyStyle
                    : null,
              ),
            );
          }).toList(),
          onChanged: (Gym? value) {
            setState(() {
              _selectedGym = value;
            });
          },
        ),
      ),
    );
  }

  Widget _buildCustomEquipmentInput() {
    return EquipmentSelector(
      selected: _equipment,
      onChanged: (updated) => setState(() => _equipment = updated),
    );
  }

  Widget _buildRoutineToggles() {
    final l10n = AppLocalizations.of(context);
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.glassDecoration(
              borderRadius: 12,
              border: Border.all(
                color: Colors.white.withOpacity(0.2),
                width: 1.5,
              ),
            )
          : BoxDecoration(
              border: Border.all(
                color: Theme.of(context).colorScheme.outline,
              ),
              borderRadius: BorderRadius.circular(12),
            ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(
            l10n.includeWarmupCooldown,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.bodyStyle
                : null,
          ),
          AdaptiveSwitch(
            value: _includeWarmupCooldown,
            onChanged: (value) => setState(() => _includeWarmupCooldown = value),
          ),
        ],
      ),
    );
  }

  Future<void> _generateTraining() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isGenerating = true;
    });

    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);

    final duration = int.parse(_durationController.text);
    final prompt = _promptController.text.trim();

    // determine gym and equipment based on selected mode
    String gymId = '';
    List<String>? equipment;
    switch (_equipmentMode) {
      case EquipmentMode.bodyweight:
        // empty gym and equipment = bodyweight only
        break;
      case EquipmentMode.gym:
        gymId = _selectedGym?.id ?? '';
        break;
      case EquipmentMode.custom:
        equipment = _equipment.isEmpty ? null : _equipment;
        break;
    }

    final response = await trainingService.generateTraining(
      duration: duration,
      gym: gymId,
      prompt: prompt.isEmpty ? null : prompt,
      equipment: equipment,
      partners: _partners.isEmpty ? null : _partners.map((p) => p.id).toList(),
      skipWarmupCooldown: !_includeWarmupCooldown ? true : null,
    );

    if (mounted) {
      setState(() {
        _isGenerating = false;
      });

      if (response.isSuccess) {
        Navigator.of(context).pop();
        AdaptiveNotification.show(
          context: context,
          message: AppLocalizations.of(context).trainingGeneratedSuccessfully,
        );
        widget.onSuccess?.call(response.data!);
      } else {
        AdaptiveNotification.showError(
          context: context,
          message: AppLocalizations.of(context).failedToGenerateTraining,
          rawError: response.error,
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 20,
                border: Border.all(
                  color: Colors.white.withOpacity(0.2),
                  width: 1.5,
                ),
              )
            : BoxDecoration(
                color: Theme.of(context).colorScheme.surface,
                borderRadius: BorderRadius.circular(20),
              ),
        child: ClipRRect(
          borderRadius: BorderRadius.circular(20),
          child: _isGenerating
              ? _buildLoadingView()
              : _buildFormView(),
        ),
      ),
    );
  }

  Widget _buildLoadingView() {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.all(32.0),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const AdaptiveLoadingIndicator(),
          const SizedBox(height: 24),
          Text(
            l10n.generatingTraining,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.headlineStyle
                : Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          Text(
            l10n.thisMayTakeAMoment,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.captionStyle
                : Theme.of(context).textTheme.bodyMedium?.copyWith(
                      color: Colors.grey,
                    ),
          ),
        ],
      ),
    );
  }

  Widget _buildFormView() {
    final l10n = AppLocalizations.of(context);
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24.0),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              l10n.generateTraining,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 24)
                  : Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 24),

            // Duration field
            AdaptiveTextField(
              controller: _durationController,
              labelText: l10n.durationMinutes,
              keyboardType: TextInputType.number,
              inputFormatters: [
                FilteringTextInputFormatter.digitsOnly,
              ],
            ),
            const SizedBox(height: 16),

            // Equipment mode selection
            _buildEquipmentSection(),
            const SizedBox(height: 16),

            // Routine skip toggles
            _buildRoutineToggles(),
            const SizedBox(height: 16),

            // Optional prompt
            AdaptiveTextField(
              controller: _promptController,
              labelText: l10n.customPromptOptional,
              placeholder: l10n.focusOnUpperBody,
              maxLines: 3,
              minLines: 1,
            ),
            const SizedBox(height: 16),

            // Partners section
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.trainingPartnersOptional,
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.captionStyle
                      : Theme.of(context).textTheme.labelMedium,
                ),
                const SizedBox(height: 8),
                AdaptiveButton(
                  onPressed: _addPartner,
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.person_add, size: 18),
                      const SizedBox(width: 8),
                      Text(l10n.addPartner),
                    ],
                  ),
                ),
                if (_partners.isNotEmpty) ...[
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: _partners.map((partner) {
                      return Chip(
                        label: Text(
                          partner.displayName,
                          style: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.captionStyle.copyWith(fontSize: 12)
                              : null,
                        ),
                        deleteIcon: const Icon(Icons.close, size: 16),
                        onDeleted: () => _removePartner(partner),
                        backgroundColor: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.primaryColor.withOpacity(0.1)
                            : null,
                      );
                    }).toList(),
                  ),
                ],
              ],
            ),
            const SizedBox(height: 24),

            // Action buttons
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                AdaptiveTextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: Text(l10n.cancel),
                ),
                const SizedBox(width: 12),
                AdaptiveButton(
                  onPressed: _generateTraining,
                  child: Text(l10n.generate),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

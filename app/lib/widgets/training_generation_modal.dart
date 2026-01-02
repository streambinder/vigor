import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/gym.dart';
import '../models/training.dart';
import '../services/service_locator.dart';
import '../services/preferences_service.dart';
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
  int? _retryAttempt;
  bool _includeWarmupCooldown = true;
  final List<UserInfo> _partners = [];
  List<String> _equipment = [];
  String? _methodology;

  static const _methodologies = [
    'strength',
    'circuit',
    'emom',
    'amrap',
    'hiit',
    'for_time',
    'endurance',
    'mobility',
  ];

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
          style: VigorTypography.caption.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        SizedBox(height: VigorSpacing.sm),
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
        SizedBox(height: VigorSpacing.sm),
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
          style: VigorTypography.caption.copyWith(
            color: VigorColors.textSecondary(context),
            fontStyle: FontStyle.italic,
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
    final isDark = Theme.of(context).brightness == Brightness.dark;
    if (widget.gyms.isEmpty) {
      return Text(
        l10n.noGymsDefinedCreateOne,
        style: VigorTypography.caption.copyWith(
          color: VigorColors.textSecondary(context),
          fontStyle: FontStyle.italic,
        ),
      );
    }
    return Container(
      decoration: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.glassDecoration(
              borderRadius: 12,
              isDark: isDark,
              border: Border.all(
                color: isDark ? Colors.white.withValues(alpha: 0.2) : Colors.black.withValues(alpha: 0.1),
                width: 1.5,
              ),
            )
          : BoxDecoration(
              border: Border.all(
                color: VigorColors.darkBorder,
              ),
              borderRadius: VigorRadius.radiusMd,
            ),
      child: DropdownButtonHideUnderline(
        child: DropdownButton<Gym>(
          value: _selectedGym,
          isExpanded: true,
          hint: Text(
            l10n.selectAGym,
            style: VigorTypography.body.copyWith(
              color: VigorColors.textMuted(context),
            ),
          ),
          padding: EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.xs),
          borderRadius: VigorRadius.radiusMd,
          items: widget.gyms.map((gym) {
            return DropdownMenuItem<Gym>(
              value: gym,
              child: Text(
                gym.name,
                style: VigorTypography.body,
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

  String _methodologyLabel(String methodology, AppLocalizations l10n) {
    switch (methodology) {
      case 'strength':
        return l10n.workoutTypeStrength;
      case 'circuit':
        return l10n.workoutTypeCircuit;
      case 'emom':
        return l10n.workoutTypeEmom;
      case 'amrap':
        return l10n.workoutTypeAmrap;
      case 'hiit':
        return l10n.workoutTypeHiit;
      case 'for_time':
        return l10n.workoutTypeForTime;
      case 'endurance':
        return l10n.workoutTypeEndurance;
      case 'mobility':
        return l10n.workoutTypeMobility;
      default:
        return methodology;
    }
  }

  Widget _buildMethodologySection() {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.methodologyOptional,
          style: VigorTypography.caption.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        SizedBox(height: VigorSpacing.sm),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            // "auto" chip for unset
            FilterChip(
              label: Text(l10n.methodologyAuto),
              selected: _methodology == null,
              onSelected: (selected) {
                if (selected) setState(() => _methodology = null);
              },
            ),
            ..._methodologies.map((m) {
              final isSelected = _methodology == m;
              return FilterChip(
                label: Text(_methodologyLabel(m, l10n)),
                selected: isSelected,
                onSelected: (selected) {
                  setState(() => _methodology = selected ? m : null);
                },
              );
            }),
          ],
        ),
      ],
    );
  }

  Widget _buildRoutineToggles() {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: () => setState(() => _includeWarmupCooldown = !_includeWarmupCooldown),
      child: Container(
        padding: EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.sm),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 12,
                isDark: isDark,
                border: Border.all(
                  color: isDark ? Colors.white.withValues(alpha: 0.2) : Colors.black.withValues(alpha: 0.1),
                  width: 1.5,
                ),
              )
            : BoxDecoration(
                border: Border.all(
                  color: VigorColors.darkBorder,
                ),
                borderRadius: VigorRadius.radiusMd,
              ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              l10n.includeWarmupCooldown,
              style: VigorTypography.body,
            ),
            AdaptiveSwitch(
              value: _includeWarmupCooldown,
              onChanged: (value) => setState(() => _includeWarmupCooldown = value),
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _generateTraining() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    setState(() {
      _isGenerating = true;
      _retryAttempt = null;
    });

    final trainingService = context.read<ServiceLocator>().trainingService;

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
      methodology: _methodology,
      onRetry: (attempt) {
        if (mounted) {
          setState(() => _retryAttempt = attempt);
        }
      },
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
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return Dialog(
      backgroundColor: Colors.transparent,
      child: Container(
        constraints: const BoxConstraints(maxWidth: 500),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 20,
                isDark: isDark,
                border: Border.all(
                  color: isDark ? Colors.white.withValues(alpha: 0.2) : Colors.black.withValues(alpha: 0.1),
                  width: 1.5,
                ),
              )
            : BoxDecoration(
                color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
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
    final statusText = _retryAttempt != null
        ? l10n.generationFailedRetrying(_retryAttempt!)
        : l10n.thisMayTakeAMoment;
    return Padding(
      padding: VigorSpacing.paddingXl,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const AdaptiveLoadingIndicator(),
          SizedBox(height: VigorSpacing.lg),
          Text(
            l10n.generatingTraining,
            style: VigorTypography.headline.copyWith(
              color: VigorColors.textPrimary(context),
            ),
          ),
          SizedBox(height: VigorSpacing.sm),
          Text(
            statusText,
            style: VigorTypography.caption.copyWith(
              color: VigorColors.textSecondary(context),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFormView() {
    final l10n = AppLocalizations.of(context);
    return SingleChildScrollView(
      padding: VigorSpacing.paddingLg,
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              l10n.generateTraining,
              style: VigorTypography.headline.copyWith(
                fontSize: 24,
                color: VigorColors.textPrimary(context),
              ),
            ),
            SizedBox(height: VigorSpacing.lg),

            // Duration field
            AdaptiveTextField(
              controller: _durationController,
              labelText: l10n.durationMinutes,
              keyboardType: TextInputType.number,
              inputFormatters: [
                FilteringTextInputFormatter.digitsOnly,
              ],
            ),
            SizedBox(height: VigorSpacing.md),

            // Equipment mode selection
            _buildEquipmentSection(),
            SizedBox(height: VigorSpacing.md),

            // Methodology selection
            _buildMethodologySection(),
            SizedBox(height: VigorSpacing.md),

            // Routine skip toggles
            _buildRoutineToggles(),
            SizedBox(height: VigorSpacing.md),

            // Optional prompt
            AdaptiveTextField(
              controller: _promptController,
              labelText: l10n.customPromptOptional,
              placeholder: l10n.focusOnUpperBody,
              maxLines: 3,
              minLines: 1,
            ),
            SizedBox(height: VigorSpacing.md),

            // Partners section
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.trainingPartnersOptional,
                  style: VigorTypography.caption.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
                SizedBox(height: VigorSpacing.sm),
                AdaptiveButton(
                  onPressed: _addPartner,
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      const Icon(Icons.person_add, size: 18),
                      SizedBox(width: VigorSpacing.sm),
                      Text(l10n.addPartner),
                    ],
                  ),
                ),
                if (_partners.isNotEmpty) ...[
                  SizedBox(height: VigorSpacing.sm),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: _partners.map((partner) {
                      return Chip(
                        label: Text(
                          partner.displayName,
                          style: VigorTypography.caption,
                        ),
                        deleteIcon: const Icon(Icons.close, size: 16),
                        onDeleted: () => _removePartner(partner),
                        backgroundColor: VigorColors.orange.withValues(alpha: 0.1),
                      );
                    }).toList(),
                  ),
                ],
              ],
            ),
            SizedBox(height: VigorSpacing.lg),

            // Action buttons
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                AdaptiveTextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: Text(l10n.cancel),
                ),
                SizedBox(width: VigorSpacing.sm),
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

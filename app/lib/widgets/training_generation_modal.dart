import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/gym.dart';
import '../models/profile_data.dart' as profile_models;
import '../models/training.dart';
import '../providers/auth_provider.dart';
import '../services/service_locator.dart';
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
  final _promptController = TextEditingController();

  late int _duration; // minutes, range: 10-180

  EquipmentMode _equipmentMode = EquipmentMode.bodyweight;
  Gym? _selectedGym;
  bool _isGenerating = false;
  int? _retryAttempt;
  bool _includeWarmupCooldown = true;
  final List<UserInfo> _partners = [];
  List<String> _equipment = [];
  String? _methodology;
  List<String> _availableGoals = [];
  Set<String> _selectedGoals = {};
  List<String> _availableMuscles = [];
  final Set<String> _selectedMuscles = {};
  List<String> _availableMethodologies = [];
  bool _advancedExpanded = false;

  @override
  void initState() {
    super.initState();
    final prefs = context.read<PreferencesService>();
    _duration = prefs.defaultDuration;
    final defaultId = prefs.defaultGymId;
    // default to gym mode if user has a default gym set, otherwise bodyweight
    if (defaultId != null && widget.gyms.any((g) => g.id == defaultId)) {
      _equipmentMode = EquipmentMode.gym;
      _selectedGym = widget.gyms.firstWhere((g) => g.id == defaultId);
    } else if (widget.gyms.isNotEmpty) {
      _selectedGym = widget.gyms.first;
    }
    _loadGoals();
    _loadMuscles();
    _loadMethodologies();
  }

  Future<void> _loadGoals() async {
    final gymService = context.read<ServiceLocator>().gymService;
    final response = await gymService.getAvailableGoals();
    if (!mounted) return;
    if (response.isSuccess && response.data != null) {
      setState(() {
        _availableGoals = response.data!;
      });
    }

    // default-select user's current profile goals
    final profile = context.read<AuthProvider>().currentUser?.profile;
    if (profile != null && profile.data.isNotEmpty) {
      try {
        final profileData = profile_models.ProfileData.fromJson(profile.data);
        setState(() {
          _selectedGoals = profileData.goals.toSet();
        });
      } catch (_) {}
    }
  }

  Future<void> _loadMuscles() async {
    final gymService = context.read<ServiceLocator>().gymService;
    final response = await gymService.getAvailableMuscles();
    if (!mounted) return;
    if (response.isSuccess && response.data != null) {
      setState(() {
        _availableMuscles = response.data!;
      });
    }
  }

  Future<void> _loadMethodologies() async {
    final gymService = context.read<ServiceLocator>().gymService;
    final response = await gymService.getAvailableMethodologies();
    if (!mounted) return;
    if (response.isSuccess && response.data != null) {
      setState(() {
        _availableMethodologies = response.data!;
      });
    }
  }

  @override
  void dispose() {
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
        const SizedBox(height: VigorSpacing.sm),
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
        const SizedBox(height: VigorSpacing.sm),
        // mode-specific content
        _buildEquipmentModeContent(),
      ],
    );
  }

  Widget _buildEquipmentModeContent() {
    switch (_equipmentMode) {
      case EquipmentMode.bodyweight:
        return const SizedBox.shrink();
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
                color: VigorColors.border(context),
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
          padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.xs),
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
            // persist the selected gym as default
            if (value != null) {
              context.read<PreferencesService>().setDefaultGymId(value.id);
            }
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
    if (_availableMethodologies.isEmpty) {
      return const SizedBox.shrink();
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.methodologyOptional,
          style: VigorTypography.caption.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        const SizedBox(height: VigorSpacing.sm),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            // "auto" chip for unset
            FilterChip(
              label: Text(
                l10n.methodologyAuto,
                style: VigorTypography.caption.copyWith(
                  color: _methodology == null ? Colors.white : VigorColors.textPrimary(context),
                ),
              ),
              selected: _methodology == null,
              selectedColor: VigorColors.indigo,
              checkmarkColor: Colors.white,
              onSelected: (selected) {
                if (selected) setState(() => _methodology = null);
              },
            ),
            ..._availableMethodologies.map((m) {
              final isSelected = _methodology == m;
              return FilterChip(
                label: Text(
                  _methodologyLabel(m, l10n),
                  style: VigorTypography.caption.copyWith(
                    color: isSelected ? Colors.white : VigorColors.textPrimary(context),
                  ),
                ),
                selected: isSelected,
                selectedColor: VigorColors.indigo,
                checkmarkColor: Colors.white,
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

  String _goalLabel(String goal) {
    // format goal id to human-readable: muscle_building -> Muscle Building
    return goal.split('_').map((w) => w.isEmpty ? '' : '${w[0].toUpperCase()}${w.substring(1)}').join(' ');
  }

  Widget _buildGoalsSection() {
    final l10n = AppLocalizations.of(context);
    if (_availableGoals.isEmpty) {
      return const SizedBox.shrink();
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.goalsOptional,
          style: VigorTypography.caption.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        const SizedBox(height: VigorSpacing.sm),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: _availableGoals.map((goal) {
            final isSelected = _selectedGoals.contains(goal);
            return FilterChip(
              label: Text(
                _goalLabel(goal),
                style: VigorTypography.caption.copyWith(
                  color: isSelected ? Colors.white : VigorColors.textPrimary(context),
                ),
              ),
              selected: isSelected,
              selectedColor: VigorColors.indigo,
              checkmarkColor: Colors.white,
              onSelected: (selected) {
                setState(() {
                  if (selected) {
                    _selectedGoals.add(goal);
                  } else {
                    _selectedGoals.remove(goal);
                  }
                });
              },
            );
          }).toList(),
        ),
      ],
    );
  }

  String _muscleLabel(String muscle) {
    // format muscle id to human-readable: chest -> Chest
    return muscle.split('_').map((w) => w.isEmpty ? '' : '${w[0].toUpperCase()}${w.substring(1)}').join(' ');
  }

  Widget _buildMusclesSection() {
    final l10n = AppLocalizations.of(context);
    if (_availableMuscles.isEmpty) {
      return const SizedBox.shrink();
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(
          l10n.musclesOptional,
          style: VigorTypography.caption.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        const SizedBox(height: VigorSpacing.sm),
        Wrap(
          spacing: 8,
          runSpacing: 8,
          children: [
            // "auto" chip for unset (no muscle filter)
            FilterChip(
              label: Text(
                l10n.musclesAuto,
                style: VigorTypography.caption.copyWith(
                  color: _selectedMuscles.isEmpty ? Colors.white : VigorColors.textPrimary(context),
                ),
              ),
              selected: _selectedMuscles.isEmpty,
              selectedColor: VigorColors.indigo,
              checkmarkColor: Colors.white,
              onSelected: (selected) {
                if (selected) setState(() => _selectedMuscles.clear());
              },
            ),
            ..._availableMuscles.map((muscle) {
              final isSelected = _selectedMuscles.contains(muscle);
              return FilterChip(
                label: Text(
                  _muscleLabel(muscle),
                  style: VigorTypography.caption.copyWith(
                    color: isSelected ? Colors.white : VigorColors.textPrimary(context),
                  ),
                ),
                selected: isSelected,
                selectedColor: VigorColors.indigo,
                checkmarkColor: Colors.white,
                onSelected: (selected) {
                  setState(() {
                    if (selected) {
                      _selectedMuscles.add(muscle);
                    } else {
                      _selectedMuscles.remove(muscle);
                    }
                  });
                },
              );
            }),
          ],
        ),
      ],
    );
  }

  Widget _buildAdvancedHeader() {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: () => setState(() => _advancedExpanded = !_advancedExpanded),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
        decoration: BoxDecoration(
          border: Border(
            top: BorderSide(
              color: isDark ? Colors.white.withValues(alpha: 0.1) : Colors.black.withValues(alpha: 0.08),
              width: 1,
            ),
          ),
        ),
        child: Row(
          children: [
            Icon(
              Icons.tune,
              size: 18,
              color: VigorColors.textSecondary(context),
            ),
            const SizedBox(width: VigorSpacing.sm),
            Text(
              l10n.advancedSettings,
              style: VigorTypography.label.copyWith(
                color: VigorColors.textSecondary(context),
              ),
            ),
            const Spacer(),
            AnimatedRotation(
              turns: _advancedExpanded ? 0.5 : 0,
              duration: VigorAnimation.fast,
              child: Icon(
                Icons.keyboard_arrow_down,
                size: 20,
                color: VigorColors.textSecondary(context),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildAdvancedContent() {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.stretch,
      children: [
        // Custom prompt
        AdaptiveTextField(
          controller: _promptController,
          labelText: l10n.customPromptOptional,
          placeholder: l10n.focusOnUpperBody,
          maxLines: 2,
          minLines: 1,
          maxLength: 500,
        ),
        const SizedBox(height: VigorSpacing.md),

        // Warmup/cooldown toggle
        _buildWarmupToggle(),
        const SizedBox(height: VigorSpacing.md),

        // Methodology selection
        _buildMethodologySection(),
        if (_availableMethodologies.isNotEmpty) const SizedBox(height: VigorSpacing.md),

        // Goals selection
        _buildGoalsSection(),
        if (_availableGoals.isNotEmpty) const SizedBox(height: VigorSpacing.md),

        // Muscles selection
        _buildMusclesSection(),
      ],
    );
  }

  Widget _buildWarmupToggle() {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: () => setState(() => _includeWarmupCooldown = !_includeWarmupCooldown),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.sm),
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
                border: Border.all(color: VigorColors.border(context)),
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

  Widget _buildPartnersSection() {
    final l10n = AppLocalizations.of(context);
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              l10n.trainingPartnersOptional,
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
              ),
            ),
            const Spacer(),
            AdaptiveButton(
              onPressed: _addPartner,
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  const Icon(Icons.person_add, size: 16),
                  const SizedBox(width: VigorSpacing.xs),
                  Text(l10n.add),
                ],
              ),
            ),
          ],
        ),
        if (_partners.isNotEmpty) ...[
          const SizedBox(height: VigorSpacing.sm),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _partners.map((partner) {
              return Chip(
                label: Text(
                  partner.displayName,
                  style: VigorTypography.caption.copyWith(
                    color: VigorColors.textPrimary(context),
                  ),
                ),
                deleteIcon: Icon(Icons.close, size: 16, color: VigorColors.textSecondary(context)),
                onDeleted: () => _removePartner(partner),
                backgroundColor: VigorColors.indigo.withValues(alpha: 0.1),
              );
            }).toList(),
          ),
        ],
      ],
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
      duration: _duration,
      gym: gymId,
      prompt: prompt.isEmpty ? null : prompt,
      equipment: equipment,
      partners: _partners.isEmpty ? null : _partners.map((p) => p.id).toList(),
      skipWarmupCooldown: !_includeWarmupCooldown ? true : null,
      methodology: _methodology,
      goals: _selectedGoals.isEmpty ? null : _selectedGoals.toList(),
      muscles: _selectedMuscles.isEmpty ? null : _selectedMuscles.toList(),
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
          AdaptiveLoadingIndicator(color: VigorColors.persimmon),
          const SizedBox(height: VigorSpacing.lg),
          Text(
            l10n.generatingTraining,
            style: VigorTypography.headline.copyWith(
              color: VigorColors.textPrimary(context),
            ),
          ),
          const SizedBox(height: VigorSpacing.sm),
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
            const SizedBox(height: VigorSpacing.lg),

            // Duration slider
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      l10n.durationMinutes,
                      style: VigorTypography.caption.copyWith(
                        color: VigorColors.textSecondary(context),
                      ),
                    ),
                    Text(
                      '$_duration min',
                      style: VigorTypography.data.copyWith(
                        color: VigorColors.indigo,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
                Slider(
                  value: _duration.toDouble(),
                  min: 10,
                  max: 180,
                  divisions: 34, // (180-10)/5 = 34 steps of 5 minutes
                  activeColor: VigorColors.indigo,
                  onChanged: (value) => setState(() => _duration = value.round()),
                ),
              ],
            ),
            const SizedBox(height: VigorSpacing.md),

            // Equipment mode selection
            _buildEquipmentSection(),
            const SizedBox(height: VigorSpacing.md),

            // Partners section
            _buildPartnersSection(),
            const SizedBox(height: VigorSpacing.md),

            // Advanced settings collapsible
            _buildAdvancedHeader(),
            AnimatedSize(
              duration: VigorAnimation.medium,
              curve: VigorAnimation.defaultCurve,
              alignment: Alignment.topCenter,
              child: _advancedExpanded
                  ? Padding(
                      padding: const EdgeInsets.only(top: VigorSpacing.md),
                      child: _buildAdvancedContent(),
                    )
                  : const SizedBox.shrink(),
            ),
            const SizedBox(height: VigorSpacing.lg),

            // Action buttons
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                AdaptiveTextButton(
                  onPressed: () => Navigator.of(context).pop(),
                  child: Text(l10n.cancel),
                ),
                const SizedBox(width: VigorSpacing.sm),
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

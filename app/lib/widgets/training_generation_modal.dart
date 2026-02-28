import 'dart:async';
import 'dart:math';

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
import '../utils/knowledge_labels.dart';
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
  final _random = Random();

  late int _duration; // minutes, range: 10-180

  EquipmentMode _equipmentMode = EquipmentMode.bodyweight;
  Gym? _selectedGym;
  bool _isGenerating = false;
  int? _retryAttempt;
  late bool _includeWarmupCooldown;
  final List<UserInfo> _partners = [];
  List<String> _equipment = [];
  String? _methodology;
  List<String> _availableGoals = [];
  Set<String> _selectedGoals = {};
  List<String> _availableMuscles = [];
  final Set<String> _selectedMuscles = {};
  List<String> _availableMethodologies = [];
  bool _advancedExpanded = false;
  List<int>? _recommendedDurationRange;
  late bool _useRecommendedDuration;

  // rotating status message state
  Timer? _messageTimer;
  int _currentMessageIndex = 0;
  List<String> _messagePool = [];
  bool _showingRetryMessage = false;
  int _retryCallbackId = 0;

  @override
  void initState() {
    super.initState();
    final prefs = context.read<PreferencesService>();
    _useRecommendedDuration = prefs.useRecommendedDuration;
    _duration = prefs.defaultDuration;
    _includeWarmupCooldown = prefs.warmupCooldown;
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
    _loadRecommendedDuration();
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

  Future<void> _loadRecommendedDuration() async {
    final progressService = context.read<ServiceLocator>().progressService;
    final response = await progressService.getWeeklyTarget();
    if (!mounted) return;
    if (response.isSuccess && response.data != null) {
      final mins = response.data!.recommendation.sessionDurationMins;
      if (mins.isNotEmpty) {
        setState(() {
          _recommendedDurationRange = mins;
          if (_useRecommendedDuration) {
            // snap to midpoint of recommended range, rounded to nearest 5 min (slider step)
            final midpoint = (mins[0] + mins[mins.length - 1]) / 2.0;
            _duration = (midpoint / 5).round() * 5;
          }
        });
      }
    }
  }

  @override
  void dispose() {
    _messageTimer?.cancel();
    _promptController.dispose();
    super.dispose();
  }

  // -- rotating status messages --

  List<String> _buildMessagePool() {
    final l10n = AppLocalizations.of(context);
    final messages = <String>[
      l10n.loadingMsg1,
      l10n.loadingMsg2,
      l10n.loadingMsg3,
      l10n.loadingMsg4,
      l10n.loadingMsg5,
      l10n.loadingMsg6,
      l10n.loadingMsg7,
      l10n.loadingMsg8,
      l10n.loadingMsg9,
      l10n.loadingMsg10,
      l10n.loadingMsg11,
      l10n.loadingMsg12,
      l10n.loadingMsg13,
      l10n.loadingMsg14,
      l10n.loadingMsg15,
      l10n.loadingMsg16,
      l10n.loadingMsg17,
      l10n.loadingMsg18,
      l10n.loadingMsg19,
      l10n.loadingMsg20,
    ];

    // contextual messages based on profile/modal state
    try {
      final profile = context.read<AuthProvider>().currentUser?.profile;
      if (profile != null && profile.data.isNotEmpty) {
        final profileData = profile_models.ProfileData.fromJson(profile.data);
        for (final goal in profileData.goals) {
          messages.add(l10n.loadingMsgGoal(KnowledgeLabels.goalLabel(goal, l10n)));
        }
        if (profileData.injuries.isNotEmpty) messages.add(l10n.loadingMsgInjuries);
        if (profileData.conditions.isNotEmpty) messages.add(l10n.loadingMsgConditions);
        if (profileData.preferences?.exercises?.isNotEmpty == true) {
          messages.add(l10n.loadingMsgFavorites);
        }
      }
    } catch (_) {}

    if (_methodology != null) {
      messages.add(l10n.loadingMsgMethodology(
        KnowledgeLabels.methodologyLabel(_methodology!, l10n),
      ));
    }
    if (_partners.isNotEmpty) messages.add(l10n.loadingMsgPartners);
    if (_equipmentMode == EquipmentMode.gym && _selectedGym != null) {
      messages.add(l10n.loadingMsgGym(_selectedGym!.name));
    }

    final trainings = context.read<ServiceLocator>().trainingsNotifier.value;
    if (trainings != null && trainings.isNotEmpty) {
      messages.add(l10n.loadingMsgHistory);
    }

    messages.shuffle(_random);
    return messages;
  }

  void _scheduleNextMessage() {
    _messageTimer?.cancel();
    final delay = 2000 + _random.nextInt(2001); // 2000-4000ms
    _messageTimer = Timer(Duration(milliseconds: delay), () {
      if (!mounted || !_isGenerating) return;
      setState(() {
        _currentMessageIndex = (_currentMessageIndex + 1) % _messagePool.length;
      });
      _scheduleNextMessage();
    });
  }

  void _startMessageRotation() {
    _currentMessageIndex = 0;
    _messagePool = _buildMessagePool();
    _scheduleNextMessage();
  }

  void _stopMessageRotation() {
    _messageTimer?.cancel();
    _messageTimer = null;
  }

  String _getRetryMessage(AppLocalizations l10n) {
    final retryMessages = [
      l10n.loadingRetryMsg1,
      l10n.loadingRetryMsg2,
      l10n.loadingRetryMsg3,
      l10n.loadingRetryMsg4,
      l10n.loadingRetryMsg5,
      l10n.loadingRetryMsg6,
    ];
    return retryMessages[(_retryAttempt ?? 0) % retryMessages.length];
  }

  // -- end rotating status messages --

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

  Widget _buildMethodologyTile(String methodology, bool isSelected, AppLocalizations l10n) {
    final isAuto = methodology == 'auto';
    return GestureDetector(
      onTap: () {
        setState(() {
          if (isAuto) {
            _methodology = null;
          } else {
            _methodology = isSelected ? null : methodology;
          }
        });
      },
      child: Container(
        padding: const EdgeInsets.all(VigorSpacing.sm),
        decoration: BoxDecoration(
          color: isSelected ? VigorColors.indigo.withValues(alpha: 0.1) : null,
          border: Border.all(
            color: isSelected ? VigorColors.indigo : VigorColors.border(context),
          ),
          borderRadius: VigorRadius.radiusMd,
        ),
        child: Row(
          children: [
            Icon(
              isSelected ? Icons.check_circle : Icons.circle_outlined,
              color: isSelected ? VigorColors.indigo : VigorColors.textMuted(context),
              size: 20,
            ),
            const SizedBox(width: VigorSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    KnowledgeLabels.methodologyLabel(methodology, l10n),
                    style: VigorTypography.body.copyWith(
                      fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                      color: VigorColors.textPrimary(context),
                    ),
                  ),
                  if (KnowledgeLabels.methodologyDescription(methodology, l10n).isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(
                      KnowledgeLabels.methodologyDescription(methodology, l10n),
                      style: VigorTypography.caption.copyWith(
                        color: VigorColors.textSecondary(context),
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
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
        // auto tile
        _buildMethodologyTile('auto', _methodology == null, l10n),
        const SizedBox(height: VigorSpacing.xs),
        ConstrainedBox(
          constraints: const BoxConstraints(maxHeight: 250),
          child: ListView.separated(
            shrinkWrap: true,
            itemCount: _availableMethodologies.length,
            separatorBuilder: (_, __) => const SizedBox(height: VigorSpacing.xs),
            itemBuilder: (context, index) {
              final methodology = _availableMethodologies[index];
              final isSelected = _methodology == methodology;
              return _buildMethodologyTile(methodology, isSelected, l10n);
            },
          ),
        ),
      ],
    );
  }

  Widget _buildGoalTile(String goal, bool isSelected, AppLocalizations l10n) {
    return GestureDetector(
      onTap: () {
        setState(() {
          if (isSelected) {
            _selectedGoals.remove(goal);
          } else if (_selectedGoals.length < 2) {
            _selectedGoals.add(goal);
          }
        });
      },
      child: Container(
        padding: const EdgeInsets.all(VigorSpacing.sm),
        decoration: BoxDecoration(
          color: isSelected ? VigorColors.indigo.withValues(alpha: 0.1) : null,
          border: Border.all(
            color: isSelected ? VigorColors.indigo : VigorColors.border(context),
          ),
          borderRadius: VigorRadius.radiusMd,
        ),
        child: Row(
          children: [
            Icon(
              isSelected ? Icons.check_circle : Icons.circle_outlined,
              color: isSelected ? VigorColors.indigo : VigorColors.textMuted(context),
              size: 20,
            ),
            const SizedBox(width: VigorSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    KnowledgeLabels.goalLabel(goal, l10n),
                    style: VigorTypography.body.copyWith(
                      fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                      color: VigorColors.textPrimary(context),
                    ),
                  ),
                  if (KnowledgeLabels.goalDescription(goal, l10n).isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(
                      KnowledgeLabels.goalDescription(goal, l10n),
                      style: VigorTypography.caption.copyWith(
                        color: VigorColors.textSecondary(context),
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildGoalsSection() {
    final l10n = AppLocalizations.of(context);
    if (_availableGoals.isEmpty) {
      return const SizedBox.shrink();
    }
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              l10n.goalsOptional,
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
              ),
            ),
            Text(
              '${_selectedGoals.length}/2',
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
              ),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.sm),
        ConstrainedBox(
          constraints: const BoxConstraints(maxHeight: 250),
          child: ListView.separated(
            shrinkWrap: true,
            itemCount: _availableGoals.length,
            separatorBuilder: (_, __) => const SizedBox(height: VigorSpacing.xs),
            itemBuilder: (context, index) {
              final goal = _availableGoals[index];
              final isSelected = _selectedGoals.contains(goal);
              return _buildGoalTile(goal, isSelected, l10n);
            },
          ),
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

        // Muscles selection
        _buildMusclesSection(),
        if (_availableMuscles.isNotEmpty) const SizedBox(height: VigorSpacing.md),

        // Methodology selection
        _buildMethodologySection(),
        if (_availableMethodologies.isNotEmpty) const SizedBox(height: VigorSpacing.md),

        // Goals selection
        _buildGoalsSection(),
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
      _showingRetryMessage = false;
    });
    _startMessageRotation();

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
        if (!mounted) return;
        setState(() {
          _retryAttempt = attempt;
          _showingRetryMessage = true;
        });
        _stopMessageRotation();
        // resume rotation after 4s, guard against stale callbacks
        final callbackId = ++_retryCallbackId;
        Timer(const Duration(seconds: 4), () {
          if (!mounted || !_isGenerating || callbackId != _retryCallbackId) return;
          setState(() => _showingRetryMessage = false);
          _startMessageRotation();
        });
      },
    );

    _stopMessageRotation();

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
    return PopScope(
      canPop: !_isGenerating,
      child: Dialog(
        backgroundColor: Colors.transparent,
        insetPadding: const EdgeInsets.symmetric(
          horizontal: VigorSpacing.xl,
          vertical: VigorSpacing.xl,
        ),
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
      ),
    );
  }

  Widget _buildDurationSlider(AppLocalizations l10n) {
    final rec = _recommendedDurationRange;
    final hasRecommendation = rec != null && rec.length >= 2;
    final isSingleValue = hasRecommendation && rec[0] == rec[1];
    // format recommendation label: "X min" if single value, "X-Y min" if range
    final recLabel = hasRecommendation
        ? (isSingleValue ? '${rec[0]} min' : '${rec[0]}-${rec[1]} min')
        : '';

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Text(
              l10n.duration,
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
              ),
            ),
            if (hasRecommendation) ...[
              const SizedBox(width: VigorSpacing.md),
              Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(
                  color: VigorColors.gold.withValues(alpha: 0.3),
                  borderRadius: BorderRadius.circular(2),
                  border: Border.all(color: VigorColors.gold, width: 1),
                ),
              ),
              const SizedBox(width: 4),
              Text(
                '${l10n.recommended}: $recLabel',
                style: VigorTypography.caption.copyWith(
                  color: VigorColors.gold,
                  fontSize: 11,
                ),
              ),
            ],
            const Spacer(),
            Text(
              '$_duration min',
              style: VigorTypography.data.copyWith(
                color: VigorColors.indigo,
                fontWeight: FontWeight.w600,
              ),
            ),
          ],
        ),
        LayoutBuilder(
          builder: (context, constraints) {
            const sliderPadding = 12.0;
            final trackWidth = constraints.maxWidth - (sliderPadding * 2);
            const minVal = 10.0;
            const maxVal = 180.0;
            const range = maxVal - minVal;

            Widget? rangeIndicator;
            if (hasRecommendation) {
              // ensure at least 20 min visual width: if range < 20, pad by 10 on each side
              final rawMin = rec[0];
              final rawMax = rec[1];
              final rangeWidth = rawMax - rawMin;
              final recMin = (rangeWidth < 20 ? rawMin - 10 : rawMin).clamp(10, 180).toDouble();
              final recMax = (rangeWidth < 20 ? rawMax + 10 : rawMax).clamp(10, 180).toDouble();
              final leftPos = sliderPadding + ((recMin - minVal) / range) * trackWidth;
              final rightPos = sliderPadding + ((recMax - minVal) / range) * trackWidth;
              final width = (rightPos - leftPos).clamp(4.0, trackWidth);

              rangeIndicator = Positioned(
                left: leftPos - 2,
                top: 12,
                child: IgnorePointer(
                  child: Container(
                    width: width + 4,
                    height: 24,
                    decoration: BoxDecoration(
                      color: VigorColors.gold.withValues(alpha: 0.25),
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: VigorColors.gold, width: 1.5),
                    ),
                  ),
                ),
              );
            }

            return Stack(
              clipBehavior: Clip.none,
              children: [
                // base slider (track only, thumb hidden)
                IgnorePointer(
                  child: SliderTheme(
                    data: SliderTheme.of(context).copyWith(
                      thumbShape: SliderComponentShape.noThumb,
                    ),
                    child: Slider(
                      value: _duration.toDouble(),
                      min: minVal,
                      max: maxVal,
                      divisions: 34,
                      activeColor: VigorColors.indigo,
                      onChanged: (_) {},
                    ),
                  ),
                ),
                // recommendation indicator (middle layer)
                if (rangeIndicator != null) rangeIndicator,
                // interactive slider with thumb on top
                SliderTheme(
                  data: SliderTheme.of(context).copyWith(
                    trackShape: const _TransparentTrackShape(),
                    thumbColor: VigorColors.indigo,
                    overlayColor: VigorColors.indigo.withValues(alpha: 0.2),
                  ),
                  child: Slider(
                    value: _duration.toDouble(),
                    min: minVal,
                    max: maxVal,
                    divisions: 34,
                    onChanged: (value) => setState(() => _duration = value.round()),
                  ),
                ),
              ],
            );
          },
        ),
      ],
    );
  }

  Widget _buildLoadingView() {
    final l10n = AppLocalizations.of(context);
    final statusText = _showingRetryMessage
        ? _getRetryMessage(l10n)
        : (_messagePool.isNotEmpty
            ? _messagePool[_currentMessageIndex % _messagePool.length]
            : l10n.thisMayTakeAMoment);

    return SizedBox(
      width: double.infinity,
      child: Padding(
        padding: VigorSpacing.paddingLg,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            AdaptiveLoadingIndicator(color: VigorColors.persimmon),
            const SizedBox(height: VigorSpacing.md),
            SizedBox(
              height: 24,
              child: ClipRect(
                child: AnimatedSwitcher(
                  duration: VigorAnimation.medium,
                  switchInCurve: VigorAnimation.entranceCurve,
                  switchOutCurve: VigorAnimation.exitCurve,
                  transitionBuilder: (child, animation) {
                    final isIncoming = child.key == ValueKey(statusText);
                    final slideOffset = Tween<Offset>(
                      begin: Offset(isIncoming ? 0.3 : 0.0, 0.0),
                      end: Offset(isIncoming ? 0.0 : -0.3, 0.0),
                    ).animate(animation);
                    return SlideTransition(
                      position: slideOffset,
                      child: FadeTransition(opacity: animation, child: child),
                    );
                  },
                  layoutBuilder: (currentChild, previousChildren) {
                    return Stack(
                      alignment: Alignment.center,
                      children: [
                        ...previousChildren,
                        if (currentChild != null) currentChild,
                      ],
                    );
                  },
                  child: Text(
                    statusText,
                    key: ValueKey(statusText),
                    style: VigorTypography.body.copyWith(
                      color: _showingRetryMessage
                          ? VigorColors.persimmon
                          : VigorColors.textSecondary(context),
                    ),
                    textAlign: TextAlign.center,
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFormView() {
    final l10n = AppLocalizations.of(context);
    return Form(
      key: _formKey,
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Flexible(
            child: SingleChildScrollView(
              padding: VigorSpacing.paddingLg.copyWith(bottom: 0),
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

                  // Duration slider with recommended range
                  _buildDurationSlider(l10n),
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
                ],
              ),
            ),
          ),
          // pinned action buttons
          Padding(
            padding: VigorSpacing.paddingLg.copyWith(top: VigorSpacing.md),
            child: Row(
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
          ),
        ],
      ),
    );
  }
}

/// Transparent track shape so only the thumb is rendered
class _TransparentTrackShape extends RoundedRectSliderTrackShape {
  const _TransparentTrackShape();

  @override
  void paint(
    PaintingContext context,
    Offset offset, {
    required RenderBox parentBox,
    required SliderThemeData sliderTheme,
    required Animation<double> enableAnimation,
    required TextDirection textDirection,
    required Offset thumbCenter,
    Offset? secondaryOffset,
    bool isDiscrete = false,
    bool isEnabled = false,
    double additionalActiveTrackHeight = 0,
  }) {
    // don't paint anything - track is invisible
  }
}

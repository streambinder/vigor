import 'dart:ui' show PlatformDispatcher;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/auth_provider.dart';
import '../providers/locale_provider.dart';
import '../models/profile.dart';
import '../models/injury.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/equipment_selector.dart';
import '../widgets/goal_selector.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

/// Full-page profile edit screen with sectioned layout
/// When missingFields is non-empty, acts as a blocking completion screen
class ProfileEditScreen extends StatefulWidget {
  final Profile profile;
  final Map<String, String> missingFields;

  const ProfileEditScreen({
    super.key,
    required this.profile,
    this.missingFields = const {},
  });

  /// Whether this screen is in completion mode (blocking, no back button)
  bool get isCompletionMode => missingFields.isNotEmpty;

  @override
  State<ProfileEditScreen> createState() => _ProfileEditScreenState();
}

class _ProfileEditScreenState extends State<ProfileEditScreen> {
  final _formKey = GlobalKey<FormState>();
  bool _isSubmitting = false;

  // form state
  String? _firstName;
  String? _lastName;
  DateTime? _birthdate;
  String? _gender;
  String? _language;
  double _height = 170;
  double _weight = 70;
  List<String> _goals = [];
  final List<Injury> _injuries = [];
  final List<String> _limitations = [];
  final List<String> _favoriteExercises = [];
  List<String> _favoriteEquipment = [];

  // controllers for text inputs
  final _injuryDescriptionController = TextEditingController();
  final _injuryYearController = TextEditingController();
  final _limitationController = TextEditingController();
  final _favoriteExerciseController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _loadExistingData();
  }

  void _loadExistingData() {
    final p = widget.profile;
    _firstName = p.firstName.isNotEmpty ? p.firstName : null;
    _lastName = p.lastName.isNotEmpty ? p.lastName : null;

    final now = DateTime.now();
    if (p.birthdate.year >= 1900 && !p.birthdate.isAfter(now)) {
      _birthdate = p.birthdate;
    }

    _gender = p.gender.isNotEmpty ? p.gender : null;
    _language = p.language.isNotEmpty ? p.language : _getSystemLanguage();
    _height = p.height > 0 ? p.height : 170;
    _weight = p.weight > 0 ? p.weight : 70;

    try {
      if (p.data.isNotEmpty) {
        final data = p.data;
        if (data['goals'] != null) {
          _goals = (data['goals'] as List).cast<String>();
        }
        if (data['injuries'] != null) {
          _injuries.addAll((data['injuries'] as List).map((i) => Injury.fromJson(i)));
        }
        if (data['limitations'] != null) {
          _limitations.addAll((data['limitations'] as List).cast<String>());
        }
        if (data['preferences'] != null) {
          final prefs = data['preferences'] as Map<String, dynamic>;
          if (prefs['exercises'] != null) {
            _favoriteExercises.addAll((prefs['exercises'] as List).cast<String>());
          }
          if (prefs['equipment'] != null) {
            _favoriteEquipment.addAll((prefs['equipment'] as List).cast<String>());
          }
        }
      }
    } catch (_) {}
  }

  @override
  void dispose() {
    _injuryDescriptionController.dispose();
    _injuryYearController.dispose();
    _limitationController.dispose();
    _favoriteExerciseController.dispose();
    super.dispose();
  }

  String? _getSystemLanguage() {
    final locale = PlatformDispatcher.instance.locale;
    const localeMap = {
      'en': 'english',
      'it': 'italiano',
      'es': 'español',
      'fr': 'français',
      'de': 'deutsch',
      'pt': 'português',
      'ru': 'русский',
      'zh': '中文',
      'ja': '日本語',
      'ko': '한국어',
    };
    return localeMap[locale.languageCode];
  }

  Future<void> _submitProfile() async {
    if (!_formKey.currentState!.validate()) return;

    final l10n = AppLocalizations.of(context);

    if (_birthdate == null) {
      AdaptiveNotification.showError(context: context, message: l10n.pleaseSelectBirthDate);
      return;
    }
    if (_goals.isEmpty) {
      AdaptiveNotification.showError(context: context, message: l10n.pleaseAddAtLeastOneGoal);
      return;
    }

    setState(() => _isSubmitting = true);

    try {
      final Map<String, dynamic> data = {
        'goals': _goals,
        'injuries': _injuries.map((i) => i.toJson()).toList(),
        'limitations': _limitations,
      };

      if (_favoriteExercises.isNotEmpty || _favoriteEquipment.isNotEmpty) {
        data['preferences'] = {
          if (_favoriteExercises.isNotEmpty) 'exercises': _favoriteExercises,
          if (_favoriteEquipment.isNotEmpty) 'equipment': _favoriteEquipment,
        };
      }

      final authProvider = context.read<AuthProvider>();
      final success = await authProvider.updateProfile(
        firstName: _firstName,
        lastName: _lastName,
        birthdate: _birthdate,
        gender: _gender,
        language: _language,
        height: _height,
        weight: _weight,
        data: data,
      );

      if (success && mounted) {
        context.read<LocaleProvider>().setFromProfileLanguage(_language);
        Navigator.of(context).pop(true);
      } else if (mounted) {
        AdaptiveNotification.showError(
          context: context,
          message: l10n.failedToUpdateProfile,
          rawError: authProvider.errorMessage,
        );
      }
    } finally {
      if (mounted) setState(() => _isSubmitting = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isCompletion = widget.isCompletionMode;

    return PopScope(
      canPop: !isCompletion,
      child: AdaptiveScaffold(
        appBar: AdaptiveAppBar(
          title: Text(isCompletion ? l10n.completeYourProfile : l10n.editProfile),
          automaticallyImplyLeading: !isCompletion,
          actions: [
            AdaptiveIconButton(
              icon: _isSubmitting
                  ? const SizedBox(width: 20, height: 20, child: AdaptiveLoadingIndicator())
                  : const Icon(Icons.check),
              onPressed: _isSubmitting ? null : _submitProfile,
              tooltip: l10n.saveChanges,
            ),
          ],
        ),
        body: Form(
          key: _formKey,
          child: ListView(
            padding: VigorSpacing.paddingLg,
            children: [
              if (isCompletion) ...[
                Text(
                  l10n.pleaseCompleteProfile,
                  style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context)),
                ),
                const SizedBox(height: VigorSpacing.lg),
              ],
              _buildSection(
                title: l10n.personalDetails,
                icon: Icons.person,
                children: [_buildPersonalDetailsContent(l10n)],
              ),
              const SizedBox(height: VigorSpacing.xl),
              _buildSection(
                title: l10n.healthAndGoals,
                icon: Icons.fitness_center,
                children: [_buildHealthGoalsContent(l10n)],
              ),
              const SizedBox(height: VigorSpacing.xl),
              _buildSection(
                title: l10n.favorites,
                icon: Icons.favorite,
                children: [_buildPreferencesContent(l10n)],
              ),
              const SizedBox(height: VigorSpacing.lg),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSection({
    required String title,
    required IconData icon,
    required List<Widget> children,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Container(
      decoration: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.glassDecoration(borderRadius: VigorRadius.lg, opacity: 0.9, isDark: isDark)
          : BoxDecoration(
              color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
              borderRadius: VigorRadius.radiusLg,
            ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.all(VigorSpacing.lg),
            child: Row(
              children: [
                Icon(icon, color: VigorColors.stone, size: 20),
                const SizedBox(width: VigorSpacing.md),
                Text(
                  title,
                  style: VigorTypography.headline.copyWith(
                    fontSize: 18,
                    color: VigorColors.textPrimary(context),
                  ),
                ),
              ],
            ),
          ),
          Divider(height: 1, color: VigorColors.border(context)),
          Padding(
            padding: const EdgeInsets.all(VigorSpacing.lg),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: children,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildPersonalDetailsContent(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // first name / last name
        Row(
          children: [
            Expanded(
              child: _buildTextField(
                label: l10n.firstName,
                value: _firstName,
                onChanged: (v) => _firstName = v.isNotEmpty ? v : null,
                required: true,
              ),
            ),
            const SizedBox(width: VigorSpacing.md),
            Expanded(
              child: _buildTextField(
                label: l10n.lastName,
                value: _lastName,
                onChanged: (v) => _lastName = v.isNotEmpty ? v : null,
                required: true,
              ),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.lg),

        // birthdate / gender
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(child: _buildBirthdatePicker(l10n)),
            const SizedBox(width: VigorSpacing.md),
            Expanded(child: _buildGenderDropdown(l10n)),
          ],
        ),
        const SizedBox(height: VigorSpacing.lg),

        // language
        _buildLanguageDropdown(l10n),
        const SizedBox(height: VigorSpacing.xl),

        // height slider
        _buildSliderField(
          label: l10n.heightCm,
          value: _height,
          min: 100,
          max: 230,
          divisions: 130,
          unit: l10n.heightUnit,
          onChanged: (v) => setState(() => _height = v),
        ),
        const SizedBox(height: VigorSpacing.lg),

        // weight slider
        _buildSliderField(
          label: l10n.weightKg,
          value: _weight,
          min: 30,
          max: 200,
          divisions: 170,
          unit: l10n.weightUnit,
          onChanged: (v) => setState(() => _weight = v),
        ),
      ],
    );
  }

  Widget _buildHealthGoalsContent(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // goals
        _buildFieldLabel(l10n.goals, required: true),
        const SizedBox(height: VigorSpacing.sm),
        GoalSelector(
          selected: _goals,
          onChanged: (updated) => setState(() => _goals = updated),
        ),
        const SizedBox(height: VigorSpacing.xl),

        // injuries
        _buildListInputSection(
          label: l10n.injuries,
          hint: l10n.optionalLeaveEmpty,
          items: _injuries.map((i) => '${i.description} (${i.year})').toList(),
          onRemove: (idx) => setState(() => _injuries.removeAt(idx)),
          inputBuilder: _buildInjuryInput,
        ),
        const SizedBox(height: VigorSpacing.xl),

        // limitations
        _buildListInputSection(
          label: l10n.limitations,
          hint: l10n.optionalLeaveEmpty,
          items: _limitations,
          onRemove: (idx) => setState(() => _limitations.removeAt(idx)),
          inputBuilder: () => _buildSimpleListInput(
            controller: _limitationController,
            placeholder: l10n.addALimitation,
            onAdd: () {
              if (_limitationController.text.isNotEmpty) {
                setState(() {
                  _limitations.add(_limitationController.text);
                  _limitationController.clear();
                });
              }
            },
          ),
        ),
      ],
    );
  }

  Widget _buildPreferencesContent(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // favorite exercises
        _buildListInputSection(
          label: l10n.favoriteExercises,
          hint: l10n.optionalExercisesPrefer,
          items: _favoriteExercises,
          onRemove: (idx) => setState(() => _favoriteExercises.removeAt(idx)),
          inputBuilder: () => _buildSimpleListInput(
            controller: _favoriteExerciseController,
            placeholder: l10n.favoriteExercisesHint,
            onAdd: () {
              if (_favoriteExerciseController.text.isNotEmpty) {
                setState(() {
                  _favoriteExercises.add(_favoriteExerciseController.text);
                  _favoriteExerciseController.clear();
                });
              }
            },
          ),
        ),
        const SizedBox(height: VigorSpacing.xl),

        // favorite equipment
        _buildFieldLabel(l10n.favoriteEquipment, hint: l10n.optionalEquipmentPrefer),
        const SizedBox(height: VigorSpacing.sm),
        EquipmentSelector(
          selected: _favoriteEquipment,
          onChanged: (updated) => setState(() => _favoriteEquipment = updated),
        ),
      ],
    );
  }

  Widget _buildTextField({
    required String label,
    String? value,
    required ValueChanged<String> onChanged,
    bool required = false,
    TextInputType? keyboardType,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFieldLabel(label, required: required),
        const SizedBox(height: VigorSpacing.sm),
        TextFormField(
          initialValue: value,
          decoration: const InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.md),
          ),
          style: VigorTypography.body,
          keyboardType: keyboardType,
          validator: required
              ? (v) {
                  if (v == null || v.isEmpty) return AppLocalizations.of(context).required;
                  return null;
                }
              : null,
          onChanged: onChanged,
        ),
      ],
    );
  }

  Widget _buildFieldLabel(String label, {bool required = false, String? hint}) {
    return Row(
      children: [
        Text(
          required ? '$label *' : label,
          style: VigorTypography.label.copyWith(
            color: VigorColors.textPrimary(context),
            fontWeight: FontWeight.w600,
          ),
        ),
        if (hint != null) ...[
          const SizedBox(width: VigorSpacing.sm),
          Text(
            hint,
            style: VigorTypography.caption.copyWith(color: VigorColors.textMuted(context)),
          ),
        ],
      ],
    );
  }

  Widget _buildBirthdatePicker(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFieldLabel(l10n.birthDate, required: true),
        const SizedBox(height: VigorSpacing.sm),
        InkWell(
          onTap: () async {
            final date = await showDatePicker(
              context: context,
              initialDate: _birthdate ?? DateTime(2000, 1, 1),
              firstDate: DateTime(1900),
              lastDate: DateTime.now(),
            );
            if (date != null) setState(() => _birthdate = date);
          },
          borderRadius: VigorRadius.radiusSm,
          child: InputDecorator(
            decoration: InputDecoration(
              border: const OutlineInputBorder(),
              contentPadding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.md),
              suffixIcon: Icon(Icons.calendar_today, size: 18, color: VigorColors.stone),
            ),
            child: Text(
              _birthdate != null
                  ? '${_birthdate!.day}/${_birthdate!.month}/${_birthdate!.year}'
                  : '',
              style: _birthdate != null
                  ? VigorTypography.data.copyWith(color: VigorColors.textPrimary(context))
                  : VigorTypography.body.copyWith(color: VigorColors.textMuted(context)),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildGenderDropdown(AppLocalizations l10n) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFieldLabel(l10n.gender, required: true),
        const SizedBox(height: VigorSpacing.sm),
        DropdownButtonFormField<String>(
          initialValue: _gender,
          decoration: const InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.md),
          ),
          style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)),
          items: [
            DropdownMenuItem(value: 'male', child: Text(l10n.male, style: VigorTypography.body)),
            DropdownMenuItem(value: 'female', child: Text(l10n.female, style: VigorTypography.body)),
          ],
          validator: (v) {
            if (v == null || v.isEmpty) return l10n.required;
            return null;
          },
          onChanged: (v) => setState(() => _gender = v),
        ),
      ],
    );
  }

  Widget _buildLanguageDropdown(AppLocalizations l10n) {
    final languages = [
      ('english', l10n.languageEnglish),
      ('italiano', l10n.languageItaliano),
      ('español', l10n.languageEspanol),
      ('français', l10n.languageFrancais),
      ('deutsch', l10n.languageDeutsch),
      ('português', l10n.languagePortugues),
      ('русский', l10n.languageRussian),
      ('中文', l10n.languageChinese),
      ('日本語', l10n.languageJapanese),
      ('한국어', l10n.languageKorean),
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFieldLabel(l10n.language, required: true),
        const SizedBox(height: VigorSpacing.sm),
        DropdownButtonFormField<String>(
          initialValue: _language,
          decoration: const InputDecoration(
            border: OutlineInputBorder(),
            contentPadding: EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.md),
          ),
          style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)),
          items: languages
              .map((l) => DropdownMenuItem(value: l.$1, child: Text(l.$2, style: VigorTypography.body)))
              .toList(),
          validator: (v) {
            if (v == null || v.isEmpty) return l10n.pleaseSelectLanguage;
            return null;
          },
          onChanged: (v) => setState(() => _language = v),
        ),
      ],
    );
  }

  Widget _buildSliderField({
    required String label,
    required double value,
    required double min,
    required double max,
    required int divisions,
    required String unit,
    required ValueChanged<double> onChanged,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              label,
              style: VigorTypography.label.copyWith(
                color: VigorColors.textPrimary(context),
                fontWeight: FontWeight.w600,
              ),
            ),
            Text(
              '${value.round()} $unit',
              style: VigorTypography.data.copyWith(color: VigorColors.textPrimary(context)),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.sm),
        SliderTheme(
          data: SliderTheme.of(context).copyWith(
            activeTrackColor: VigorColors.stone,
            inactiveTrackColor: VigorColors.stone.withValues(alpha: 0.2),
            thumbColor: VigorColors.stone,
            overlayColor: VigorColors.stone.withValues(alpha: 0.2),
            trackHeight: 4,
          ),
          child: Slider(
            value: value,
            min: min,
            max: max,
            divisions: divisions,
            onChanged: onChanged,
          ),
        ),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(
              '${min.round()}',
              style: VigorTypography.caption.copyWith(color: VigorColors.textMuted(context)),
            ),
            Text(
              '${max.round()}',
              style: VigorTypography.caption.copyWith(color: VigorColors.textMuted(context)),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildListInputSection({
    required String label,
    String? hint,
    required List<String> items,
    required void Function(int) onRemove,
    required Widget Function() inputBuilder,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        _buildFieldLabel(label, hint: hint),
        const SizedBox(height: VigorSpacing.sm),
        // existing items as chips
        if (items.isNotEmpty)
          Wrap(
            spacing: VigorSpacing.sm,
            runSpacing: VigorSpacing.sm,
            children: items.asMap().entries.map((entry) {
              return Chip(
                label: Text(
                  entry.value,
                  style: VigorTypography.caption.copyWith(color: VigorColors.textPrimary(context)),
                ),
                deleteIcon: Icon(Icons.close, size: 16, color: VigorColors.stone),
                onDeleted: () => onRemove(entry.key),
                backgroundColor: VigorColors.surfaceElevated(context),
                side: BorderSide.none,
              );
            }).toList(),
          ),
        const SizedBox(height: VigorSpacing.md),
        inputBuilder(),
      ],
    );
  }

  Widget _buildInjuryInput() {
    final l10n = AppLocalizations.of(context);

    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          flex: 3,
          child: TextField(
            controller: _injuryDescriptionController,
            decoration: InputDecoration(
              hintText: l10n.injuryDescription,
              hintStyle: VigorTypography.body.copyWith(color: VigorColors.textMuted(context)),
              border: const OutlineInputBorder(),
              contentPadding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.md),
            ),
            style: VigorTypography.body,
          ),
        ),
        const SizedBox(width: VigorSpacing.sm),
        Expanded(
          flex: 1,
          child: TextField(
            controller: _injuryYearController,
            decoration: InputDecoration(
              hintText: l10n.year,
              hintStyle: VigorTypography.body.copyWith(color: VigorColors.textMuted(context)),
              border: const OutlineInputBorder(),
              contentPadding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.md),
            ),
            style: VigorTypography.data,
            keyboardType: TextInputType.number,
          ),
        ),
        const SizedBox(width: VigorSpacing.sm),
        SizedBox(
          height: 48,
          child: AdaptiveButton(
            onPressed: () {
              final desc = _injuryDescriptionController.text;
              final year = int.tryParse(_injuryYearController.text);
              if (desc.isNotEmpty && year != null) {
                setState(() {
                  _injuries.add(Injury(description: desc, year: year));
                  _injuryDescriptionController.clear();
                  _injuryYearController.clear();
                });
              }
            },
            child: Text(l10n.add),
          ),
        ),
      ],
    );
  }

  Widget _buildSimpleListInput({
    required TextEditingController controller,
    required String placeholder,
    required VoidCallback onAdd,
  }) {
    final l10n = AppLocalizations.of(context);

    return Row(
      children: [
        Expanded(
          child: TextField(
            controller: controller,
            decoration: InputDecoration(
              hintText: placeholder,
              hintStyle: VigorTypography.body.copyWith(color: VigorColors.textMuted(context)),
              border: const OutlineInputBorder(),
              contentPadding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.md),
            ),
            style: VigorTypography.body,
            onSubmitted: (_) => onAdd(),
          ),
        ),
        const SizedBox(width: VigorSpacing.sm),
        SizedBox(
          height: 48,
          child: AdaptiveButton(
            onPressed: onAdd,
            child: Text(l10n.add),
          ),
        ),
      ],
    );
  }
}

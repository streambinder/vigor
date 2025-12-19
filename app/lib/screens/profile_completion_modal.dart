import 'dart:ui' show PlatformDispatcher;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../generated/app_localizations.dart';
import '../providers/auth_provider.dart';
import '../providers/locale_provider.dart';
import '../models/profile.dart';
import '../models/goal.dart';
import '../models/injury.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

/// Modal for completing missing profile fields
/// This modal cannot be dismissed and blocks user from proceeding until complete
class ProfileCompletionModal extends StatefulWidget {
  final Profile profile;
  final Map<String, String> missingFields;

  const ProfileCompletionModal({
    super.key,
    required this.profile,
    required this.missingFields,
  });

  @override
  State<ProfileCompletionModal> createState() => _ProfileCompletionModalState();
}

class _ProfileCompletionModalState extends State<ProfileCompletionModal> {
  final _formKey = GlobalKey<FormState>();
  bool _isSubmitting = false;

  // Form field controllers
  String? _first_name;
  String? _last_name;
  DateTime? _birthdate;
  String? _gender;
  String? _language;
  double? _height;
  double? _weight;
  final List<Goal> _goals = [];
  final List<Injury> _injuries = [];
  final List<String> _limitations = [];
  final List<String> _favoriteExercises = [];
  final List<String> _favoriteEquipment = [];
  final List<String> _favoriteWorkoutTypes = [];

  // Available workout types (must match backend enum)
  static const _workoutTypes = [
    'strength',
    'circuit',
    'emom',
    'amrap',
    'hiit',
    'for_time',
    'endurance',
    'mobility',
  ];

  // Controllers for dynamic list inputs
  final _goalDescriptionController = TextEditingController();
  final _injuryDescriptionController = TextEditingController();
  int? _injuryYear;
  final _limitationController = TextEditingController();
  final _favoriteExerciseController = TextEditingController();
  final _favoriteEquipmentController = TextEditingController();

  @override
  void initState() {
    super.initState();
    // Pre-fill with existing values if they exist and are valid
    _first_name = widget.profile.firstName.isNotEmpty ? widget.profile.firstName : null;
    _last_name = widget.profile.lastName.isNotEmpty ? widget.profile.lastName : null;
    final now = DateTime.now();
    final birthdate = widget.profile.birthdate;
    // Only use birthdate if it's valid (between 1900 and today)
    if (birthdate.year >= 1900 && !birthdate.isAfter(now)) {
      _birthdate = birthdate;
    } else {
      _birthdate = null; // Will use default in date picker
    }

    _gender = widget.profile.gender.isNotEmpty ? widget.profile.gender : null;
    _language = widget.profile.language.isNotEmpty ? widget.profile.language : _getSystemLanguage();
    _height = widget.profile.height > 0 ? widget.profile.height : null;
    _weight = widget.profile.weight > 0 ? widget.profile.weight : null;

    // Try to load existing data
    try {
      if (widget.profile.data.isNotEmpty) {
        final data = widget.profile.data;
        if (data['goals'] != null) {
          _goals.addAll((data['goals'] as List).map((g) => Goal.fromJson(g)));
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
          if (prefs['workout_types'] != null) {
            _favoriteWorkoutTypes.addAll((prefs['workout_types'] as List).cast<String>());
          }
        }
      }
    } catch (e) {
      // Ignore parsing errors, start fresh
    }
  }

  @override
  void dispose() {
    _goalDescriptionController.dispose();
    _injuryDescriptionController.dispose();
    _limitationController.dispose();
    _favoriteExerciseController.dispose();
    _favoriteEquipmentController.dispose();
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
    if (!_formKey.currentState!.validate()) {
      return;
    }

    // Validate required fields that aren't in the form
    if (widget.missingFields.containsKey('birthdate') && _birthdate == null) {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).pleaseSelectBirthDate,
      );
      return;
    }

    // Ensure required list fields have at least one entry
    if (widget.missingFields.containsKey('goals') && _goals.isEmpty) {
      AdaptiveNotification.showError(
        context: context,
        message: AppLocalizations.of(context).pleaseAddAtLeastOneGoal,
      );
      return;
    }

    setState(() {
      _isSubmitting = true;
    });

    try {
      // Build profile data
      final Map<String, dynamic> data = {
        'goals': _goals.map((g) => g.toJson()).toList(),
        'injuries': _injuries.map((i) => i.toJson()).toList(),
        'limitations': _limitations,
      };

      // Only include preferences if user has set any
      if (_favoriteExercises.isNotEmpty || _favoriteEquipment.isNotEmpty || _favoriteWorkoutTypes.isNotEmpty) {
        data['preferences'] = {
          if (_favoriteExercises.isNotEmpty) 'exercises': _favoriteExercises,
          if (_favoriteEquipment.isNotEmpty) 'equipment': _favoriteEquipment,
          if (_favoriteWorkoutTypes.isNotEmpty) 'workout_types': _favoriteWorkoutTypes,
        };
      }

      final authProvider = context.read<AuthProvider>();
      final success = await authProvider.updateProfile(
        firstName: _first_name,
        lastName: _last_name,
        birthdate: _birthdate,
        gender: _gender,
        language: _language,
        height: _height,
        weight: _weight,
        data: data,
      );

      if (success && mounted) {
        // update app locale to match profile language
        context.read<LocaleProvider>().setFromProfileLanguage(_language);
        // Close modal on success
        Navigator.of(context).pop();
      } else if (mounted) {
        AdaptiveNotification.showError(
          context: context,
          message: AppLocalizations.of(context).failedToUpdateProfile,
          rawError: authProvider.errorMessage,
        );
      }
    } finally {
      if (mounted) {
        setState(() {
          _isSubmitting = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return PopScope(
      canPop: widget.missingFields.isEmpty, // Allow dismissal when editing
      child: Dialog(
        backgroundColor: Colors.transparent,
        child: Container(
          decoration: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.glassDecoration(
                  borderRadius: 20,
                  opacity: 0.95,
                )
              : BoxDecoration(
                  color: Theme.of(context).colorScheme.surface,
                  borderRadius: BorderRadius.circular(20),
                ),
          child: SingleChildScrollView(
            child: Padding(
              padding: const EdgeInsets.all(16.0),
              child: Form(
                key: _formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Header
                    Text(
                      widget.missingFields.isEmpty ? l10n.editProfile : l10n.completeYourProfile,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.titleStyle
                          : const TextStyle(
                              fontSize: 24,
                              fontWeight: FontWeight.bold,
                            ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      widget.missingFields.isEmpty
                          ? l10n.updateYourProfileInfo
                          : l10n.pleaseCompleteProfile,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.captionStyle
                          : const TextStyle(color: Colors.grey),
                    ),
                    const SizedBox(height: 24),

                    // First Name and Last Name (side by side)
                    _buildNameFields(),

                    // Birthdate and Gender (side by side)
                    _buildBirthdateGenderFields(),

                    // Height and Weight (side by side)
                    _buildHeightWeightFields(),

                    // Language
                    _buildLanguageField(),

                    // Goals
                    _buildGoalsSection(),

                    // Injuries
                    _buildInjuriesSection(),

                    // Limitations
                    _buildLimitationsSection(),

                    // Favorite Exercises
                    _buildFavoriteExercisesSection(),

                    // Favorite Equipment
                    _buildFavoriteEquipmentSection(),

                    // Favorite Workout Types
                    _buildFavoriteWorkoutTypesSection(),

                    const SizedBox(height: 24),

                    // Submit button
                    AdaptiveButton(
                      onPressed: _isSubmitting ? null : _submitProfile,
                      child: _isSubmitting
                          ? const SizedBox(
                              height: 20,
                              width: 20,
                              child: AdaptiveLoadingIndicator(),
                            )
                          : Text(widget.missingFields.isEmpty ? l10n.saveChanges : l10n.saveProfile),
                    ),

                    // Cancel button (only when editing)
                    if (widget.missingFields.isEmpty) ...[
                      const SizedBox(height: 12),
                      TextButton(
                        onPressed: _isSubmitting ? null : () => Navigator.of(context).pop(),
                        child: Text(l10n.cancel),
                      ),
                    ],
                  ],
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildNameFields() {
    final l10n = AppLocalizations.of(context);
    final isFirstNameRequired = widget.missingFields.containsKey('first_name');
    final isLastNameRequired = widget.missingFields.containsKey('last_name');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        children: [
          Expanded(
            child: TextFormField(
              initialValue: _first_name,
              decoration: InputDecoration(
                labelText: isFirstNameRequired ? '${l10n.firstName} *' : l10n.firstName,
                border: const OutlineInputBorder(),
              ),
              validator: isFirstNameRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.required;
                      }
                      return null;
                    }
                  : null,
              onChanged: (value) {
                _first_name = value.isNotEmpty ? value : null;
              },
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: TextFormField(
              initialValue: _last_name,
              decoration: InputDecoration(
                labelText: isLastNameRequired ? '${l10n.lastName} *' : l10n.lastName,
                border: const OutlineInputBorder(),
              ),
              validator: isLastNameRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.required;
                      }
                      return null;
                    }
                  : null,
              onChanged: (value) {
                _last_name = value.isNotEmpty ? value : null;
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildBirthdateGenderFields() {
    final l10n = AppLocalizations.of(context);
    final isBirthdateRequired = widget.missingFields.containsKey('birthdate');
    final isGenderRequired = widget.missingFields.containsKey('gender');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: InkWell(
              onTap: () async {
                final date = await showDatePicker(
                  context: context,
                  initialDate: _birthdate ?? DateTime(2000, 1, 1),
                  firstDate: DateTime(1900),
                  lastDate: DateTime.now(),
                );
                if (date != null) {
                  setState(() {
                    _birthdate = date;
                  });
                }
              },
              child: InputDecorator(
                decoration: InputDecoration(
                  labelText: isBirthdateRequired ? '${l10n.birthDate} *' : l10n.birthDate,
                  border: const OutlineInputBorder(),
                  suffixIcon: const Icon(Icons.calendar_today),
                ),
                child: Text(
                  _birthdate != null
                      ? '${_birthdate!.day}/${_birthdate!.month}/${_birthdate!.year}'
                      : '',
                  style: TextStyle(
                    color: _birthdate != null
                        ? Theme.of(context).textTheme.bodyLarge?.color
                        : Theme.of(context).hintColor,
                  ),
                ),
              ),
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: DropdownButtonFormField<String>(
              value: _gender,
              decoration: InputDecoration(
                labelText: isGenderRequired ? '${l10n.gender} *' : l10n.gender,
                border: const OutlineInputBorder(),
              ),
              items: [
                DropdownMenuItem(value: 'male', child: Text(l10n.male)),
                DropdownMenuItem(value: 'female', child: Text(l10n.female)),
              ],
              validator: isGenderRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.required;
                      }
                      return null;
                    }
                  : null,
              onChanged: (value) {
                setState(() {
                  _gender = value;
                });
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLanguageField() {
    final l10n = AppLocalizations.of(context);
    final isRequired = widget.missingFields.containsKey('language');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? '${l10n.language} *' : l10n.language,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            value: _language,
            decoration: InputDecoration(
              border: const OutlineInputBorder(),
              hintText: l10n.selectLanguage,
            ),
            items: [
              DropdownMenuItem(value: 'english', child: Text(l10n.languageEnglish)),
              DropdownMenuItem(value: 'italiano', child: Text(l10n.languageItaliano)),
              DropdownMenuItem(value: 'español', child: Text(l10n.languageEspanol)),
              DropdownMenuItem(value: 'français', child: Text(l10n.languageFrancais)),
              DropdownMenuItem(value: 'deutsch', child: Text(l10n.languageDeutsch)),
              DropdownMenuItem(value: 'português', child: Text(l10n.languagePortugues)),
              DropdownMenuItem(value: 'русский', child: Text(l10n.languageRussian)),
              DropdownMenuItem(value: '中文', child: Text(l10n.languageChinese)),
              DropdownMenuItem(value: '日本語', child: Text(l10n.languageJapanese)),
              DropdownMenuItem(value: '한국어', child: Text(l10n.languageKorean)),
            ],
            validator: isRequired ? (value) {
              if (value == null || value.isEmpty) {
                return l10n.pleaseSelectLanguage;
              }
              return null;
            } : null,
            onChanged: (value) {
              setState(() {
                _language = value;
              });
            },
          ),
        ],
      ),
    );
  }

  Widget _buildHeightWeightFields() {
    final l10n = AppLocalizations.of(context);
    final isHeightRequired = widget.missingFields.containsKey('height');
    final isWeightRequired = widget.missingFields.containsKey('weight');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Row(
        children: [
          Expanded(
            child: TextFormField(
              initialValue: _height?.toString(),
              decoration: InputDecoration(
                labelText: isHeightRequired ? '${l10n.heightCm} *' : l10n.heightCm,
                border: const OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
              validator: isHeightRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.required;
                      }
                      final height = double.tryParse(value);
                      if (height == null || height <= 0) {
                        return l10n.invalid;
                      }
                      return null;
                    }
                  : null,
              onChanged: (value) {
                _height = double.tryParse(value);
              },
            ),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: TextFormField(
              initialValue: _weight?.toString(),
              decoration: InputDecoration(
                labelText: isWeightRequired ? '${l10n.weightKg} *' : l10n.weightKg,
                border: const OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
              validator: isWeightRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return l10n.required;
                      }
                      final weight = double.tryParse(value);
                      if (weight == null || weight <= 0) {
                        return l10n.invalid;
                      }
                      return null;
                    }
                  : null,
              onChanged: (value) {
                _weight = double.tryParse(value);
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildGoalsSection() {
    final l10n = AppLocalizations.of(context);
    final isRequired = widget.missingFields.containsKey('goals');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? '${l10n.goals} *' : l10n.goals,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          ..._goals.map((goal) => ListTile(
                title: Text(goal.description),
                subtitle: Text(l10n.startedDate(_formatDate(goal.startDate))),
                trailing: IconButton(
                  icon: const Icon(Icons.delete),
                  onPressed: () {
                    setState(() {
                      _goals.remove(goal);
                    });
                  },
                ),
              )),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _goalDescriptionController,
                  decoration: InputDecoration(
                    hintText: l10n.addAGoal,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.add),
                onPressed: () {
                  if (_goalDescriptionController.text.isNotEmpty) {
                    setState(() {
                      _goals.add(Goal(
                        description: _goalDescriptionController.text,
                        startDate: DateTime.now(),
                      ));
                      _goalDescriptionController.clear();
                    });
                  }
                },
              ),
            ],
          ),
        ],
      ),
    );
  }

  String _formatDate(DateTime date) {
    return '${date.day}/${date.month}/${date.year}';
  }

  Widget _buildInjuriesSection() {
    final l10n = AppLocalizations.of(context);
    final isRequired = widget.missingFields.containsKey('injuries');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? '${l10n.injuries} *' : l10n.injuries,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          if (!isRequired)
            Text(
              l10n.optionalLeaveEmpty,
              style: const TextStyle(fontSize: 12, color: Colors.grey),
            ),
          const SizedBox(height: 8),
          ..._injuries.map((injury) => ListTile(
                title: Text(injury.description),
                subtitle: Text(l10n.yearLabel(injury.year)),
                trailing: IconButton(
                  icon: const Icon(Icons.delete),
                  onPressed: () {
                    setState(() {
                      _injuries.remove(injury);
                    });
                  },
                ),
              )),
          Row(
            children: [
              Expanded(
                flex: 2,
                child: TextField(
                  controller: _injuryDescriptionController,
                  decoration: InputDecoration(
                    hintText: l10n.injuryDescription,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: TextField(
                  decoration: InputDecoration(
                    hintText: l10n.year,
                    border: const OutlineInputBorder(),
                  ),
                  keyboardType: TextInputType.number,
                  onChanged: (value) {
                    _injuryYear = int.tryParse(value);
                  },
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.add),
                onPressed: () {
                  if (_injuryDescriptionController.text.isNotEmpty && _injuryYear != null) {
                    setState(() {
                      _injuries.add(Injury(
                        description: _injuryDescriptionController.text,
                        year: _injuryYear!,
                      ));
                      _injuryDescriptionController.clear();
                      _injuryYear = null;
                    });
                  }
                },
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildLimitationsSection() {
    final l10n = AppLocalizations.of(context);
    final isRequired = widget.missingFields.containsKey('limitations');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? '${l10n.limitations} *' : l10n.limitations,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          if (!isRequired)
            Text(
              l10n.optionalLeaveEmpty,
              style: const TextStyle(fontSize: 12, color: Colors.grey),
            ),
          const SizedBox(height: 8),
          ..._limitations.map((limitation) => ListTile(
                title: Text(limitation),
                trailing: IconButton(
                  icon: const Icon(Icons.delete),
                  onPressed: () {
                    setState(() {
                      _limitations.remove(limitation);
                    });
                  },
                ),
              )),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _limitationController,
                  decoration: InputDecoration(
                    hintText: l10n.addALimitation,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.add),
                onPressed: () {
                  if (_limitationController.text.isNotEmpty) {
                    setState(() {
                      _limitations.add(_limitationController.text);
                      _limitationController.clear();
                    });
                  }
                },
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFavoriteExercisesSection() {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.favoriteExercises,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          Text(
            l10n.optionalExercisesPrefer,
            style: const TextStyle(fontSize: 12, color: Colors.grey),
          ),
          const SizedBox(height: 8),
          ..._favoriteExercises.map((exercise) => ListTile(
                title: Text(exercise),
                trailing: IconButton(
                  icon: const Icon(Icons.delete),
                  onPressed: () {
                    setState(() {
                      _favoriteExercises.remove(exercise);
                    });
                  },
                ),
              )),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _favoriteExerciseController,
                  decoration: InputDecoration(
                    hintText: l10n.favoriteExercisesHint,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.add),
                onPressed: () {
                  if (_favoriteExerciseController.text.isNotEmpty) {
                    setState(() {
                      _favoriteExercises.add(_favoriteExerciseController.text);
                      _favoriteExerciseController.clear();
                    });
                  }
                },
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFavoriteEquipmentSection() {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.favoriteEquipment,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          Text(
            l10n.optionalEquipmentPrefer,
            style: const TextStyle(fontSize: 12, color: Colors.grey),
          ),
          const SizedBox(height: 8),
          ..._favoriteEquipment.map((equipment) => ListTile(
                title: Text(equipment),
                trailing: IconButton(
                  icon: const Icon(Icons.delete),
                  onPressed: () {
                    setState(() {
                      _favoriteEquipment.remove(equipment);
                    });
                  },
                ),
              )),
          Row(
            children: [
              Expanded(
                child: TextField(
                  controller: _favoriteEquipmentController,
                  decoration: InputDecoration(
                    hintText: l10n.favoriteEquipmentHint,
                    border: const OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.add),
                onPressed: () {
                  if (_favoriteEquipmentController.text.isNotEmpty) {
                    setState(() {
                      _favoriteEquipment.add(_favoriteEquipmentController.text);
                      _favoriteEquipmentController.clear();
                    });
                  }
                },
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFavoriteWorkoutTypesSection() {
    final l10n = AppLocalizations.of(context);
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            l10n.favoriteWorkoutTypes,
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          Text(
            l10n.optionalWorkoutTypesPrefer,
            style: const TextStyle(fontSize: 12, color: Colors.grey),
          ),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: _workoutTypes.map((type) {
              final isSelected = _favoriteWorkoutTypes.contains(type);
              return FilterChip(
                label: Text(_workoutTypeLabel(type, l10n)),
                selected: isSelected,
                onSelected: (selected) {
                  setState(() {
                    if (selected) {
                      _favoriteWorkoutTypes.add(type);
                    } else {
                      _favoriteWorkoutTypes.remove(type);
                    }
                  });
                },
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  String _workoutTypeLabel(String type, AppLocalizations l10n) {
    switch (type) {
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
        return type;
    }
  }
}

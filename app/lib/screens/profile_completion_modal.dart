import 'dart:ui' show PlatformDispatcher;

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
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
        message: 'Please select your birth date',
      );
      return;
    }

    // Ensure required list fields have at least one entry
    if (widget.missingFields.containsKey('goals') && _goals.isEmpty) {
      AdaptiveNotification.showError(
        context: context,
        message: 'Please add at least one goal',
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
      if (_favoriteExercises.isNotEmpty || _favoriteEquipment.isNotEmpty) {
        data['preferences'] = {
          if (_favoriteExercises.isNotEmpty) 'exercises': _favoriteExercises,
          if (_favoriteEquipment.isNotEmpty) 'equipment': _favoriteEquipment,
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
        // Close modal on success
        Navigator.of(context).pop();
      } else if (mounted) {
        AdaptiveNotification.showError(
          context: context,
          message: 'Failed to update profile',
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
              padding: const EdgeInsets.all(24.0),
              child: Form(
                key: _formKey,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.stretch,
                  children: [
                    // Header
                    Text(
                      widget.missingFields.isEmpty ? 'Edit Profile' : 'Complete Your Profile',
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
                          ? 'Update your profile information below.'
                          : 'Please complete your profile. Fields marked with * are required.',
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
                          : Text(widget.missingFields.isEmpty ? 'Save Changes' : 'Save Profile'),
                    ),

                    // Cancel button (only when editing)
                    if (widget.missingFields.isEmpty) ...[
                      const SizedBox(height: 12),
                      TextButton(
                        onPressed: _isSubmitting ? null : () => Navigator.of(context).pop(),
                        child: const Text('Cancel'),
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
                labelText: isFirstNameRequired ? 'First Name *' : 'First Name',
                border: const OutlineInputBorder(),
              ),
              validator: isFirstNameRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return 'Required';
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
                labelText: isLastNameRequired ? 'Last Name *' : 'Last Name',
                border: const OutlineInputBorder(),
              ),
              validator: isLastNameRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return 'Required';
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
                  labelText: isBirthdateRequired ? 'Birth Date *' : 'Birth Date',
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
                labelText: isGenderRequired ? 'Gender *' : 'Gender',
                border: const OutlineInputBorder(),
              ),
              items: const [
                DropdownMenuItem(value: 'male', child: Text('Male')),
                DropdownMenuItem(value: 'female', child: Text('Female')),
              ],
              validator: isGenderRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return 'Required';
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
    final isRequired = widget.missingFields.containsKey('language');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? 'Language *' : 'Language',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            value: _language,
            decoration: const InputDecoration(
              border: OutlineInputBorder(),
              hintText: 'Select language',
            ),
            items: const [
              DropdownMenuItem(value: 'english', child: Text('English')),
              DropdownMenuItem(value: 'italiano', child: Text('Italiano')),
              DropdownMenuItem(value: 'español', child: Text('Español')),
              DropdownMenuItem(value: 'français', child: Text('Français')),
              DropdownMenuItem(value: 'deutsch', child: Text('Deutsch')),
              DropdownMenuItem(value: 'português', child: Text('Português')),
              DropdownMenuItem(value: 'русский', child: Text('Русский')),
              DropdownMenuItem(value: '中文', child: Text('中文')),
              DropdownMenuItem(value: '日本語', child: Text('日本語')),
              DropdownMenuItem(value: '한국어', child: Text('한국어')),
            ],
            validator: isRequired ? (value) {
              if (value == null || value.isEmpty) {
                return 'Please select your language';
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
                labelText: isHeightRequired ? 'Height (cm) *' : 'Height (cm)',
                border: const OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
              validator: isHeightRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return 'Required';
                      }
                      final height = double.tryParse(value);
                      if (height == null || height <= 0) {
                        return 'Invalid';
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
                labelText: isWeightRequired ? 'Weight (kg) *' : 'Weight (kg)',
                border: const OutlineInputBorder(),
              ),
              keyboardType: TextInputType.number,
              validator: isWeightRequired
                  ? (value) {
                      if (value == null || value.isEmpty) {
                        return 'Required';
                      }
                      final weight = double.tryParse(value);
                      if (weight == null || weight <= 0) {
                        return 'Invalid';
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
    final isRequired = widget.missingFields.containsKey('goals');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? 'Goals *' : 'Goals',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          ..._goals.map((goal) => ListTile(
                title: Text(goal.description),
                subtitle: Text('Started: ${goal.startDate.day}/${goal.startDate.month}/${goal.startDate.year}'),
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
                  decoration: const InputDecoration(
                    hintText: 'Add a goal',
                    border: OutlineInputBorder(),
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

  Widget _buildInjuriesSection() {
    final isRequired = widget.missingFields.containsKey('injuries');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? 'Injuries *' : 'Injuries',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          if (!isRequired)
            const Text(
              '(Optional - leave empty if none)',
              style: TextStyle(fontSize: 12, color: Colors.grey),
            ),
          const SizedBox(height: 8),
          ..._injuries.map((injury) => ListTile(
                title: Text(injury.description),
                subtitle: Text('Year: ${injury.year}'),
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
                  decoration: const InputDecoration(
                    hintText: 'Injury description',
                    border: OutlineInputBorder(),
                  ),
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: TextField(
                  decoration: const InputDecoration(
                    hintText: 'Year',
                    border: OutlineInputBorder(),
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
    final isRequired = widget.missingFields.containsKey('limitations');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? 'Limitations *' : 'Limitations',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          if (!isRequired)
            const Text(
              '(Optional - leave empty if none)',
              style: TextStyle(fontSize: 12, color: Colors.grey),
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
                  decoration: const InputDecoration(
                    hintText: 'Add a limitation',
                    border: OutlineInputBorder(),
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
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Favorite Exercises',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
          const Text(
            '(Optional - exercises you enjoy or prefer)',
            style: TextStyle(fontSize: 12, color: Colors.grey),
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
                  decoration: const InputDecoration(
                    hintText: 'e.g., squats, pull-ups, running',
                    border: OutlineInputBorder(),
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
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text(
            'Favorite Equipment',
            style: TextStyle(fontWeight: FontWeight.bold),
          ),
          const Text(
            '(Optional - equipment you prefer using)',
            style: TextStyle(fontSize: 12, color: Colors.grey),
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
                  decoration: const InputDecoration(
                    hintText: 'e.g., dumbbells, barbell, kettlebell',
                    border: OutlineInputBorder(),
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
}

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
  DateTime? _birthdate;
  String? _gender;
  String? _language;
  double? _height;
  double? _weight;
  final List<Goal> _goals = [];
  final List<Injury> _injuries = [];
  final List<String> _limitations = [];

  // Controllers for dynamic list inputs
  final _goalDescriptionController = TextEditingController();
  final _injuryDescriptionController = TextEditingController();
  int? _injuryYear;
  final _limitationController = TextEditingController();

  @override
  void initState() {
    super.initState();
    // Pre-fill with existing values if they exist and are valid
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
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please select your birth date')),
      );
      return;
    }

    // Ensure required list fields have at least one entry
    if (widget.missingFields.containsKey('goals') && _goals.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please add at least one goal')),
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

      final authProvider = context.read<AuthProvider>();
      final success = await authProvider.updateProfile(
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
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(authProvider.errorMessage ?? 'Failed to update profile'),
            backgroundColor: Colors.red,
          ),
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

                    // Birthdate
                    _buildDateField(),

                    // Gender
                    _buildGenderField(),

                    // Language
                    _buildLanguageField(),

                    // Height
                    _buildHeightField(),

                    // Weight
                    _buildWeightField(),

                    // Goals
                    _buildGoalsSection(),

                    // Injuries
                    _buildInjuriesSection(),

                    // Limitations
                    _buildLimitationsSection(),

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

  Widget _buildDateField() {
    final isRequired = widget.missingFields.containsKey('birthdate');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? 'Birth Date *' : 'Birth Date',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          InkWell(
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
              decoration: const InputDecoration(
                border: OutlineInputBorder(),
                suffixIcon: Icon(Icons.calendar_today),
              ),
              child: Text(
                _birthdate != null
                    ? '${_birthdate!.day}/${_birthdate!.month}/${_birthdate!.year}'
                    : 'Select date',
                style: TextStyle(
                  color: _birthdate != null
                      ? Theme.of(context).textTheme.bodyLarge?.color
                      : Theme.of(context).hintColor,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildGenderField() {
    final isRequired = widget.missingFields.containsKey('gender');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            isRequired ? 'Gender *' : 'Gender',
            style: const TextStyle(fontWeight: FontWeight.bold),
          ),
          const SizedBox(height: 8),
          DropdownButtonFormField<String>(
            value: _gender,
            decoration: const InputDecoration(
              border: OutlineInputBorder(),
              hintText: 'Select gender',
            ),
            items: const [
              DropdownMenuItem(value: 'male', child: Text('Male')),
              DropdownMenuItem(value: 'female', child: Text('Female')),
            ],
            validator: isRequired ? (value) {
              if (value == null || value.isEmpty) {
                return 'Please select your gender';
              }
              return null;
            } : null,
            onChanged: (value) {
              setState(() {
                _gender = value;
              });
            },
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

  Widget _buildHeightField() {
    final isRequired = widget.missingFields.containsKey('height');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: TextFormField(
        initialValue: _height?.toString(),
        decoration: InputDecoration(
          labelText: isRequired ? 'Height (cm) *' : 'Height (cm)',
          border: const OutlineInputBorder(),
        ),
        keyboardType: TextInputType.number,
        validator: isRequired ? (value) {
          if (value == null || value.isEmpty) {
            return 'Please enter your height';
          }
          final height = double.tryParse(value);
          if (height == null || height <= 0) {
            return 'Please enter a valid height';
          }
          return null;
        } : null,
        onChanged: (value) {
          _height = double.tryParse(value);
        },
      ),
    );
  }

  Widget _buildWeightField() {
    final isRequired = widget.missingFields.containsKey('weight');
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: TextFormField(
        initialValue: _weight?.toString(),
        decoration: InputDecoration(
          labelText: isRequired ? 'Weight (kg) *' : 'Weight (kg)',
          border: const OutlineInputBorder(),
        ),
        keyboardType: TextInputType.number,
        validator: isRequired ? (value) {
          if (value == null || value.isEmpty) {
            return 'Please enter your weight';
          }
          final weight = double.tryParse(value);
          if (weight == null || weight <= 0) {
            return 'Please enter a valid weight';
          }
          return null;
        } : null,
        onChanged: (value) {
          _weight = double.tryParse(value);
        },
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
}

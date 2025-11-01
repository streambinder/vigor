import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../providers/auth_provider.dart';
import '../models/profile.dart';
import '../models/goal.dart';
import '../models/injury.dart';

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
    // Pre-fill with existing values if they exist
    _birthdate = widget.profile.birthdate;
    _language = widget.profile.language.isNotEmpty ? widget.profile.language : null;
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

  Future<void> _submitProfile() async {
    if (!_formKey.currentState!.validate()) {
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
      canPop: false, // Prevent back button dismissal
      child: Dialog(
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
                  const Text(
                    'Complete Your Profile',
                    style: TextStyle(
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Please fill in the following required information to continue:',
                    style: TextStyle(color: Colors.grey),
                  ),
                  const SizedBox(height: 24),

                  // Birthdate
                  if (widget.missingFields.containsKey('birthdate'))
                    _buildDateField(),

                  // Language
                  if (widget.missingFields.containsKey('language'))
                    _buildLanguageField(),

                  // Height
                  if (widget.missingFields.containsKey('height'))
                    _buildHeightField(),

                  // Weight
                  if (widget.missingFields.containsKey('weight'))
                    _buildWeightField(),

                  // Goals
                  if (widget.missingFields.containsKey('goals'))
                    _buildGoalsSection(),

                  // Injuries
                  if (widget.missingFields.containsKey('injuries'))
                    _buildInjuriesSection(),

                  // Limitations
                  if (widget.missingFields.containsKey('limitations'))
                    _buildLimitationsSection(),

                  const SizedBox(height: 24),

                  // Submit button
                  ElevatedButton(
                    onPressed: _isSubmitting ? null : _submitProfile,
                    style: ElevatedButton.styleFrom(
                      padding: const EdgeInsets.all(16),
                    ),
                    child: _isSubmitting
                        ? const SizedBox(
                            height: 20,
                            width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('Save Profile'),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildDateField() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Birth Date *', style: TextStyle(fontWeight: FontWeight.bold)),
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
                  color: _birthdate != null ? Colors.black : Colors.grey,
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildLanguageField() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: TextFormField(
        initialValue: _language,
        decoration: const InputDecoration(
          labelText: 'Language *',
          hintText: 'e.g., en, it, es',
          border: OutlineInputBorder(),
        ),
        validator: (value) {
          if (value == null || value.isEmpty) {
            return 'Please enter your language (ISO 639-1 code)';
          }
          return null;
        },
        onChanged: (value) {
          _language = value;
        },
      ),
    );
  }

  Widget _buildHeightField() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: TextFormField(
        initialValue: _height?.toString(),
        decoration: const InputDecoration(
          labelText: 'Height (cm) *',
          border: OutlineInputBorder(),
        ),
        keyboardType: TextInputType.number,
        validator: (value) {
          if (value == null || value.isEmpty) {
            return 'Please enter your height';
          }
          final height = double.tryParse(value);
          if (height == null || height <= 0) {
            return 'Please enter a valid height';
          }
          return null;
        },
        onChanged: (value) {
          _height = double.tryParse(value);
        },
      ),
    );
  }

  Widget _buildWeightField() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: TextFormField(
        initialValue: _weight?.toString(),
        decoration: const InputDecoration(
          labelText: 'Weight (kg) *',
          border: OutlineInputBorder(),
        ),
        keyboardType: TextInputType.number,
        validator: (value) {
          if (value == null || value.isEmpty) {
            return 'Please enter your weight';
          }
          final weight = double.tryParse(value);
          if (weight == null || weight <= 0) {
            return 'Please enter a valid weight';
          }
          return null;
        },
        onChanged: (value) {
          _weight = double.tryParse(value);
        },
      ),
    );
  }

  Widget _buildGoalsSection() {
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Goals *', style: TextStyle(fontWeight: FontWeight.bold)),
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
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Injuries', style: TextStyle(fontWeight: FontWeight.bold)),
          const Text('(Optional - leave empty if none)', style: TextStyle(fontSize: 12, color: Colors.grey)),
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
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Text('Limitations', style: TextStyle(fontWeight: FontWeight.bold)),
          const Text('(Optional - leave empty if none)', style: TextStyle(fontSize: 12, color: Colors.grey)),
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

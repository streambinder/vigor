import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../models/gym.dart';
import '../models/training.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';
import '../services/preferences_service.dart';
import '../services/user_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/user_select_dialog.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

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
  final _equipmentInputController = TextEditingController();

  Gym? _selectedGym;
  bool _isGenerating = false;
  final List<UserInfo> _partners = [];
  final List<String> _equipment = [];

  @override
  void initState() {
    super.initState();
    if (widget.gyms.isNotEmpty) {
      final prefs = context.read<PreferencesService>();
      final defaultName = prefs.defaultGymName;
      _selectedGym = widget.gyms.firstWhere(
        (g) => g.name == defaultName,
        orElse: () => widget.gyms.first,
      );
    }
  }

  @override
  void dispose() {
    _durationController.dispose();
    _promptController.dispose();
    _equipmentInputController.dispose();
    super.dispose();
  }

  void _addEquipment() {
    final equipment = _equipmentInputController.text.trim();
    if (equipment.isNotEmpty && !_equipment.contains(equipment)) {
      setState(() {
        _equipment.add(equipment);
        _equipmentInputController.clear();
      });
    }
  }

  void _removeEquipment(String equipment) {
    setState(() {
      _equipment.remove(equipment);
    });
  }

  Future<void> _addPartner() async {
    final user = await showUserSelectDialog(
      context: context,
      title: 'Add Partner',
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
    final gym = _selectedGym?.name;
    final prompt = _promptController.text.trim();

    final response = await trainingService.generateTraining(
      duration: duration,
      gym: gym ?? '',
      prompt: prompt.isEmpty ? null : prompt,
      equipment: _equipment.isEmpty ? null : _equipment,
      partners: _partners.isEmpty ? null : _partners.map((p) => p.id).toList(),
    );

    if (mounted) {
      setState(() {
        _isGenerating = false;
      });

      if (response.isSuccess) {
        Navigator.of(context).pop();
        AdaptiveNotification.show(
          context: context,
          message: 'Training generated successfully!',
        );
        widget.onSuccess?.call(response.data!);
      } else {
        AdaptiveNotification.showError(
          context: context,
          message: 'Failed to generate training',
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
    return Padding(
      padding: const EdgeInsets.all(32.0),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const AdaptiveLoadingIndicator(),
          const SizedBox(height: 24),
          Text(
            'Generating your training...',
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.headlineStyle
                : Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 8),
          Text(
            'This may take a moment',
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
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24.0),
      child: Form(
        key: _formKey,
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            Text(
              'Generate Training',
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 24)
                  : Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 24),

            // Duration field
            AdaptiveTextField(
              controller: _durationController,
              labelText: 'Duration (minutes)',
              keyboardType: TextInputType.number,
              inputFormatters: [
                FilteringTextInputFormatter.digitsOnly,
              ],
            ),
            const SizedBox(height: 16),

            // Gym selection
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Gym (optional)',
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.captionStyle
                      : Theme.of(context).textTheme.labelMedium,
                ),
                const SizedBox(height: 8),
                Opacity(
                  opacity: widget.gyms.isEmpty ? 0.5 : 1.0,
                  child: Container(
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
                          widget.gyms.isEmpty ? 'No gyms available' : 'Select a gym',
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
                        onChanged: widget.gyms.isEmpty
                            ? null
                            : (Gym? value) {
                                setState(() {
                                  _selectedGym = value;
                                });
                              },
                      ),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 16),

            // Equipment section
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Equipment',
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.captionStyle
                      : Theme.of(context).textTheme.labelMedium,
                ),
                const SizedBox(height: 8),
                Row(
                  children: [
                    Expanded(
                      child: AdaptiveTextField(
                        controller: _equipmentInputController,
                        labelText: 'Add Equipment',
                        placeholder: 'e.g., Barbell, Dumbbells',
                        onSubmitted: (_) => _addEquipment(),
                      ),
                    ),
                    const SizedBox(width: 8),
                    AdaptiveButton(
                      onPressed: _addEquipment,
                      child: const Icon(Icons.add),
                    ),
                  ],
                ),
                if (_equipment.isEmpty)
                  Padding(
                    padding: const EdgeInsets.symmetric(vertical: 12),
                    child: Text(
                      _selectedGym != null
                          ? 'Using equipment from "${_selectedGym!.name}"'
                          : 'Bodyweight training (no equipment)',
                      style: TextStyle(
                        color: Colors.grey[600],
                        fontStyle: FontStyle.italic,
                        fontSize: 12,
                      ),
                    ),
                  )
                else ...[
                  const SizedBox(height: 12),
                  Wrap(
                    spacing: 8,
                    runSpacing: 8,
                    children: _equipment.map((equipment) {
                      return Chip(
                        label: Text(
                          equipment,
                          style: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.captionStyle.copyWith(fontSize: 12)
                              : null,
                        ),
                        deleteIcon: const Icon(Icons.close, size: 16),
                        onDeleted: () => _removeEquipment(equipment),
                        backgroundColor: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.primaryColor.withOpacity(0.1)
                            : null,
                      );
                    }).toList(),
                  ),
                ],
              ],
            ),
            const SizedBox(height: 16),

            // Optional prompt
            AdaptiveTextField(
              controller: _promptController,
              labelText: 'Custom Prompt (optional)',
              placeholder: 'e.g., Focus on upper body',
              maxLines: 3,
              minLines: 1,
            ),
            const SizedBox(height: 16),

            // Partners section
            Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  'Training Partners (optional)',
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.captionStyle
                      : Theme.of(context).textTheme.labelMedium,
                ),
                const SizedBox(height: 8),
                AdaptiveButton(
                  onPressed: _addPartner,
                  child: const Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Icon(Icons.person_add, size: 18),
                      SizedBox(width: 8),
                      Text('Add Partner'),
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
                  child: const Text('Cancel'),
                ),
                const SizedBox(width: 12),
                AdaptiveButton(
                  onPressed: _generateTraining,
                  child: const Text('Generate'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}

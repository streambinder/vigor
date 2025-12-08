import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import '../models/gym.dart';
import '../models/training.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';
import '../services/preferences_service.dart';
import '../widgets/adaptive/adaptive.dart';
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

  Gym? _selectedGym;
  bool _isGenerating = false;

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
    super.dispose();
  }

  Future<void> _generateTraining() async {
    if (!_formKey.currentState!.validate()) {
      return;
    }

    if (_selectedGym == null) {
      AdaptiveNotification.showError(
        context: context,
        message: 'Please select a gym',
      );
      return;
    }

    setState(() {
      _isGenerating = true;
    });

    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);

    final duration = int.parse(_durationController.text);
    final gym = _selectedGym!.name;
    final prompt = _promptController.text.trim();

    final response = await trainingService.generateTraining(
      duration: duration,
      gym: gym,
      prompt: prompt.isEmpty ? null : prompt,
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
          message: response.error ?? 'Failed to generate training',
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
            'Generating your workout...',
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
                  'Gym',
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.captionStyle
                      : Theme.of(context).textTheme.labelMedium,
                ),
                const SizedBox(height: 8),
                Container(
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
                ),
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

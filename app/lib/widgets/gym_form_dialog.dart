import 'package:flutter/material.dart';
import '../models/gym.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

class GymFormDialog extends StatefulWidget {
  final Gym? gym;

  const GymFormDialog({super.key, this.gym});

  @override
  State<GymFormDialog> createState() => _GymFormDialogState();
}

class _GymFormDialogState extends State<GymFormDialog> {
  late TextEditingController _nameController;

  List<String> _equipment = [];
  final TextEditingController _equipmentInputController = TextEditingController();

  @override
  void initState() {
    super.initState();
    _nameController = TextEditingController(text: widget.gym?.name ?? '');
    _equipment = List.from(widget.gym?.equipment ?? []);
  }

  @override
  void dispose() {
    _nameController.dispose();
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

  void _submit() {
    final name = _nameController.text.trim();
    if (name.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Please enter a gym name')),
      );
      return;
    }

    Navigator.of(context).pop({
      'name': name,
      'equipment': _equipment,
    });
  }

  @override
  Widget build(BuildContext context) {
    final isEditing = widget.gym != null;

    return Dialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(20),
      ),
      child: Container(
        width: MediaQuery.of(context).size.width * 0.9,
        constraints: const BoxConstraints(maxWidth: 500),
        padding: const EdgeInsets.all(24),
        decoration: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.glassDecoration(
                borderRadius: 20,
              )
            : null,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    isEditing ? 'Edit Gym' : 'Add Gym',
                    style: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.headlineStyle
                        : const TextStyle(
                            fontSize: 20,
                            fontWeight: FontWeight.bold,
                          ),
                  ),
                  IconButton(
                    icon: const Icon(Icons.close),
                    onPressed: () => Navigator.of(context).pop(),
                  ),
                ],
              ),
              const SizedBox(height: 24),
              AdaptiveTextField(
                controller: _nameController,
                labelText: 'Gym Name',
                placeholder: 'e.g., Home Gym, LA Fitness',
              ),
              const SizedBox(height: 24),
              Text(
                'Equipment',
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 16)
                    : const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
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
              const SizedBox(height: 16),
              if (_equipment.isEmpty)
                Padding(
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  child: Center(
                    child: Text(
                      'No equipment added yet',
                      style: TextStyle(
                        color: Colors.grey[600],
                        fontStyle: FontStyle.italic,
                      ),
                    ),
                  ),
                )
              else
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: _equipment.map((equipment) {
                    return Chip(
                      label: Text(equipment),
                      deleteIcon: const Icon(Icons.close, size: 18),
                      onDeleted: () => _removeEquipment(equipment),
                      backgroundColor: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                          : Theme.of(context).colorScheme.primaryContainer,
                    );
                  }).toList(),
                ),
              const SizedBox(height: 24),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TextButton(
                    onPressed: () => Navigator.of(context).pop(),
                    child: const Text('Cancel'),
                  ),
                  const SizedBox(width: 8),
                  AdaptiveButton(
                    onPressed: _submit,
                    child: Text(isEditing ? 'Update' : 'Add'),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

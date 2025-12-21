import 'package:flutter/material.dart';
import '../generated/app_localizations.dart';
import '../services/gym_service.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import 'adaptive/adaptive.dart';

/// searchable multi-select equipment picker that fetches from /equipment
class EquipmentSelector extends StatefulWidget {
  final List<String> selected;
  final ValueChanged<List<String>> onChanged;

  const EquipmentSelector({
    super.key,
    required this.selected,
    required this.onChanged,
  });

  @override
  State<EquipmentSelector> createState() => _EquipmentSelectorState();
}

class _EquipmentSelectorState extends State<EquipmentSelector> {
  final GymService _gymService = GymService();
  final TextEditingController _searchController = TextEditingController();

  List<String>? _allEquipment;
  String? _error;
  bool _loading = true;
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    _loadEquipment();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadEquipment() async {
    final response = await _gymService.getAvailableEquipment();
    if (!mounted) return;

    setState(() {
      _loading = false;
      if (response.isSuccess && response.data != null) {
        _allEquipment = response.data!..sort();
      } else {
        _error = response.error ?? 'Failed to load equipment';
      }
    });
  }

  List<String> get _filteredEquipment {
    if (_allEquipment == null) return [];
    if (_searchQuery.isEmpty) return _allEquipment!;
    final query = _searchQuery.toLowerCase();
    return _allEquipment!.where((e) => e.toLowerCase().contains(query)).toList();
  }

  void _toggle(String equipment) {
    final updated = List<String>.from(widget.selected);
    if (updated.contains(equipment)) {
      updated.remove(equipment);
    } else {
      updated.add(equipment);
    }
    widget.onChanged(updated);
  }

  void _selectAll() {
    final visible = _filteredEquipment;
    final updated = List<String>.from(widget.selected);
    for (final e in visible) {
      if (!updated.contains(e)) updated.add(e);
    }
    widget.onChanged(updated);
  }

  void _deselectAll() {
    // wipe entire list to help migration from legacy free-text equipment
    widget.onChanged([]);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    if (_loading) {
      return const Center(
        child: Padding(
          padding: EdgeInsets.all(24),
          child: AdaptiveLoadingIndicator(),
        ),
      );
    }

    if (_error != null) {
      return Center(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                _error!,
                style: TextStyle(color: Colors.red[400]),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: 8),
              AdaptiveTextButton(
                onPressed: () {
                  setState(() {
                    _loading = true;
                    _error = null;
                  });
                  _loadEquipment();
                },
                child: Text(l10n.refresh),
              ),
            ],
          ),
        ),
      );
    }

    final filtered = _filteredEquipment;
    final selectedSet = Set<String>.from(widget.selected);
    final allVisibleSelected = filtered.isNotEmpty && filtered.every(selectedSet.contains);

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        // search bar + select all/none
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4),
          child: Row(
            children: [
              Expanded(
                child: AdaptiveTextField(
                  controller: _searchController,
                  labelText: l10n.searchByName,
                  onChanged: (value) => setState(() => _searchQuery = value),
                  prefix: const Icon(Icons.search, size: 20),
                ),
              ),
              const SizedBox(width: 8),
              IconButton(
                onPressed: allVisibleSelected ? _deselectAll : _selectAll,
                icon: Icon(
                  allVisibleSelected ? Icons.deselect : Icons.select_all,
                  size: 20,
                ),
                tooltip: allVisibleSelected ? l10n.noEquipment : l10n.addAllEquipment,
              ),
            ],
          ),
        ),
        const SizedBox(height: 12),
        // equipment chips
        if (filtered.isEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 16),
            child: Text(
              l10n.noMatchingUsers, // reusing "no matching" string
              style: TextStyle(
                color: Colors.grey[600],
                fontStyle: FontStyle.italic,
              ),
            ),
          )
        else
          ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 200),
            child: SingleChildScrollView(
              child: Wrap(
                spacing: 8,
                runSpacing: 8,
                children: filtered.map((equipment) {
                  final isSelected = selectedSet.contains(equipment);
                  return FilterChip(
                    label: Text(
                      equipment,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.captionStyle.copyWith(
                              fontSize: 12,
                              color: isSelected ? Colors.white : null,
                            )
                          : null,
                    ),
                    selected: isSelected,
                    onSelected: (_) => _toggle(equipment),
                    selectedColor: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.primaryColor
                        : Theme.of(context).colorScheme.primaryContainer,
                    checkmarkColor: Colors.white,
                  );
                }).toList(),
              ),
            ),
          ),
        // selected count
        if (widget.selected.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: 8),
            child: Text(
              '${widget.selected.length} selected',
              style: TextStyle(
                fontSize: 12,
                color: Colors.grey[600],
              ),
            ),
          ),
      ],
    );
  }
}

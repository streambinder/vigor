import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/equipment_info.dart';
import '../services/gym_service.dart';
import '../utils/knowledge_labels.dart';
import 'adaptive/adaptive.dart';

/// searchable multi-select equipment picker that fetches from /equipment
class EquipmentSelector extends StatefulWidget {
  final List<String> selected;
  final ValueChanged<List<String>> onChanged;
  /// called when equipment data loads, exposing which items are weighted
  final ValueChanged<Set<String>>? onWeightedModifiersLoaded;

  const EquipmentSelector({
    super.key,
    required this.selected,
    required this.onChanged,
    this.onWeightedModifiersLoaded,
  });

  @override
  State<EquipmentSelector> createState() => _EquipmentSelectorState();
}

class _EquipmentSelectorState extends State<EquipmentSelector> {
  final GymService _gymService = GymService();
  final TextEditingController _searchController = TextEditingController();

  List<EquipmentInfo>? _allEquipment;
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
        _allEquipment = response.data!..sort((a, b) => a.id.compareTo(b.id));
        // notify parent about weighted modifiers
        final weighted = _allEquipment!.where((e) => e.isWeighted).map((e) => e.id).toSet();
        widget.onWeightedModifiersLoaded?.call(weighted);
      } else {
        _error = response.error ?? 'Failed to load equipment';
      }
    });
  }

  List<EquipmentInfo> get _filteredEquipment {
    if (_allEquipment == null) return [];
    if (_searchQuery.isEmpty) return _allEquipment!;
    final query = _searchQuery.toLowerCase();
    final l10n = AppLocalizations.of(context);
    return _allEquipment!.where((e) =>
      e.id.toLowerCase().contains(query) ||
      KnowledgeLabels.equipmentLabel(e.id, l10n).toLowerCase().contains(query),
    ).toList();
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
      if (!updated.contains(e.id)) updated.add(e.id);
    }
    widget.onChanged(updated);
  }

  void _deselectAll() {
    widget.onChanged([]);
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    if (_loading) {
      return const Center(
        child: Padding(
          padding: VigorSpacing.paddingLg,
          child: AdaptiveLoadingIndicator(),
        ),
      );
    }

    if (_error != null) {
      return Center(
        child: Padding(
          padding: VigorSpacing.paddingMd,
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                _error!,
                style: VigorTypography.body.copyWith(color: VigorColors.error),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: VigorSpacing.sm),
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
    final allVisibleSelected = filtered.isNotEmpty && filtered.every((e) => selectedSet.contains(e.id));

    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        // search bar + select all/none
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.xs),
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
              const SizedBox(width: VigorSpacing.sm),
              IconButton(
                onPressed: allVisibleSelected ? _deselectAll : _selectAll,
                icon: Icon(
                  allVisibleSelected ? Icons.deselect : Icons.select_all,
                  size: 20,
                  color: VigorColors.indigo,
                ),
                tooltip: allVisibleSelected ? l10n.noEquipment : l10n.addAllEquipment,
              ),
            ],
          ),
        ),
        const SizedBox(height: VigorSpacing.sm),
        // equipment chips
        if (filtered.isEmpty)
          Padding(
            padding: VigorSpacing.paddingMd,
            child: Text(
              l10n.noMatchingUsers,
              style: VigorTypography.body.copyWith(
                color: VigorColors.textMuted(context),
                fontStyle: FontStyle.italic,
              ),
            ),
          )
        else
          ConstrainedBox(
            constraints: const BoxConstraints(maxHeight: 200),
            child: SingleChildScrollView(
              child: Wrap(
                spacing: VigorSpacing.xs,
                runSpacing: VigorSpacing.xs,
                children: filtered.map((item) {
                  final isSelected = selectedSet.contains(item.id);
                  final label = KnowledgeLabels.equipmentLabel(item.id, l10n);
                  return FilterChip(
                    label: Text(
                      label,
                      style: VigorTypography.caption.copyWith(
                        color: isSelected ? Colors.white : VigorColors.textPrimary(context),
                      ),
                    ),
                    selected: isSelected,
                    onSelected: (_) => _toggle(item.id),
                    selectedColor: VigorColors.indigo,
                    checkmarkColor: Colors.white,
                    visualDensity: VisualDensity.compact,
                    materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.xs),
                  );
                }).toList(),
              ),
            ),
          ),
        // selected count
        if (widget.selected.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: VigorSpacing.sm),
            child: Text(
              '${widget.selected.length} selected',
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
              ),
            ),
          ),
      ],
    );
  }
}

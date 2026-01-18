import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../services/gym_service.dart';
import 'adaptive/adaptive.dart';

/// searchable multi-select goal picker that fetches from /goals
class GoalSelector extends StatefulWidget {
  final List<String> selected;
  final ValueChanged<List<String>> onChanged;

  const GoalSelector({
    super.key,
    required this.selected,
    required this.onChanged,
  });

  @override
  State<GoalSelector> createState() => _GoalSelectorState();
}

class _GoalSelectorState extends State<GoalSelector> {
  final GymService _gymService = GymService();
  final TextEditingController _searchController = TextEditingController();

  List<String>? _allGoals;
  String? _error;
  bool _loading = true;
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    _loadGoals();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _loadGoals() async {
    final response = await _gymService.getAvailableGoals();
    if (!mounted) return;

    setState(() {
      _loading = false;
      if (response.isSuccess && response.data != null) {
        _allGoals = response.data!..sort();
      } else {
        _error = response.error ?? 'Failed to load goals';
      }
    });
  }

  List<String> get _filteredGoals {
    if (_allGoals == null) return [];
    if (_searchQuery.isEmpty) return _allGoals!;
    final query = _searchQuery.toLowerCase();
    return _allGoals!.where((g) => g.toLowerCase().contains(query)).toList();
  }

  void _toggle(String goal) {
    final updated = List<String>.from(widget.selected);
    if (updated.contains(goal)) {
      updated.remove(goal);
    } else {
      updated.add(goal);
    }
    widget.onChanged(updated);
  }

  void _selectAll() {
    final visible = _filteredGoals;
    final updated = List<String>.from(widget.selected);
    for (final g in visible) {
      if (!updated.contains(g)) updated.add(g);
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
                  _loadGoals();
                },
                child: Text(l10n.refresh),
              ),
            ],
          ),
        ),
      );
    }

    final filtered = _filteredGoals;
    final selectedSet = Set<String>.from(widget.selected);
    final allVisibleSelected = filtered.isNotEmpty && filtered.every(selectedSet.contains);

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
        // goal chips
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
                spacing: VigorSpacing.sm,
                runSpacing: VigorSpacing.sm,
                children: filtered.map((goal) {
                  final isSelected = selectedSet.contains(goal);
                  return FilterChip(
                    label: Text(
                      goal,
                      style: VigorTypography.caption.copyWith(
                        color: isSelected ? Colors.white : VigorColors.textPrimary(context),
                      ),
                    ),
                    selected: isSelected,
                    onSelected: (_) => _toggle(goal),
                    selectedColor: VigorColors.indigo,
                    checkmarkColor: Colors.white,
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

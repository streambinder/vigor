import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../services/gym_service.dart';
import 'adaptive/adaptive.dart';

/// searchable multi-select goal picker that fetches from /goals
class GoalSelector extends StatefulWidget {
  final List<String> selected;
  final ValueChanged<List<String>> onChanged;
  final int? maxSelection;

  const GoalSelector({
    super.key,
    required this.selected,
    required this.onChanged,
    this.maxSelection,
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
    final l10n = AppLocalizations.of(context);
    return _allGoals!.where((g) {
      // match against goal id or localized label
      return g.toLowerCase().contains(query) ||
          _goalLabel(g, l10n).toLowerCase().contains(query);
    }).toList();
  }

  void _toggle(String goal) {
    final updated = List<String>.from(widget.selected);
    if (updated.contains(goal)) {
      updated.remove(goal);
    } else {
      if (widget.maxSelection != null && updated.length >= widget.maxSelection!) {
        return;
      }
      updated.add(goal);
    }
    widget.onChanged(updated);
  }

  void _selectAll() {
    if (widget.maxSelection != null) return;
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

  String _goalLabel(String goalId, AppLocalizations l10n) {
    return switch (goalId) {
      'hypertrophy' => l10n.goalHypertrophy,
      'fat loss' => l10n.goalFatLoss,
      'toning' => l10n.goalToning,
      'posture' => l10n.goalPosture,
      'rehabilitation' => l10n.goalRehabilitation,
      'wellness' => l10n.goalWellness,
      'flexibility' => l10n.goalFlexibility,
      'sports' => l10n.goalSports,
      _ => goalId.split(' ').map((w) => w.isEmpty ? '' : '${w[0].toUpperCase()}${w.substring(1)}').join(' '),
    };
  }

  String _goalDescription(String goalId, AppLocalizations l10n) {
    return switch (goalId) {
      'hypertrophy' => l10n.goalHypertrophyDescription,
      'fat loss' => l10n.goalFatLossDescription,
      'toning' => l10n.goalToningDescription,
      'posture' => l10n.goalPostureDescription,
      'rehabilitation' => l10n.goalRehabilitationDescription,
      'wellness' => l10n.goalWellnessDescription,
      'flexibility' => l10n.goalFlexibilityDescription,
      'sports' => l10n.goalSportsDescription,
      _ => '',
    };
  }

  Widget _buildGoalTile(String goal, bool isSelected, AppLocalizations l10n) {
    return GestureDetector(
      onTap: () => _toggle(goal),
      child: Container(
        padding: const EdgeInsets.all(VigorSpacing.sm),
        decoration: BoxDecoration(
          color: isSelected ? VigorColors.indigo.withValues(alpha: 0.1) : null,
          border: Border.all(
            color: isSelected ? VigorColors.indigo : VigorColors.border(context),
          ),
          borderRadius: VigorRadius.radiusMd,
        ),
        child: Row(
          children: [
            Icon(
              isSelected ? Icons.check_circle : Icons.circle_outlined,
              color: isSelected ? VigorColors.indigo : VigorColors.textMuted(context),
              size: 20,
            ),
            const SizedBox(width: VigorSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    _goalLabel(goal, l10n),
                    style: VigorTypography.body.copyWith(
                      fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                      color: VigorColors.textPrimary(context),
                    ),
                  ),
                  if (_goalDescription(goal, l10n).isNotEmpty) ...[
                    const SizedBox(height: 2),
                    Text(
                      _goalDescription(goal, l10n),
                      style: VigorTypography.caption.copyWith(
                        color: VigorColors.textSecondary(context),
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ],
              ),
            ),
          ],
        ),
      ),
    );
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
              if (widget.maxSelection == null) ...[
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
            ],
          ),
        ),
        const SizedBox(height: VigorSpacing.sm),
        // goal tiles
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
            constraints: const BoxConstraints(maxHeight: 300),
            child: ListView.separated(
              shrinkWrap: true,
              itemCount: filtered.length,
              separatorBuilder: (_, __) => const SizedBox(height: VigorSpacing.xs),
              itemBuilder: (context, index) {
                final goal = filtered[index];
                final isSelected = selectedSet.contains(goal);
                return _buildGoalTile(goal, isSelected, l10n);
              },
            ),
          ),
        // selected count
        Padding(
          padding: const EdgeInsets.only(top: VigorSpacing.sm),
          child: Text(
            widget.maxSelection != null
                ? '${widget.selected.length}/${widget.maxSelection} selected'
                : '${widget.selected.length} selected',
            style: VigorTypography.caption.copyWith(
              color: VigorColors.textSecondary(context),
            ),
          ),
        ),
      ],
    );
  }
}

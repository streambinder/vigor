import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import 'package:cached_network_image/cached_network_image.dart';
import '../config/api_config.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity.dart';
import '../providers/auth_provider.dart';
import '../services/service_locator.dart';
import '../services/share_service.dart';
import '../widgets/adaptive/adaptive.dart';

class SharedTrainingScreen extends StatefulWidget {
  final String token;

  const SharedTrainingScreen({super.key, required this.token});

  @override
  State<SharedTrainingScreen> createState() => _SharedTrainingScreenState();
}

class _SharedTrainingScreenState extends State<SharedTrainingScreen> {
  final ShareService _shareService = ShareService();
  bool _loading = true;
  String? _error;
  Training? _training;
  String? _ownerName;
  String? _ownerUserId;
  bool _claiming = false;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    final response = await _shareService.getSharedTraining(widget.token);
    if (!mounted) return;

    if (response.isSuccess && response.data != null) {
      try {
        final trainingJson = response.data!['training'] as Map<String, dynamic>;
        final ownerJson = response.data!['owner'] as Map<String, dynamic>;
        setState(() {
          _training = Training.fromJson(trainingJson);
          _ownerName = '${ownerJson['first_name']} ${ownerJson['last_name']}'.trim();
          _ownerUserId = ownerJson['user_id'] as String?;
          _loading = false;
        });
      } catch (e) {
        setState(() {
          _error = 'Failed to load training';
          _loading = false;
        });
      }
    } else {
      setState(() {
        _error = response.error ?? 'Training not found';
        _loading = false;
      });
    }
  }

  Future<void> _claimTraining() async {
    final l10n = AppLocalizations.of(context);
    final authProvider = context.read<AuthProvider>();

    if (authProvider.state != AuthState.authenticated) {
      AdaptiveNotification.show(context: context, message: l10n.loginToAdd);
      return;
    }

    setState(() => _claiming = true);

    final response = await context.read<ServiceLocator>().trainingService.claimSharedTraining(widget.token);

    if (!mounted) return;
    setState(() => _claiming = false);

    if (response.isSuccess) {
      AdaptiveNotification.show(context: context, message: l10n.trainingAddedSuccessfully);
      Navigator.of(context).pop();
    } else {
      AdaptiveNotification.showError(context: context, message: l10n.failedToAddTraining, rawError: response.error);
    }
  }

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    if (minutes < 60) return '${minutes}m';
    final hours = minutes ~/ 60;
    final remaining = minutes % 60;
    return remaining == 0 ? '${hours}h' : '${hours}h ${remaining}m';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(title: Text(_training?.name ?? '')),
      body: _loading
          ? const AdaptiveLoadingScreen()
          : _error != null
              ? _buildError(l10n)
              : _buildContent(l10n, isDark),
    );
  }

  Widget _buildError(AppLocalizations l10n) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.error_outline, size: 64, color: VigorColors.error),
          const SizedBox(height: VigorSpacing.md),
          Text(l10n.trainingNotFound, style: VigorTypography.title.copyWith(color: VigorColors.textPrimary(context))),
        ],
      ),
    );
  }

  Widget _buildContent(AppLocalizations l10n, bool isDark) {
    final training = _training!;
    final isAuthenticated = context.watch<AuthProvider>().state == AuthState.authenticated;

    return Column(
      children: [
        Expanded(
          child: ListView(
            padding: VigorSpacing.paddingLg,
            children: [
              // owner card
              if (_ownerName != null && _ownerName!.isNotEmpty)
                _buildOwnerCard(isDark),
              const SizedBox(height: VigorSpacing.md),

              // training info card
              Container(
                padding: VigorSpacing.paddingLg,
                decoration: BoxDecoration(
                  color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
                  borderRadius: VigorRadius.radiusLg,
                ),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // metadata chips
                    Wrap(
                      spacing: VigorSpacing.xs,
                      runSpacing: VigorSpacing.xs,
                      children: [
                        _buildChip(Icons.timer, _formatDuration(training.duration)),
                        _buildChip(Icons.sports_martial_arts, training.methodology),
                      ],
                    ),
                    const SizedBox(height: VigorSpacing.md),
                    Text(training.description, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context), height: 1.5)),
                    if (training.equipment.isNotEmpty) ...[
                      const SizedBox(height: VigorSpacing.md),
                      Wrap(
                        spacing: VigorSpacing.xs,
                        runSpacing: VigorSpacing.xs,
                        children: training.equipment.map((eq) => _buildChip(Icons.fitness_center, eq)).toList(),
                      ),
                    ],
                    if (training.goals.isNotEmpty) ...[
                      const SizedBox(height: VigorSpacing.md),
                      Wrap(
                        spacing: VigorSpacing.xs,
                        runSpacing: VigorSpacing.xs,
                        children: training.goals.map((g) => _buildChip(Icons.track_changes, g)).toList(),
                      ),
                    ],
                    if (training.muscles.isNotEmpty) ...[
                      const SizedBox(height: VigorSpacing.md),
                      Wrap(
                        spacing: VigorSpacing.xs,
                        runSpacing: VigorSpacing.xs,
                        children: training.muscles.map((m) => _buildChip(Icons.accessibility_new, m)).toList(),
                      ),
                    ],
                  ],
                ),
              ),
              const SizedBox(height: VigorSpacing.lg),

              // routines
              for (final routine in training.routines) ...[
                _buildRoutineCard(routine, isDark),
                const SizedBox(height: VigorSpacing.md),
              ],
              const SizedBox(height: VigorSpacing.xl),
            ],
          ),
        ),

        // bottom CTA
        SafeArea(
          child: Padding(
            padding: VigorSpacing.paddingLg,
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: _claiming ? null : _claimTraining,
                icon: _claiming
                    ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                    : Icon(isAuthenticated ? Icons.add : Icons.login),
                label: Text(isAuthenticated ? l10n.addToMyTrainings : l10n.loginToAdd),
                style: ElevatedButton.styleFrom(
                  backgroundColor: VigorColors.persimmon,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: VigorSpacing.md),
                  shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildOwnerCard(bool isDark) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusMd,
      ),
      child: Row(
        children: [
          if (_ownerUserId != null)
            CircleAvatar(
              radius: 18,
              backgroundImage: CachedNetworkImageProvider(ApiConfig.avatarUrl(_ownerUserId!)),
              backgroundColor: VigorColors.stone.withValues(alpha: 0.2),
            ),
          const SizedBox(width: VigorSpacing.sm),
          Text(
            AppLocalizations.of(context).sharedBy(_ownerName!),
            style: VigorTypography.label.copyWith(color: VigorColors.textSecondary(context)),
          ),
        ],
      ),
    );
  }

  Widget _buildChip(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.15),
        borderRadius: VigorRadius.radiusSm,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 14, color: VigorColors.stone),
          const SizedBox(width: 4),
          Text(label, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
        ],
      ),
    );
  }

  Widget _buildRoutineCard(Routine routine, bool isDark) {
    return Container(
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusMd,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Padding(
            padding: const EdgeInsets.fromLTRB(VigorSpacing.md, VigorSpacing.md, VigorSpacing.md, VigorSpacing.sm),
            child: Text(
              routine.type.toUpperCase(),
              style: VigorTypography.label.copyWith(color: VigorColors.indigoAdaptive(context), fontWeight: FontWeight.w600),
            ),
          ),
          for (final block in routine.blocks)
            _buildBlockRow(block),
        ],
      ),
    );
  }

  Widget _buildBlockRow(Block block) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.xs),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (block.repeats > 1)
            Padding(
              padding: const EdgeInsets.only(bottom: VigorSpacing.xs),
              child: Text('${block.repeats}x', style: VigorTypography.caption.copyWith(color: VigorColors.stone, fontWeight: FontWeight.w600)),
            ),
          for (final activity in block.activities)
            Padding(
              padding: const EdgeInsets.only(bottom: VigorSpacing.xs),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      activity.name,
                      style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)),
                    ),
                  ),
                  if (activity.reps > 0)
                    Text('${activity.reps} reps', style: VigorTypography.caption.copyWith(color: VigorColors.stone))
                  else if (activity.duration > 0)
                    Text('${activity.duration}s', style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
                ],
              ),
            ),
          const Divider(height: 1),
        ],
      ),
    );
  }
}

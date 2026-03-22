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
import '../models/activity_ext.dart';
import '../models/exercise.dart';
import '../providers/auth_provider.dart';
import '../services/service_locator.dart';
import '../services/share_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/cached_exercise_image.dart';
import '../widgets/marquee_text.dart';
import '../utils/exercise_modal.dart';
import '../main.dart';
import 'training_details_screen.dart';

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
        final training = Training.fromJson(trainingJson);
        final ownerUserId = ownerJson['user_id'] as String?;

        setState(() {
          _training = training;
          _ownerName = '${ownerJson['first_name']} ${ownerJson['last_name']}'.trim();
          _ownerUserId = ownerUserId;
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

  Future<void> _goToOwnTraining() async {
    final serviceLocator = context.read<ServiceLocator>();

    // try to find the training in the cache first
    var cached = serviceLocator.trainingsNotifier.value
        ?.where((t) => t.id == _training!.id).firstOrNull;

    if (cached == null) {
      await serviceLocator.refreshTrainings();
      cached = serviceLocator.trainingsNotifier.value
          ?.where((t) => t.id == _training!.id).firstOrNull;
    }

    if (!mounted) return;
    Navigator.of(context).pushReplacement(
      MaterialPageRoute(builder: (_) => TrainingDetailsScreen(training: cached ?? _training!)),
    );
  }

  Future<void> _claimTraining() async {
    final l10n = AppLocalizations.of(context);
    final authProvider = context.read<AuthProvider>();

    if (authProvider.state != AuthState.authenticated) {
      // store token for post-login auto-claim, then navigate to login
      final serviceLocator = context.read<ServiceLocator>();
      serviceLocator.pendingShareToken = widget.token;
      serviceLocator.pendingShareAutoClaim = true;
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => const AuthenticationWrapper()),
      );
      return;
    }

    setState(() => _claiming = true);

    final response = await context.read<ServiceLocator>().trainingService.claimSharedTraining(widget.token);

    if (!mounted) return;
    setState(() => _claiming = false);

    if (response.isSuccess && response.data != null) {
      AdaptiveNotification.show(context: context, message: l10n.trainingAddedSuccessfully);
      Navigator.of(context).pushReplacement(
        MaterialPageRoute(builder: (_) => TrainingDetailsScreen(training: response.data!)),
      );
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

  String _formatTime(int seconds) {
    if (seconds < 60) return '${seconds}s';
    final minutes = seconds ~/ 60;
    final remainingSeconds = seconds % 60;
    return remainingSeconds == 0 ? '${minutes}m' : '${minutes}m ${remainingSeconds}s';
  }

  Exercise? _parseExercise(Map<String, dynamic> detail) {
    if (detail.isEmpty) return null;
    try {
      return Exercise.fromJson(detail);
    } catch (_) {
      return null;
    }
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
          const Icon(Icons.error_outline, size: 64, color: VigorColors.error),
          const SizedBox(height: VigorSpacing.md),
          Text(l10n.trainingNotFound, style: VigorTypography.title.copyWith(color: VigorColors.textPrimary(context))),
        ],
      ),
    );
  }

  Widget _buildContent(AppLocalizations l10n, bool isDark) {
    final training = _training!;
    final authState = context.watch<AuthProvider>();
    final isAuthenticated = authState.state == AuthState.authenticated;
    final isOwner = isAuthenticated && _ownerUserId != null && authState.currentUser?.id == _ownerUserId;

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
                onPressed: _claiming ? null : (isOwner ? _goToOwnTraining : _claimTraining),
                icon: _claiming
                    ? const SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                    : Icon(isOwner ? Icons.open_in_new : (isAuthenticated ? Icons.add : Icons.login)),
                label: Text(isOwner ? l10n.goToTraining : (isAuthenticated ? l10n.addToMyTrainings : l10n.loginToAdd)),
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
        borderRadius: VigorRadius.radiusLg,
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // routine header
          Stack(
            children: [
              Container(
                width: double.infinity,
                padding: VigorSpacing.paddingMd,
                decoration: BoxDecoration(
                  color: VigorColors.stone.withValues(alpha: 0.05),
                  borderRadius: const BorderRadius.only(topLeft: Radius.circular(VigorRadius.lg), topRight: Radius.circular(VigorRadius.lg)),
                ),
                child: Center(
                  child: Text(routine.type.toUpperCase(), style: VigorTypography.headline.copyWith(color: VigorColors.stone, fontSize: 14)),
                ),
              ),
              if (routine.rest > 0)
                Positioned(
                  top: VigorSpacing.sm,
                  right: VigorSpacing.sm,
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusFull,
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.timer, size: 12, color: VigorColors.stone),
                        const SizedBox(width: VigorSpacing.xs),
                        Text('${routine.rest}s', style: VigorTypography.data.copyWith(color: VigorColors.stone, fontWeight: FontWeight.w600)),
                      ],
                    ),
                  ),
                ),
            ],
          ),
          // blocks
          Column(
            children: routine.blocks.asMap().entries.expand((entry) => [
              if (entry.key > 0) Divider(height: 1, thickness: 1, color: VigorColors.stone.withValues(alpha: 0.2)),
              _buildBlockCard(entry.value, entry.key + 1, routine.blocks.length, isDark),
            ]).toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildBlockCard(Block block, int blockNumber, int totalBlocks, bool isDark) {
    final showBlockLabel = totalBlocks > 1;
    final hasChips = block.repeats > 1 || block.rest > 0;

    return Stack(
      clipBehavior: Clip.none,
      children: [
        Padding(
          padding: VigorSpacing.paddingMd,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              if (showBlockLabel) ...[
                Center(
                  child: Text('Block $blockNumber', style: VigorTypography.caption.copyWith(fontWeight: FontWeight.bold, color: VigorColors.stone)),
                ),
                const SizedBox(height: VigorSpacing.md),
              ],
              ...block.activities.asMap().entries.map((entry) => Padding(
                padding: EdgeInsets.only(bottom: entry.key < block.activities.length - 1 ? VigorSpacing.md : 0),
                child: _buildActivityRow(entry.value, isDark, isFirstInBlock: entry.key == 0 && hasChips && totalBlocks == 1),
              )),
            ],
          ),
        ),
        if (hasChips)
          Positioned(
            top: VigorSpacing.sm,
            right: VigorSpacing.sm,
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                if (block.repeats > 1)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusSm,
                    ),
                    child: Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        const Icon(Icons.repeat, size: 12, color: VigorColors.stone),
                        const SizedBox(width: VigorSpacing.xs),
                        Text('${block.repeats}x', style: VigorTypography.data.copyWith(fontWeight: FontWeight.bold, color: VigorColors.stone)),
                      ],
                    ),
                  ),
                if (block.repeats > 1 && block.rest > 0)
                  const SizedBox(width: VigorSpacing.xs),
                if (block.rest > 0)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusSm,
                    ),
                    child: Text('${block.rest}s', style: VigorTypography.data.copyWith(fontWeight: FontWeight.bold, color: VigorColors.stone)),
                  ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _buildActivityRow(Activity activity, bool isDark, {bool isFirstInBlock = false}) {
    final exercise = _parseExercise(activity.detail);
    final hasValidImage = exercise != null && CachedExerciseImage.isValidUrl(exercise.reference);

    return GestureDetector(
      onTap: exercise != null ? () => ExerciseModal.show(context, exercise) : null,
      behavior: HitTestBehavior.opaque,
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // thumbnail
          SizedBox(
            width: 72,
            height: 72,
            child: hasValidImage
                ? CachedExerciseImage(
                    imageUrl: exercise.reference,
                    width: 72,
                    height: 72,
                    borderRadius: VigorRadius.radiusMd,
                  )
                : Container(
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusMd,
                    ),
                    child: Center(child: Icon(Icons.fitness_center, size: 72 * 0.4, color: VigorColors.stone.withValues(alpha: 0.5))),
                  ),
          ),
          const SizedBox(width: VigorSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                MarqueeText(text: activity.name, style: VigorTypography.body.copyWith(fontWeight: FontWeight.w600, color: VigorColors.textPrimary(context))),
                if (exercise != null && (exercise.equipment.isNotEmpty || exercise.muscles.isNotEmpty || activity.modifiers.isNotEmpty)) ...[
                  const SizedBox(height: VigorSpacing.xs),
                  _buildExerciseDetails(exercise, activity.modifiers),
                ],
                const SizedBox(height: VigorSpacing.sm),
                _buildActivityTags(activity),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildExerciseDetails(Exercise exercise, List<String> modifiers) {
    return Wrap(
      spacing: VigorSpacing.sm,
      runSpacing: VigorSpacing.xs,
      children: [
        if (exercise.muscles.isNotEmpty) _buildDetailChip(Icons.accessibility_new, exercise.muscles.take(2).join(' · ')),
        if (exercise.equipment.isNotEmpty) _buildDetailChip(Icons.fitness_center, exercise.equipment.join(' · ')),
        if (modifiers.isNotEmpty) _buildDetailChip(Icons.tune, modifiers.join(' · ')),
      ],
    );
  }

  Widget _buildDetailChip(IconData icon, String text) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 12, color: VigorColors.stone),
        const SizedBox(width: VigorSpacing.xs),
        Flexible(child: Text(text, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context)), overflow: TextOverflow.ellipsis)),
      ],
    );
  }

  Widget _buildActivityTags(Activity activity) {
    return Wrap(
      spacing: VigorSpacing.xs,
      runSpacing: VigorSpacing.xs,
      children: [
        if (activity.reps > 0) _buildTag(Icons.repeat, '${activity.reps}'),
        if (activity.weightKg > 0) _buildTag(Icons.fitness_center, '${activity.weightKgDisplay}kg'),
        if (activity.duration > 0) _buildTag(Icons.timer, _formatTime(activity.duration)),
        if (activity.rest > 0) _buildTag(Icons.hourglass_bottom, '${activity.rest}s'),
      ],
    );
  }

  Widget _buildTag(IconData icon, String label) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.sm, vertical: VigorSpacing.xs),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusSm,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 11, color: VigorColors.stone),
          const SizedBox(width: VigorSpacing.xs),
          Text(label, style: VigorTypography.data.copyWith(fontSize: 11, fontWeight: FontWeight.w600, color: VigorColors.stone)),
        ],
      ),
    );
  }
}

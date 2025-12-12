import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../config/api_config.dart';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity.dart';
import '../models/exercise.dart';
import '../providers/auth_provider.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/user_select_dialog.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../utils/exercise_modal.dart';
import '../utils/feedback_modal.dart';
import 'main_navigation.dart';
import 'tabata_timer_screen.dart';

class TrainingDetailsScreen extends StatefulWidget {
  final Training training;

  const TrainingDetailsScreen({
    super.key,
    required this.training,
  });

  @override
  State<TrainingDetailsScreen> createState() => _TrainingDetailsScreenState();
}

class _TrainingDetailsScreenState extends State<TrainingDetailsScreen> {
  Training get training => widget.training;
  int _partnerCount = 0;

  @override
  void initState() {
    super.initState();
    _loadPartners();
  }

  Future<void> _loadPartners() async {
    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);
    final response = await trainingService.getPartners(training.id);
    if (response.isSuccess && mounted) {
      setState(() {
        _partnerCount = response.data?.length ?? 0;
      });
    }
  }

  String _formatDate(DateTime date) {
    return '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}';
  }

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    if (minutes < 60) {
      return '$minutes min';
    }
    final hours = minutes ~/ 60;
    final remainingMinutes = minutes % 60;
    if (remainingMinutes == 0) {
      return '$hours hr';
    }
    return '$hours hr $remainingMinutes min';
  }

  String _formatTime(int seconds) {
    if (seconds < 60) {
      return '${seconds}s';
    }
    final minutes = seconds ~/ 60;
    final remainingSeconds = seconds % 60;
    if (remainingSeconds == 0) {
      return '${minutes}m';
    }
    return '${minutes}m ${remainingSeconds}s';
  }

  Exercise? _parseExercise(Map<String, dynamic> detail) {
    try {
      if (detail.isEmpty) return null;
      return Exercise.fromJson(detail);
    } catch (e) {
      return null;
    }
  }

  bool _isValidImageUrl(String? url) {
    if (url == null || url.isEmpty) return false;
    final uri = Uri.tryParse(url);
    if (uri == null) return false;
    return uri.hasScheme && (uri.scheme == 'http' || uri.scheme == 'https');
  }

  // proxy external image URLs through backend to avoid CORS issues on web
  String _proxyImageUrl(String url) {
    return '${ApiConfig.baseUrl}/proxy/image?url=${Uri.encodeComponent(url)}';
  }

  Future<void> _deleteTraining(BuildContext context) async {
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    final isOwner = training.userId == currentUserId;
    final title = isOwner ? 'Delete Training' : 'Leave Training';
    final content = isOwner
        ? 'Are you sure you want to delete "${training.name}"? This action cannot be undone.'
        : 'Are you sure you want to leave "${training.name}"? You will no longer see this training.';
    final actionLabel = isOwner ? 'Delete' : 'Leave';
    final successMessage = isOwner ? 'Training deleted successfully' : 'Left training successfully';

    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: title,
      content: content,
      actions: [
        AdaptiveDialogAction(
          label: 'Cancel',
          onPressed: () => Navigator.of(context).pop(false),
        ),
        AdaptiveDialogAction(
          label: actionLabel,
          isDestructive: true,
          onPressed: () => Navigator.of(context).pop(true),
        ),
      ],
    );

    if (shouldDelete == true && context.mounted) {
      final storage = context.read<SecureStorageService>();
      final trainingService = TrainingService(storageService: storage);

      final response = await trainingService.deleteTraining(training.id);

      if (context.mounted) {
        if (response.isSuccess) {
          Navigator.of(context).pop(true); // Return true to indicate deletion
          AdaptiveNotification.show(
            context: context,
            message: successMessage,
          );
        } else {
          AdaptiveNotification.showError(
            context: context,
            message: response.error ?? 'Failed to delete training',
          );
        }
      }
    }
  }

  Future<void> _completeTraining(BuildContext context) async {
    final result = await FeedbackModal.show(context, training);
    if (result == null) return; // user cancelled

    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);

    final response = await trainingService.completeTraining(
      training.id,
      feedback: result.feedback,
      activityFeedback: result.activityFeedback,
    );

    if (context.mounted) {
      if (response.isSuccess) {
        Navigator.of(context).pop(true); // Return true to refresh the list
        AdaptiveNotification.show(
          context: context,
          message: 'Training marked as complete',
        );
      } else {
        AdaptiveNotification.showError(
          context: context,
          message: response.error ?? 'Failed to complete training',
        );
      }
    }
  }

  void _navigateToActivityScreen(BuildContext context) {
    Navigator.of(context).popUntil((route) => route.isFirst);
    MainNavigation.navigateToTab(1); // Activity tab
  }

  Future<void> _showAddPartnerDialog(BuildContext context) async {
    final user = await showUserSelectDialog(
      context: context,
      title: 'Add Partner',
    );

    if (user == null || !context.mounted) return;

    final shouldAdd = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: 'Add Partner',
      content: 'Add ${user.email} as a partner to "${training.name}"?',
      actions: [
        AdaptiveDialogAction(
          label: 'Cancel',
          onPressed: () => Navigator.of(context).pop(false),
        ),
        AdaptiveDialogAction(
          label: 'Add',
          onPressed: () => Navigator.of(context).pop(true),
        ),
      ],
    );

    if (shouldAdd != true || !context.mounted) return;

    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);
    final response = await trainingService.addPartner(training.id, user.id);

    if (context.mounted) {
      if (response.isSuccess) {
        _loadPartners();
        AdaptiveNotification.show(
          context: context,
          message: 'Partner added successfully',
        );
      } else {
        AdaptiveNotification.showError(
          context: context,
          message: response.error ?? 'Failed to add partner',
        );
      }
    }
  }

  Future<void> _cloneTraining(BuildContext context) async {
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    if (currentUserId.isEmpty) return;

    final shouldClone = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: 'Clone Training',
      content: 'Clone "${training.name}" to your trainings?',
      actions: [
        AdaptiveDialogAction(
          label: 'Cancel',
          onPressed: () => Navigator.of(context).pop(false),
        ),
        AdaptiveDialogAction(
          label: 'Clone',
          onPressed: () => Navigator.of(context).pop(true),
        ),
      ],
    );

    if (shouldClone != true || !context.mounted) return;

    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);
    final response = await trainingService.copyTraining(training.id, currentUserId);

    if (context.mounted) {
      if (response.isSuccess) {
        AdaptiveNotification.show(
          context: context,
          message: 'Training cloned',
        );
        Navigator.of(context).pop(true);
      } else {
        AdaptiveNotification.showError(
          context: context,
          message: response.error ?? 'Failed to clone training',
        );
      }
    }
  }

  Future<void> _showCopyTrainingDialog(BuildContext context) async {
    final user = await showUserSelectDialog(
      context: context,
      title: 'Share with User',
    );

    if (user == null || !context.mounted) return;

    final shouldShare = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: 'Share with User',
      content: 'Share "${training.name}" with ${user.email}?',
      actions: [
        AdaptiveDialogAction(
          label: 'Cancel',
          onPressed: () => Navigator.of(context).pop(false),
        ),
        AdaptiveDialogAction(
          label: 'Share',
          onPressed: () => Navigator.of(context).pop(true),
        ),
      ],
    );

    if (shouldShare != true || !context.mounted) return;

    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);
    final response = await trainingService.copyTraining(training.id, user.id);

    if (context.mounted) {
      if (response.isSuccess) {
        AdaptiveNotification.show(
          context: context,
          message: 'Training shared successfully',
        );
      } else {
        AdaptiveNotification.showError(
          context: context,
          message: response.error ?? 'Failed to share training',
        );
      }
    }
  }

  void _showReasoningDialog(BuildContext context) {
    final r = training.reasoning;

    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: const Text('Reasoning'),
        content: SizedBox(
          width: double.maxFinite,
          child: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                _buildReasoningSection(
                  title: 'Strategy',
                  child: Text(r.strategy),
                ),
                if (r.constraints.isNotEmpty)
                  _buildReasoningSection(
                    title: 'Constraints',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: r.constraints.map((c) => Padding(
                        padding: const EdgeInsets.only(bottom: 4),
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('• '),
                            Expanded(child: Text(c)),
                          ],
                        ),
                      )).toList(),
                    ),
                  ),
                if (r.targetMuscles.isNotEmpty)
                  _buildReasoningSection(
                    title: 'Target Muscles',
                    child: Wrap(
                      spacing: 8,
                      runSpacing: 4,
                      children: r.targetMuscles.map((m) => Chip(
                        label: Text(m),
                        visualDensity: VisualDensity.compact,
                      )).toList(),
                    ),
                  ),
                if (r.exercises.isNotEmpty)
                  _buildReasoningSection(
                    title: 'Exercises',
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: r.exercises.map((e) => Padding(
                        padding: const EdgeInsets.only(bottom: 4),
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            const Text('• '),
                            Expanded(child: Text(e)),
                          ],
                        ),
                      )).toList(),
                    ),
                  ),
                _buildReasoningSection(
                  title: 'Naming',
                  child: Text(r.namingLogic),
                ),
              ],
            ),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(ctx).pop(),
            child: const Text('Close'),
          ),
        ],
      ),
    );
  }

  Widget _buildReasoningSection({required String title, required Widget child}) {
    return Theme(
      data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
      child: ExpansionTile(
        title: Text(title, style: const TextStyle(fontWeight: FontWeight.w600)),
        initiallyExpanded: false,
        tilePadding: EdgeInsets.zero,
        childrenPadding: const EdgeInsets.only(bottom: 8),
        children: [child],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    final isOwner = training.userId == currentUserId;

    return PopScope(
      canPop: false,
      onPopInvokedWithResult: (didPop, result) {
        if (!didPop) {
          _navigateToActivityScreen(context);
        }
      },
      child: AdaptiveScaffold(
        appBar: AdaptiveAppBar(
          title: Text(training.name),
          automaticallyImplyLeading: false,
          leading: IconButton(
            icon: Icon(
              PlatformHelper.useLiquidGlass ? Icons.arrow_back_ios : Icons.arrow_back,
              color: PlatformHelper.useLiquidGlass ? LiquidGlassTheme.primaryColor : null,
            ),
            onPressed: () => _navigateToActivityScreen(context),
          ),
          actions: [
            AdaptiveIconButton(
              icon: const Icon(Icons.copy),
              tooltip: 'Clone Training',
              onPressed: () => _cloneTraining(context),
            ),
            // only owner can add partners
            if (isOwner)
              AdaptiveIconButton(
                icon: const Icon(Icons.person_add),
                tooltip: 'Add Partner',
                onPressed: () => _showAddPartnerDialog(context),
              ),
            AdaptiveIconButton(
              icon: const Icon(Icons.share),
              tooltip: 'Share with User',
              onPressed: () => _showCopyTrainingDialog(context),
            ),
            AdaptiveIconButton(
              icon: Icon(
                Icons.delete,
                color: isOwner ? null : Colors.grey,
              ),
              tooltip: isOwner ? 'Delete Training' : 'Leave Training',
              onPressed: () => _deleteTraining(context),
            ),
          ],
        ),
        body: SingleChildScrollView(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Training header card
            AdaptiveCard(
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: Text(
                            training.name,
                            style: PlatformHelper.useLiquidGlass
                                ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 24)
                                : Theme.of(context).textTheme.headlineMedium,
                          ),
                        ),
                        IconButton(
                          icon: Icon(
                            Icons.help_outline,
                            size: 20,
                            color: PlatformHelper.useLiquidGlass
                                ? LiquidGlassTheme.captionStyle.color
                                : Colors.grey,
                          ),
                          tooltip: 'Show AI reasoning',
                          onPressed: () => _showReasoningDialog(context),
                        ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Text(
                      training.description,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.bodyStyle
                          : Theme.of(context).textTheme.bodyLarge,
                    ),
                    const SizedBox(height: 16),
                    Wrap(
                      spacing: 16,
                      runSpacing: 8,
                      children: [
                        if (_partnerCount > 0)
                          _buildInfoLabel(
                            context,
                            icon: Icons.people,
                            text: '${1 + _partnerCount}',
                          ),
                        if (training.gym != null)
                          _buildInfoLabel(
                            context,
                            icon: Icons.location_on,
                            text: training.gym!.name,
                          ),
                        _buildInfoLabel(
                          context,
                          icon: Icons.tune,
                          text: training.type,
                        ),
                        _buildInfoLabel(
                          context,
                          icon: Icons.schedule,
                          text: _formatDuration(training.duration),
                        ),
                        _buildInfoLabel(
                          context,
                          icon: Icons.calendar_today,
                          text: _formatDate(training.completedAt ?? training.createdAt),
                        ),
                        if (training.parentId != null)
                          _buildInfoLabel(
                            context,
                            icon: Icons.copy,
                            text: 'Copied',
                          ),
                      ],
                    ),
                    const SizedBox(height: 12),
                    Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Icon(
                          Icons.fitness_center,
                          size: 18,
                          color: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.captionStyle.color
                              : Colors.grey.shade600,
                        ),
                        const SizedBox(width: 6),
                        Expanded(
                          child: Text(
                            training.equipment.isEmpty
                                ? 'No equipment'
                                : training.equipment.join(' · '),
                            style: PlatformHelper.useLiquidGlass
                                ? LiquidGlassTheme.captionStyle
                                : Theme.of(context).textTheme.bodyMedium?.copyWith(
                                      color: Colors.grey.shade600,
                                    ),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            const SizedBox(height: 16),

            // Start Tabata Timer button
            SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: () async {
                  final completed = await Navigator.of(context).push<bool>(
                    MaterialPageRoute(
                      builder: (context) => TabataTimerScreen(training: training),
                    ),
                  );
                  // If the timer returned true (workout completed), refresh
                  if (completed == true && context.mounted) {
                    Navigator.of(context).pop(true);
                  }
                },
                icon: const Icon(Icons.timer),
                label: const Text('Start Workout Timer'),
                style: ElevatedButton.styleFrom(
                  backgroundColor: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.successColor
                      : Colors.green[600],
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 16),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(12),
                  ),
                ),
              ),
            ),
            // Mark as Complete button (only show if not already completed, only owner can complete)
            if (training.completedAt == null) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: isOwner ? () => _completeTraining(context) : null,
                  icon: Icon(
                    Icons.check_circle_outline,
                    color: isOwner ? null : Colors.grey,
                  ),
                  label: Text(
                    'Mark as Complete',
                    style: TextStyle(
                      color: isOwner ? null : Colors.grey,
                    ),
                  ),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.successColor
                        : Colors.green[600],
                    side: BorderSide(
                      color: isOwner
                          ? (PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.successColor
                              : Colors.green[600]!)
                          : Colors.grey,
                    ),
                    padding: const EdgeInsets.symmetric(vertical: 16),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(12),
                    ),
                  ),
                ),
              ),
            ],
            const SizedBox(height: 24),

            // Routines
            Text(
              'Workout Routines',
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 20)
                  : Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: 12),
            ...training.routines.map((routine) {
              return Padding(
                padding: const EdgeInsets.only(bottom: 16.0),
                child: _buildRoutineCard(context, routine),
              );
            }),
          ],
        ),
      ),
    ),
    );
  }

  Widget _buildRoutineCard(BuildContext context, Routine routine) {
    return AdaptiveCard(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: Text(
                    routine.type.toUpperCase(),
                    style: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 18)
                        : Theme.of(context).textTheme.titleMedium?.copyWith(
                              fontWeight: FontWeight.bold,
                            ),
                  ),
                ),
                if (routine.rest > 0)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.orange.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(6),
                      border: Border.all(
                        color: Colors.orange.withOpacity(0.3),
                      ),
                    ),
                    child: Row(
                      children: [
                        Icon(
                          Icons.timer,
                          size: 14,
                          color: Colors.orange.shade700,
                        ),
                        const SizedBox(width: 4),
                        Text(
                          '${routine.rest}s rest',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.orange.shade700,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
                  ),
              ],
            ),
            const SizedBox(height: 16),
            ...routine.blocks.asMap().entries.map((entry) {
              final blockIndex = entry.key;
              final block = entry.value;
              return Padding(
                padding: const EdgeInsets.only(bottom: 12.0),
                child: _buildBlockCard(context, block, blockIndex + 1, routine.blocks.length),
              );
            }),
          ],
        ),
      ),
    );
  }

  Widget _buildBlockCard(BuildContext context, Block block, int blockNumber, int totalBlocks) {
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final showBlockLabel = totalBlocks > 1;

    return Container(
      decoration: BoxDecoration(
        color: PlatformHelper.useLiquidGlass
            ? Colors.white.withOpacity(0.05)
            : isDark
                ? Colors.white.withOpacity(0.08)
                : Colors.grey[50],
        borderRadius: BorderRadius.circular(12),
        border: Border.all(
          color: PlatformHelper.useLiquidGlass
              ? Colors.white.withOpacity(0.1)
              : isDark
                  ? Colors.white.withOpacity(0.15)
                  : Colors.grey[200]!,
        ),
      ),
      padding: const EdgeInsets.all(12.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          if (showBlockLabel || block.repeats > 1 || block.rest > 0)
            Row(
              children: [
                if (showBlockLabel)
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.primaryColor.withOpacity(0.15)
                          : Theme.of(context).colorScheme.secondaryContainer,
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Text(
                      'Block $blockNumber',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.primaryColor
                            : Theme.of(context).colorScheme.secondary,
                      ),
                    ),
                  ),
                const Spacer(),
              if (block.repeats > 1)
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.blue.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Row(
                    children: [
                      Icon(
                        Icons.repeat,
                        size: 12,
                        color: Colors.blue.shade700,
                      ),
                      const SizedBox(width: 4),
                      Text(
                        '${block.repeats}x',
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: Colors.blue.shade700,
                        ),
                      ),
                    ],
                  ),
                ),
              if (block.rest > 0) ...[
                const SizedBox(width: 8),
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.orange.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(4),
                  ),
                  child: Text(
                    '${block.rest}s',
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      color: Colors.orange.shade700,
                    ),
                  ),
                ),
              ],
              ],
            ),
          if (showBlockLabel || block.repeats > 1 || block.rest > 0)
            const SizedBox(height: 12),
          ...block.activities.map((activity) => _buildActivityRow(context, activity)),
        ],
      ),
    );
  }

  Widget _buildActivityRow(BuildContext context, Activity activity) {
    final exercise = _parseExercise(activity.detail);

    return Padding(
      padding: const EdgeInsets.only(bottom: 12.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Exercise image thumbnail (if available)
          if (exercise != null && _isValidImageUrl(exercise.reference)) ...[
            GestureDetector(
              onTap: () => ExerciseModal.show(context, exercise),
              child: Container(
                width: 80,
                height: 80,
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: PlatformHelper.useLiquidGlass
                        ? Colors.white.withOpacity(0.2)
                        : Theme.of(context).brightness == Brightness.dark
                            ? Colors.white.withOpacity(0.2)
                            : Colors.grey.shade300,
                  ),
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(8),
                  child: Image.network(
                    _proxyImageUrl(exercise.reference),
                    fit: BoxFit.cover,
                    errorBuilder: (context, error, stackTrace) {
                      final isDark = Theme.of(context).brightness == Brightness.dark;
                      return Container(
                        color: isDark
                            ? Colors.white.withOpacity(0.1)
                            : Colors.grey.shade200,
                        child: Icon(
                          Icons.broken_image,
                          size: 24,
                          color: isDark
                              ? Colors.white.withOpacity(0.5)
                              : Colors.grey.shade600,
                        ),
                      );
                    },
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12),
          ] else ...[
            Container(
              width: 80,
              height: 80,
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                color: PlatformHelper.useLiquidGlass
                    ? Colors.white.withOpacity(0.1)
                    : Theme.of(context).brightness == Brightness.dark
                        ? Colors.white.withOpacity(0.1)
                        : Colors.grey.shade200,
                border: Border.all(
                  color: PlatformHelper.useLiquidGlass
                      ? Colors.white.withOpacity(0.2)
                      : Theme.of(context).brightness == Brightness.dark
                          ? Colors.white.withOpacity(0.2)
                          : Colors.grey.shade300,
                ),
              ),
              child: Icon(
                Icons.fitness_center,
                size: 24,
                color: PlatformHelper.useLiquidGlass
                    ? Colors.white.withOpacity(0.4)
                    : Theme.of(context).brightness == Brightness.dark
                        ? Colors.white.withOpacity(0.4)
                        : Colors.grey.shade500,
              ),
            ),
            const SizedBox(width: 12),
          ],
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  activity.name,
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.bodyStyle.copyWith(
                          fontWeight: FontWeight.w600,
                          fontSize: 15,
                        )
                      : Theme.of(context).textTheme.bodyMedium?.copyWith(
                            fontWeight: FontWeight.w600,
                          ),
                ),
                // Exercise details from parsed detail field
                if (exercise != null &&
                    (exercise.equipment.isNotEmpty ||
                        exercise.muscles.isNotEmpty ||
                        activity.modifiers.isNotEmpty)) ...[
                  const SizedBox(height: 8),
                  _buildExerciseDetailsRow(context, exercise, activity.modifiers),
                ],
                const SizedBox(height: 6),
                Wrap(
                  spacing: 8,
                  runSpacing: 4,
                  children: [
                    if (activity.reps > 0)
                      _buildActivityTag(
                        context,
                        icon: Icons.repeat,
                        label: '${activity.reps} reps',
                        color: Colors.purple.shade400,
                      ),
                    if (activity.weightKg > 0)
                      _buildActivityTag(
                        context,
                        icon: Icons.fitness_center,
                        label: '${activity.weightKg} kg',
                        color: Colors.red.shade700,
                      ),
                    if (activity.duration > 0)
                      _buildActivityTag(
                        context,
                        icon: Icons.timer,
                        label: _formatTime(activity.duration),
                        color: Colors.blue.shade700,
                      ),
                    if (activity.rest > 0)
                      _buildActivityTag(
                        context,
                        icon: Icons.hourglass_bottom,
                        label: '${activity.rest}s rest',
                        color: Colors.orange.shade700,
                      ),
                  ],
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildExerciseDetailsRow(
      BuildContext context, Exercise exercise, List<String> modifiers) {
    final items = <Widget>[];

    if (exercise.muscles.isNotEmpty) {
      items.add(_buildDetailChip(context, Icons.accessibility_new,
          exercise.muscles.take(3).join(' · '), Colors.red.shade700));
    }
    if (exercise.equipment.isNotEmpty) {
      items.add(_buildDetailChip(
          context, Icons.fitness_center, exercise.equipment.join(' · '), Colors.blue.shade700));
    }
    if (modifiers.isNotEmpty) {
      items.add(_buildDetailChip(
          context, Icons.tune, modifiers.join(' · '), Colors.green.shade700));
    }

    return Wrap(
      spacing: 8,
      runSpacing: 4,
      children: items,
    );
  }

  Widget _buildDetailChip(BuildContext context, IconData icon, String text, Color color) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.only(top: 2),
          child: Icon(icon, size: 14, color: color),
        ),
        const SizedBox(width: 4),
        Flexible(
          child: Text(
            text,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.captionStyle.copyWith(fontSize: 12)
                : Theme.of(context).textTheme.bodySmall?.copyWith(
                      fontSize: 12,
                      color: Colors.grey.shade600,
                    ),
          ),
        ),
      ],
    );
  }

  Widget _buildInfoLabel(
    BuildContext context, {
    required IconData icon,
    required String text,
    Color? color,
  }) {
    final defaultColor = PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.captionStyle.color
        : Colors.grey.shade600;
    final labelColor = color ?? defaultColor;

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 18, color: labelColor),
        const SizedBox(width: 6),
        Text(
          text,
          style: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.captionStyle.copyWith(
                  color: labelColor,
                  fontWeight: color != null ? FontWeight.w600 : null,
                )
              : Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: labelColor,
                    fontWeight: color != null ? FontWeight.w600 : null,
                  ),
        ),
      ],
    );
  }

  Widget _buildActivityTag(
    BuildContext context, {
    required IconData icon,
    required String label,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(
          color: color.withOpacity(0.3),
          width: 1,
        ),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(
            icon,
            size: 12,
            color: color,
          ),
          const SizedBox(width: 4),
          Text(
            label,
            style: TextStyle(
              fontSize: 11,
              fontWeight: FontWeight.w600,
              color: color,
            ),
          ),
        ],
      ),
    );
  }
}

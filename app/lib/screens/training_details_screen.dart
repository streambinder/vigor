import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../config/api_config.dart';
import '../models/training.dart';
import '../models/routine.dart';
import '../models/block.dart';
import '../models/activity.dart';
import '../models/exercise.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import '../utils/exercise_modal.dart';
import 'tabata_timer_screen.dart';

class TrainingDetailsScreen extends StatelessWidget {
  final Training training;

  const TrainingDetailsScreen({
    super.key,
    required this.training,
  });

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
    final shouldDelete = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: 'Delete Training',
      content: 'Are you sure you want to delete "${training.name}"? This action cannot be undone.',
      actions: [
        AdaptiveDialogAction(
          label: 'Cancel',
          onPressed: () => Navigator.of(context).pop(false),
        ),
        AdaptiveDialogAction(
          label: 'Delete',
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
            message: 'Training deleted successfully',
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
    final storage = context.read<SecureStorageService>();
    final trainingService = TrainingService(storageService: storage);

    final response = await trainingService.completeTraining(training.id);

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

  @override
  Widget build(BuildContext context) {
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: Text(training.name),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.delete),
            tooltip: 'Delete Training',
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
                    // Workout type badge
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                      decoration: BoxDecoration(
                        color: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                            : Theme.of(context).colorScheme.primaryContainer,
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Text(
                        training.type,
                        style: TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                          color: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.primaryColor
                              : Theme.of(context).colorScheme.primary,
                        ),
                      ),
                    ),
                    const SizedBox(height: 12),
                    // Workout name
                    Text(
                      training.name,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 24)
                          : Theme.of(context).textTheme.headlineMedium,
                    ),
                    const SizedBox(height: 12),
                    Text(
                      training.description,
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.bodyStyle
                          : Theme.of(context).textTheme.bodyLarge,
                    ),
                    const SizedBox(height: 16),
                    Row(
                      children: [
                        Icon(
                          Icons.schedule,
                          size: 18,
                          color: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.captionStyle.color
                              : Colors.grey.shade600,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          _formatDuration(training.duration),
                          style: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.captionStyle
                              : Theme.of(context).textTheme.bodyMedium?.copyWith(
                                    color: Colors.grey.shade600,
                                  ),
                        ),
                        const SizedBox(width: 20),
                        Icon(
                          Icons.calendar_today,
                          size: 18,
                          color: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.captionStyle.color
                              : Colors.grey.shade600,
                        ),
                        const SizedBox(width: 6),
                        Text(
                          _formatDate(training.completedAt ?? training.createdAt),
                          style: PlatformHelper.useLiquidGlass
                              ? LiquidGlassTheme.captionStyle
                              : Theme.of(context).textTheme.bodyMedium?.copyWith(
                                    color: Colors.grey.shade600,
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
            // Mark as Complete button (only show if not already completed)
            if (training.completedAt == null) ...[
              const SizedBox(height: 12),
              SizedBox(
                width: double.infinity,
                child: OutlinedButton.icon(
                  onPressed: () => _completeTraining(context),
                  icon: const Icon(Icons.check_circle_outline),
                  label: const Text('Mark as Complete'),
                  style: OutlinedButton.styleFrom(
                    foregroundColor: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.successColor
                        : Colors.green[600],
                    side: BorderSide(
                      color: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.successColor
                          : Colors.green[600]!,
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
            ...training.routines.asMap().entries.map((entry) {
              final index = entry.key;
              final routine = entry.value;
              return Padding(
                padding: const EdgeInsets.only(bottom: 16.0),
                child: _buildRoutineCard(context, routine, index + 1),
              );
            }),
          ],
        ),
      ),
    );
  }

  Widget _buildRoutineCard(BuildContext context, Routine routine, int routineNumber) {
    return AdaptiveCard(
      child: Padding(
        padding: const EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  width: 32,
                  height: 32,
                  decoration: BoxDecoration(
                    color: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                        : Theme.of(context).colorScheme.primaryContainer,
                    shape: BoxShape.circle,
                  ),
                  child: Center(
                    child: Text(
                      routineNumber.toString(),
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.primaryColor
                            : Theme.of(context).colorScheme.primary,
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 12),
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
                child: _buildBlockCard(context, block, blockIndex + 1),
              );
            }),
          ],
        ),
      ),
    );
  }

  Widget _buildBlockCard(BuildContext context, Block block, int blockNumber) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

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
          Row(
            children: [
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
                width: 60,
                height: 60,
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
              width: 60,
              height: 60,
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
                if (exercise != null) ...[
                  const SizedBox(height: 8),
                  if (exercise.equipment.isNotEmpty)
                    _buildExerciseDetailRow(
                      context,
                      icon: Icons.fitness_center,
                      label: 'Equipment',
                      values: exercise.equipment,
                      color: Colors.blue.shade700,
                    ),
                  if (exercise.muscles.isNotEmpty)
                    _buildExerciseDetailRow(
                      context,
                      icon: Icons.accessibility_new,
                      label: 'Muscles',
                      values: exercise.muscles,
                      color: Colors.red.shade700,
                    ),
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

  Widget _buildExerciseDetailRow(
    BuildContext context, {
    required IconData icon,
    required String label,
    required List<String> values,
    required Color color,
  }) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 4.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            icon,
            size: 14,
            color: color,
          ),
          const SizedBox(width: 6),
          Expanded(
            child: Text(
              values.join(', '),
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.captionStyle.copyWith(fontSize: 12)
                  : Theme.of(context).textTheme.bodySmall?.copyWith(
                        fontSize: 12,
                        color: Colors.grey.shade600,
                      ),
            ),
          ),
        ],
      ),
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

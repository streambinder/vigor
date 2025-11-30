import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../widgets/adaptive/adaptive.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';
import '../models/training.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import 'training_details_screen.dart';
import 'tabata_timer_screen.dart';

class ActivityScreen extends StatefulWidget {
  const ActivityScreen({super.key});

  @override
  State<ActivityScreen> createState() => _ActivityScreenState();
}

class _ActivityScreenState extends State<ActivityScreen> {
  TrainingService? _trainingService;
  List<Training>? _trainings;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final storage = context.read<SecureStorageService>();
      _trainingService = TrainingService(storageService: storage);
      _loadTrainings();
    });
  }

  Future<void> _loadTrainings() async {
    if (_trainingService == null) return;

    setState(() {
      _isLoading = true;
    });

    final response = await _trainingService!.getTrainings();
    if (response.isSuccess && mounted) {
      setState(() {
        _trainings = response.data;
        _isLoading = false;
      });
    } else if (mounted) {
      setState(() {
        _isLoading = false;
      });
      if (response.error != null) {
        AdaptiveNotification.showError(
          context: context,
          message: response.error!,
        );
      }
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

  bool _isCompletedWorkout(Training training) {
    final completedAt = training.completedAt;
    if (completedAt == null) {
      return false;
    }
    final now = DateTime.now();
    return completedAt.isBefore(now);
  }

  bool _isStaleWorkout(Training training) {
    // Consider a workout stale if it's older than 7 days and not completed
    if (_isCompletedWorkout(training)) {
      return false;
    }
    final now = DateTime.now();
    final daysSinceCreation = now.difference(training.createdAt).inDays;
    return daysSinceCreation >= 7;
  }

  @override
  Widget build(BuildContext context) {
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: const Text('Activity'),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.refresh),
            tooltip: 'Refresh',
            onPressed: _loadTrainings,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _loadTrainings,
        child: _isLoading
            ? const Center(child: AdaptiveLoadingIndicator())
            : _trainings == null || _trainings!.isEmpty
                ? _buildEmptyState()
                : _buildTrainingsList(),
      ),
    );
  }

  Widget _buildEmptyState() {
    return ListView(
      padding: const EdgeInsets.all(24.0),
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.2),
        Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.fitness_center,
                size: 64,
                color: Colors.grey[400],
              ),
              const SizedBox(height: 16),
              Text(
                'No workouts yet',
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.headlineStyle
                    : Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              Text(
                'Generate your first workout from the Home tab',
                textAlign: TextAlign.center,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.captionStyle
                    : Theme.of(context).textTheme.bodyMedium?.copyWith(
                          color: Colors.grey[600],
                        ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildTrainingsList() {
    // Separate workouts into available and past
    final availableWorkouts = _trainings!
        .where((t) => !_isCompletedWorkout(t))
        .toList()
      ..sort((a, b) => b.createdAt.compareTo(a.createdAt)); // Most recent first

    final pastWorkouts = _trainings!
        .where((t) => _isCompletedWorkout(t))
        .toList()
      ..sort((a, b) {
        // Both should have non-null completedAt after filtering, but add safety
        final aDate = a.completedAt ?? a.createdAt;
        final bDate = b.completedAt ?? b.createdAt;
        return bDate.compareTo(aDate);
      }); // Most recent first

    return ListView(
      padding: const EdgeInsets.all(16.0),
      children: [
        // Available Workouts Section
        Text(
          'Available workouts',
          style: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 20)
              : Theme.of(context).textTheme.titleLarge,
        ),
        const SizedBox(height: 12),

        if (availableWorkouts.isEmpty)
          _buildEmptyAvailableState()
        else
          ...availableWorkouts.map((training) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 12.0),
              child: _buildTrainingCard(training),
            );
          }),

        const SizedBox(height: 32),

        // Past Workouts Section (only show if there are past workouts)
        if (pastWorkouts.isNotEmpty) ...[
          Text(
            'Past workouts',
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 20)
                : Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 12),
          ...pastWorkouts.map((training) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 12.0),
              child: _buildTrainingCard(training),
            );
          }),
        ],
      ],
    );
  }

  Widget _buildEmptyAvailableState() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 24.0),
      child: Center(
        child: Column(
          children: [
            Icon(
              Icons.fitness_center,
              size: 48,
              color: Colors.grey[400],
            ),
            const SizedBox(height: 12),
            Text(
              'No workout available. Start generating one.',
              textAlign: TextAlign.center,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.bodyStyle.copyWith(
                      color: Colors.grey[600],
                    )
                  : Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Colors.grey[600],
                      ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildTrainingCard(Training training) {
    final isCompleted = _isCompletedWorkout(training);
    final isStale = _isStaleWorkout(training);
    final opacity = isCompleted ? 0.5 : 1.0;

    return Opacity(
      opacity: opacity,
      child: AdaptiveCard(
        child: InkWell(
          onTap: () async {
            final changed = await Navigator.of(context).push<bool>(
              MaterialPageRoute(
                builder: (context) => TrainingDetailsScreen(training: training),
              ),
            );
            // Refresh the list if training was deleted or completed
            if (changed == true) {
              _loadTrainings();
            }
          },
          borderRadius: BorderRadius.circular(12),
          child: Padding(
            padding: const EdgeInsets.all(16.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Workout type badge
                Container(
                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                  decoration: BoxDecoration(
                    color: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.primaryColor.withOpacity(0.2)
                        : Theme.of(context).colorScheme.primaryContainer,
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Text(
                    training.type,
                    style: TextStyle(
                      fontSize: 12,
                      fontWeight: FontWeight.w600,
                      color: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.primaryColor
                          : Theme.of(context).colorScheme.primary,
                    ),
                  ),
                ),
                const SizedBox(height: 8),
                // Workout name (uppercase)
                Text(
                  training.name.toUpperCase(),
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 18)
                      : Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: 8),
                Text(
                  training.description,
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.bodyStyle
                      : Theme.of(context).textTheme.bodyMedium,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 12),
                Row(
                  children: [
                    if (isCompleted) ...[
                      Icon(
                        Icons.check_circle,
                        size: 16,
                        color: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.successColor
                            : Colors.green[600],
                      ),
                      const SizedBox(width: 4),
                      Text(
                        'Completed: ${_formatDate(training.completedAt ?? training.createdAt)}',
                        style: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.captionStyle.copyWith(
                                color: LiquidGlassTheme.successColor,
                              )
                            : Theme.of(context).textTheme.bodySmall?.copyWith(
                                  color: Colors.green[600],
                                ),
                      ),
                      const SizedBox(width: 16),
                    ],
                    if (!isCompleted && isStale) ...[
                      Icon(
                        Icons.warning,
                        size: 16,
                        color: PlatformHelper.useLiquidGlass
                            ? Colors.orange[700]
                            : Colors.orange[600],
                      ),
                      const SizedBox(width: 4),
                      Text(
                        'Stale',
                        style: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.captionStyle.copyWith(
                                color: Colors.orange[700],
                                fontWeight: FontWeight.w600,
                              )
                            : Theme.of(context).textTheme.bodySmall?.copyWith(
                                  color: Colors.orange[600],
                                  fontWeight: FontWeight.w600,
                                ),
                      ),
                      const SizedBox(width: 16),
                    ],
                    Icon(
                      Icons.calendar_today,
                      size: 16,
                      color: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.captionStyle.color
                          : Colors.grey[600],
                    ),
                    const SizedBox(width: 4),
                    Text(
                      'Created: ${_formatDate(training.createdAt)}',
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.captionStyle
                          : Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: Colors.grey[600],
                              ),
                    ),
                    const SizedBox(width: 16),
                    Icon(
                      Icons.schedule,
                      size: 16,
                      color: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.captionStyle.color
                          : Colors.grey[600],
                    ),
                    const SizedBox(width: 4),
                    Text(
                      _formatDuration(training.duration),
                      style: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.captionStyle
                          : Theme.of(context).textTheme.bodySmall?.copyWith(
                                color: Colors.grey[600],
                              ),
                    ),
                  ],
                ),
                const SizedBox(height: 12),
                // Timer button
                SizedBox(
                  width: double.infinity,
                  child: ElevatedButton.icon(
                    onPressed: () async {
                      final completed = await Navigator.of(context).push<bool>(
                        MaterialPageRoute(
                          builder: (context) => TabataTimerScreen(training: training),
                        ),
                      );
                      // Reload trainings if workout was completed
                      if (completed == true) {
                        _loadTrainings();
                      }
                    },
                    icon: const Icon(Icons.timer),
                    label: const Text('Start Workout Timer'),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: PlatformHelper.useLiquidGlass
                          ? LiquidGlassTheme.successColor
                          : Colors.green[600],
                      foregroundColor: Colors.white,
                      padding: const EdgeInsets.symmetric(vertical: 12),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../generated/app_localizations.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/training_generation_modal.dart';
import '../services/training_service.dart';
import '../services/gym_service.dart';
import '../services/secure_storage_service.dart';
import '../models/training.dart';
import '../models/gym.dart';
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
  GymService? _gymService;
  List<Training>? _trainings;
  List<Gym>? _gyms;
  bool _isLoading = false;
  bool _isLoadingGyms = false;
  Map<String, int> _partnerCounts = {};

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final storage = context.read<SecureStorageService>();
      _trainingService = TrainingService(storageService: storage);
      _gymService = GymService(storageService: storage);
      _loadTrainings();
      _loadGyms();
    });
  }

  Future<void> _loadGyms() async {
    if (_gymService == null) return;

    setState(() {
      _isLoadingGyms = true;
    });

    final response = await _gymService!.getGyms();
    if (response.isSuccess && mounted) {
      setState(() {
        _gyms = response.data;
        _isLoadingGyms = false;
      });
    } else if (mounted) {
      setState(() {
        _isLoadingGyms = false;
      });
    }
  }

  void _showTrainingGenerationModal() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (context) => TrainingGenerationModal(
        gyms: _gyms ?? [],
        onSuccess: (training) {
          _loadTrainings(); // refresh the list
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => TrainingDetailsScreen(training: training),
            ),
          );
        },
      ),
    );
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
      // load partner counts in background
      _loadPartnerCounts();
    } else if (mounted) {
      setState(() {
        _isLoading = false;
      });
      if (response.error != null) {
        AdaptiveNotification.showError(
          context: context,
          message: AppLocalizations.of(context).failedToLoadTrainings,
          rawError: response.error,
        );
      }
    }
  }

  Future<void> _loadPartnerCounts() async {
    if (_trainingService == null || _trainings == null) return;
    for (final training in _trainings!) {
      final response = await _trainingService!.getPartners(training.id);
      if (response.isSuccess && mounted) {
        setState(() {
          _partnerCounts[training.id] = response.data?.length ?? 0;
        });
      }
    }
  }

  String _formatDate(DateTime date) {
    return '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}';
  }

  String _formatDuration(AppLocalizations l10n, int seconds) {
    final minutes = seconds ~/ 60;
    if (minutes < 60) {
      return l10n.durationMin(minutes);
    }
    final hours = minutes ~/ 60;
    final remainingMinutes = minutes % 60;
    if (remainingMinutes == 0) {
      return l10n.durationHr(hours);
    }
    return l10n.durationHrMin(hours, remainingMinutes);
  }

  bool _isCompletedTraining(Training training) {
    final completedAt = training.completedAt;
    if (completedAt == null) {
      return false;
    }
    final now = DateTime.now();
    return completedAt.isBefore(now);
  }

  bool _isStaleTraining(Training training) {
    // consider a training stale if it's older than 7 days and not completed
    if (_isCompletedTraining(training)) {
      return false;
    }
    final now = DateTime.now();
    final daysSinceCreation = now.difference(training.createdAt).inDays;
    return daysSinceCreation >= 7;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        title: Text(l10n.activity),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.refresh),
            tooltip: l10n.refresh,
            onPressed: _loadTrainings,
          ),
        ],
      ),
      floatingActionButton: FloatingActionButton.extended(
        onPressed: _isLoadingGyms ? null : _showTrainingGenerationModal,
        icon: const Icon(Icons.add),
        label: Text(l10n.generateTraining),
        backgroundColor: PlatformHelper.useLiquidGlass
            ? LiquidGlassTheme.primaryColor
            : null,
      ),
      body: RefreshIndicator(
        onRefresh: _loadTrainings,
        child: _isLoading
            ? const Center(child: AdaptiveLoadingIndicator())
            : _trainings == null || _trainings!.isEmpty
                ? _buildEmptyState(l10n)
                : _buildTrainingsList(l10n),
      ),
    );
  }

  Widget _buildEmptyState(AppLocalizations l10n) {
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
                l10n.noTrainingsYet,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.headlineStyle
                    : Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              Text(
                l10n.generateFirstTraining,
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

  Widget _buildTrainingsList(AppLocalizations l10n) {
    // separate trainings into available and past
    final availableTrainings = _trainings!
        .where((t) => !_isCompletedTraining(t))
        .toList()
      ..sort((a, b) => b.createdAt.compareTo(a.createdAt)); // Most recent first

    final pastTrainings = _trainings!
        .where((t) => _isCompletedTraining(t))
        .toList()
      ..sort((a, b) {
        // Both should have non-null completedAt after filtering, but add safety
        final aDate = a.completedAt ?? a.createdAt;
        final bDate = b.completedAt ?? b.createdAt;
        return bDate.compareTo(aDate);
      }); // Most recent first

    return ListView(
      padding: const EdgeInsets.all(20.0),
      children: [
        // Available Trainings Section
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 32),
          child: Text(
            l10n.availableTrainings,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 20)
                : Theme.of(context).textTheme.titleLarge,
          ),
        ),
        const SizedBox(height: 12),

        if (availableTrainings.isEmpty)
          _buildEmptyAvailableState(l10n)
        else
          ...availableTrainings.map((training) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 12.0),
              child: _buildTrainingCard(training, l10n),
            );
          }),

        const SizedBox(height: 32),

        // Past Trainings Section (only show if there are past trainings)
        if (pastTrainings.isNotEmpty) ...[
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Text(
              l10n.pastTrainings,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 20)
                  : Theme.of(context).textTheme.titleLarge,
            ),
          ),
          const SizedBox(height: 12),
          ...pastTrainings.map((training) {
            return Padding(
              padding: const EdgeInsets.only(bottom: 12.0),
              child: _buildTrainingCard(training, l10n),
            );
          }),
        ],
      ],
    );
  }

  Widget _buildEmptyAvailableState(AppLocalizations l10n) {
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
              l10n.noTrainingAvailable,
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

  Widget _buildTrainingCard(Training training, AppLocalizations l10n) {
    final isCompleted = _isCompletedTraining(training);
    final isStale = _isStaleTraining(training);
    final partnerCount = _partnerCounts[training.id] ?? 0;
    final peopleCount = 1 + partnerCount; // owner + partners
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
                Text(
                  training.name,
                  style: PlatformHelper.useLiquidGlass
                      ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 18)
                      : Theme.of(context).textTheme.titleLarge,
                ),
                if (isCompleted) ...[
                  // for completed trainings: show completion time first, then other labels
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 12,
                    runSpacing: 4,
                    children: [
                      _buildMetaLabel(
                        context,
                        icon: Icons.check_circle,
                        text: _formatDate(training.completedAt ?? training.createdAt),
                        color: PlatformHelper.useLiquidGlass
                            ? LiquidGlassTheme.successColor
                            : Colors.green[600],
                      ),
                      if (partnerCount > 0)
                        _buildMetaLabel(context, icon: Icons.people, text: '$peopleCount'),
                      if (training.gym != null)
                        _buildMetaLabel(context, icon: Icons.location_on, text: training.gym!.name),
                      _buildMetaLabel(context, icon: Icons.fitness_center, text: '${training.equipment.length}'),
                      _buildMetaLabel(context, icon: Icons.tune, text: training.type),
                      if (training.parentId != null)
                        _buildMetaLabel(context, icon: Icons.copy, text: l10n.copied),
                      _buildMetaLabel(context, icon: Icons.schedule, text: _formatDuration(l10n, training.duration)),
                    ],
                  ),
                ] else ...[
                  // for available trainings: show full details
                  const SizedBox(height: 6),
                  Text(
                    training.description,
                    style: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.bodyStyle
                        : Theme.of(context).textTheme.bodyMedium,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 8),
                  Wrap(
                    spacing: 12,
                    runSpacing: 4,
                    children: [
                      if (partnerCount > 0)
                        _buildMetaLabel(context, icon: Icons.people, text: '$peopleCount'),
                      if (training.gym != null)
                        _buildMetaLabel(context, icon: Icons.location_on, text: training.gym!.name),
                      _buildMetaLabel(context, icon: Icons.fitness_center, text: '${training.equipment.length}'),
                      _buildMetaLabel(context, icon: Icons.tune, text: training.type),
                      if (isStale)
                        _buildMetaLabel(
                          context,
                          icon: Icons.warning,
                          text: l10n.stale,
                          color: PlatformHelper.useLiquidGlass ? Colors.orange[700] : Colors.orange[600],
                          bold: true,
                        ),
                      if (training.parentId != null)
                        _buildMetaLabel(context, icon: Icons.copy, text: l10n.copied),
                      _buildMetaLabel(context, icon: Icons.calendar_today, text: _formatDate(training.createdAt)),
                      _buildMetaLabel(context, icon: Icons.schedule, text: _formatDuration(l10n, training.duration)),
                    ],
                  ),
                  const SizedBox(height: 10),
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
                        // reload trainings if training was completed
                        if (completed == true) {
                          _loadTrainings();
                        }
                      },
                      icon: const Icon(Icons.timer),
                      label: Text(l10n.startTraining),
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
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildMetaLabel(
    BuildContext context, {
    required IconData icon,
    required String text,
    Color? color,
    bool bold = false,
  }) {
    final defaultColor = PlatformHelper.useLiquidGlass
        ? LiquidGlassTheme.captionStyle.color
        : Colors.grey[600];
    final labelColor = color ?? defaultColor;
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14, color: labelColor),
        const SizedBox(width: 3),
        Text(
          text,
          style: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.captionStyle.copyWith(
                  color: labelColor,
                  fontWeight: bold ? FontWeight.w600 : null,
                )
              : Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: labelColor,
                    fontWeight: bold ? FontWeight.w600 : null,
                  ),
        ),
      ],
    );
  }
}

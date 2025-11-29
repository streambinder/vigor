import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../widgets/adaptive/adaptive.dart';
import '../services/training_service.dart';
import '../services/secure_storage_service.dart';
import '../models/training.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';
import 'training_details_screen.dart';

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

  String _formatDuration(int minutes) {
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
      body: _isLoading
          ? const Center(child: AdaptiveLoadingIndicator())
          : _trainings == null || _trainings!.isEmpty
              ? _buildEmptyState()
              : _buildTrainingsList(),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Padding(
        padding: const EdgeInsets.all(24.0),
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
    );
  }

  Widget _buildTrainingsList() {
    return RefreshIndicator(
      onRefresh: _loadTrainings,
      child: ListView.builder(
        padding: const EdgeInsets.all(16.0),
        itemCount: _trainings!.length,
        itemBuilder: (context, index) {
          final training = _trainings![index];
          return Padding(
            padding: const EdgeInsets.only(bottom: 12.0),
            child: _buildTrainingCard(training),
          );
        },
      ),
    );
  }

  Widget _buildTrainingCard(Training training) {
    return AdaptiveCard(
      child: InkWell(
        onTap: () async {
          final deleted = await Navigator.of(context).push<bool>(
            MaterialPageRoute(
              builder: (context) => TrainingDetailsScreen(training: training),
            ),
          );
          // Refresh the list if training was deleted
          if (deleted == true) {
            _loadTrainings();
          }
        },
        borderRadius: BorderRadius.circular(12),
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
                          ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 18)
                          : Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
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
                ],
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
                  const SizedBox(width: 16),
                  Icon(
                    Icons.calendar_today,
                    size: 16,
                    color: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.captionStyle.color
                        : Colors.grey[600],
                  ),
                  const SizedBox(width: 4),
                  Text(
                    _formatDate(training.completedAt),
                    style: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.captionStyle
                        : Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey[600],
                            ),
                  ),
                  const Spacer(),
                  Text(
                    '${training.routines.length} routines',
                    style: PlatformHelper.useLiquidGlass
                        ? LiquidGlassTheme.captionStyle
                        : Theme.of(context).textTheme.bodySmall?.copyWith(
                              color: Colors.grey[600],
                            ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }
}

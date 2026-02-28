import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/auth_provider.dart';
import '../services/secure_storage_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/training_generation_modal.dart';
import '../services/service_locator.dart';
import '../models/training.dart';
import '../models/gym.dart';
import 'training_details_screen.dart';
import 'workout_timer_screen.dart';

class ActivityScreen extends StatefulWidget {
  const ActivityScreen({super.key});

  @override
  State<ActivityScreen> createState() => _ActivityScreenState();
}

class _ActivityScreenState extends State<ActivityScreen> with SingleTickerProviderStateMixin {
  bool _isLoading = false;
  bool _hasLoadedOnce = false;
  final Map<String, int> _partnerCounts = {};

  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _tabController.addListener(() {
      if (_tabController.indexIsChanging) return;
      setState(() {});
    });
  }

  @override
  void dispose() {
    _tabController.dispose();
    super.dispose();
  }

  void _showTrainingGenerationModal(List<Gym> gyms) {
    final locator = context.read<ServiceLocator>();
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (dialogContext) => ValueListenableBuilder<List<Gym>?>(
        valueListenable: locator.gymsNotifier,
        builder: (context, currentGyms, _) => TrainingGenerationModal(
          gyms: currentGyms ?? gyms,
          onSuccess: (training) {
            Navigator.of(dialogContext).push(
              MaterialPageRoute(
                builder: (context) => TrainingDetailsScreen(training: training),
              ),
            );
          },
        ),
      ),
    );
  }

  Future<void> _loadData({int retryCount = 0}) async {
    if (_isLoading) return;
    setState(() => _isLoading = true);
    _hasLoadedOnce = true;

    final storage = context.read<SecureStorageService>();
    final locator = context.read<ServiceLocator>();
    if (!await storage.hasTokens()) {
      if (retryCount < 3) {
        await Future.delayed(const Duration(milliseconds: 100));
        if (mounted) {
          setState(() => _isLoading = false);
          _hasLoadedOnce = false;
          _loadData(retryCount: retryCount + 1);
        }
        return;
      }
    }

    await Future.wait([locator.refreshTrainings(), locator.refreshGyms()]);
    if (mounted) {
      setState(() => _isLoading = false);
      _loadPartnerCounts(locator.trainingsNotifier.value);
    }
  }

  Future<void> _loadPartnerCounts(List<Training>? trainings) async {
    if (trainings == null) return;
    final trainingService = context.read<ServiceLocator>().trainingService;
    // batch all partner counts and update state once to avoid multiple rebuilds
    final Map<String, int> newCounts = {};
    for (final training in trainings) {
      final response = await trainingService.getPartners(training.id);
      if (response.isSuccess) {
        newCounts[training.id] = response.data?.length ?? 0;
      }
    }
    if (mounted && newCounts.isNotEmpty) {
      setState(() => _partnerCounts.addAll(newCounts));
    }
  }

  String _formatDate(DateTime date) {
    final now = DateTime.now();
    final diff = now.difference(date).inDays;
    if (diff == 0) return AppLocalizations.of(context).today;
    if (diff == 1) return AppLocalizations.of(context).yesterday;
    if (diff < 7) return '${diff}d ago';
    return '${date.day}/${date.month}';
  }

  String _formatFullDate(DateTime date) {
    return '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}';
  }

  String _formatDuration(int seconds) {
    final minutes = seconds ~/ 60;
    if (minutes < 60) return '${minutes}m';
    final hours = minutes ~/ 60;
    final remainingMinutes = minutes % 60;
    return remainingMinutes == 0 ? '${hours}h' : '${hours}h ${remainingMinutes}m';
  }

  bool _isCompletedTraining(Training training) {
    final completedAt = training.completedAt;
    return completedAt != null && completedAt.isBefore(DateTime.now());
  }

  bool _isStaleTraining(Training training) {
    if (_isCompletedTraining(training)) return false;
    return DateTime.now().difference(training.createdAt).inDays >= 7;
  }

  List<Training> _availableTrainings(List<Training>? trainings) => (trainings ?? [])
      .where((t) => !_isCompletedTraining(t))
      .toList()
    ..sort((a, b) => b.createdAt.compareTo(a.createdAt));

  List<Training> _pastTrainings(List<Training>? trainings) => (trainings ?? [])
      .where((t) => _isCompletedTraining(t))
      .toList()
    ..sort((a, b) => (b.completedAt ?? b.createdAt).compareTo(a.completedAt ?? a.createdAt));

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    final locator = context.read<ServiceLocator>();

    final authState = context.watch<AuthProvider>().state;
    if (authState == AuthState.authenticated && !_hasLoadedOnce && !_isLoading) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _loadData());
    }

    return ValueListenableBuilder<List<Training>?>(
      valueListenable: locator.trainingsNotifier,
      builder: (context, trainings, _) => ValueListenableBuilder<List<Gym>?>(
        valueListenable: locator.gymsNotifier,
        builder: (context, gyms, _) => AdaptiveScaffold(
          appBar: AdaptiveAppBar(
            title: Text(l10n.activity),
            actions: [
              AdaptiveIconButton(
                icon: const Icon(Icons.refresh),
                tooltip: l10n.refresh,
                onPressed: _loadData,
              ),
            ],
          ),
          body: RefreshIndicator(
            onRefresh: _loadData,
            color: VigorColors.persimmon,
            child: _isLoading
                ? const Center(child: AdaptiveLoadingIndicator())
                : trainings == null || trainings.isEmpty
                    ? _buildEmptyState(l10n, gyms ?? [])
                    : _buildContent(l10n, isDark, trainings),
          ),
        ),
      ),
    );
  }

  // kanso: minimal empty state, no decorative gradients
  Widget _buildEmptyState(AppLocalizations l10n, List<Gym> gyms) {
    return ListView(
      padding: VigorSpacing.paddingLg,
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.15),
        Center(
          child: Icon(
            Icons.fitness_center,
            size: 56,
            color: VigorColors.stone,
          ),
        ),
        const SizedBox(height: VigorSpacing.lg),
        Text(
          l10n.noTrainingsYet,
          style: VigorTypography.title.copyWith(color: VigorColors.textPrimary(context)),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: VigorSpacing.sm),
        Text(
          l10n.generateFirstTraining,
          textAlign: TextAlign.center,
          style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context)),
        ),
        const SizedBox(height: VigorSpacing.xl),
        // solid persimmon CTA button
        Center(
          child: Material(
            color: VigorColors.persimmon,
            borderRadius: VigorRadius.radiusSm,
            child: InkWell(
              onTap: () => _showTrainingGenerationModal(gyms),
              borderRadius: VigorRadius.radiusSm,
              child: Padding(
                padding: VigorSpacing.buttonPadding,
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Icon(Icons.bolt, color: Colors.white),
                    const SizedBox(width: 8),
                    Text(
                      l10n.generateTraining,
                      style: VigorTypography.label.copyWith(
                        color: Colors.white,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildContent(AppLocalizations l10n, bool isDark, List<Training> trainings) {
    final available = _availableTrainings(trainings);
    final past = _pastTrainings(trainings);

    return Column(
      children: [
        _buildSegmentedControl(l10n, available.length, past.length, isDark),
        Expanded(
          child: TabBarView(
            controller: _tabController,
            children: [
              _buildTrainingList(available, l10n, isAvailable: true),
              _buildTrainingList(past, l10n, isAvailable: false),
            ],
          ),
        ),
      ],
    );
  }

  // seijaku: calm segmented control, no gradient indicator
  Widget _buildSegmentedControl(AppLocalizations l10n, int available, int past, bool isDark) {
    return Container(
      margin: const EdgeInsets.fromLTRB(VigorSpacing.lg, VigorSpacing.md, VigorSpacing.lg, VigorSpacing.md),
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusSm,
      ),
      child: TabBar(
        controller: _tabController,
        indicator: BoxDecoration(
          color: VigorColors.indigoAdaptive(context),
          borderRadius: VigorRadius.radiusSm,
        ),
        indicatorSize: TabBarIndicatorSize.tab,
        dividerHeight: 0,
        labelColor: Colors.white,
        unselectedLabelColor: VigorColors.textSecondary(context),
        labelStyle: VigorTypography.label.copyWith(fontWeight: FontWeight.w600),
        unselectedLabelStyle: VigorTypography.label,
        tabs: [
          Tab(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.play_arrow, size: 18),
                const SizedBox(width: 6),
                Text('$available ${l10n.available}'),
              ],
            ),
          ),
          Tab(
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.check_circle, size: 18),
                const SizedBox(width: 6),
                Text('$past ${l10n.completed}'),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTrainingList(List<Training> trainings, AppLocalizations l10n, {required bool isAvailable}) {
    if (trainings.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              isAvailable ? Icons.fitness_center : Icons.history,
              size: 48,
              color: VigorColors.stone,
            ),
            const SizedBox(height: VigorSpacing.sm),
            Text(
              isAvailable ? l10n.noTrainingAvailable : l10n.noPastTrainings,
              style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context)),
            ),
          ],
        ),
      );
    }

    return ListView.separated(
      padding: VigorSpacing.paddingLg,
      itemCount: trainings.length,
      separatorBuilder: (_, __) => const SizedBox(height: VigorSpacing.sm),
      itemBuilder: (context, index) {
        final training = trainings[index];
        return _buildTrainingCard(training, l10n, isAvailable: isAvailable, key: ValueKey(training.id));
      },
    );
  }

  Widget _buildTrainingCard(Training training, AppLocalizations l10n, {required bool isAvailable, Key? key}) {
    final isStale = _isStaleTraining(training);
    final isCompleted = _isCompletedTraining(training);
    final partnerCount = _partnerCounts[training.id] ?? 0;

    return Container(
      key: key,
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusMd,
      ),
      child: InkWell(
        onTap: () => Navigator.of(context).push(
          MaterialPageRoute(builder: (context) => TrainingDetailsScreen(training: training)),
        ),
        borderRadius: VigorRadius.radiusMd,
        child: Padding(
          padding: VigorSpacing.paddingMd,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // header row
              Row(
                children: [
                  // methodology badge
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusXs,
                    ),
                    child: Text(
                      training.methodology.toUpperCase(),
                      style: VigorTypography.caption.copyWith(
                        color: VigorColors.stone,
                        fontWeight: FontWeight.w600,
                        fontSize: 10,
                      ),
                    ),
                  ),
                  const SizedBox(width: VigorSpacing.sm),
                  Expanded(
                    child: Text(
                      training.name,
                      style: VigorTypography.headline.copyWith(
                        fontSize: 16,
                        color: VigorColors.textPrimary(context),
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  // status badges
                  if (isCompleted)
                    _buildStatusBadge(l10n.completed, VigorColors.gold, Icons.check_circle),
                  if (isStale)
                    _buildStatusBadge(l10n.stale, VigorColors.stone, Icons.schedule),
                ],
              ),
              if (isAvailable) ...[
                const SizedBox(height: VigorSpacing.xs),
                Text(
                  training.description,
                  style: VigorTypography.body.copyWith(
                    color: VigorColors.textSecondary(context),
                    fontSize: 13,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ],
              const SizedBox(height: VigorSpacing.sm),
              // metadata row - use data typography for durations/counts
              Row(
                children: [
                  _buildDataChip(Icons.schedule, _formatDuration(training.completedIn ?? training.duration)),
                  const SizedBox(width: VigorSpacing.sm),
                  _buildDataChip(Icons.calendar_today, isAvailable
                      ? _formatDate(training.createdAt)
                      : _formatFullDate(training.completedAt ?? training.createdAt)),
                  if (partnerCount > 0) ...[
                    const SizedBox(width: VigorSpacing.sm),
                    _buildDataChip(Icons.people, '${1 + partnerCount}'),
                  ],
                  if (training.gym != null) ...[
                    const SizedBox(width: VigorSpacing.sm),
                    Flexible(child: _buildDataChip(Icons.location_on, training.gym!.name)),
                  ],
                ],
              ),
              // start button - indigo for secondary action
              if (isAvailable) ...[
                const SizedBox(height: VigorSpacing.md),
                SizedBox(
                  width: double.infinity,
                  child: Material(
                    color: VigorColors.indigo,
                    borderRadius: VigorRadius.radiusSm,
                    child: InkWell(
                      onTap: () => Navigator.of(context).push(
                        MaterialPageRoute(builder: (context) => WorkoutTimerScreen(training: training)),
                      ),
                      borderRadius: VigorRadius.radiusSm,
                      child: Padding(
                        padding: const EdgeInsets.symmetric(vertical: VigorSpacing.sm),
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            const Icon(Icons.play_arrow, color: Colors.white, size: 20),
                            const SizedBox(width: 6),
                            Text(
                              l10n.startTraining,
                              style: VigorTypography.label.copyWith(
                                color: Colors.white,
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  // kintsugi: gold for completed, stone for stale
  Widget _buildStatusBadge(String text, Color color, IconData icon) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: VigorRadius.radiusXs,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: color),
          const SizedBox(width: 2),
          Text(
            text,
            style: VigorTypography.caption.copyWith(
              color: color,
              fontWeight: FontWeight.w600,
              fontSize: 10,
            ),
          ),
        ],
      ),
    );
  }

  // use VigorTypography.data for durations and counts per IDENTITY.md
  Widget _buildDataChip(IconData icon, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
      decoration: BoxDecoration(
        color: VigorColors.stone.withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusXs,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: VigorColors.stone),
          const SizedBox(width: 4),
          Flexible(
            child: Text(
              text,
              style: VigorTypography.data.copyWith(
                color: VigorColors.textSecondary(context),
                fontSize: 11,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      ),
    );
  }
}

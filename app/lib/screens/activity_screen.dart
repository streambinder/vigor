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
import 'tabata_timer_screen.dart';

class ActivityScreen extends StatefulWidget {
  const ActivityScreen({super.key});

  @override
  State<ActivityScreen> createState() => _ActivityScreenState();
}

class _ActivityScreenState extends State<ActivityScreen> with SingleTickerProviderStateMixin {
  List<Training>? _trainings;
  List<Gym>? _gyms;
  bool _isLoading = false;
  bool _isLoadingGyms = false;
  bool _hasLoadedOnce = false;
  final Map<String, int> _partnerCounts = {};

  // tab controller for available/past toggle
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

  Future<void> _loadGyms() async {
    setState(() => _isLoadingGyms = true);
    final response = await context.read<ServiceLocator>().gymService.getGyms();
    if (response.isSuccess && mounted) {
      setState(() {
        _gyms = response.data;
        _isLoadingGyms = false;
      });
    } else if (mounted) {
      setState(() => _isLoadingGyms = false);
    }
  }

  void _showTrainingGenerationModal() {
    showDialog(
      context: context,
      barrierDismissible: true,
      builder: (context) => TrainingGenerationModal(
        gyms: _gyms ?? [],
        onSuccess: (training) {
          _loadTrainings();
          Navigator.of(context).push(
            MaterialPageRoute(
              builder: (context) => TrainingDetailsScreen(training: training),
            ),
          );
        },
      ),
    );
  }

  Future<void> _loadTrainings({int retryCount = 0}) async {
    if (_isLoading) return;
    setState(() => _isLoading = true);
    _hasLoadedOnce = true;

    // on web, storage may need a moment to persist after login
    final storage = context.read<SecureStorageService>();
    final trainingService = context.read<ServiceLocator>().trainingService;
    if (!await storage.hasTokens()) {
      if (retryCount < 3) {
        await Future.delayed(const Duration(milliseconds: 100));
        if (mounted) {
          setState(() => _isLoading = false);
          _hasLoadedOnce = false;
          _loadTrainings(retryCount: retryCount + 1);
        }
        return;
      }
    }

    final response = await trainingService.getTrainings();
    if (response.isSuccess && mounted) {
      setState(() {
        _trainings = response.data;
        _isLoading = false;
      });
      _loadPartnerCounts();
    } else if (mounted) {
      setState(() => _isLoading = false);
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
    if (_trainings == null) return;
    final trainingService = context.read<ServiceLocator>().trainingService;
    for (final training in _trainings!) {
      final response = await trainingService.getPartners(training.id);
      if (response.isSuccess && mounted) {
        setState(() => _partnerCounts[training.id] = response.data?.length ?? 0);
      }
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

  List<Training> get _availableTrainings => (_trainings ?? [])
      .where((t) => !_isCompletedTraining(t))
      .toList()
    ..sort((a, b) => b.createdAt.compareTo(a.createdAt));

  List<Training> get _pastTrainings => (_trainings ?? [])
      .where((t) => _isCompletedTraining(t))
      .toList()
    ..sort((a, b) => (b.completedAt ?? b.createdAt).compareTo(a.completedAt ?? a.createdAt));

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    // watch auth state to trigger load when authenticated
    final authState = context.watch<AuthProvider>().state;
    if (authState == AuthState.authenticated && !_hasLoadedOnce && !_isLoading) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        _loadTrainings();
        _loadGyms();
      });
    }

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
      floatingActionButton: _buildFAB(l10n),
      body: RefreshIndicator(
        onRefresh: _loadTrainings,
        color: VigorColors.orange,
        child: _isLoading
            ? const Center(child: AdaptiveLoadingIndicator())
            : _trainings == null || _trainings!.isEmpty
                ? _buildEmptyState(l10n)
                : _buildContent(l10n, isDark),
      ),
    );
  }

  Widget _buildFAB(AppLocalizations l10n) {
    return Container(
      decoration: BoxDecoration(
        gradient: const LinearGradient(
          colors: [VigorColors.orange, VigorColors.electricBlue],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16),
        boxShadow: VigorShadows.orangeGlow,
      ),
      child: Material(
        color: Colors.transparent,
        child: InkWell(
          onTap: _isLoadingGyms ? null : _showTrainingGenerationModal,
          borderRadius: BorderRadius.circular(16),
          child: Padding(
            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.bolt, color: Colors.white, size: 22),
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
    );
  }

  Widget _buildEmptyState(AppLocalizations l10n) {
    return ListView(
      padding: VigorSpacing.paddingLg,
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.15),
        // gradient icon
        Center(
          child: Container(
            width: 120,
            height: 120,
            decoration: BoxDecoration(
              gradient: LinearGradient(
                colors: [
                  VigorColors.orange.withValues(alpha: 0.2),
                  VigorColors.electricBlue.withValues(alpha: 0.2),
                ],
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
              ),
              shape: BoxShape.circle,
            ),
            child: ShaderMask(
              shaderCallback: (bounds) => const LinearGradient(
                colors: [VigorColors.orange, VigorColors.electricBlue],
              ).createShader(bounds),
              child: const Icon(Icons.fitness_center, size: 56, color: Colors.white),
            ),
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
        // CTA button
        Center(
          child: AdaptiveButton(
            onPressed: _isLoadingGyms ? null : _showTrainingGenerationModal,
            useGradient: true,
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                const Icon(Icons.bolt, color: Colors.white),
                const SizedBox(width: 8),
                Text(l10n.generateTraining, style: const TextStyle(color: Colors.white)),
              ],
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildContent(AppLocalizations l10n, bool isDark) {
    final availableCount = _availableTrainings.length;
    final pastCount = _pastTrainings.length;

    return Column(
      children: [
        // stats header
        _buildStatsHeader(l10n, availableCount, pastCount),
        // segmented control tabs
        _buildSegmentedControl(l10n, availableCount, pastCount, isDark),
        // training list
        Expanded(
          child: TabBarView(
            controller: _tabController,
            children: [
              _buildTrainingList(_availableTrainings, l10n, isAvailable: true),
              _buildTrainingList(_pastTrainings, l10n, isAvailable: false),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildStatsHeader(AppLocalizations l10n, int available, int past) {
    return Container(
      padding: const EdgeInsets.fromLTRB(VigorSpacing.lg, VigorSpacing.sm, VigorSpacing.lg, VigorSpacing.md),
      child: Row(
        children: [
          _buildStatCard(
            count: available,
            label: l10n.availableTrainings,
            color: VigorColors.orange,
            icon: Icons.play_arrow,
          ),
          const SizedBox(width: VigorSpacing.md),
          _buildStatCard(
            count: past,
            label: l10n.pastTrainings,
            color: VigorColors.success,
            icon: Icons.check_circle,
          ),
        ],
      ),
    );
  }

  Widget _buildStatCard({
    required int count,
    required String label,
    required Color color,
    required IconData icon,
  }) {
    return Expanded(
      child: Container(
        padding: VigorSpacing.paddingMd,
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.1),
          borderRadius: VigorRadius.radiusMd,
          border: Border.all(color: color.withValues(alpha: 0.2)),
        ),
        child: Row(
          children: [
            Icon(icon, color: color, size: 28),
            const SizedBox(width: VigorSpacing.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    '$count',
                    style: VigorTypography.title.copyWith(
                      color: color,
                      fontWeight: FontWeight.w700,
                    ),
                  ),
                  Text(
                    label,
                    style: VigorTypography.caption.copyWith(
                      color: color.withValues(alpha: 0.8),
                    ),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSegmentedControl(AppLocalizations l10n, int available, int past, bool isDark) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: VigorSpacing.lg),
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusSm,
        border: Border.all(
          color: isDark ? VigorColors.darkBorder : VigorColors.lightBorder,
        ),
      ),
      child: TabBar(
        controller: _tabController,
        indicator: const BoxDecoration(
          gradient: LinearGradient(
            colors: [VigorColors.orange, VigorColors.electricBlue],
          ),
          borderRadius: VigorRadius.radiusSm,
        ),
        indicatorSize: TabBarIndicatorSize.tab,
        dividerHeight: 0,
        labelColor: Colors.white,
        unselectedLabelColor: VigorColors.textSecondary(context),
        labelStyle: VigorTypography.label.copyWith(fontWeight: FontWeight.w600),
        unselectedLabelStyle: VigorTypography.label,
        tabs: [
          Tab(text: '${l10n.available} ($available)'),
          Tab(text: '${l10n.completed} ($past)'),
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
              color: VigorColors.textMuted(context),
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
      separatorBuilder: (_, _) => const SizedBox(height: VigorSpacing.sm),
      itemBuilder: (context, index) {
        final training = trainings[index];
        return _buildTrainingCard(training, l10n, isAvailable: isAvailable, key: ValueKey(training.id));
      },
    );
  }

  Widget _buildTrainingCard(Training training, AppLocalizations l10n, {required bool isAvailable, Key? key}) {
    final isStale = _isStaleTraining(training);
    final partnerCount = _partnerCounts[training.id] ?? 0;

    return AdaptiveCard(
      key: key,
      child: InkWell(
        onTap: () async {
          final changed = await Navigator.of(context).push<bool>(
            MaterialPageRoute(builder: (context) => TrainingDetailsScreen(training: training)),
          );
          if (changed == true) _loadTrainings();
        },
        borderRadius: VigorRadius.radiusMd,
        child: Padding(
          padding: VigorSpacing.paddingMd,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // header row with name and status
              Row(
                children: [
                  // type indicator
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: VigorColors.electricBlue.withValues(alpha: 0.15),
                      borderRadius: VigorRadius.radiusXs,
                    ),
                    child: Text(
                      training.type.toUpperCase(),
                      style: VigorTypography.caption.copyWith(
                        color: VigorColors.electricBlue,
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
                  if (isStale)
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                      decoration: BoxDecoration(
                        color: VigorColors.warning.withValues(alpha: 0.15),
                        borderRadius: VigorRadius.radiusXs,
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          const Icon(Icons.schedule, size: 12, color: VigorColors.warning),
                          const SizedBox(width: 2),
                          Text(
                            l10n.stale,
                            style: VigorTypography.caption.copyWith(
                              color: VigorColors.warning,
                              fontWeight: FontWeight.w600,
                              fontSize: 10,
                            ),
                          ),
                        ],
                      ),
                    ),
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
              // metadata row
              Row(
                children: [
                  _buildChip(Icons.schedule, _formatDuration(training.duration)),
                  const SizedBox(width: VigorSpacing.sm),
                  _buildChip(Icons.calendar_today, _formatDate(
                    isAvailable ? training.createdAt : (training.completedAt ?? training.createdAt),
                  )),
                  if (partnerCount > 0) ...[
                    const SizedBox(width: VigorSpacing.sm),
                    _buildChip(Icons.people, '${1 + partnerCount}'),
                  ],
                  if (training.gym != null) ...[
                    const SizedBox(width: VigorSpacing.sm),
                    Flexible(child: _buildChip(Icons.location_on, training.gym!.name)),
                  ],
                ],
              ),
              // start button for available trainings
              if (isAvailable) ...[
                const SizedBox(height: VigorSpacing.md),
                SizedBox(
                  width: double.infinity,
                  child: Container(
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: [VigorColors.success, VigorColors.success.withValues(alpha: 0.8)],
                      ),
                      borderRadius: VigorRadius.radiusSm,
                    ),
                    child: Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () async {
                          final completed = await Navigator.of(context).push<bool>(
                            MaterialPageRoute(builder: (context) => TabataTimerScreen(training: training)),
                          );
                          if (completed == true) _loadTrainings();
                        },
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
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildChip(IconData icon, String text) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
      decoration: BoxDecoration(
        color: VigorColors.textMuted(context).withValues(alpha: 0.1),
        borderRadius: VigorRadius.radiusXs,
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, size: 12, color: VigorColors.textSecondary(context)),
          const SizedBox(width: 4),
          Flexible(
            child: Text(
              text,
              style: VigorTypography.caption.copyWith(
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

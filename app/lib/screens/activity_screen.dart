import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:provider/provider.dart';
import 'package:share_plus/share_plus.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../dto/partner_info.dart';
import '../providers/auth_provider.dart';
import '../services/secure_storage_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../utils/knowledge_labels.dart';
import '../widgets/training_generation_modal.dart';
import '../services/service_locator.dart';
import '../models/training.dart';
import '../models/gym.dart';
import 'training_details_screen.dart';

class ActivityScreen extends StatefulWidget {
  const ActivityScreen({super.key});

  @override
  State<ActivityScreen> createState() => _ActivityScreenState();
}

class _ActivityScreenState extends State<ActivityScreen> with SingleTickerProviderStateMixin {
  bool _isLoading = false;
  bool _hasLoadedOnce = false;
  bool _isDeleting = false;
  bool _isSharing = false;
  final Map<String, List<PartnerInfo>> _partnerData = {};
  final Set<String> _selectedIds = {};
  bool get _isSelecting => _selectedIds.isNotEmpty;

  late TabController _tabController;

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 2, vsync: this);
    _tabController.addListener(() {
      if (_tabController.indexIsChanging) return;
      setState(() => _selectedIds.clear());
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
    // batch all partner data and update state once to avoid multiple rebuilds
    final Map<String, List<PartnerInfo>> newData = {};
    for (final training in trainings) {
      final response = await trainingService.getPartners(training.id);
      if (response.isSuccess) {
        newData[training.id] = response.data ?? [];
      }
    }
    if (mounted && newData.isNotEmpty) {
      setState(() => _partnerData.addAll(newData));
    }
  }

  Future<void> _deleteSelected() async {
    final l10n = AppLocalizations.of(context);
    final count = _selectedIds.length;

    final confirmed = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.deleteTraining,
      content: l10n.deleteSelectedTrainings(count),
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.delete, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (confirmed != true || !mounted) return;
    setState(() => _isDeleting = true);

    final trainingService = context.read<ServiceLocator>().trainingService;
    bool hadError = false;
    for (final id in _selectedIds.toList()) {
      final response = await trainingService.deleteTraining(id);
      if (!response.isSuccess) hadError = true;
    }

    if (!mounted) return;
    setState(() {
      _selectedIds.clear();
      _isDeleting = false;
    });
    await context.read<ServiceLocator>().refreshTrainings();
    if (mounted) {
      if (hadError) {
        AdaptiveNotification.showError(context: context, message: l10n.failedToDeleteTraining);
      } else {
        AdaptiveNotification.show(context: context, message: l10n.trainingsDeletedSuccessfully);
      }
    }
  }

  Future<void> _shareSelected() async {
    final l10n = AppLocalizations.of(context);
    final trainingId = _selectedIds.first;
    setState(() => _isSharing = true);

    final response = await context.read<ServiceLocator>().trainingService.shareTraining(trainingId);

    if (!mounted) return;
    setState(() => _isSharing = false);

    if (response.isSuccess && response.data != null) {
      final url = response.data!['url']!;
      try {
        await Share.share(url);
      } catch (_) {
        await Clipboard.setData(ClipboardData(text: url));
        if (mounted) AdaptiveNotification.show(context: context, message: l10n.linkCopied);
      }
    } else {
      AdaptiveNotification.showError(context: context, message: l10n.failedToShareTraining, rawError: response.error);
    }
  }

  String _formatDate(DateTime date) {
    final l10n = AppLocalizations.of(context);
    final diff = DateTime.now().difference(date).inDays;
    if (diff == 0) return l10n.today;
    if (diff == 1) return l10n.yesterday;
    if (diff < 7) return l10n.daysAgo(diff);
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
        builder: (context, gyms, _) => PopScope(
          canPop: !_isSelecting,
          onPopInvokedWithResult: (didPop, _) {
            if (!didPop) setState(() => _selectedIds.clear());
          },
          child: AdaptiveScaffold(
            appBar: _isSelecting
                ? AdaptiveAppBar(
                    leading: AdaptiveIconButton(
                      icon: const Icon(Icons.close),
                      onPressed: () => setState(() => _selectedIds.clear()),
                    ),
                    title: Text(l10n.nSelected(_selectedIds.length)),
                    actions: [
                      if (_selectedIds.length == 1)
                        _isSharing
                            ? const Padding(
                                padding: EdgeInsets.all(12),
                                child: SizedBox(width: 20, height: 20, child: AdaptiveLoadingIndicator()),
                              )
                            : AdaptiveIconButton(
                                icon: const Icon(Icons.share),
                                tooltip: l10n.share,
                                onPressed: _shareSelected,
                              ),
                      if (_isDeleting)
                        const Padding(
                          padding: EdgeInsets.all(12),
                          child: SizedBox(width: 20, height: 20, child: AdaptiveLoadingIndicator()),
                        )
                      else
                        AdaptiveIconButton(
                          icon: const Icon(Icons.delete),
                          tooltip: l10n.delete,
                          onPressed: _deleteSelected,
                        ),
                    ],
                  )
                : AdaptiveAppBar(
                    automaticallyImplyLeading: false,
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
    final currentUserId = context.read<AuthProvider>().currentUser?.id ?? '';
    final internalPastCount = past.where((t) => t.userId == currentUserId).length;

    return Column(
      children: [
        _buildSegmentedControl(l10n, available.length, internalPastCount, isDark),
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
    final externalSessions = !isAvailable ? _getExternalSessions() : <Map<String, dynamic>>[];

    if (trainings.isEmpty && externalSessions.isEmpty) {
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

    // for available tab, just list trainings
    if (isAvailable) {
      return ListView.separated(
        padding: VigorSpacing.paddingLg,
        itemCount: trainings.length,
        separatorBuilder: (_, __) => const SizedBox(height: VigorSpacing.sm),
        itemBuilder: (context, index) {
          final training = trainings[index];
          return _buildTrainingCard(training, l10n, isAvailable: true, key: ValueKey(training.id));
        },
      );
    }

    // completed tab: merge trainings and day-grouped external sessions, sort chronologically
    // group external sessions by day into a single item per date
    final Map<String, List<Map<String, dynamic>>> sessionsByDay = {};
    for (final s in externalSessions) {
      final date = DateTime.parse(s['started_at'] as String);
      final key = '${date.year}-${date.month}-${date.day}';
      (sessionsByDay[key] ??= []).add(s);
    }

    final List<_CompletedItem> items = [
      ...trainings.map((t) => _CompletedItem(
        date: t.completedAt ?? t.createdAt,
        training: t,
      )),
      ...sessionsByDay.entries.map((e) {
        // use the latest session's start time as the group date
        final sorted = e.value..sort((a, b) =>
          DateTime.parse(b['started_at'] as String).compareTo(DateTime.parse(a['started_at'] as String)));
        return _CompletedItem(
          date: DateTime.parse(sorted.first['started_at'] as String),
          externalSessions: sorted,
        );
      }),
    ]..sort((a, b) => b.date.compareTo(a.date));

    return ListView.separated(
      padding: VigorSpacing.paddingLg,
      itemCount: items.length,
      separatorBuilder: (_, __) => const SizedBox(height: VigorSpacing.sm),
      itemBuilder: (context, index) {
        final item = items[index];
        if (item.training != null) {
          return _buildTrainingCard(item.training!, l10n, isAvailable: false, key: ValueKey(item.training!.id));
        }
        return _buildExternalSessionsCard(item.externalSessions!, l10n);
      },
    );
  }

  List<Map<String, dynamic>> _getExternalSessions() {
    final healthDaily = context.read<ServiceLocator>().healthDailyNotifier.value;
    if (healthDaily == null) return [];
    final sessions = healthDaily['sessions'] as List?;
    if (sessions == null) return [];
    return sessions.cast<Map<String, dynamic>>();
  }

  /// day-grouped external sessions card — merges all sessions for the same day
  Widget _buildExternalSessionsCard(List<Map<String, dynamic>> sessions, AppLocalizations l10n) {
    final date = DateTime.parse(sessions.first['started_at'] as String);
    final totalMins = sessions.fold<int>(0, (sum, s) {
      final start = DateTime.parse(s['started_at'] as String);
      final end = DateTime.parse(s['ended_at'] as String);
      return sum + end.difference(start).inMinutes;
    });

    // collect unique exercise types for the summary
    final types = sessions
        .map((s) => _localizedExerciseType(s['exercise_type'] as String? ?? 'workout', l10n))
        .toSet()
        .toList();

    return Container(
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusMd,
      ),
      child: Opacity(
        opacity: 0.7,
        child: Padding(
          padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  sessions.length == 1
                      ? types.first
                      : '${sessions.length}x — ${types.join(', ')}',
                  style: VigorTypography.body.copyWith(
                    color: VigorColors.stone,
                    fontSize: 13,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              const SizedBox(width: VigorSpacing.sm),
              Text(
                _formatDuration(totalMins * 60),
                style: VigorTypography.data.copyWith(
                  color: VigorColors.stone,
                  fontSize: 11,
                ),
              ),
              const SizedBox(width: VigorSpacing.sm),
              Text(
                _formatFullDate(date),
                style: VigorTypography.data.copyWith(
                  color: VigorColors.stone,
                  fontSize: 11,
                ),
              ),
              const Padding(
                padding: EdgeInsets.only(left: VigorSpacing.xs),
                child: Icon(Icons.monitor_heart, size: 16, color: VigorColors.stone),
              ),
            ],
          ),
        ),
      ),
    );
  }

  String _localizedExerciseType(String type, AppLocalizations l10n) {
    final labels = {
      'running': l10n.exerciseTypeRunning,
      'walking': l10n.exerciseTypeWalking,
      'biking': l10n.exerciseTypeBiking,
      'yoga': l10n.exerciseTypeYoga,
      'swimming': l10n.exerciseTypeSwimming,
      'hiking': l10n.exerciseTypeHiking,
      'strength_training': l10n.exerciseTypeStrengthTraining,
      'functional_strength_training': l10n.exerciseTypeFunctionalStrengthTraining,
      'traditional_strength_training': l10n.exerciseTypeTraditionalStrengthTraining,
      'running_treadmill': l10n.exerciseTypeRunningTreadmill,
      'biking_stationary': l10n.exerciseTypeBikingStationary,
      'walking_treadmill': l10n.exerciseTypeWalkingTreadmill,
      'rowing': l10n.exerciseTypeRowing,
      'pilates': l10n.exerciseTypePilates,
      'dancing': l10n.exerciseTypeDancing,
      'elliptical': l10n.exerciseTypeElliptical,
      'stair_climbing': l10n.exerciseTypeStairClimbing,
    };
    return labels[type.toLowerCase()] ?? _capitalizeExerciseType(type);
  }

  String _capitalizeExerciseType(String type) {
    if (type.isEmpty) return type;
    // "strength_training" → "Strength Training"
    return type.split('_').map((w) => w.isEmpty ? w : '${w[0].toUpperCase()}${w.substring(1)}').join(' ');
  }

  Widget _buildTrainingCard(Training training, AppLocalizations l10n, {required bool isAvailable, Key? key}) {
    final isStale = _isStaleTraining(training);
    final partners = _partnerData[training.id] ?? [];
    final partnerCount = partners.length;
    final isSelected = _selectedIds.contains(training.id);

    return Container(
      key: key,
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusMd,
        border: isSelected ? Border.all(color: VigorColors.indigo, width: 2) : null,
      ),
      child: InkWell(
        onTap: _isSelecting
            ? () => setState(() {
                  if (isSelected) {
                    _selectedIds.remove(training.id);
                  } else {
                    _selectedIds.add(training.id);
                  }
                })
            : () => Navigator.of(context).push(
                  MaterialPageRoute(builder: (context) => TrainingDetailsScreen(training: training)),
                ),
        onLongPress: _isSelecting
            ? null
            : () => setState(() => _selectedIds.add(training.id)),
        borderRadius: VigorRadius.radiusMd,
        child: Padding(
          padding: VigorSpacing.paddingMd,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // header row
              Row(
                children: [
                  if (isSelected)
                    Padding(
                      padding: const EdgeInsets.only(right: VigorSpacing.sm),
                      child: Icon(Icons.check_circle, size: 20, color: VigorColors.indigo),
                    ),
                  // methodology badge
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                    decoration: BoxDecoration(
                      color: VigorColors.stone.withValues(alpha: 0.1),
                      borderRadius: VigorRadius.radiusXs,
                    ),
                    child: Text(
                      KnowledgeLabels.methodologyLabel(training.methodology, l10n).toUpperCase(),
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
                  if (training.hasHealthSession)
                    Padding(
                      padding: const EdgeInsets.only(left: VigorSpacing.xs),
                      child: ShaderMask(
                        shaderCallback: (bounds) => const LinearGradient(
                          colors: [VigorColors.persimmon, VigorColors.crimson],
                        ).createShader(bounds),
                        child: const Icon(Icons.monitor_heart, size: 16, color: Colors.white),
                      ),
                    ),
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
                    _buildDataChip(Icons.people, partnerCount == 1
                        ? partners.first.firstName
                        : '${1 + partnerCount}'),
                  ],
                  if (training.gym != null) ...[
                    const SizedBox(width: VigorSpacing.sm),
                    Flexible(child: _buildDataChip(Icons.location_on, training.gym!.name)),
                  ],
                ],
              ),
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

/// unified item for chronological sorting of vigor trainings and day-grouped external HC sessions
class _CompletedItem {
  final DateTime date;
  final Training? training;
  final List<Map<String, dynamic>>? externalSessions;
  const _CompletedItem({required this.date, this.training, this.externalSessions});
}
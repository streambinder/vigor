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

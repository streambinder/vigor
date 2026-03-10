import 'dart:math';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/family_progress.dart';
import '../models/muscle_impact.dart';
import '../models/weekly_target.dart';
import '../models/week_progress.dart';
import '../models/week_summary.dart';
import '../providers/auth_provider.dart';
import '../services/preferences_service.dart';
import '../services/secure_storage_service.dart';
import '../utils/knowledge_labels.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/progress/progress.dart';
import '../models/progress.dart';
import '../models/training.dart';
import '../services/progress_service.dart';
import '../services/service_locator.dart';
import '../widgets/vigor_logo.dart';
import 'health_permissions_screen.dart';
import 'training_details_screen.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  Progress? _progress;
  WeeklyTarget? _weeklyTarget;
  Map<String, dynamic>? _healthDaily;
  bool _isLoading = false;
  bool _hasLoadedOnce = false;
  bool _consumedInitialData = false;

  void _consumePreloadedData() {
    if (_consumedInitialData) return;
    _consumedInitialData = true;
    final serviceLocator = context.read<ServiceLocator>();
    if (serviceLocator.initialDataLoaded) {
      _progress = serviceLocator.initialProgress;
      _weeklyTarget = serviceLocator.initialWeeklyTarget;
      _healthDaily = serviceLocator.healthDailyNotifier.value;
      _hasLoadedOnce = true;
      serviceLocator.initialProgress = null;
      serviceLocator.initialWeeklyTarget = null;
    }
  }

  Future<void> _loadProgress({int retryCount = 0}) async {
    if (_isLoading) return;
    setState(() => _isLoading = true);
    _hasLoadedOnce = true;

    // on web, storage may need a moment to persist after login
    final storage = context.read<SecureStorageService>();
    final progressService = context.read<ServiceLocator>().progressService;
    if (!await storage.hasTokens()) {
      if (retryCount < 3) {
        await Future.delayed(const Duration(milliseconds: 100));
        if (mounted) {
          setState(() => _isLoading = false);
          _hasLoadedOnce = false;
          _loadProgress(retryCount: retryCount + 1);
        }
        return;
      }
    }

    try {
      final prefs = context.read<PreferencesService>();
      final locator = context.read<ServiceLocator>();
      final futures = <Future>[
        progressService.getProgress(),
        progressService.getWeeklyTarget(),
      ];
      if (prefs.hcConnected) {
        futures.add(locator.trainingService.getHealthDaily());
      }

      final results = await Future.wait(futures);

      if (mounted) {
        final progressResponse = results[0];
        final weeklyTargetResponse = results[1];

        setState(() {
          if (progressResponse.isSuccess) {
            _progress = progressResponse.data as Progress?;
          }
          if (weeklyTargetResponse.isSuccess) {
            _weeklyTarget = weeklyTargetResponse.data as WeeklyTarget?;
          }
          if (prefs.hcConnected && results.length > 2 && results[2].isSuccess) {
            _healthDaily = results[2].data as Map<String, dynamic>?;
            locator.healthDailyNotifier.value = _healthDaily;
          }
          _isLoading = false;
        });

        if (!progressResponse.isSuccess && progressResponse.error != null) {
          AdaptiveNotification.showError(
            context: context,
            message: AppLocalizations.of(context).failedToLoadProgress,
            rawError: progressResponse.error,
          );
        }
      }
    } catch (e) {
      if (mounted) {
        setState(() => _isLoading = false);
        AdaptiveNotification.showError(
          context: context,
          message: AppLocalizations.of(context).failedToLoadProgress,
          rawError: e.toString(),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    _consumePreloadedData();
    // watch auth state to trigger load when authenticated
    final authState = context.watch<AuthProvider>().state;
    if (authState == AuthState.authenticated && !_hasLoadedOnce && !_isLoading) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _loadProgress());
    }

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
        automaticallyImplyLeading: false,
        title: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            const VigorLogo(size: 24),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.appName.toUpperCase()),
          ],
        ),
        actions: [
          AdaptiveIconButton(
            icon: const Icon(Icons.refresh),
            tooltip: l10n.refresh,
            onPressed: _loadProgress,
          ),
        ],
      ),
      body: RefreshIndicator(
        onRefresh: _loadProgress,
        color: VigorColors.stone,
        child: _isLoading
            ? const Center(child: AdaptiveLoadingIndicator())
            : _buildContent(l10n),
      ),
    );
  }

  Widget _buildContent(AppLocalizations l10n) {
    if (_progress == null) {
      return _buildEmptyState(l10n);
    }

    final families = ProgressService.parseFamilies(_progress!.families);
    final muscles = ProgressService.parseMuscles(_progress!.muscles);
    final trainings = _progress!.trainings;

    // get gender from user profile for body figure
    final authProvider = context.read<AuthProvider>();
    final gender = authProvider.currentUser?.profile.gender ?? '';

    if (trainings == 0) {
      return _buildWelcomeState(l10n);
    }

    // check if any family is still under 100% calibration
    final isCalibrating = families.values.any((fp) => fp.calibration < 100.0);
    context.read<ServiceLocator>().updateCalibrationFromProgress(families);

    final sections = <Widget>[
      // hero stats
      Padding(
        padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
        child: _buildHeroStats(l10n, families),
      ),
      // pending feedback card for partnered trainings missing user feedback
      if (_progress!.pendingFeedback.isNotEmpty)
        Padding(
          padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
          child: _buildPendingFeedbackCard(l10n),
        ),
      // calibration card — only during calibration phase
      if (isCalibrating)
        Padding(
          padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
          child: _buildCalibrationCard(l10n, families),
        ),
      // health onboarding card — post-first-training, non-blocking
      if (_shouldShowHealthOnboarding())
        Padding(
          padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
          child: _buildHealthOnboardingCard(l10n),
        ),
      // weekly target card
      if (_weeklyTarget != null && _weeklyTarget!.goals.isNotEmpty)
        Padding(
          padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
          child: _buildWeeklyTargetCard(context, l10n, _weeklyTarget!),
        ),
      // health metrics card — only when HC connected and data exists
      if (_healthDaily != null && _hasHealthMetrics())
        Padding(
          padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
          child: _buildHealthMetricsCard(l10n),
        ),
      // muscle map section
      Padding(
        padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
        child: _buildMuscleMapSection(context, l10n, muscles),
      ),
      // capabilities section
      Padding(
        padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
        child: _buildCapabilitiesSection(l10n, families),
      ),
    ];

    return ListView.builder(
      padding: VigorSpacing.paddingLg,
      itemCount: sections.length,
      itemBuilder: (context, index) => sections[index],
    );
  }

  bool _shouldShowHealthOnboarding() {
    final locator = context.read<ServiceLocator>();
    final prefs = context.read<PreferencesService>();
    // don't show if no health service (web), already connected, or recently dismissed
    if (locator.healthDataService == null) return false;
    if (prefs.hcConnected) return false;
    if (prefs.hcOnboardingRecentlyDismissed) return false;
    return true;
  }

  Widget _buildHealthOnboardingCard(AppLocalizations l10n) {
    final accentColor = VigorColors.indigoAdaptive(context);
    return AdaptiveCard(
      padding: VigorSpacing.paddingLg,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.monitor_heart_outlined, size: 22, color: accentColor),
              const SizedBox(width: VigorSpacing.sm),
              Expanded(
                child: Text(
                  l10n.healthOnboardingTitle,
                  style: VigorTypography.headline.copyWith(fontSize: 16, color: VigorColors.textPrimary(context)),
                ),
              ),
            ],
          ),
          const SizedBox(height: VigorSpacing.sm),
          Text(
            l10n.healthOnboardingDescription,
            style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context), fontSize: 13),
          ),
          const SizedBox(height: VigorSpacing.md),
          Row(
            children: [
              Material(
                color: accentColor,
                borderRadius: VigorRadius.radiusSm,
                child: InkWell(
                  onTap: () => Navigator.of(context).push(
                    MaterialPageRoute(builder: (_) => const HealthPermissionsScreen()),
                  ),
                  borderRadius: VigorRadius.radiusSm,
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
                    child: Text(
                      l10n.healthOnboardingConnect,
                      style: VigorTypography.label.copyWith(color: Colors.white, fontWeight: FontWeight.w600),
                    ),
                  ),
                ),
              ),
              const SizedBox(width: VigorSpacing.sm),
              TextButton(
                onPressed: () async {
                  final prefs = context.read<PreferencesService>();
                  await prefs.setHcOnboardingDismissedMs(DateTime.now().millisecondsSinceEpoch);
                  if (mounted) setState(() {});
                },
                child: Text(l10n.healthOnboardingDismiss, style: TextStyle(color: VigorColors.stone)),
              ),
            ],
          ),
        ],
      ),
    );
  }

  bool _hasHealthMetrics() {
    final metrics = _healthDaily?['metrics'] as List?;
    return metrics != null && metrics.isNotEmpty;
  }

  Widget _buildHealthMetricsCard(AppLocalizations l10n) {
    final metrics = _healthDaily!['metrics'] as List;
    final today = metrics.first as Map<String, dynamic>;

    final sleepHours = (today['sleep_hours'] as num?)?.toDouble() ?? 0;
    final restingHR = (today['resting_hr'] as num?)?.toInt() ?? 0;
    final hrv = (today['hrv_rmssd'] as num?)?.toDouble() ?? 0;
    final steps = (today['steps'] as num?)?.toInt() ?? 0;
    final calories = (today['active_calories'] as num?)?.toDouble() ?? 0;

    // always show all tiles — use — for missing values
    final tiles = <Widget>[
      _buildMetricTile(l10n.healthDailySleep, sleepHours > 0 ? _formatSleepHours(sleepHours) : '—', 'h'),
      _buildMetricTile(l10n.healthDailyRestingHr, restingHR > 0 ? '$restingHR' : '—', 'bpm'),
      _buildMetricTile(l10n.healthDailyHrv, hrv > 0 ? '${hrv.toInt()}' : '—', 'ms'),
      _buildMetricTile(l10n.healthDailySteps, steps > 0 ? _formatSteps(steps) : '—', ''),
      _buildMetricTile(l10n.healthDailyCalories, calories > 0 ? '${calories.toInt()}' : '—', 'kcal'),
    ];

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.monitor_heart, size: 24, color: VigorColors.stone),
            const SizedBox(width: VigorSpacing.sm),
            Text(
              l10n.healthMetrics,
              style: VigorTypography.headline.copyWith(
                fontSize: 18,
                color: VigorColors.textPrimary(context),
              ),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.sm),
        AdaptiveCard(
          padding: VigorSpacing.paddingMd,
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: tiles,
          ),
        ),
      ],
    );
  }

  Widget _buildMetricTile(String label, String value, String unit) {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Row(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.baseline,
          textBaseline: TextBaseline.alphabetic,
          children: [
            Text(value, style: VigorTypography.data.copyWith(
              fontSize: 20,
              fontWeight: FontWeight.w600,
              color: VigorColors.textPrimary(context),
            )),
            if (unit.isNotEmpty) ...[
              const SizedBox(width: 2),
              Text(unit, style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
                fontSize: 10,
              )),
            ],
          ],
        ),
        const SizedBox(height: 2),
        Text(label, style: VigorTypography.caption.copyWith(
          color: VigorColors.textSecondary(context),
        )),
      ],
    );
  }

  String _formatSleepHours(double hours) {
    final h = hours.toInt();
    final m = ((hours - h) * 60).round();
    return '$h:${m.toString().padLeft(2, '0')}';
  }

  String _formatSteps(int steps) {
    if (steps >= 1000) return '${(steps / 1000).toStringAsFixed(1)}k';
    return '$steps';
  }

  Widget _buildPendingFeedbackCard(AppLocalizations l10n) {
    final pending = _progress!.pendingFeedback;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.rate_review, color: VigorColors.persimmon, size: 24),
            const SizedBox(width: VigorSpacing.sm),
            Text(
              l10n.pendingFeedbacks,
              style: VigorTypography.headline.copyWith(
                fontSize: 18,
                color: VigorColors.textPrimary(context),
              ),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.sm),
        AdaptiveCard(
          padding: VigorSpacing.paddingMd,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                l10n.pendingFeedbacksDescription,
                style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context)),
              ),
              const SizedBox(height: VigorSpacing.sm),
              for (final t in pending)
                InkWell(
                  onTap: () => _openPendingTraining(t.id),
                  child: Padding(
                    padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
                    child: Row(
                      children: [
                        Expanded(
                          child: Text(
                            t.name,
                            style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        Icon(Icons.chevron_right, color: VigorColors.stone, size: 20),
                      ],
                    ),
                  ),
                ),
            ],
          ),
        ),
      ],
    );
  }

  Future<void> _openPendingTraining(String trainingId) async {
    final response = await context.read<ServiceLocator>().trainingService.getTrainings();
    if (!response.isSuccess || !mounted) return;
    final training = response.data?.where((t) => t.id == trainingId).firstOrNull;
    if (training == null || !mounted) return;
    Navigator.of(context).push(
      MaterialPageRoute(builder: (context) => TrainingDetailsScreen(training: training)),
    );
  }

  Widget _buildHeroStats(AppLocalizations l10n, Map<String, FamilyProgress> families) {
    final trainings = _progress?.trainings ?? 0;
    final partnered = _progress?.trainingsPartnered ?? 0;

    // calculate overall calibration
    double calibration = 0;
    if (families.isNotEmpty) {
      final sum = families.values.fold(0.0, (acc, fp) => acc + fp.calibration);
      calibration = (sum / families.length).clamp(0.0, 100.0);
    }

    return Center(
      child: LayoutBuilder(
        builder: (context, constraints) {
          // size circle based on content height, not viewport width
          const minSize = 220.0;
          final size = minSize;
          return SizedBox(
            width: size,
            height: size,
            child: Stack(
              clipBehavior: Clip.none,
              children: [
                Container(
                  padding: VigorSpacing.paddingLg,
                  decoration: BoxDecoration(
                    color: VigorColors.surface(context),
                    shape: BoxShape.circle,
                  ),
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // main stat — JetBrains Mono for data
                      Text(
                        '$trainings',
                        style: VigorTypography.dataDisplay.copyWith(
                          fontSize: 72,
                          fontWeight: FontWeight.w700,
                          color: VigorColors.persimmon,
                          height: 1,
                        ),
                      ),
                      const SizedBox(height: VigorSpacing.xs),
                      Text(
                        l10n.completedTrainings,
                        style: VigorTypography.body.copyWith(
                          color: VigorColors.textSecondary(context),
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      const SizedBox(height: VigorSpacing.sm),
                      // secondary stat - compact
                      Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.people, color: VigorColors.indigoAdaptive(context), size: 18),
                          const SizedBox(width: VigorSpacing.xs),
                          Text(
                            '$partnered',
                            style: VigorTypography.data.copyWith(
                              color: VigorColors.indigoAdaptive(context),
                              fontSize: 18,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                // calibration badge - top right
                Positioned(
                  top: 8,
                  right: 8,
                  child: GestureDetector(
                    onTap: () => _showCalibrationModal(context, l10n, families, calibration),
                    child: Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: VigorColors.surface(context),
                        borderRadius: VigorRadius.radiusFull,
                        border: Border.all(color: VigorColors.border(context)),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.tune, size: 14, color: VigorColors.indigoAdaptive(context)),
                          const SizedBox(width: 4),
                          Text(
                            calibration > 0 ? '${calibration.toInt()}%' : '–',
                            style: VigorTypography.caption.copyWith(
                              color: VigorColors.indigoAdaptive(context),
                              fontWeight: FontWeight.w600,
                            ),
                          ),
                          const SizedBox(width: 2),
                          Icon(Icons.help_outline, size: 12, color: VigorColors.stone),
                        ],
                      ),
                    ),
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildCalibrationCard(AppLocalizations l10n, Map<String, FamilyProgress> families) {
    final total = families.length;
    final calibrated = families.values.where((fp) => fp.calibration >= 100.0).length;
    final overallCalibration = families.values.fold(0.0, (acc, fp) => acc + fp.calibration) / total;

    // build segment data in display order
    final segments = KnowledgeLabels.familyDisplayOrder
        .where((f) => families.containsKey(f))
        .map((f) => families[f]!.calibration >= 100.0)
        .toList();
    // append any families not in the predefined order
    for (final entry in families.entries) {
      if (!KnowledgeLabels.familyDisplayOrder.contains(entry.key)) {
        segments.add(entry.value.calibration >= 100.0);
      }
    }

    return GestureDetector(
      onTap: () => _showCalibrationModal(context, l10n, families, overallCalibration),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.tune, color: VigorColors.indigoAdaptive(context), size: 24),
              const SizedBox(width: VigorSpacing.sm),
              Text(
                l10n.calibration,
                style: VigorTypography.headline.copyWith(
                  fontSize: 18,
                  color: VigorColors.textPrimary(context),
                ),
              ),
            ],
          ),
          const SizedBox(height: VigorSpacing.sm),
          AdaptiveCard(
            padding: VigorSpacing.paddingMd,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                // left column: text + click icon
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        l10n.calibrationFamiliesLearned(calibrated, total),
                        style: VigorTypography.data.copyWith(
                          color: VigorColors.textPrimary(context),
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      const SizedBox(height: VigorSpacing.xs),
                      Text(
                        l10n.calibrationDescription,
                        style: VigorTypography.caption.copyWith(
                          color: VigorColors.textSecondary(context),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(width: VigorSpacing.lg),
                // segmented arc on right
                Padding(
                  padding: const EdgeInsets.all(4),
                  child: _buildCalibrationRing(context, segments),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCalibrationRing(BuildContext context, List<bool> segments) {
    const size = 72.0;
    const strokeWidth = 6.0;
    return SizedBox(
      width: size,
      height: size,
      child: CustomPaint(
        painter: _CalibrationRingPainter(
          segments: segments,
          activeColor: VigorColors.indigoAdaptive(context),
          inactiveColor: Theme.of(context).brightness == Brightness.dark
              ? Colors.white.withValues(alpha: 0.1)
              : VigorColors.stone.withValues(alpha: 0.2),
          strokeWidth: strokeWidth,
        ),
        child: Center(
          child: Icon(Icons.tune, size: 20, color: VigorColors.indigoAdaptive(context)),
        ),
      ),
    );
  }

  void _showCalibrationModal(BuildContext context, AppLocalizations l10n, Map<String, FamilyProgress> families, double calibration) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => _CalibrationModal(
        l10n: l10n,
        families: families,
        calibration: calibration,
      ),
    );
  }

  Widget _buildMuscleMapSection(BuildContext context, AppLocalizations l10n, Map<String, MuscleImpact> muscles) {
    final gender = context.read<AuthProvider>().currentUser?.profile.gender;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.accessibility_new, color: VigorColors.stone, size: 24),
            const SizedBox(width: VigorSpacing.sm),
            Text(
              l10n.muscleHeatMap,
              style: VigorTypography.headline.copyWith(
                fontSize: 18,
                color: VigorColors.textPrimary(context),
              ),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.sm),
        AdaptiveCard(
          padding: VigorSpacing.paddingMd,
          child: Column(
            children: [
              MuscleMapWidget(
                muscles: muscles,
                showToggle: false,
                height: 280,
                gender: gender,
              ),
              const SizedBox(height: VigorSpacing.md),
              _buildHeatLegend(context),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildHeatLegend(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Wrap(
      alignment: WrapAlignment.center,
      spacing: VigorSpacing.md,
      runSpacing: VigorSpacing.xs,
      children: [
        _buildLegendItem(context, Colors.transparent, l10n.heatResting, showBorder: true),
        _buildLegendItem(context, VigorColors.persimmon.withValues(alpha: 0.90), l10n.heatHot),
      ],
    );
  }

  Widget _buildLegendItem(BuildContext context, Color color, String label, {bool showBorder = false}) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 12,
          height: 12,
          decoration: BoxDecoration(
            color: color,
            borderRadius: VigorRadius.radiusXs,
            border: showBorder ? Border.all(color: VigorColors.border(context)) : null,
          ),
        ),
        const SizedBox(width: VigorSpacing.xs),
        Text(
          label,
          style: VigorTypography.caption.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
      ],
    );
  }

  Widget _buildCapabilitiesSection(AppLocalizations l10n, Map<String, FamilyProgress> families) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.show_chart, color: VigorColors.stone, size: 24),
            const SizedBox(width: VigorSpacing.sm),
            Text(
              l10n.capabilities,
              style: VigorTypography.headline.copyWith(
                fontSize: 18,
                color: VigorColors.textPrimary(context),
              ),
            ),
          ],
        ),
        const SizedBox(height: VigorSpacing.sm),
        AdaptiveCard(
          padding: VigorSpacing.paddingMd,
          child: FamilyProgressWidget(families: families),
        ),
      ],
    );
  }

  Widget _buildWeeklyTargetCard(BuildContext context, AppLocalizations l10n, WeeklyTarget weeklyTarget) {
    final current = weeklyTarget.currentWeek;
    final rec = weeklyTarget.recommendation;
    final minTarget = rec.sessionsPerWeek.isNotEmpty ? rec.sessionsPerWeek[0] : 0;
    final maxTarget = rec.sessionsPerWeek.length > 1 ? rec.sessionsPerWeek[1] : minTarget;
    final progress = maxTarget > 0
        ? (current.sessionsCompleted / maxTarget).clamp(0.0, 1.0)
        : 0.0;
    final targetAvg = ((minTarget + maxTarget) / 2).round();

    return GestureDetector(
      onTap: () => _showWeeklyTargetModal(context, l10n, weeklyTarget),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.event_note, color: VigorColors.stone, size: 24),
              const SizedBox(width: VigorSpacing.sm),
              Text(
                l10n.weeklyTarget,
                style: VigorTypography.headline.copyWith(
                  fontSize: 18,
                  color: VigorColors.textPrimary(context),
                ),
              ),
            ],
          ),
          const SizedBox(height: VigorSpacing.sm),
          AdaptiveCard(
            padding: VigorSpacing.paddingMd,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                // circular progress on left with sessions count inside
                Padding(
                  padding: const EdgeInsets.all(4),
                  child: SizedBox(
                    width: 72,
                    height: 72,
                    child: Stack(
                      alignment: Alignment.center,
                      children: [
                        SizedBox(
                          width: 72,
                          height: 72,
                          child: CircularProgressIndicator(
                            value: progress,
                            strokeWidth: 6,
                            backgroundColor: VigorColors.stone.withValues(alpha: 0.2),
                            valueColor: AlwaysStoppedAnimation(
                              current.sessionsCompleted >= minTarget
                                  ? VigorColors.persimmon
                                  : VigorColors.stone,
                            ),
                          ),
                        ),
                        if (current.sessionsCompleted > targetAvg)
                          Icon(Icons.check, size: 28, color: VigorColors.persimmon)
                        else
                          Text.rich(
                            TextSpan(
                              children: [
                                TextSpan(
                                  text: '${current.sessionsCompleted}',
                                  style: VigorTypography.data.copyWith(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w600,
                                    color: VigorColors.textPrimary(context),
                                  ),
                                ),
                                TextSpan(
                                  text: '/$targetAvg',
                                  style: VigorTypography.data.copyWith(
                                    fontSize: 14,
                                    fontWeight: FontWeight.w400,
                                    color: VigorColors.textSecondary(context),
                                  ),
                                ),
                              ],
                            ),
                          ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(width: VigorSpacing.lg),
                // right column: chips row, week days below
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // goals chips left, days left chip right
                      Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Wrap(
                              alignment: WrapAlignment.start,
                              spacing: 6,
                              runSpacing: 4,
                              children: weeklyTarget.goals.map((g) => Container(
                                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                                decoration: BoxDecoration(
                                  color: VigorColors.surfaceElevated(context),
                                  borderRadius: VigorRadius.radiusFull,
                                ),
                                child: Text(
                                  KnowledgeLabels.goalLabel(g, l10n),
                                  style: VigorTypography.caption.copyWith(
                                    color: VigorColors.textSecondary(context),
                                    fontSize: 12,
                                  ),
                                ),
                              )).toList(),
                            ),
                          ),
                          const SizedBox(width: 6),
                          Icon(
                            Icons.open_in_new,
                            size: 14,
                            color: VigorColors.textSecondary(context),
                          ),
                        ],
                      ),
                      const SizedBox(height: VigorSpacing.md),
                      // week days indicator
                      _buildWeekDays(context, current.completedDays),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWeekDays(BuildContext context, List<int> completedDays) {
    final locale = Localizations.localeOf(context).toString();
    // generate localized single-letter weekday labels starting from Monday
    final dayLetters = List.generate(7, (i) {
      // DateTime weekday: 1=Mon...7=Sun; find a known Monday and offset
      final monday = DateTime(2024, 1, 1); // 2024-01-01 is a Monday
      return DateFormat.E(locale).format(monday.add(Duration(days: i)))[0].toUpperCase();
    });
    final today = (DateTime.now().weekday - 1) % 7;

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: List.generate(7, (i) {
        final isCompleted = completedDays.contains(i);
        final isToday = i == today;
        final isPast = i < today;

        return Container(
          width: 24,
          height: 24,
          decoration: BoxDecoration(
            color: isCompleted
                ? VigorColors.persimmon
                : isToday
                    ? VigorColors.surface(context)
                    : Colors.transparent,
            shape: BoxShape.circle,
            border: isToday && !isCompleted
                ? Border.all(color: VigorColors.stone)
                : null,
          ),
          child: Center(
            child: isCompleted
                ? const Icon(Icons.check, size: 12, color: Colors.white)
                : Text(
                    dayLetters[i],
                    style: VigorTypography.caption.copyWith(
                      fontSize: 10,
                      color: isPast && !isCompleted
                          ? VigorColors.stone.withValues(alpha: 0.5)
                          : VigorColors.textSecondary(context),
                      fontWeight: isToday ? FontWeight.w600 : FontWeight.normal,
                    ),
                  ),
          ),
        );
      }),
    );
  }

  void _showWeeklyTargetModal(BuildContext context, AppLocalizations l10n, WeeklyTarget weeklyTarget) {
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      backgroundColor: Colors.transparent,
      builder: (context) => _WeeklyTargetModal(l10n: l10n, weeklyTarget: weeklyTarget),
    );
  }

  Widget _buildEmptyState(AppLocalizations l10n) {
    return ListView(
      padding: VigorSpacing.paddingLg,
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.15),
        // simple icon container
        Center(
          child: Container(
            width: 120,
            height: 120,
            decoration: BoxDecoration(
              color: VigorColors.surface(context),
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.trending_up, size: 56, color: VigorColors.stone),
          ),
        ),
        const SizedBox(height: VigorSpacing.lg),
        Text(
          l10n.yourProgress,
          style: VigorTypography.title.copyWith(color: VigorColors.textPrimary(context)),
          textAlign: TextAlign.center,
        ),
        const SizedBox(height: VigorSpacing.sm),
        Text(
          l10n.noProgressYet,
          textAlign: TextAlign.center,
          style: VigorTypography.body.copyWith(color: VigorColors.textSecondary(context)),
        ),
      ],
    );
  }

  Widget _buildWelcomeState(AppLocalizations l10n) {
    return ListView(
      padding: VigorSpacing.paddingLg,
      children: [
        SizedBox(height: MediaQuery.of(context).size.height * 0.1),
        // welcome hero
        Center(
          child: Column(
            children: [
              // vigor logo — clean surface container
              Container(
                width: 100,
                height: 100,
                decoration: BoxDecoration(
                  color: VigorColors.surface(context),
                  shape: BoxShape.circle,
                ),
                child: const Center(child: VigorLogo(size: 48)),
              ),
              const SizedBox(height: VigorSpacing.lg),
              Text(
                l10n.readyToTrain,
                style: VigorTypography.title.copyWith(
                  color: VigorColors.textPrimary(context),
                ),
                textAlign: TextAlign.center,
              ),
              const SizedBox(height: VigorSpacing.sm),
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.xl),
                child: Text(
                  l10n.noTrainingsCompletedYet,
                  textAlign: TextAlign.center,
                  style: VigorTypography.body.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
              ),
            ],
          ),
        ),
        const SizedBox(height: VigorSpacing.xxl),
        // info cards — stone borders, stone icons
        _buildInfoCard(
          icon: Icons.auto_awesome,
          title: 'AI-Powered',
          description: 'Personalized workouts generated by AI based on your goals and equipment',
        ),
        const SizedBox(height: VigorSpacing.md),
        _buildInfoCard(
          icon: Icons.trending_up,
          title: 'Track Progress',
          description: 'Monitor your capabilities across movement families as you train',
        ),
        const SizedBox(height: VigorSpacing.md),
        _buildInfoCard(
          icon: Icons.people,
          title: 'Train Together',
          description: 'Add partners to adjust workouts for group training sessions',
        ),
      ],
    );
  }

  Widget _buildInfoCard({
    required IconData icon,
    required String title,
    required String description,
  }) {
    return Container(
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusMd,
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              color: VigorColors.surfaceElevated(context),
              borderRadius: VigorRadius.radiusSm,
            ),
            child: Icon(icon, color: VigorColors.stone, size: 24),
          ),
          const SizedBox(width: VigorSpacing.md),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  title,
                  style: VigorTypography.headline.copyWith(
                    fontSize: 16,
                    color: VigorColors.textPrimary(context),
                  ),
                ),
                const SizedBox(height: VigorSpacing.xs),
                Text(
                  description,
                  style: VigorTypography.caption.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}

/// Modal for calibration details
class _CalibrationModal extends StatelessWidget {
  final AppLocalizations l10n;
  final Map<String, FamilyProgress> families;
  final double calibration;

  const _CalibrationModal({
    required this.l10n,
    required this.families,
    required this.calibration,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Container(
      margin: const EdgeInsets.all(VigorSpacing.md),
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusLg,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // header
          Padding(
            padding: VigorSpacing.paddingMd,
            child: Row(
              children: [
                Icon(Icons.tune, color: VigorColors.indigoAdaptive(context), size: 24),
                const SizedBox(width: VigorSpacing.sm),
                Text(
                  l10n.calibration,
                  style: VigorTypography.headline.copyWith(
                    color: VigorColors.textPrimary(context),
                  ),
                ),
                const Spacer(),
                GestureDetector(
                  onTap: () => Navigator.pop(context),
                  child: Icon(Icons.close, color: VigorColors.stone, size: 24),
                ),
              ],
            ),
          ),
          Divider(height: 1, color: VigorColors.border(context)),
          // description
          Padding(
            padding: VigorSpacing.paddingMd,
            child: Text(
              l10n.calibrationDescription,
              style: VigorTypography.body.copyWith(
                color: VigorColors.textSecondary(context),
              ),
            ),
          ),
          // overall progress bar with Global label
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md),
            child: Row(
              children: [
                SizedBox(
                  width: 80,
                  child: Text(
                    l10n.calibrationGlobal,
                    style: VigorTypography.caption.copyWith(
                      color: VigorColors.textSecondary(context),
                    ),
                  ),
                ),
                Expanded(
                  child: _buildProgressBar(context, calibration, isDark),
                ),
                const SizedBox(width: 8),
                SizedBox(
                  width: 36,
                  child: Text(
                    calibration > 0 ? '${calibration.toInt()}%' : '–',
                    textAlign: TextAlign.right,
                    style: VigorTypography.data.copyWith(
                      color: VigorColors.textPrimary(context),
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(height: VigorSpacing.md),
          Divider(height: 1, color: VigorColors.border(context)),
          const SizedBox(height: VigorSpacing.md),
          // per-family bars
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md),
            child: Column(
              children: _buildFamilyBars(context, isDark),
            ),
          ),
          const SizedBox(height: VigorSpacing.lg),
        ],
      ),
    );
  }

  Widget _buildProgressBar(BuildContext context, double value, bool isDark) {
    return Stack(
      children: [
        Container(
          height: 8,
          decoration: BoxDecoration(
            color: isDark ? Colors.white.withValues(alpha: 0.1) : VigorColors.stone.withValues(alpha: 0.2),
            borderRadius: BorderRadius.circular(4),
          ),
        ),
        FractionallySizedBox(
          widthFactor: value / 100,
          child: Container(
            height: 8,
            decoration: BoxDecoration(
              color: VigorColors.indigoAdaptive(context),
              borderRadius: BorderRadius.circular(4),
            ),
          ),
        ),
      ],
    );
  }

  List<Widget> _buildFamilyBars(BuildContext context, bool isDark) {
    final sortedFamilies = KnowledgeLabels.familyDisplayOrder
        .where((f) => families.containsKey(f))
        .map((f) => MapEntry(f, families[f]!))
        .toList();

    // add any families not in the predefined order
    for (final entry in families.entries) {
      if (!KnowledgeLabels.familyDisplayOrder.contains(entry.key)) {
        sortedFamilies.add(entry);
      }
    }

    return sortedFamilies.map((entry) {
      final label = KnowledgeLabels.familyLabel(entry.key, l10n);
      final cal = entry.value.calibration.clamp(0.0, 100.0);

      return Padding(
        padding: const EdgeInsets.only(bottom: 8),
        child: Row(
          children: [
            SizedBox(
              width: 80,
              child: Text(
                label,
                style: VigorTypography.caption.copyWith(
                  color: VigorColors.textSecondary(context),
                ),
              ),
            ),
            Expanded(
              child: _buildProgressBar(context, cal, isDark),
            ),
            const SizedBox(width: 8),
            SizedBox(
              width: 36,
              child: Text(
                cal > 0 ? '${cal.toInt()}%' : '–',
                textAlign: TextAlign.right,
                style: VigorTypography.data.copyWith(
                  color: VigorColors.textSecondary(context),
                ),
              ),
            ),
          ],
        ),
      );
    }).toList();
  }

}

/// Modal for weekly target details
class _WeeklyTargetModal extends StatelessWidget {
  final AppLocalizations l10n;
  final WeeklyTarget weeklyTarget;

  const _WeeklyTargetModal({required this.l10n, required this.weeklyTarget});

  @override
  Widget build(BuildContext context) {
    final rec = weeklyTarget.recommendation;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Container(
      margin: const EdgeInsets.all(VigorSpacing.md),
      decoration: BoxDecoration(
        color: VigorColors.surface(context),
        borderRadius: VigorRadius.radiusLg,
      ),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          // header
          Padding(
            padding: VigorSpacing.paddingMd,
            child: Row(
              children: [
                Icon(Icons.event_note, color: VigorColors.stone, size: 24),
                const SizedBox(width: VigorSpacing.sm),
                Text(
                  l10n.weeklyTarget,
                  style: VigorTypography.headline.copyWith(
                    color: VigorColors.textPrimary(context),
                  ),
                ),
                const Spacer(),
                GestureDetector(
                  onTap: () => Navigator.pop(context),
                  child: Icon(Icons.close, color: VigorColors.stone, size: 24),
                ),
              ],
            ),
          ),
          Divider(height: 1, color: VigorColors.border(context)),

          const SizedBox(height: VigorSpacing.md),

          // goals as chips
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md),
            child: Row(
              children: [
                Icon(Icons.flag, color: VigorColors.stone, size: 18),
                const SizedBox(width: VigorSpacing.sm),
                Text(
                  l10n.goals,
                  style: VigorTypography.body.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
                const Spacer(),
                ...weeklyTarget.goals.map((g) => Padding(
                  padding: const EdgeInsets.only(left: 6),
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                    decoration: BoxDecoration(
                      color: VigorColors.surfaceElevated(context),
                      borderRadius: VigorRadius.radiusFull,
                    ),
                    child: Text(
                      KnowledgeLabels.goalLabel(g, l10n),
                      style: VigorTypography.caption.copyWith(
                        color: VigorColors.textSecondary(context),
                        fontSize: 12,
                      ),
                    ),
                  ),
                )),
              ],
            ),
          ),
          const SizedBox(height: VigorSpacing.sm),

          // recommendations
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md),
            child: Column(
              children: [
                _buildRecommendationRow(
                  context,
                  Icons.repeat,
                  l10n.sessionsPerWeek,
                  _formatRange(rec.sessionsPerWeek),
                ),
                const SizedBox(height: VigorSpacing.sm),
                _buildRecommendationRow(
                  context,
                  Icons.timer,
                  l10n.sessionDuration,
                  '${_formatDurationRange(rec.sessionDurationMins)} min',
                ),
                if (rec.preferredHours.isNotEmpty) ...[
                  const SizedBox(height: VigorSpacing.sm),
                  _buildRecommendationRow(
                    context,
                    Icons.schedule,
                    l10n.recommendedTime,
                    _formatTimeRange(rec.preferredHours),
                  ),
                ],
              ],
            ),
          ),
          const SizedBox(height: VigorSpacing.sm),

          // methodology mix
          if (rec.methodologyMix.isNotEmpty)
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    l10n.methodologyMix,
                    style: VigorTypography.caption.copyWith(
                      color: VigorColors.textSecondary(context),
                    ),
                  ),
                  const SizedBox(height: VigorSpacing.sm),
                  _buildMethodologyBars(context, _toDoubleMap(rec.methodologyMix), isDark),
                ],
              ),
            ),

          const SizedBox(height: VigorSpacing.md),
          Divider(height: 1, color: VigorColors.border(context)),
          const SizedBox(height: VigorSpacing.md),

          // current week
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.thisWeek,
                  style: VigorTypography.caption.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
                const SizedBox(height: VigorSpacing.sm),
                _buildWeekCalendar(context, weeklyTarget.currentWeek),
              ],
            ),
          ),

          // history
          if (weeklyTarget.history.isNotEmpty) ...[
            const SizedBox(height: VigorSpacing.md),
            Divider(height: 1, color: VigorColors.border(context)),
            const SizedBox(height: VigorSpacing.md),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    l10n.pastWeeks,
                    style: VigorTypography.caption.copyWith(
                      color: VigorColors.textSecondary(context),
                    ),
                  ),
                  const SizedBox(height: VigorSpacing.sm),
                  ...weeklyTarget.history.map((week) => _buildHistoryRow(
                        context,
                        week,
                        rec.sessionsPerWeek.length > 1 ? rec.sessionsPerWeek[1] : (rec.sessionsPerWeek.isNotEmpty ? rec.sessionsPerWeek[0] : 1),
                        isDark,
                      )),
                ],
              ),
            ),
          ],

          const SizedBox(height: VigorSpacing.lg),
        ],
      ),
    );
  }

  Widget _buildRecommendationRow(
    BuildContext context,
    IconData icon,
    String label,
    String value,
  ) {
    return Row(
      children: [
        Icon(icon, color: VigorColors.stone, size: 18),
        const SizedBox(width: VigorSpacing.sm),
        Text(
          label,
          style: VigorTypography.body.copyWith(
            color: VigorColors.textSecondary(context),
          ),
        ),
        const Spacer(),
        Text(
          value,
          style: VigorTypography.data.copyWith(
            color: VigorColors.textPrimary(context),
            fontWeight: FontWeight.w500,
          ),
        ),
      ],
    );
  }

  Widget _buildMethodologyBars(
    BuildContext context,
    Map<String, double> mix,
    bool isDark,
  ) {
    final sorted = mix.entries.toList()
      ..sort((a, b) => b.value.compareTo(a.value));

    return Column(
      children: sorted.take(4).map((entry) {
        return Padding(
          padding: const EdgeInsets.only(bottom: 6),
          child: Row(
            children: [
              SizedBox(
                width: 80,
                child: Text(
                  KnowledgeLabels.methodologyLabel(entry.key, l10n),
                  style: VigorTypography.caption.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
              ),
              Expanded(
                child: Stack(
                  children: [
                    Container(
                      height: 6,
                      decoration: BoxDecoration(
                        color: isDark
                            ? Colors.white.withValues(alpha: 0.1)
                            : VigorColors.stone.withValues(alpha: 0.2),
                        borderRadius: BorderRadius.circular(3),
                      ),
                    ),
                    FractionallySizedBox(
                      widthFactor: entry.value,
                      child: Container(
                        height: 6,
                        decoration: BoxDecoration(
                          color: VigorColors.stone,
                          borderRadius: BorderRadius.circular(3),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(width: 8),
              SizedBox(
                width: 36,
                child: Text(
                  '${(entry.value * 100).round()}%',
                  textAlign: TextAlign.right,
                  style: VigorTypography.caption.copyWith(
                    color: VigorColors.textSecondary(context),
                  ),
                ),
              ),
            ],
          ),
        );
      }).toList(),
    );
  }

  Widget _buildWeekCalendar(BuildContext context, WeekProgress week) {
    final locale = Localizations.localeOf(context).toString();
    final monday = DateTime(2024, 1, 1);
    final days = List.generate(7, (i) => DateFormat.E(locale).format(monday.add(Duration(days: i))));
    final today = (DateTime.now().weekday - 1) % 7;

    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: List.generate(7, (i) {
        final isCompleted = week.completedDays.contains(i);
        final isToday = i == today;

        return Column(
          children: [
            Text(
              days[i],
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
                fontSize: 10,
              ),
            ),
            const SizedBox(height: 4),
            Container(
              width: 36,
              height: 36,
              decoration: BoxDecoration(
                color: isCompleted
                    ? VigorColors.persimmon
                    : VigorColors.surfaceElevated(context),
                borderRadius: VigorRadius.radiusSm,
                border: isToday && !isCompleted
                    ? Border.all(color: VigorColors.persimmon, width: 2)
                    : null,
              ),
              child: isCompleted
                  ? const Icon(Icons.check, size: 18, color: Colors.white)
                  : null,
            ),
          ],
        );
      }),
    );
  }

  Widget _buildHistoryRow(
    BuildContext context,
    WeekSummary week,
    int maxTarget,
    bool isDark,
  ) {
    final progress =
        maxTarget > 0 ? (week.sessionsCompleted / maxTarget).clamp(0.0, 1.0) : 0.0;
    final dateStr = _formatWeekDate(week.weekStart);

    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: Row(
        children: [
          SizedBox(
            width: 110,
            child: Text(
              dateStr,
              style: VigorTypography.caption.copyWith(
                color: VigorColors.textSecondary(context),
                fontFamily: 'monospace',
                fontFamilyFallback: const ['Courier', 'Courier New'],
              ),
            ),
          ),
          Expanded(
            child: Stack(
              children: [
                Container(
                  height: 8,
                  decoration: BoxDecoration(
                    color: isDark
                        ? Colors.white.withValues(alpha: 0.1)
                        : VigorColors.stone.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(4),
                  ),
                ),
                FractionallySizedBox(
                  widthFactor: progress,
                  child: Container(
                    height: 8,
                    decoration: BoxDecoration(
                      color:
                          week.onTarget ? VigorColors.persimmon : VigorColors.stone,
                      borderRadius: BorderRadius.circular(4),
                    ),
                  ),
                ),
              ],
            ),
          ),
          const SizedBox(width: 8),
          SizedBox(
            width: 24,
            child: Text(
              '${week.sessionsCompleted}',
              textAlign: TextAlign.right,
              style: VigorTypography.data.copyWith(
                color: VigorColors.textPrimary(context),
              ),
            ),
          ),
          const SizedBox(width: 4),
          if (week.onTarget)
            Icon(Icons.check_circle, size: 16, color: VigorColors.persimmon)
          else
            const SizedBox(width: 16),
        ],
      ),
    );
  }

  String _formatTimeRange(List<int> hours) {
    if (hours.length < 2) return '';
    return '${hours[0].toString().padLeft(2, '0')}:00 - ${hours[1].toString().padLeft(2, '0')}:00';
  }

  String _formatWeekDate(DateTime weekStart) {
    final weekEnd = weekStart.add(const Duration(days: 6));
    return '${weekStart.day.toString().padLeft(2, '0')}/${weekStart.month.toString().padLeft(2, '0')} – ${weekEnd.day.toString().padLeft(2, '0')}/${weekEnd.month.toString().padLeft(2, '0')}';
  }

  String _formatRange(List<int> values) {
    if (values.isEmpty) return '0';
    if (values.length == 1) return '${values[0]}';
    return '${values[0]}-${values[1]}';
  }

  String _formatDurationRange(List<int> values) {
    if (values.isEmpty) return '0';
    if (values.length == 1) return '${values[0]}';
    if (values[0] == values[1]) return '${values[0]}';
    return '${values[0]}-${values[1]}';
  }

  Map<String, double> _toDoubleMap(Map<String, dynamic> map) {
    return map.map((k, v) => MapEntry(k, (v as num).toDouble()));
  }
}

/// Draws a segmented ring where each segment represents a movement family.
/// Calibrated segments use [activeColor], uncalibrated use [inactiveColor].
class _CalibrationRingPainter extends CustomPainter {
  final List<bool> segments;
  final Color activeColor;
  final Color inactiveColor;
  final double strokeWidth;

  _CalibrationRingPainter({
    required this.segments,
    required this.activeColor,
    required this.inactiveColor,
    required this.strokeWidth,
  });

  @override
  void paint(Canvas canvas, Size size) {
    if (segments.isEmpty) return;

    final center = Offset(size.width / 2, size.height / 2);
    final radius = (size.shortestSide - strokeWidth) / 2;
    final count = segments.length;

    // gap between segments in radians
    const gapRadians = 0.06;
    final sweepPerSegment = (2 * pi - count * gapRadians) / count;
    // start from top (-pi/2)
    var startAngle = -pi / 2;

    final paint = Paint()
      ..style = PaintingStyle.stroke
      ..strokeWidth = strokeWidth
      ..strokeCap = StrokeCap.round;

    for (var i = 0; i < count; i++) {
      paint.color = segments[i] ? activeColor : inactiveColor;
      canvas.drawArc(
        Rect.fromCircle(center: center, radius: radius),
        startAngle,
        sweepPerSegment,
        false,
        paint,
      );
      startAngle += sweepPerSegment + gapRadians;
    }
  }

  @override
  bool shouldRepaint(_CalibrationRingPainter oldDelegate) =>
      !listEquals(segments, oldDelegate.segments) ||
      activeColor != oldDelegate.activeColor ||
      inactiveColor != oldDelegate.inactiveColor;
}

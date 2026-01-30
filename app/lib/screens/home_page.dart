import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/family_progress.dart';
import '../models/muscle_impact.dart';
import '../providers/auth_provider.dart';
import '../services/secure_storage_service.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/progress/progress.dart';
import '../models/progress.dart';
import '../services/progress_service.dart';
import '../services/service_locator.dart';
import '../widgets/vigor_logo.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  Progress? _progress;
  bool _isLoading = false;
  bool _hasLoadedOnce = false;

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
      final response = await progressService.getProgress();
      if (response.isSuccess && mounted) {
        setState(() {
          _progress = response.data;
          _isLoading = false;
        });
      } else if (mounted) {
        setState(() => _isLoading = false);
        if (response.error != null) {
          AdaptiveNotification.showError(
            context: context,
            message: AppLocalizations.of(context).failedToLoadProgress,
            rawError: response.error,
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
    // watch auth state to trigger load when authenticated
    final authState = context.watch<AuthProvider>().state;
    if (authState == AuthState.authenticated && !_hasLoadedOnce && !_isLoading) {
      WidgetsBinding.instance.addPostFrameCallback((_) => _loadProgress());
    }

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(
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

    return ListView.builder(
      padding: VigorSpacing.paddingLg,
      itemCount: 4,
      itemBuilder: (context, index) {
        switch (index) {
          case 0:
            // hero stats section with calibration badge
            return Padding(
              padding: const EdgeInsets.only(bottom: VigorSpacing.xl),
              child: _buildHeroStats(l10n, families),
            );
          case 1:
            // muscle map section
            return Padding(
              padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
              child: _buildMuscleMapSection(context, l10n, muscles),
            );
          case 2:
            // capabilities section
            return Padding(
              padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
              child: _buildCapabilitiesSection(l10n, families),
            );
          default:
            return const SizedBox.shrink();
        }
      },
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
                          color: VigorColors.textPrimary(context),
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
                          Icon(Icons.people, color: VigorColors.stone, size: 18),
                          const SizedBox(width: VigorSpacing.xs),
                          Text(
                            '$partnered',
                            style: VigorTypography.data.copyWith(
                              color: VigorColors.textSecondary(context),
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
    return Row(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        _buildLegendItem(context, Colors.transparent, 'Cool', showBorder: true),
        const SizedBox(width: VigorSpacing.md),
        _buildLegendItem(context, VigorColors.persimmon.withValues(alpha: 0.25), 'Warm'),
        const SizedBox(width: VigorSpacing.md),
        _buildLegendItem(context, VigorColors.persimmon.withValues(alpha: 0.60), 'Hot'),
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

  static const _familyLabels = {
    'horizontal_push': 'Push',
    'horizontal_pull': 'Pull',
    'vertical_push': 'Overhead',
    'vertical_pull': 'Pull-up',
    'squat': 'Squat',
    'hinge': 'Hinge',
    'core': 'Core',
    'carry': 'Carry',
    'balance': 'Balance',
    'cardio': 'Cardio',
    'mobility': 'Mobility',
  };

  static const _familyOrder = [
    'horizontal_push',
    'horizontal_pull',
    'vertical_push',
    'vertical_pull',
    'squat',
    'hinge',
    'core',
    'cardio',
    'mobility',
    'balance',
    'carry',
  ];

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
    final sortedFamilies = _familyOrder
        .where((f) => families.containsKey(f))
        .map((f) => MapEntry(f, families[f]!))
        .toList();

    // add any families not in the predefined order
    for (final entry in families.entries) {
      if (!_familyOrder.contains(entry.key)) {
        sortedFamilies.add(entry);
      }
    }

    return sortedFamilies.map((entry) {
      final label = _familyLabels[entry.key] ?? _formatFamilyName(entry.key);
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

  String _formatFamilyName(String family) {
    return family
        .split('_')
        .map((w) => w.isNotEmpty ? '${w[0].toUpperCase()}${w.substring(1)}' : w)
        .join(' ');
  }
}

import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../models/family_progress.dart';
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
        title: Text(l10n.appName),
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
        color: VigorColors.orange,
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
    final trainings = _progress!.trainings;

    if (trainings == 0) {
      return _buildWelcomeState(l10n);
    }

    return ListView(
      padding: VigorSpacing.paddingLg,
      children: [
        // hero stats section
        _buildHeroStats(l10n),
        const SizedBox(height: VigorSpacing.xl),
        // calibration section
        CalibrationWidget(families: families),
        const SizedBox(height: VigorSpacing.lg),
        // capabilities section
        _buildCapabilitiesSection(l10n, families),
      ],
    );
  }

  Widget _buildHeroStats(AppLocalizations l10n) {
    final trainings = _progress?.trainings ?? 0;
    final partnered = _progress?.trainingsPartnered ?? 0;
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Container(
      padding: VigorSpacing.paddingLg,
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [
            VigorColors.orange.withValues(alpha: 0.15),
            VigorColors.electricBlue.withValues(alpha: 0.15),
          ],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: VigorRadius.radiusLg,
        border: Border.all(
          color: isDark
              ? Colors.white.withValues(alpha: 0.1)
              : Colors.black.withValues(alpha: 0.05),
        ),
      ),
      child: Column(
        children: [
          // main stat with gradient text
          RepaintBoundary(
            child: ShaderMask(
              shaderCallback: (bounds) => const LinearGradient(
                colors: [VigorColors.orange, VigorColors.electricBlue],
              ).createShader(bounds),
              child: Text(
                '$trainings',
                style: VigorTypography.display.copyWith(
                  fontSize: 80,
                  fontWeight: FontWeight.w800,
                  color: Colors.white,
                  height: 1,
                ),
              ),
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
          const SizedBox(height: VigorSpacing.lg),
          // secondary stat
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              _buildSecondaryStatBadge(
                icon: Icons.people,
                value: partnered,
                label: l10n.partneredTrainings,
                color: VigorColors.electricBlue,
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildSecondaryStatBadge({
    required IconData icon,
    required int value,
    required String label,
    required Color color,
  }) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: VigorSpacing.md, vertical: VigorSpacing.sm),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: VigorRadius.radiusFull,
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(icon, color: color, size: 18),
          const SizedBox(width: VigorSpacing.sm),
          Text(
            '$value',
            style: VigorTypography.headline.copyWith(
              color: color,
              fontWeight: FontWeight.w700,
            ),
          ),
          const SizedBox(width: VigorSpacing.xs),
          Text(
            label,
            style: VigorTypography.caption.copyWith(color: color),
          ),
        ],
      ),
    );
  }

  Widget _buildCapabilitiesSection(AppLocalizations l10n, Map<String, FamilyProgress> families) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            RepaintBoundary(
              child: ShaderMask(
                shaderCallback: (bounds) => const LinearGradient(
                  colors: [VigorColors.orange, VigorColors.electricBlue],
                ).createShader(bounds),
                child: const Icon(Icons.show_chart, color: Colors.white, size: 24),
              ),
            ),
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
              child: const Icon(Icons.trending_up, size: 56, color: Colors.white),
            ),
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
              // vigor logo icon
              Container(
                width: 100,
                height: 100,
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
        // info cards
        _buildInfoCard(
          icon: Icons.auto_awesome,
          title: 'AI-Powered',
          description: 'Personalized workouts generated by AI based on your goals and equipment',
          gradient: [VigorColors.orange, VigorColors.orange.withValues(alpha: 0.7)],
        ),
        const SizedBox(height: VigorSpacing.md),
        _buildInfoCard(
          icon: Icons.trending_up,
          title: 'Track Progress',
          description: 'Monitor your capabilities across movement families as you train',
          gradient: [VigorColors.electricBlue, VigorColors.electricBlue.withValues(alpha: 0.7)],
        ),
        const SizedBox(height: VigorSpacing.md),
        _buildInfoCard(
          icon: Icons.people,
          title: 'Train Together',
          description: 'Add partners to adjust workouts for group training sessions',
          gradient: [VigorColors.success, VigorColors.success.withValues(alpha: 0.7)],
        ),
      ],
    );
  }

  Widget _buildInfoCard({
    required IconData icon,
    required String title,
    required String description,
    required List<Color> gradient,
  }) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return Container(
      padding: VigorSpacing.paddingMd,
      decoration: BoxDecoration(
        color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
        borderRadius: VigorRadius.radiusMd,
        border: Border.all(
          color: gradient[0].withValues(alpha: 0.3),
        ),
      ),
      child: Row(
        children: [
          Container(
            width: 48,
            height: 48,
            decoration: BoxDecoration(
              gradient: LinearGradient(colors: gradient),
              borderRadius: VigorRadius.radiusSm,
            ),
            child: Icon(icon, color: Colors.white, size: 24),
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

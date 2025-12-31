import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../generated/app_localizations.dart';
import '../widgets/adaptive/adaptive.dart';
import '../widgets/progress/progress.dart';
import '../models/progress.dart';
import '../services/progress_service.dart';
import '../services/secure_storage_service.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

class HomePage extends StatefulWidget {
  const HomePage({super.key});

  @override
  State<HomePage> createState() => _HomePageState();
}

class _HomePageState extends State<HomePage> {
  ProgressService? _progressService;
  Progress? _progress;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      final storage = context.read<SecureStorageService>();
      _progressService = ProgressService(storageService: storage);
      _loadProgress();
    });
  }

  Future<void> _loadProgress() async {
    if (_progressService == null) return;

    setState(() {
      _isLoading = true;
    });

    final response = await _progressService!.getProgress();
    if (response.isSuccess && mounted) {
      setState(() {
        _progress = response.data;
        _isLoading = false;
      });
    } else if (mounted) {
      setState(() {
        _isLoading = false;
      });
      if (response.error != null) {
        AdaptiveNotification.showError(
          context: context,
          message: AppLocalizations.of(context).failedToLoadProgress,
          rawError: response.error,
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
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
    final trainingsComplete = _progress!.trainingsComplete ?? 0;

    // show calibration message when no trainings completed
    if (trainingsComplete == 0) {
      return ListView(
        padding: const EdgeInsets.all(20.0),
        children: [
          _buildTrainingCount(l10n),
          const SizedBox(height: 32),
          _buildCalibrationMessage(l10n),
        ],
      );
    }

    return ListView(
      padding: const EdgeInsets.all(20.0),
      children: [
        // big training count
        _buildTrainingCount(l10n),
        const SizedBox(height: 32),

        // overall calibration bar (expandable)
        CalibrationWidget(families: families),
        const SizedBox(height: 24),

        // family capability bars
        _buildSection(
          l10n.capabilities,
          FamilyProgressWidget(families: families),
        ),
      ],
    );
  }

  Widget _buildTrainingCount(AppLocalizations l10n) {
    final count = _progress?.trainingsComplete ?? 0;

    return Column(
      children: [
        Text(
          '$count',
          style: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.titleStyle.copyWith(
                  fontSize: 72,
                  fontWeight: FontWeight.w700,
                )
              : Theme.of(context).textTheme.displayLarge?.copyWith(
                    fontSize: 72,
                    fontWeight: FontWeight.w700,
                  ),
        ),
        const SizedBox(height: 4),
        Text(
          l10n.completedTrainings,
          style: PlatformHelper.useLiquidGlass
              ? LiquidGlassTheme.captionStyle
              : Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
        ),
      ],
    );
  }

  Widget _buildSection(String title, Widget child) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 4),
          child: Text(
            title,
            style: PlatformHelper.useLiquidGlass
                ? LiquidGlassTheme.headlineStyle.copyWith(fontSize: 18)
                : Theme.of(context).textTheme.titleMedium,
          ),
        ),
        const SizedBox(height: 12),
        AdaptiveCard(
          padding: const EdgeInsets.all(16),
          child: child,
        ),
      ],
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
                Icons.trending_up,
                size: 64,
                color: Colors.grey[400],
              ),
              const SizedBox(height: 16),
              Text(
                l10n.yourProgress,
                style: PlatformHelper.useLiquidGlass
                    ? LiquidGlassTheme.headlineStyle
                    : Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: 8),
              Text(
                l10n.noProgressYet,
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

  Widget _buildCalibrationMessage(AppLocalizations l10n) {
    return Container(
      padding: const EdgeInsets.all(20),
      decoration: PlatformHelper.useLiquidGlass
          ? LiquidGlassTheme.glassDecoration()
          : BoxDecoration(
              color: Theme.of(context).cardColor,
              borderRadius: BorderRadius.circular(12),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withOpacity(0.05),
                  blurRadius: 10,
                ),
              ],
            ),
      child: Row(
        children: [
          Icon(
            Icons.tune,
            size: 32,
            color: Colors.grey.shade500,
          ),
          const SizedBox(width: 16),
          Expanded(
            child: Text(
              l10n.calibrationNeeded,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.bodyStyle
                  : Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Colors.grey[600],
                      ),
            ),
          ),
        ],
      ),
    );
  }
}

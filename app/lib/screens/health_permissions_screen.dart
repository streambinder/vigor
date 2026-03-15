import 'package:flutter/material.dart';
import 'package:health/health.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../services/android_health_data_service.dart';
import '../services/app_logger.dart';
import '../services/preferences_service.dart';
import '../services/service_locator.dart';
import '../widgets/adaptive/adaptive.dart';

class HealthPermissionsScreen extends StatelessWidget {
  const HealthPermissionsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(title: Text(l10n.healthPermissionsTitle)),
      body: Padding(
        padding: VigorSpacing.paddingLg,
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.stretch,
          children: [
            const Spacer(flex: 1),
            Icon(Icons.monitor_heart_outlined, size: 64, color: VigorColors.indigoAdaptive(context)),
            const SizedBox(height: VigorSpacing.lg),
            Text(
              l10n.healthPermissionsDescription,
              style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context), height: 1.6),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: VigorSpacing.xl),
            // data types list
            _buildPermissionItem(context, Icons.bedtime_outlined, l10n.healthPermissionsSleep),
            _buildPermissionItem(context, Icons.show_chart, l10n.healthPermissionsHrv),
            _buildPermissionItem(context, Icons.favorite_outline, l10n.healthPermissionsRhr),
            _buildPermissionItem(context, Icons.directions_walk, l10n.healthPermissionsSteps),
            _buildPermissionItem(context, Icons.fitness_center, l10n.healthPermissionsWorkouts),
            const Spacer(flex: 2),
            // read-only disclaimer — centered, right above connect button
            Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.lock_outline, size: 14, color: VigorColors.stone),
                const SizedBox(width: VigorSpacing.xs),
                Text(l10n.healthPermissionsReadOnly, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
              ],
            ),
            const SizedBox(height: VigorSpacing.md),
            // connect button
            SizedBox(
              width: double.infinity,
              child: Material(
                color: VigorColors.persimmon,
                borderRadius: VigorRadius.radiusSm,
                child: InkWell(
                  onTap: () => _requestPermissions(context, l10n),
                  borderRadius: VigorRadius.radiusSm,
                  child: Padding(
                    padding: VigorSpacing.buttonPadding,
                    child: Text(
                      l10n.healthPermissionsGrant,
                      style: VigorTypography.label.copyWith(color: Colors.white, fontWeight: FontWeight.w600),
                      textAlign: TextAlign.center,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(height: VigorSpacing.sm),
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: Text(l10n.healthPermissionsSkip, style: TextStyle(color: VigorColors.stone)),
            ),
            const SizedBox(height: VigorSpacing.lg),
          ],
        ),
      ),
    );
  }

  Widget _buildPermissionItem(BuildContext context, IconData icon, String label) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: VigorSpacing.xs),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: VigorColors.indigoAdaptive(context).withValues(alpha: 0.1),
              borderRadius: VigorRadius.radiusSm,
            ),
            child: Icon(icon, size: 18, color: VigorColors.indigoAdaptive(context)),
          ),
          const SizedBox(width: VigorSpacing.md),
          Text(label, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
        ],
      ),
    );
  }

  Future<void> _requestPermissions(BuildContext context, AppLocalizations l10n) async {
    AppLogger.info('[HealthPermissions] user tapped connect');
    final locator = context.read<ServiceLocator>();
    final healthService = locator.healthDataService;
    if (healthService == null) {
      AppLogger.warning('[HealthPermissions] no health service available');
      return;
    }

    // on android, check HC SDK availability before requesting permissions
    if (healthService is AndroidHealthDataService) {
      AppLogger.debug('[HealthPermissions] checking Android Health Connect SDK status');
      final status = await healthService.getSdkStatus();
      if (status != HealthConnectSdkStatus.sdkAvailable) {
        AppLogger.warning('[HealthPermissions] Health Connect SDK not available: $status');
        if (!context.mounted) return;
        await AdaptiveAlertDialog.show(
          context: context,
          title: l10n.healthInstallHcTitle,
          content: l10n.healthInstallHcDescription,
          actions: [
            AdaptiveDialogAction(label: l10n.ok, onPressed: () => Navigator.of(context).pop()),
          ],
        );
        return;
      }
      AppLogger.debug('[HealthPermissions] Health Connect SDK available');
    }

    final granted = await healthService.requestPermissions();

    if (!context.mounted) return;

    if (granted) {
      AppLogger.info('[HealthPermissions] connected — setting hcConnected=true, triggering initial sync');
      final prefs = context.read<PreferencesService>();
      await prefs.setHcConnected(true);
      // trigger initial sync immediately
      healthService.syncToBackend();
      if (context.mounted) {
        AdaptiveNotification.show(context: context, message: l10n.healthPermissionsGranted);
        // pop back to settings — setState in settings rebuilds with syncing state
        Navigator.of(context).pop(true);
      }
    } else {
      AppLogger.info('[HealthPermissions] user denied permissions');
      AdaptiveNotification.showError(context: context, message: l10n.healthPermissionsDenied);
    }
  }
}

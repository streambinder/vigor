import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/theme_provider.dart';
import '../services/app_logger.dart';
import '../services/authenticated_api_service.dart';
import '../services/preferences_service.dart';
import '../services/secure_storage_service.dart';
import '../services/health_data_service.dart';
import '../services/service_locator.dart';
import '../utils/platform_helper.dart';
import '../widgets/adaptive/adaptive.dart';
import 'health_permissions_screen.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  late int _defaultDuration;
  late bool _intervalJingle;
  late bool _duckOtherAudio;
  late bool _warmupCooldown;
  late bool _useRecommendedDuration;
  late bool _liveTimerNotification;

  @override
  void initState() {
    super.initState();
    final prefs = context.read<PreferencesService>();
    _defaultDuration = prefs.defaultDuration;
    _intervalJingle = prefs.intervalJingle;
    _duckOtherAudio = prefs.duckOtherAudio;
    _warmupCooldown = prefs.warmupCooldown;
    _useRecommendedDuration = prefs.useRecommendedDuration;
    _liveTimerNotification = prefs.liveTimerNotification;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(title: Text(l10n.settings)),
      body: ListView.builder(
        padding: VigorSpacing.paddingLg,
        itemCount: 5,
        itemBuilder: (context, index) {
          switch (index) {
            case 0:
              return Padding(
                padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
                child: _buildAppearanceSection(context, l10n, isDark),
              );
            case 1:
              return Padding(
                padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
                child: _buildTimerSection(context, l10n, isDark),
              );
            case 2:
              return Padding(
                padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
                child: _buildTrainingSection(context, l10n, isDark),
              );
            case 3:
              return Padding(
                padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
                child: _buildHealthSection(context, l10n, isDark),
              );
            case 4:
              return Padding(
                padding: const EdgeInsets.only(bottom: VigorSpacing.lg),
                child: _buildLogsSection(context, l10n, isDark),
              );
            default:
              return const SizedBox.shrink();
          }
        },
      ),
    );
  }

  Widget _buildAppearanceSection(BuildContext context, AppLocalizations l10n, bool isDark) {
    // use Selector to only rebuild when themeModeString changes, not on every provider update
    return Selector<ThemeProvider, String>(
      selector: (_, provider) => provider.themeModeString,
      builder: (context, currentMode, child) => Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(
                color: VigorColors.indigoAdaptive(context).withValues(alpha: 0.15),
                borderRadius: VigorRadius.radiusSm,
              ),
              child: Icon(Icons.palette, color: VigorColors.indigoAdaptive(context), size: 20),
            ),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.appearance, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
          ],
        ),
        const SizedBox(height: VigorSpacing.md),
        Container(
          decoration: BoxDecoration(
            color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
            borderRadius: VigorRadius.radiusMd,
          ),
          child: Column(
            children: [
              _buildThemeOption(
                context: context,
                value: 'system',
                currentValue: currentMode,
                title: l10n.themeAuto,
                subtitle: l10n.themeAutoDescription,
                icon: Icons.brightness_auto,
                onTap: () => context.read<ThemeProvider>().setThemeMode('system'),
              ),
              Divider(height: 1, color: VigorColors.border(context)),
              _buildThemeOption(
                context: context,
                value: 'light',
                currentValue: currentMode,
                title: l10n.themeLight,
                icon: Icons.light_mode,
                onTap: () => context.read<ThemeProvider>().setThemeMode('light'),
              ),
              Divider(height: 1, color: VigorColors.border(context)),
              _buildThemeOption(
                context: context,
                value: 'dark',
                currentValue: currentMode,
                title: l10n.themeDark,
                icon: Icons.dark_mode,
                onTap: () => context.read<ThemeProvider>().setThemeMode('dark'),
              ),
            ],
          ),
        ),
        ],
      ),
    );
  }

  Widget _buildTimerSection(BuildContext context, AppLocalizations l10n, bool isDark) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(
                color: VigorColors.indigoAdaptive(context).withValues(alpha: 0.15),
                borderRadius: VigorRadius.radiusSm,
              ),
              child: Icon(Icons.timer, color: VigorColors.indigoAdaptive(context), size: 20),
            ),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.timer, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
          ],
        ),
        const SizedBox(height: VigorSpacing.md),
        Container(
          decoration: BoxDecoration(
            color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
            borderRadius: VigorRadius.radiusMd,
          ),
          child: Column(
            children: [
              SwitchListTile(
                title: Text(l10n.intervalJingle, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                value: _intervalJingle,
                activeThumbColor: VigorColors.indigo,
                onChanged: (value) {
                  setState(() => _intervalJingle = value);
                  context.read<PreferencesService>().setIntervalJingle(value);
                },
              ),
              Divider(height: 1, color: VigorColors.border(context)),
              IgnorePointer(
                ignoring: !_intervalJingle,
                child: AnimatedOpacity(
                  opacity: _intervalJingle ? 1.0 : 0.3,
                  duration: VigorAnimation.fast,
                  child: SwitchListTile(
                    title: Text(l10n.duckOtherAudio, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                    subtitle: Text(l10n.duckOtherAudioDescription, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                    value: _duckOtherAudio,
                    activeThumbColor: VigorColors.indigo,
                    onChanged: (value) {
                      setState(() => _duckOtherAudio = value);
                      context.read<PreferencesService>().setDuckOtherAudio(value);
                    },
                  ),
                ),
              ),
              Divider(height: 1, color: VigorColors.border(context)),
              SwitchListTile(
                title: Text(l10n.liveTimerNotification, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                subtitle: Text(l10n.liveTimerNotificationDescription, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                value: _liveTimerNotification,
                activeThumbColor: VigorColors.indigo,
                onChanged: (value) {
                  setState(() => _liveTimerNotification = value);
                  context.read<PreferencesService>().setLiveTimerNotification(value);
                },
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildTrainingSection(BuildContext context, AppLocalizations l10n, bool isDark) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(
                color: VigorColors.indigoAdaptive(context).withValues(alpha: 0.15),
                borderRadius: VigorRadius.radiusSm,
              ),
              child: Icon(Icons.fitness_center, color: VigorColors.indigoAdaptive(context), size: 20),
            ),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.trainingDefaults, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
          ],
        ),
        const SizedBox(height: VigorSpacing.md),
        Container(
          padding: VigorSpacing.paddingMd,
          decoration: BoxDecoration(
            color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
            borderRadius: VigorRadius.radiusMd,
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Text(l10n.defaultDuration, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                  ),
                  Text(
                    _useRecommendedDuration ? l10n.recommended : '$_defaultDuration min',
                    style: VigorTypography.data.copyWith(
                      color: VigorColors.indigo,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: VigorSpacing.xs),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Text(
                      l10n.recommended,
                      style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context)),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Switch(
                    value: _useRecommendedDuration,
                    activeThumbColor: VigorColors.indigo,
                    onChanged: (value) {
                      setState(() => _useRecommendedDuration = value);
                      context.read<PreferencesService>().setUseRecommendedDuration(value);
                    },
                  ),
                ],
              ),
              IgnorePointer(
                ignoring: _useRecommendedDuration,
                child: AnimatedOpacity(
                  opacity: _useRecommendedDuration ? 0.3 : 1.0,
                  duration: VigorAnimation.fast,
                  child: Slider(
                    value: _defaultDuration.toDouble(),
                    min: 10,
                    max: 180,
                    divisions: 34,
                    activeColor: VigorColors.indigo,
                    onChanged: (value) {
                      setState(() => _defaultDuration = value.round());
                      context.read<PreferencesService>().setDefaultDuration(value.round());
                    },
                  ),
                ),
              ),
              const SizedBox(height: VigorSpacing.sm),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Expanded(
                    child: Text(
                      l10n.warmupCooldown,
                      style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Switch(
                    value: _warmupCooldown,
                    activeThumbColor: VigorColors.indigo,
                    onChanged: (value) {
                      setState(() => _warmupCooldown = value);
                      context.read<PreferencesService>().setWarmupCooldown(value);
                    },
                  ),
                ],
              ),
            ],
          ),
        ),
      ],
    );
  }

  String _formatShortDate(DateTime dt) => DateFormat.MMMd().format(dt);

  Widget _buildHealthSection(BuildContext context, AppLocalizations l10n, bool isDark) {
    final prefs = context.read<PreferencesService>();
    final healthService = context.read<ServiceLocator>().healthDataService;
    final isConnected = prefs.hcConnected;
    final isNative = !PlatformHelper.isWeb;

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(
                color: VigorColors.indigoAdaptive(context).withValues(alpha: 0.15),
                borderRadius: VigorRadius.radiusSm,
              ),
              child: Icon(Icons.monitor_heart, color: VigorColors.indigoAdaptive(context), size: 20),
            ),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.healthData, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
          ],
        ),
        const SizedBox(height: VigorSpacing.md),
        IgnorePointer(
          ignoring: !isNative,
          child: AnimatedOpacity(
            opacity: isNative ? 1.0 : 0.3,
            duration: VigorAnimation.fast,
            child: Container(
              decoration: BoxDecoration(
                color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
                borderRadius: VigorRadius.radiusMd,
              ),
              child: Column(
                children: [
                  if (isConnected && healthService != null)
                    ValueListenableBuilder<bool>(
                      valueListenable: healthService.syncing,
                      builder: (context, isSyncing, _) => ValueListenableBuilder<HealthSyncResult?>(
                        valueListenable: healthService.lastSyncResult,
                        builder: (context, syncResult, _) {
                          final hasDeviceData = syncResult != null && (syncResult.deviceMetrics > 0 || syncResult.deviceSessions > 0);
                          final showNoDataWarning = syncResult != null && !hasDeviceData && syncResult.wasForced;
                          final hasBackendData = syncResult != null && (syncResult.totalMetrics > 0 || syncResult.totalSessions > 0);
                          final hasSyncError = syncResult?.syncError != null;

                          String? backendDateRange;
                          if (hasBackendData) {
                            final dates = [syncResult.metricsFrom, syncResult.sessionsFrom].whereType<DateTime>();
                            final endDates = [syncResult.metricsTo, syncResult.sessionsTo].whereType<DateTime>();
                            if (dates.isNotEmpty && endDates.isNotEmpty) {
                              final from = dates.reduce((a, b) => a.isBefore(b) ? a : b);
                              final to = endDates.reduce((a, b) => a.isAfter(b) ? a : b);
                              backendDateRange = l10n.healthDateRange(_formatShortDate(from), _formatShortDate(to));
                            }
                          }

                          return Column(children: [
                            // status row
                            ListTile(
                              leading: isSyncing
                                  ? SizedBox(width: 22, height: 22, child: CircularProgressIndicator(strokeWidth: 2, color: VigorColors.indigoAdaptive(context)))
                                  : hasSyncError
                                      ? const Icon(Icons.error_outline, color: VigorColors.warning, size: 22)
                                      : Icon(Icons.check_circle, color: VigorColors.indigoAdaptive(context), size: 22),
                              title: Text(
                                isSyncing ? l10n.healthSynchronizing : hasSyncError ? l10n.healthSyncFailed : l10n.healthSynchronized,
                                style: VigorTypography.body.copyWith(color: isSyncing ? VigorColors.textPrimary(context) : hasSyncError ? VigorColors.warning : VigorColors.textPrimary(context)),
                              ),
                              subtitle: !isSyncing && syncResult != null
                                  ? Text(
                                      hasSyncError
                                          ? l10n.healthSyncFailedDetail(syncResult.syncError!)
                                          : syncResult.wasForced ? l10n.healthSyncTypeFull : l10n.healthSyncTypeIncremental,
                                      style: VigorTypography.caption.copyWith(color: VigorColors.stone),
                                    )
                                  : null,
                            ),
                            // device + backend data block (compact, no dividers between them)
                            if (!isSyncing && hasDeviceData) ...[
                              Divider(height: 1, color: VigorColors.border(context)),
                              ...syncResult.deviceSources.entries.map((entry) => ListTile(
                                dense: true,
                                visualDensity: VisualDensity.compact,
                                leading: const Icon(Icons.smartphone, color: VigorColors.stone, size: 22),
                                title: Text(entry.key, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
                                subtitle: Text(
                                  l10n.healthSourceData(entry.value.metrics, entry.value.sessions),
                                  style: VigorTypography.caption.copyWith(color: VigorColors.stone.withValues(alpha: 0.6)),
                                ),
                              )),
                            ] else if (!isSyncing && showNoDataWarning) ...[
                              Divider(height: 1, color: VigorColors.border(context)),
                              ListTile(
                                dense: true,
                                visualDensity: VisualDensity.compact,
                                leading: const Icon(Icons.warning_amber_rounded, color: VigorColors.warning, size: 22),
                                title: Text(l10n.healthSyncNoData, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
                              ),
                            ],
                            if (!isSyncing && hasBackendData)
                              ListTile(
                                dense: true,
                                visualDensity: VisualDensity.compact,
                                leading: const Icon(Icons.cloud_done_outlined, color: VigorColors.stone, size: 22),
                                title: Text(l10n.healthBackend, style: VigorTypography.caption.copyWith(color: VigorColors.stone)),
                                subtitle: Text(
                                  backendDateRange != null
                                      ? '${l10n.healthBackendData(syncResult.totalMetrics, syncResult.totalSessions)} · $backendDateRange'
                                      : l10n.healthBackendData(syncResult.totalMetrics, syncResult.totalSessions),
                                  style: VigorTypography.caption.copyWith(color: VigorColors.stone.withValues(alpha: 0.6)),
                                ),
                              ),
                            // sync now
                            Divider(height: 1, color: VigorColors.border(context)),
                            ListTile(
                              leading: Icon(Icons.sync, color: isSyncing ? VigorColors.stone : VigorColors.indigoAdaptive(context), size: 22),
                              title: Text(l10n.healthSynchronize, style: VigorTypography.body.copyWith(color: isSyncing ? VigorColors.stone : VigorColors.textPrimary(context))),
                              onTap: isSyncing ? null : () {
                                AppLogger.info('[Settings] manual sync triggered');
                                final healthService = context.read<ServiceLocator>().healthDataService;
                                if (healthService == null) return;
                                healthService.syncToBackend(fullRescan: true);
                              },
                            ),
                            // disconnect
                            Divider(height: 1, color: VigorColors.border(context)),
                            ListTile(
                              leading: const Icon(Icons.delete_outline, color: VigorColors.crimson, size: 22),
                              title: Text(l10n.healthDisconnect, style: VigorTypography.body.copyWith(color: VigorColors.crimson)),
                              onTap: () => _showDisconnectDialog(context, l10n),
                            ),
                          ]);
                        },
                      ),
                    )
                  else
                    ListTile(
                      leading: const Icon(Icons.link_off, color: VigorColors.stone, size: 22),
                      title: Text(l10n.healthNotConnected, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                      subtitle: !isNative ? Text(l10n.healthNativeOnly, style: VigorTypography.caption.copyWith(color: VigorColors.stone)) : null,
                      trailing: healthService != null
                          ? TextButton(
                              onPressed: () async {
                                await Navigator.of(context).push(
                                  MaterialPageRoute(builder: (_) => const HealthPermissionsScreen()),
                                );
                                if (context.mounted) setState(() {});
                              },
                              child: Text(l10n.healthConnect, style: TextStyle(color: VigorColors.indigoAdaptive(context))),
                            )
                          : null,
                    ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }

  Future<void> _showDisconnectDialog(BuildContext context, AppLocalizations l10n) async {
    final confirmed = await AdaptiveAlertDialog.show<bool>(
      context: context,
      title: l10n.healthDisconnect,
      content: l10n.healthDisconnectConfirmation,
      actions: [
        AdaptiveDialogAction(label: l10n.cancel, onPressed: () => Navigator.of(context).pop(false)),
        AdaptiveDialogAction(label: l10n.delete, isDestructive: true, onPressed: () => Navigator.of(context).pop(true)),
      ],
    );

    if (confirmed != true || !context.mounted) return;

    AppLogger.info('[Settings] disconnecting health — calling POST /health/disconnect');
    final response = await AuthenticatedApiService(
      storageService: context.read<SecureStorageService>(),
    ).post('/health/disconnect');
    if (!context.mounted) return;

    if (response.isSuccess) {
      AppLogger.info('[Settings] health disconnected — clearing local data');
      final prefs = context.read<PreferencesService>();
      await prefs.clearHealthData();
      await prefs.setHcConnected(false);
      if (context.mounted) {
        setState(() {});
        AdaptiveNotification.show(context: context, message: l10n.healthDisconnectedSuccessfully);
      }
    } else {
      AppLogger.error('[Settings] health disconnect failed: ${response.error}');
      AdaptiveNotification.showError(context: context, message: l10n.failedToDisconnectHealth, rawError: response.error);
    }
  }

  Widget _buildLogsSection(BuildContext context, AppLocalizations l10n, bool isDark) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: VigorSpacing.paddingSm,
              decoration: BoxDecoration(
                color: VigorColors.indigoAdaptive(context).withValues(alpha: 0.15),
                borderRadius: VigorRadius.radiusSm,
              ),
              child: Icon(Icons.article_outlined, color: VigorColors.indigoAdaptive(context), size: 20),
            ),
            const SizedBox(width: VigorSpacing.sm),
            Text(l10n.appLogs, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
          ],
        ),
        const SizedBox(height: VigorSpacing.md),
        Container(
          decoration: BoxDecoration(
            color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
            borderRadius: VigorRadius.radiusMd,
          ),
          child: Column(
            children: [
              ListTile(
                leading: Icon(Icons.visibility_outlined, color: VigorColors.indigoAdaptive(context), size: 22),
                title: Text(l10n.viewLogs, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                onTap: () => _showLogsModal(context, l10n),
              ),
            ],
          ),
        ),
      ],
    );
  }

  void _showLogsModal(BuildContext context, AppLocalizations l10n) {
    final logs = AppLogger.logs;
    showModalBottomSheet(
      context: context,
      isScrollControlled: true,
      useSafeArea: true,
      backgroundColor: Theme.of(context).scaffoldBackgroundColor,
      shape: const RoundedRectangleBorder(borderRadius: BorderRadius.vertical(top: Radius.circular(VigorSpacing.md))),
      builder: (context) => DraggableScrollableSheet(
        initialChildSize: 0.9,
        minChildSize: 0.5,
        maxChildSize: 1.0,
        expand: false,
        builder: (context, scrollController) => Column(
          children: [
            Padding(
              padding: VigorSpacing.paddingMd,
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(l10n.appLogs, style: VigorTypography.headline.copyWith(fontSize: 18, color: VigorColors.textPrimary(context))),
                  Row(
                    children: [
                      IconButton(
                        icon: Icon(Icons.copy, size: 20, color: VigorColors.textSecondary(context)),
                        onPressed: () {
                          if (logs.isEmpty) return;
                          Clipboard.setData(ClipboardData(text: logs.join('\n')));
                          AdaptiveNotification.show(context: context, message: l10n.copied);
                        },
                      ),
                      IconButton(icon: const Icon(Icons.close), onPressed: () => Navigator.of(context).pop()),
                    ],
                  ),
                ],
              ),
            ),
            Divider(height: 1, color: VigorColors.border(context)),
            Expanded(
              child: logs.isEmpty
                  ? Center(child: Text(l10n.noLogsYet, style: VigorTypography.body.copyWith(color: VigorColors.stone)))
                  : ListView.builder(
                      controller: scrollController,
                      padding: VigorSpacing.paddingSm,
                      itemCount: logs.length,
                      itemBuilder: (context, index) => Padding(
                        padding: const EdgeInsets.symmetric(vertical: 1),
                        child: Text(
                          logs[index],
                          style: VigorTypography.caption.copyWith(
                            fontFamily: 'monospace',
                            fontSize: 9,
                            color: VigorColors.textSecondary(context),
                          ),
                        ),
                      ),
                    ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildThemeOption({
    required BuildContext context,
    required String value,
    required String currentValue,
    required String title,
    String? subtitle,
    required IconData icon,
    required VoidCallback onTap,
  }) {
    final isSelected = value == currentValue;
    final accentColor = VigorColors.indigoAdaptive(context);
    return ListTile(
      leading: Icon(icon, color: isSelected ? accentColor : VigorColors.stone, size: 22),
      title: Text(title, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context), fontWeight: isSelected ? FontWeight.w600 : FontWeight.w400)),
      subtitle: subtitle != null ? Text(subtitle, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))) : null,
      trailing: isSelected ? Icon(Icons.check, color: accentColor, size: 20) : null,
      onTap: onTap,
    );
  }
}

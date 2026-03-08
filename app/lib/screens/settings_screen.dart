import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/theme_provider.dart';
import '../services/authenticated_api_service.dart';
import '../services/preferences_service.dart';
import '../services/secure_storage_service.dart';
import '../services/service_locator.dart';
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

  @override
  void initState() {
    super.initState();
    final prefs = context.read<PreferencesService>();
    _defaultDuration = prefs.defaultDuration;
    _intervalJingle = prefs.intervalJingle;
    _duckOtherAudio = prefs.duckOtherAudio;
    _warmupCooldown = prefs.warmupCooldown;
    _useRecommendedDuration = prefs.useRecommendedDuration;
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
                activeColor: VigorColors.indigo,
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
                    activeColor: VigorColors.indigo,
                    onChanged: (value) {
                      setState(() => _duckOtherAudio = value);
                      context.read<PreferencesService>().setDuckOtherAudio(value);
                    },
                  ),
                ),
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
                  Text(l10n.recommended, style: VigorTypography.caption.copyWith(color: VigorColors.textSecondary(context))),
                  Switch(
                    value: _useRecommendedDuration,
                    activeColor: VigorColors.indigo,
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
                  Text(l10n.warmupCooldown, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                  Switch(
                    value: _warmupCooldown,
                    activeColor: VigorColors.indigo,
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

  Widget _buildHealthSection(BuildContext context, AppLocalizations l10n, bool isDark) {
    final prefs = context.read<PreferencesService>();
    final healthService = context.read<ServiceLocator>().healthDataService;
    final isConnected = prefs.hcConnected;

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
        Container(
          decoration: BoxDecoration(
            color: isDark ? VigorColors.darkSurface : VigorColors.lightSurface,
            borderRadius: VigorRadius.radiusMd,
          ),
          child: Column(
            children: [
              if (isConnected && healthService != null)
                ValueListenableBuilder<bool>(
                  valueListenable: healthService.syncing,
                  builder: (context, isSyncing, _) => Column(
                    children: [
                      ListTile(
                        leading: isSyncing
                            ? SizedBox(width: 22, height: 22, child: CircularProgressIndicator(strokeWidth: 2, color: VigorColors.indigoAdaptive(context)))
                            : Icon(Icons.check_circle, color: VigorColors.indigoAdaptive(context), size: 22),
                        title: Text(
                          isSyncing ? l10n.healthSynchronizing : l10n.healthSynchronized,
                          style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context)),
                        ),
                      ),
                      Divider(height: 1, color: VigorColors.border(context)),
                      ListTile(
                        leading: Icon(Icons.sync, color: isSyncing ? VigorColors.stone : VigorColors.indigoAdaptive(context), size: 22),
                        title: Text(l10n.healthSynchronize, style: VigorTypography.body.copyWith(color: isSyncing ? VigorColors.stone : VigorColors.textPrimary(context))),
                        onTap: isSyncing ? null : () {
                          final healthService = context.read<ServiceLocator>().healthDataService;
                          if (healthService == null) return;
                          healthService.syncToBackend(force: true);
                        },
                      ),
                      Divider(height: 1, color: VigorColors.border(context)),
                      ListTile(
                        leading: const Icon(Icons.delete_outline, color: VigorColors.crimson, size: 22),
                        title: Text(l10n.healthDisconnect, style: VigorTypography.body.copyWith(color: VigorColors.crimson)),
                        onTap: () => _showDisconnectDialog(context, l10n),
                      ),
                    ],
                  ),
                )
              else if (!isConnected)
                ListTile(
                  leading: Icon(Icons.link_off, color: VigorColors.stone, size: 22),
                  title: Text(l10n.healthNotConnected, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
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

    final response = await AuthenticatedApiService(
      storageService: context.read<SecureStorageService>(),
    ).post('/health/disconnect');
    if (!context.mounted) return;

    if (response.isSuccess) {
      final prefs = context.read<PreferencesService>();
      await prefs.clearHealthData();
      await prefs.setHcConnected(false);
      if (context.mounted) {
        setState(() {});
        AdaptiveNotification.show(context: context, message: l10n.healthDisconnectedSuccessfully);
      }
    } else {
      AdaptiveNotification.showError(context: context, message: l10n.failedToDisconnectHealth, rawError: response.error);
    }
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

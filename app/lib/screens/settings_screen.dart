import 'package:flutter/material.dart';
import 'package:provider/provider.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../providers/theme_provider.dart';
import '../services/preferences_service.dart';
import '../widgets/adaptive/adaptive.dart';

class SettingsScreen extends StatefulWidget {
  const SettingsScreen({super.key});

  @override
  State<SettingsScreen> createState() => _SettingsScreenState();
}

class _SettingsScreenState extends State<SettingsScreen> {
  late int _defaultDuration;

  @override
  void initState() {
    super.initState();
    _defaultDuration = context.read<PreferencesService>().defaultDuration;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;

    return AdaptiveScaffold(
      appBar: AdaptiveAppBar(title: Text(l10n.settings)),
      body: ListView(
        padding: VigorSpacing.paddingLg,
        children: [
          _buildAppearanceSection(context, l10n, isDark),
          const SizedBox(height: VigorSpacing.lg),
          _buildTrainingSection(context, l10n, isDark),
          const SizedBox(height: VigorSpacing.lg),
        ],
      ),
    );
  }

  Widget _buildAppearanceSection(BuildContext context, AppLocalizations l10n, bool isDark) {
    final themeProvider = context.watch<ThemeProvider>();
    final currentMode = themeProvider.themeModeString;

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
                onTap: () => themeProvider.setThemeMode('system'),
              ),
              Divider(height: 1, color: VigorColors.border(context)),
              _buildThemeOption(
                context: context,
                value: 'light',
                currentValue: currentMode,
                title: l10n.themeLight,
                icon: Icons.light_mode,
                onTap: () => themeProvider.setThemeMode('light'),
              ),
              Divider(height: 1, color: VigorColors.border(context)),
              _buildThemeOption(
                context: context,
                value: 'dark',
                currentValue: currentMode,
                title: l10n.themeDark,
                icon: Icons.dark_mode,
                onTap: () => themeProvider.setThemeMode('dark'),
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
                  Text(l10n.defaultDuration, style: VigorTypography.body.copyWith(color: VigorColors.textPrimary(context))),
                  Text('$_defaultDuration min', style: VigorTypography.data.copyWith(color: VigorColors.indigo, fontWeight: FontWeight.w600)),
                ],
              ),
              Slider(
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
            ],
          ),
        ),
      ],
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

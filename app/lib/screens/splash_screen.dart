import 'package:flutter/material.dart';
import '../generated/app_localizations.dart';
import '../widgets/adaptive/adaptive.dart';
import '../theme/liquid_glass_theme.dart';
import '../utils/platform_helper.dart';

class SplashScreen extends StatelessWidget {
  const SplashScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return AdaptiveScaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              Icons.fitness_center,
              size: 100,
              color: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.primaryColor
                  : Theme.of(context).colorScheme.primary,
            ),
            const SizedBox(height: 24),
            Text(
              l10n.appName,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.titleStyle
                  : const TextStyle(
                      fontSize: 36,
                      fontWeight: FontWeight.bold,
                    ),
            ),
            const SizedBox(height: 16),
            const AdaptiveLoadingIndicator(),
          ],
        ),
      ),
    );
  }
}

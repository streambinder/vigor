import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import '../widgets/adaptive/adaptive.dart';

class SplashScreen extends StatelessWidget {
  const SplashScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final textColor = VigorColors.textPrimary(context);

    return AdaptiveScaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            // lightning bolt icon with gradient effect
            ShaderMask(
              shaderCallback: (bounds) => LinearGradient(
                begin: Alignment.topLeft,
                end: Alignment.bottomRight,
                colors: [VigorColors.orange, VigorColors.electricBlue],
              ).createShader(bounds),
              child: const Icon(
                Icons.bolt,
                size: 100,
                color: Colors.white,
              ),
            ),
            SizedBox(height: VigorSpacing.lg),
            Text(
              l10n.appName.toUpperCase(),
              style: VigorTypography.display.copyWith(
                color: textColor,
                letterSpacing: 4,
              ),
            ),
            SizedBox(height: VigorSpacing.md),
            const AdaptiveLoadingIndicator(),
          ],
        ),
      ),
    );
  }
}

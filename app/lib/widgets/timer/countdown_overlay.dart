import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';

class CountdownOverlay extends StatelessWidget {
  final int remainingSeconds;
  final VoidCallback onTap;

  const CountdownOverlay({
    super.key,
    required this.remainingSeconds,
    required this.onTap,
  });

  String _formatTime(int seconds) {
    final minutes = seconds ~/ 60;
    final secs = seconds % 60;
    return '${minutes.toString().padLeft(2, '0')}:${secs.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final isDark = Theme.of(context).brightness == Brightness.dark;
    return GestureDetector(
      onTap: onTap,
      child: Container(
        color: isDark ? VigorColors.darkBackground : VigorColors.lightBackground,
        child: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(
                Icons.campaign,
                size: 120,
                color: VigorColors.stone,
              ),
              const SizedBox(height: VigorSpacing.lg),
              Text(
                l10n.tapToStart,
                style: VigorTypography.headline.copyWith(
                  color: VigorColors.textPrimary(context),
                ),
              ),
              const SizedBox(height: VigorSpacing.xxl),
              Text(
                _formatTime(remainingSeconds),
                style: VigorTypography.dataDisplay.copyWith(
                  color: VigorColors.persimmon.withValues(alpha: 0.5),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

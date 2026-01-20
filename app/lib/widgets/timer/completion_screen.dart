import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../generated/app_localizations.dart';
import '../adaptive/adaptive.dart';

class CompletionScreen extends StatelessWidget {
  final String trainingName;
  final VoidCallback onDone;

  const CompletionScreen({
    super.key,
    required this.trainingName,
    required this.onDone,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    return Center(
      child: Padding(
        padding: VigorSpacing.paddingLg,
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Container(
              width: 120,
              height: 120,
              decoration: BoxDecoration(
                shape: BoxShape.circle,
                color: VigorColors.gold.withValues(alpha: 0.15),
              ),
              child: const Icon(
                Icons.check_circle,
                size: 80,
                color: VigorColors.gold,
              ),
            ),
            const SizedBox(height: VigorSpacing.xl),
            Text(
              l10n.trainingCompleted,
              style: VigorTypography.title.copyWith(
                color: VigorColors.textPrimary(context),
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: VigorSpacing.md),
            Text(
              l10n.greatJobCompleting(trainingName),
              style: VigorTypography.body.copyWith(
                color: VigorColors.textSecondary(context),
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: VigorSpacing.xxl),
            AdaptiveButton(
              onPressed: onDone,
              child: Text(l10n.done),
            ),
          ],
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../generated/app_localizations.dart';
import 'adaptive/adaptive_card.dart';
import 'adaptive/adaptive_dialog.dart';

/// Daily training-readiness hint shown on the homepage: a colored chip with the
/// verdict for the day (train / easy / rest). Tap opens the readines summary.
/// Rendered only when the backend produced a readiness score for today; when
/// recovery data is missing the caller hides the widget entirely.
class ReadinessHintCard extends StatelessWidget {
  final int score;
  final String level; // green | yellow | red
  final String summary;

  const ReadinessHintCard({
    super.key,
    required this.score,
    required this.level,
    required this.summary,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context);
    final (color, icon, label) = switch (level) {
      'green' => (
        VigorColors.success,
        Icons.check_circle_outline,
        l10n.readinessGreen,
      ),
      'yellow' => (VigorColors.warning, Icons.speed, l10n.readinessYellow),
      _ => (VigorColors.error, Icons.hotel_outlined, l10n.readinessRed),
    };

    return AdaptiveCard(
      onTap: () => AdaptiveAlertDialog.show<void>(
        context: context,
        title: l10n.readinessTitle,
        content: '$score/100\n\n$summary',
        actions: [
          AdaptiveDialogAction(
            label: l10n.close,
            onPressed: () => Navigator.of(context).pop(),
          ),
        ],
      ),
      child: Row(
        children: [
          Icon(icon, size: 22, color: color),
          const SizedBox(width: VigorSpacing.sm),
          Expanded(
            child: Text(
              label,
              style: VigorTypography.headline.copyWith(
                fontSize: 16,
                color: VigorColors.textPrimary(context),
              ),
            ),
          ),
          Text(
            '$score',
            style: VigorTypography.headline.copyWith(
              fontSize: 20,
              color: color,
            ),
          ),
          const SizedBox(width: VigorSpacing.xs),
          Icon(
            Icons.chevron_right,
            size: 20,
            color: VigorColors.textSecondary(context),
          ),
        ],
      ),
    );
  }
}

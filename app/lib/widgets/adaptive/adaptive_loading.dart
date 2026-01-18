import 'package:flutter/material.dart';
import '../../design/tokens.dart';
import '../../utils/platform_helper.dart';

/// Platform-adaptive loading indicator
class AdaptiveLoadingIndicator extends StatelessWidget {
  final double? value;
  final Color? color;

  const AdaptiveLoadingIndicator({
    super.key,
    this.value,
    this.color,
  });

  @override
  Widget build(BuildContext context) {
    final indicatorColor = color ?? VigorColors.indigoAdaptive(context);

    if (PlatformHelper.useLiquidGlass) {
      return SizedBox(
        width: 24,
        height: 24,
        child: CircularProgressIndicator(
          value: value,
          strokeWidth: 2.5,
          color: indicatorColor,
        ),
      );
    }

    return CircularProgressIndicator(
      value: value,
      color: indicatorColor,
    );
  }
}

/// Platform-adaptive centered loading screen
class AdaptiveLoadingScreen extends StatelessWidget {
  final String? message;

  const AdaptiveLoadingScreen({super.key, this.message});

  @override
  Widget build(BuildContext context) {
    final textColor = VigorColors.textPrimary(context);

    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const AdaptiveLoadingIndicator(),
          if (message != null) ...[
            const SizedBox(height: VigorSpacing.md),
            Text(
              message!,
              style: VigorTypography.body.copyWith(color: textColor),
            ),
          ],
        ],
      ),
    );
  }
}

import 'package:flutter/material.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

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
    if (PlatformHelper.useLiquidGlass) {
      return SizedBox(
        width: 24,
        height: 24,
        child: CircularProgressIndicator(
          value: value,
          strokeWidth: 2.5,
          color: color ?? LiquidGlassTheme.primaryColor,
        ),
      );
    }

    return CircularProgressIndicator(
      value: value,
      color: color,
    );
  }
}

/// Platform-adaptive centered loading screen
class AdaptiveLoadingScreen extends StatelessWidget {
  final String? message;

  const AdaptiveLoadingScreen({super.key, this.message});

  @override
  Widget build(BuildContext context) {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const AdaptiveLoadingIndicator(),
          if (message != null) ...[
            const SizedBox(height: 16),
            Text(
              message!,
              style: PlatformHelper.useLiquidGlass
                  ? LiquidGlassTheme.bodyStyle
                  : Theme.of(context).textTheme.bodyMedium,
            ),
          ],
        ],
      ),
    );
  }
}

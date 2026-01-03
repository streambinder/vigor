import 'package:flutter/material.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive scaffold
/// Uses Liquid Glass gradient background on iOS and Material Scaffold on other platforms
class AdaptiveScaffold extends StatelessWidget {
  final PreferredSizeWidget? appBar;
  final Widget body;
  final Widget? bottomNavigationBar;
  final Widget? floatingActionButton;
  final Color? backgroundColor;
  final bool useLiquidGlassBackground;

  const AdaptiveScaffold({
    super.key,
    this.appBar,
    required this.body,
    this.bottomNavigationBar,
    this.floatingActionButton,
    this.backgroundColor,
    this.useLiquidGlassBackground = true,
  });

  @override
  Widget build(BuildContext context) {
    final isDark = Theme.of(context).brightness == Brightness.dark;

    if (PlatformHelper.useLiquidGlass && useLiquidGlassBackground) {
      return Container(
        decoration: BoxDecoration(
          gradient: isDark
              ? LiquidGlassTheme.darkBackgroundGradient
              : LiquidGlassTheme.backgroundGradient,
        ),
        child: Scaffold(
          backgroundColor: Colors.transparent,
          appBar: appBar,
          body: body,
          bottomNavigationBar: bottomNavigationBar,
          floatingActionButton: floatingActionButton,
        ),
      );
    }

    return Scaffold(
      appBar: appBar,
      body: body,
      bottomNavigationBar: bottomNavigationBar,
      floatingActionButton: floatingActionButton,
      backgroundColor: backgroundColor,
    );
  }
}

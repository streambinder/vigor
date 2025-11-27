import 'dart:ui';
import 'package:flutter/material.dart';
import '../../utils/platform_helper.dart';
import '../../theme/liquid_glass_theme.dart';

/// Platform-adaptive card
/// Uses Liquid Glass effect on iOS and Material Card on other platforms
class AdaptiveCard extends StatelessWidget {
  final Widget child;
  final EdgeInsetsGeometry? margin;
  final EdgeInsetsGeometry? padding;
  final VoidCallback? onTap;
  final Color? glassColor;
  final bool useVibrantGradient;
  final List<Color>? gradientColors;

  const AdaptiveCard({
    super.key,
    required this.child,
    this.margin,
    this.padding,
    this.onTap,
    this.glassColor,
    this.useVibrantGradient = false,
    this.gradientColors,
  });

  @override
  Widget build(BuildContext context) {
    if (PlatformHelper.useLiquidGlass) {
      return _buildLiquidGlassCard(context);
    }
    return _buildMaterialCard(context);
  }

  Widget _buildLiquidGlassCard(BuildContext context) {
    final cardChild = Container(
      margin: margin ?? const EdgeInsets.all(12),
      padding: padding ?? const EdgeInsets.all(16),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(16),
        child: BackdropFilter(
          filter: ImageFilter.blur(
            sigmaX: LiquidGlassTheme.glassBlur,
            sigmaY: LiquidGlassTheme.glassBlur,
          ),
          child: Container(
            decoration: useVibrantGradient
                ? LiquidGlassTheme.vibrantGradient(
                    colors: gradientColors ??
                        [
                          LiquidGlassTheme.primaryColor.withOpacity(0.8),
                          LiquidGlassTheme.accentColor.withOpacity(0.8),
                        ],
                  )
                : LiquidGlassTheme.glassDecoration(
                    baseColor: glassColor,
                  ),
            child: child,
          ),
        ),
      ),
    );

    if (onTap != null) {
      return GestureDetector(
        onTap: onTap,
        child: cardChild,
      );
    }

    return cardChild;
  }

  Widget _buildMaterialCard(BuildContext context) {
    if (useVibrantGradient) {
      return Container(
        margin: margin ?? const EdgeInsets.all(8),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: gradientColors ??
                [
                  Theme.of(context).colorScheme.primaryContainer,
                  Theme.of(context).colorScheme.secondaryContainer,
                ],
          ),
          borderRadius: BorderRadius.circular(16),
        ),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onTap,
            borderRadius: BorderRadius.circular(16),
            child: Padding(
              padding: padding ?? const EdgeInsets.all(16),
              child: child,
            ),
          ),
        ),
      );
    }

    return Card(
      margin: margin ?? const EdgeInsets.all(8),
      color: glassColor,
      child: InkWell(
        onTap: onTap,
        borderRadius: BorderRadius.circular(16),
        child: Padding(
          padding: padding ?? const EdgeInsets.all(16),
          child: child,
        ),
      ),
    );
  }
}

import 'dart:ui';
import 'package:flutter/material.dart';
import '../../design/tokens.dart';
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
    final isDark = Theme.of(context).brightness == Brightness.dark;

    final cardChild = Container(
      margin: margin ?? VigorSpacing.paddingXs,
      padding: padding ?? VigorSpacing.paddingMd,
      child: ClipRRect(
        borderRadius: VigorRadius.card,
        child: BackdropFilter(
          filter: ImageFilter.blur(
            sigmaX: VigorColors.glassBlur,
            sigmaY: VigorColors.glassBlur,
          ),
          child: Container(
            decoration: useVibrantGradient
                ? LiquidGlassTheme.vibrantGradient(
                    colors: gradientColors ??
                        [
                          VigorColors.orange.withValues(alpha: 0.8),
                          VigorColors.electricBlue.withValues(alpha: 0.8),
                        ],
                    borderRadius: VigorRadius.md,
                  )
                : LiquidGlassTheme.glassDecoration(
                    baseColor: glassColor,
                    isDark: isDark,
                    borderRadius: VigorRadius.md,
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
        margin: margin ?? VigorSpacing.paddingXs,
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: gradientColors ??
                [
                  VigorColors.orange.withValues(alpha: 0.8),
                  VigorColors.electricBlue.withValues(alpha: 0.8),
                ],
          ),
          borderRadius: VigorRadius.card,
        ),
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: onTap,
            borderRadius: VigorRadius.card,
            child: Padding(
              padding: padding ?? VigorSpacing.paddingMd,
              child: child,
            ),
          ),
        ),
      );
    }

    return Card(
      margin: margin ?? VigorSpacing.paddingXs,
      color: glassColor,
      child: InkWell(
        onTap: onTap,
        borderRadius: VigorRadius.card,
        child: Padding(
          padding: padding ?? VigorSpacing.paddingMd,
          child: child,
        ),
      ),
    );
  }
}

import 'package:flutter/material.dart';
import '../design/tokens.dart';

/// Liquid Glass design system for iOS
/// Features glassmorphism with frosted blur effects, transparency, and depth
/// Now using Vigor design tokens for consistency
class LiquidGlassTheme {
  // Primary colors from design tokens
  static Color get primaryColor => VigorColors.orange;
  static Color get accentColor => VigorColors.electricBlue;
  static Color get successColor => VigorColors.success;
  static Color get errorColor => VigorColors.error;
  static Color get warningColor => VigorColors.warning;

  // Glass effect parameters from design tokens
  static double get glassOpacity => VigorColors.glassOpacityDark;
  static double get glassBlur => VigorColors.glassBlur;
  static double get glassBorderOpacity => VigorColors.glassBorderOpacity;

  // Shadows for depth
  static List<BoxShadow> softShadow(BuildContext context) {
    return VigorShadows.elevation1(context);
  }

  static List<BoxShadow> get glowShadow => VigorShadows.orangeGlow;

  // Background gradients
  static LinearGradient get backgroundGradient => const LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [
          VigorColors.lightBackground,
          Color(0xFFE8EEF5),
        ],
      );

  static LinearGradient get darkBackgroundGradient => const LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [
          VigorColors.darkBackground,
          VigorColors.darkSurface,
        ],
      );

  /// Creates a glass container decoration
  static BoxDecoration glassDecoration({
    Color? baseColor,
    double? opacity,
    double borderRadius = 16,
    List<BoxShadow>? shadows,
    Border? border,
    bool isDark = false,
  }) {
    final effectiveOpacity = opacity ??
        (isDark ? VigorColors.glassOpacityDark : VigorColors.glassOpacityLight);
    final effectiveBaseColor = baseColor ?? (isDark ? Colors.black : Colors.white);

    return BoxDecoration(
      color: effectiveBaseColor.withValues(alpha: effectiveOpacity),
      borderRadius: BorderRadius.circular(borderRadius),
      border: border ??
          Border.all(
            color: Colors.white.withValues(alpha: glassBorderOpacity),
            width: 1.5,
          ),
      boxShadow: shadows,
    );
  }

  /// Creates a vibrant gradient decoration
  static BoxDecoration vibrantGradient({
    required List<Color> colors,
    double borderRadius = 16,
    List<BoxShadow>? shadows,
  }) {
    return BoxDecoration(
      gradient: LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: colors,
      ),
      borderRadius: BorderRadius.circular(borderRadius),
      boxShadow: shadows ?? glowShadow,
    );
  }

  /// Primary button gradient (orange energy)
  static BoxDecoration get primaryButtonDecoration => vibrantGradient(
        colors: [
          VigorColors.orange,
          VigorColors.orange.withRed(240),
        ],
        borderRadius: VigorRadius.sm,
        shadows: VigorShadows.orangeGlow,
      );

  /// Secondary button gradient (electric blue)
  static BoxDecoration get secondaryButtonDecoration => vibrantGradient(
        colors: [
          VigorColors.electricBlue,
          VigorColors.electricBlue.withBlue(240),
        ],
        borderRadius: VigorRadius.sm,
        shadows: VigorShadows.blueGlow,
      );

  /// Text styles for Liquid Glass (using design tokens)
  static TextStyle get titleStyle => VigorTypography.title;
  static TextStyle get headlineStyle => VigorTypography.headline;
  static TextStyle get bodyStyle => VigorTypography.bodyLarge;
  static TextStyle get captionStyle => VigorTypography.caption;

  /// Get theme-aware text styles
  static TextStyle titleStyleColored(BuildContext context) {
    return VigorTypography.titleColored(context);
  }

  static TextStyle headlineStyleColored(BuildContext context) {
    return VigorTypography.headlineColored(context);
  }

  static TextStyle bodyStyleColored(BuildContext context) {
    return VigorTypography.bodyLargeColored(context);
  }

  static TextStyle captionStyleColored(BuildContext context) {
    return VigorTypography.captionColored(context);
  }

  /// Material Theme (not used on iOS but kept for compatibility)
  static ThemeData get materialTheme {
    return ThemeData(
      colorScheme: ColorScheme.fromSeed(
        seedColor: primaryColor,
        brightness: Brightness.light,
      ),
      useMaterial3: true,
    );
  }
}

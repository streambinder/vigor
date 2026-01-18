import 'package:flutter/material.dart';
import '../design/tokens.dart';

/// Liquid Glass design system for iOS
/// Features glassmorphism with frosted blur effects, transparency, and depth
/// Now using Vigor design tokens (IDENTITY.md) for consistency
class LiquidGlassTheme {
  // Primary colors from design tokens
  // Note: primaryColor is used for data viz/progress bars (indigo)
  // ctaColor is used for buttons/actions (persimmon)
  static Color get primaryColor => VigorColors.indigo;
  static Color get ctaColor => VigorColors.persimmon;
  static Color get accentColor => VigorColors.indigo;
  static Color get successColor => VigorColors.success;
  static Color get errorColor => VigorColors.crimson;
  static Color get warningColor => VigorColors.gold;
  static Color get achievementColor => VigorColors.gold;

  // Glass effect parameters from design tokens
  static double get glassOpacity => VigorColors.glassOpacityDark;
  static double get glassBlur => VigorColors.glassBlur;
  static double get glassBorderOpacity => VigorColors.glassBorderOpacity;

  // Shadows for depth
  static List<BoxShadow> softShadow(BuildContext context) => VigorShadows.elevation1(context);

  static List<BoxShadow> get glowShadow => VigorShadows.persimmonGlow;

  // Background gradients (Sumi-based for dark, Washi-based for light)
  static LinearGradient get backgroundGradient => const LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [VigorColors.lightBackground, Color(0xFFE8EAE5)],
      );

  static LinearGradient get darkBackgroundGradient => const LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [VigorColors.darkBackground, VigorColors.darkSurface],
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
    final effectiveOpacity =
        opacity ?? (isDark ? VigorColors.glassOpacityDark : VigorColors.glassOpacityLight);
    final effectiveBaseColor = baseColor ?? (isDark ? Colors.black : Colors.white);

    return BoxDecoration(
      color: effectiveBaseColor.withValues(alpha: effectiveOpacity),
      borderRadius: BorderRadius.circular(borderRadius),
      border: border ??
          Border.all(color: Colors.white.withValues(alpha: glassBorderOpacity), width: 1.5),
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
      gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: colors),
      borderRadius: BorderRadius.circular(borderRadius),
      boxShadow: shadows ?? glowShadow,
    );
  }

  /// Primary button gradient (persimmon = warm/active energy)
  static BoxDecoration get primaryButtonDecoration => vibrantGradient(
        colors: [VigorColors.persimmon, VigorColors.persimmon.withRed(240)],
        borderRadius: VigorRadius.sm,
        shadows: VigorShadows.persimmonGlow,
      );

  /// Secondary button gradient (indigo = cool/recovered)
  static BoxDecoration get secondaryButtonDecoration => vibrantGradient(
        colors: [VigorColors.indigo, VigorColors.indigo.withBlue(100)],
        borderRadius: VigorRadius.sm,
        shadows: VigorShadows.indigoGlow,
      );

  /// Achievement button gradient (gold = kintsugi/mastery)
  static BoxDecoration get achievementButtonDecoration => vibrantGradient(
        colors: [VigorColors.gold, VigorColors.gold.withRed(220)],
        borderRadius: VigorRadius.sm,
        shadows: VigorShadows.goldGlow,
      );

  /// Text styles for Liquid Glass (using design tokens)
  static TextStyle get titleStyle => VigorTypography.title;
  static TextStyle get headlineStyle => VigorTypography.headline;
  static TextStyle get bodyStyle => VigorTypography.bodyLarge;
  static TextStyle get captionStyle => VigorTypography.caption;
  static TextStyle get dataStyle => VigorTypography.data;

  static TextStyle titleStyleColored(BuildContext context) => VigorTypography.titleColored(context);
  static TextStyle headlineStyleColored(BuildContext context) => VigorTypography.headlineColored(context);
  static TextStyle bodyStyleColored(BuildContext context) => VigorTypography.bodyLargeColored(context);
  static TextStyle captionStyleColored(BuildContext context) => VigorTypography.captionColored(context);
  static TextStyle dataStyleColored(BuildContext context) => VigorTypography.dataColored(context);

  /// Material Theme (not used on iOS but kept for compatibility)
  static ThemeData get materialTheme {
    return ThemeData(
      colorScheme: ColorScheme.fromSeed(seedColor: primaryColor, brightness: Brightness.light),
      useMaterial3: true,
    );
  }
}

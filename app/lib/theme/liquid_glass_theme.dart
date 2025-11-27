import 'dart:ui';
import 'package:flutter/material.dart';

/// Liquid Glass design system constants and utilities
/// Features glassmorphism with frosted blur effects, transparency, and depth
class LiquidGlassTheme {
  // Primary colors with vibrant tones
  static const Color primaryColor = Color(0xFF007AFF); // iOS blue
  static const Color accentColor = Color(0xFF5E5CE6); // Purple accent
  static const Color successColor = Color(0xFF34C759); // Green
  static const Color errorColor = Color(0xFFFF3B30); // Red
  static const Color warningColor = Color(0xFFFF9500); // Orange

  // Glass effect parameters
  static const double glassOpacity = 0.15;
  static const double glassBlur = 20.0;
  static const double glassBorderOpacity = 0.2;

  // Shadows for depth
  static List<BoxShadow> get softShadow => [
        BoxShadow(
          color: Colors.black.withOpacity(0.1),
          blurRadius: 20,
          offset: const Offset(0, 10),
        ),
        BoxShadow(
          color: Colors.black.withOpacity(0.05),
          blurRadius: 10,
          offset: const Offset(0, 5),
        ),
      ];

  static List<BoxShadow> get glowShadow => [
        BoxShadow(
          color: primaryColor.withOpacity(0.3),
          blurRadius: 20,
          offset: const Offset(0, 5),
        ),
      ];

  // Background gradients
  static LinearGradient get backgroundGradient => LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [
          const Color(0xFFF5F7FA),
          const Color(0xFFE8EEF5),
        ],
      );

  static LinearGradient get darkBackgroundGradient => LinearGradient(
        begin: Alignment.topLeft,
        end: Alignment.bottomRight,
        colors: [
          const Color(0xFF1C1C1E),
          const Color(0xFF2C2C2E),
        ],
      );

  /// Creates a glass container decoration
  static BoxDecoration glassDecoration({
    Color? baseColor,
    double opacity = glassOpacity,
    double borderRadius = 16,
    List<BoxShadow>? shadows,
    Border? border,
  }) {
    return BoxDecoration(
      color: (baseColor ?? Colors.white).withOpacity(opacity),
      borderRadius: BorderRadius.circular(borderRadius),
      border: border ??
          Border.all(
            color: Colors.white.withOpacity(glassBorderOpacity),
            width: 1.5,
          ),
      boxShadow: shadows ?? softShadow,
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

  /// Text styles for Liquid Glass
  static const TextStyle titleStyle = TextStyle(
    fontSize: 28,
    fontWeight: FontWeight.bold,
    letterSpacing: -0.5,
    color: Color(0xFF1C1C1E),
  );

  static const TextStyle headlineStyle = TextStyle(
    fontSize: 22,
    fontWeight: FontWeight.w600,
    letterSpacing: -0.3,
    color: Color(0xFF1C1C1E),
  );

  static const TextStyle bodyStyle = TextStyle(
    fontSize: 17,
    fontWeight: FontWeight.normal,
    letterSpacing: -0.4,
    color: Color(0xFF3C3C43),
  );

  static const TextStyle captionStyle = TextStyle(
    fontSize: 13,
    fontWeight: FontWeight.normal,
    letterSpacing: -0.1,
    color: Color(0xFF8E8E93),
  );

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

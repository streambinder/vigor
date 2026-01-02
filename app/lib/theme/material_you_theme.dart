import 'package:flutter/material.dart';
import '../design/tokens.dart';
import '../design/vigor_theme.dart';

/// Material You (Material Design 3) theme configuration
/// Used on Android, Web, and all non-iOS platforms
/// Now using Vigor design tokens for consistency
class MaterialYouTheme {
  // Seed color from design tokens
  static Color get seedColor => VigorColors.orange;
  static Color get accentColor => VigorColors.electricBlue;

  /// Material 3 light theme - delegates to VigorTheme
  static ThemeData get lightTheme => VigorTheme.light;

  /// Material 3 dark theme - delegates to VigorTheme
  static ThemeData get darkTheme => VigorTheme.dark;
}

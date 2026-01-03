import 'package:flutter/material.dart';
import 'tokens.dart';

/// Vigor unified theme builder
/// Creates consistent Material themes for both light and dark modes
/// Uses design tokens for all values

class VigorTheme {
  /// Light theme
  static ThemeData get light {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      colorScheme: _lightColorScheme,
      fontFamily: VigorTypography.fontFamily,
      scaffoldBackgroundColor: VigorColors.lightBackground,
      cardColor: VigorColors.lightSurface,
      dividerColor: VigorColors.lightBorder,
      textTheme: _textTheme(Brightness.light),
      cardTheme: _cardTheme(),
      inputDecorationTheme: _inputDecorationTheme(Brightness.light),
      elevatedButtonTheme: _elevatedButtonTheme(),
      filledButtonTheme: _filledButtonTheme(),
      textButtonTheme: _textButtonTheme(),
      outlinedButtonTheme: _outlinedButtonTheme(Brightness.light),
      dialogTheme: _dialogTheme(),
      bottomSheetTheme: _bottomSheetTheme(),
      appBarTheme: _appBarTheme(Brightness.light),
      listTileTheme: _listTileTheme(),
      navigationBarTheme: _navigationBarTheme(Brightness.light),
      switchTheme: _switchTheme(),
      chipTheme: _chipTheme(Brightness.light),
      progressIndicatorTheme: _progressIndicatorTheme(),
    );
  }

  /// Dark theme
  static ThemeData get dark {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.dark,
      colorScheme: _darkColorScheme,
      fontFamily: VigorTypography.fontFamily,
      scaffoldBackgroundColor: VigorColors.darkBackground,
      cardColor: VigorColors.darkSurface,
      dividerColor: VigorColors.darkBorder,
      textTheme: _textTheme(Brightness.dark),
      cardTheme: _cardTheme(),
      inputDecorationTheme: _inputDecorationTheme(Brightness.dark),
      elevatedButtonTheme: _elevatedButtonTheme(),
      filledButtonTheme: _filledButtonTheme(),
      textButtonTheme: _textButtonTheme(),
      outlinedButtonTheme: _outlinedButtonTheme(Brightness.dark),
      dialogTheme: _dialogTheme(),
      bottomSheetTheme: _bottomSheetTheme(),
      appBarTheme: _appBarTheme(Brightness.dark),
      listTileTheme: _listTileTheme(),
      navigationBarTheme: _navigationBarTheme(Brightness.dark),
      switchTheme: _switchTheme(),
      chipTheme: _chipTheme(Brightness.dark),
      progressIndicatorTheme: _progressIndicatorTheme(),
    );
  }

  // Color schemes
  static ColorScheme get _lightColorScheme => ColorScheme(
        brightness: Brightness.light,
        primary: VigorColors.orange,
        onPrimary: Colors.white,
        secondary: VigorColors.electricBlue,
        onSecondary: Colors.white,
        tertiary: VigorColors.info,
        onTertiary: Colors.white,
        error: VigorColors.error,
        onError: Colors.white,
        surface: VigorColors.lightSurface,
        onSurface: VigorColors.lightTextPrimary,
        surfaceContainerHighest: VigorColors.lightSurfaceElevated,
        onSurfaceVariant: VigorColors.lightTextSecondary,
        outline: VigorColors.lightBorder,
        outlineVariant: VigorColors.lightBorder.withValues(alpha: 0.5),
      );

  static ColorScheme get _darkColorScheme => ColorScheme(
        brightness: Brightness.dark,
        primary: VigorColors.orange,
        onPrimary: Colors.white,
        secondary: VigorColors.electricBlue,
        onSecondary: Colors.white,
        tertiary: VigorColors.info,
        onTertiary: Colors.white,
        error: VigorColors.error,
        onError: Colors.white,
        surface: VigorColors.darkSurface,
        onSurface: VigorColors.darkTextPrimary,
        surfaceContainerHighest: VigorColors.darkSurfaceElevated,
        onSurfaceVariant: VigorColors.darkTextSecondary,
        outline: VigorColors.darkBorder,
        outlineVariant: VigorColors.darkBorder.withValues(alpha: 0.5),
      );

  // Text theme
  static TextTheme _textTheme(Brightness brightness) {
    final textColor = brightness == Brightness.dark
        ? VigorColors.darkTextPrimary
        : VigorColors.lightTextPrimary;
    final secondaryColor = brightness == Brightness.dark
        ? VigorColors.darkTextSecondary
        : VigorColors.lightTextSecondary;

    return TextTheme(
      displayLarge: VigorTypography.display.copyWith(color: textColor),
      displayMedium: VigorTypography.display.copyWith(color: textColor, fontSize: 28),
      displaySmall: VigorTypography.display.copyWith(color: textColor, fontSize: 24),
      headlineLarge: VigorTypography.title.copyWith(color: textColor),
      headlineMedium: VigorTypography.headline.copyWith(color: textColor),
      headlineSmall: VigorTypography.headline.copyWith(color: textColor, fontSize: 18),
      titleLarge: VigorTypography.title.copyWith(color: textColor),
      titleMedium: VigorTypography.headline.copyWith(color: textColor),
      titleSmall: VigorTypography.headline.copyWith(color: textColor, fontSize: 16),
      bodyLarge: VigorTypography.bodyLarge.copyWith(color: textColor),
      bodyMedium: VigorTypography.body.copyWith(color: textColor),
      bodySmall: VigorTypography.caption.copyWith(color: secondaryColor),
      labelLarge: VigorTypography.label.copyWith(color: textColor),
      labelMedium: VigorTypography.label.copyWith(color: textColor, fontSize: 12),
      labelSmall: VigorTypography.caption.copyWith(color: secondaryColor),
    );
  }

  // Card theme
  static CardThemeData _cardTheme() {
    return const CardThemeData(
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.card),
      clipBehavior: Clip.antiAlias,
      margin: EdgeInsets.zero,
    );
  }

  // Input decoration theme
  static InputDecorationTheme _inputDecorationTheme(Brightness brightness) {
    final borderColor = brightness == Brightness.dark
        ? VigorColors.darkBorder
        : VigorColors.lightBorder;
    final fillColor = brightness == Brightness.dark
        ? VigorColors.darkSurface
        : VigorColors.lightSurface;

    return InputDecorationTheme(
      filled: true,
      fillColor: fillColor,
      border: OutlineInputBorder(
        borderRadius: VigorRadius.input,
        borderSide: BorderSide(color: borderColor),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: VigorRadius.input,
        borderSide: BorderSide(color: borderColor),
      ),
      focusedBorder: const OutlineInputBorder(
        borderRadius: VigorRadius.input,
        borderSide: BorderSide(color: VigorColors.orange, width: 2),
      ),
      errorBorder: const OutlineInputBorder(
        borderRadius: VigorRadius.input,
        borderSide: BorderSide(color: VigorColors.error),
      ),
      focusedErrorBorder: const OutlineInputBorder(
        borderRadius: VigorRadius.input,
        borderSide: BorderSide(color: VigorColors.error, width: 2),
      ),
      contentPadding: VigorSpacing.inputPadding,
      labelStyle: VigorTypography.body,
      hintStyle: VigorTypography.body.copyWith(
        color: brightness == Brightness.dark
            ? VigorColors.darkTextMuted
            : VigorColors.lightTextMuted,
      ),
    );
  }

  // Elevated button theme
  static ElevatedButtonThemeData _elevatedButtonTheme() {
    return ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: VigorColors.orange,
        foregroundColor: Colors.white,
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: VigorSpacing.buttonPadding,
        elevation: 0,
        textStyle: VigorTypography.label,
      ),
    );
  }

  // Filled button theme
  static FilledButtonThemeData _filledButtonTheme() {
    return FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: VigorColors.orange,
        foregroundColor: Colors.white,
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: VigorSpacing.buttonPadding,
        textStyle: VigorTypography.label,
      ),
    );
  }

  // Text button theme
  static TextButtonThemeData _textButtonTheme() {
    return TextButtonThemeData(
      style: TextButton.styleFrom(
        foregroundColor: VigorColors.orange,
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: const EdgeInsets.symmetric(
          horizontal: VigorSpacing.md,
          vertical: VigorSpacing.sm,
        ),
        textStyle: VigorTypography.label,
      ),
    );
  }

  // Outlined button theme
  static OutlinedButtonThemeData _outlinedButtonTheme(Brightness brightness) {
    final borderColor = brightness == Brightness.dark
        ? VigorColors.darkBorder
        : VigorColors.lightBorder;

    return OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: VigorColors.orange,
        side: BorderSide(color: borderColor),
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: VigorSpacing.buttonPadding,
        textStyle: VigorTypography.label,
      ),
    );
  }

  // Dialog theme
  static DialogThemeData _dialogTheme() {
    return const DialogThemeData(
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.modal),
      elevation: 0,
    );
  }

  // Bottom sheet theme
  static BottomSheetThemeData _bottomSheetTheme() {
    return const BottomSheetThemeData(
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.bottomSheet),
      clipBehavior: Clip.antiAlias,
    );
  }

  // App bar theme
  static AppBarTheme _appBarTheme(Brightness brightness) {
    return AppBarTheme(
      centerTitle: false,
      elevation: 0,
      scrolledUnderElevation: 0,
      backgroundColor: brightness == Brightness.dark
          ? VigorColors.darkBackground
          : VigorColors.lightBackground,
      foregroundColor: brightness == Brightness.dark
          ? VigorColors.darkTextPrimary
          : VigorColors.lightTextPrimary,
      titleTextStyle: VigorTypography.headline.copyWith(
        color: brightness == Brightness.dark
            ? VigorColors.darkTextPrimary
            : VigorColors.lightTextPrimary,
      ),
    );
  }

  // List tile theme
  static ListTileThemeData _listTileTheme() {
    return const ListTileThemeData(
      contentPadding: VigorSpacing.listTilePadding,
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.radiusSm),
    );
  }

  // Navigation bar theme
  static NavigationBarThemeData _navigationBarTheme(Brightness brightness) {
    final bgColor = brightness == Brightness.dark
        ? VigorColors.darkSurface
        : VigorColors.lightSurface;
    final indicatorColor = VigorColors.orange.withValues(alpha: 0.15);

    return NavigationBarThemeData(
      height: 70,
      backgroundColor: bgColor,
      indicatorColor: indicatorColor,
      labelTextStyle: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return VigorTypography.label.copyWith(
            color: VigorColors.orange,
            fontWeight: FontWeight.w600,
          );
        }
        return VigorTypography.label.copyWith(
          color: brightness == Brightness.dark
              ? VigorColors.darkTextSecondary
              : VigorColors.lightTextSecondary,
        );
      }),
      iconTheme: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return const IconThemeData(color: VigorColors.orange, size: 24);
        }
        return IconThemeData(
          color: brightness == Brightness.dark
              ? VigorColors.darkTextSecondary
              : VigorColors.lightTextSecondary,
          size: 24,
        );
      }),
    );
  }

  // Switch theme
  static SwitchThemeData _switchTheme() {
    return SwitchThemeData(
      thumbColor: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return Colors.white;
        }
        return VigorColors.darkTextSecondary;
      }),
      trackColor: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return VigorColors.orange;
        }
        return VigorColors.darkBorder;
      }),
      trackOutlineColor: WidgetStateProperty.all(Colors.transparent),
    );
  }

  // Chip theme
  static ChipThemeData _chipTheme(Brightness brightness) {
    return ChipThemeData(
      backgroundColor: brightness == Brightness.dark
          ? VigorColors.darkSurface
          : VigorColors.lightSurface,
      selectedColor: VigorColors.orange.withValues(alpha: 0.15),
      labelStyle: VigorTypography.label,
      shape: const RoundedRectangleBorder(borderRadius: VigorRadius.chip),
      side: BorderSide(
        color: brightness == Brightness.dark
            ? VigorColors.darkBorder
            : VigorColors.lightBorder,
      ),
    );
  }

  // Progress indicator theme
  static ProgressIndicatorThemeData _progressIndicatorTheme() {
    return const ProgressIndicatorThemeData(
      color: VigorColors.orange,
      linearTrackColor: VigorColors.darkBorder,
      circularTrackColor: VigorColors.darkBorder,
    );
  }
}

/// Extension for quick access to Vigor colors from context
extension VigorThemeExtension on BuildContext {
  bool get isDarkMode => Theme.of(this).brightness == Brightness.dark;

  Color get vigorOrange => VigorColors.orange;
  Color get vigorBlue => VigorColors.electricBlue;
  Color get vigorSuccess => VigorColors.success;
  Color get vigorWarning => VigorColors.warning;
  Color get vigorError => VigorColors.error;

  Color get background => VigorColors.background(this);
  Color get surface => VigorColors.surface(this);
  Color get surfaceElevated => VigorColors.surfaceElevated(this);
  Color get borderColor => VigorColors.border(this);

  Color get textPrimary => VigorColors.textPrimary(this);
  Color get textSecondary => VigorColors.textSecondary(this);
  Color get textMuted => VigorColors.textMuted(this);
}
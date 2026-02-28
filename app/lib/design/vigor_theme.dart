import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'tokens.dart';

/// Vigor unified theme builder
/// Based on IDENTITY.md: Japanese Aesthetics × Data Science
/// Uses design tokens for all values

class VigorTheme {
  /// Light theme
  static ThemeData get light {
    return ThemeData(
      useMaterial3: true,
      brightness: Brightness.light,
      colorScheme: _lightColorScheme,
      fontFamily: VigorTypography.fontFamilyBody,
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
      fontFamily: VigorTypography.fontFamilyBody,
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

  // Color schemes — using IDENTITY.md palette
  // Note: primary is indigo (for data viz, selections)
  // Buttons explicitly use persimmon for CTAs
  static ColorScheme get _lightColorScheme => ColorScheme(
        brightness: Brightness.light,
        primary: VigorColors.indigo,
        onPrimary: VigorColors.washi,
        secondary: VigorColors.persimmon,
        onSecondary: VigorColors.washi,
        tertiary: VigorColors.gold,
        onTertiary: VigorColors.sumi,
        error: VigorColors.crimson,
        onError: VigorColors.washi,
        surface: VigorColors.lightSurface,
        onSurface: VigorColors.lightTextPrimary,
        surfaceContainerHighest: VigorColors.lightSurfaceElevated,
        onSurfaceVariant: VigorColors.lightTextSecondary,
        outline: VigorColors.lightBorder,
        outlineVariant: VigorColors.lightBorder.withValues(alpha: 0.5),
      );

  static ColorScheme get _darkColorScheme => ColorScheme(
        brightness: Brightness.dark,
        primary: VigorColors.indigoLight,
        onPrimary: VigorColors.sumi,
        secondary: VigorColors.persimmon,
        onSecondary: VigorColors.washi,
        tertiary: VigorColors.gold,
        onTertiary: VigorColors.sumi,
        error: VigorColors.crimson,
        onError: VigorColors.washi,
        surface: VigorColors.darkSurface,
        onSurface: VigorColors.darkTextPrimary,
        surfaceContainerHighest: VigorColors.darkSurfaceElevated,
        onSurfaceVariant: VigorColors.darkTextSecondary,
        outline: VigorColors.darkBorder,
        outlineVariant: VigorColors.darkBorder.withValues(alpha: 0.5),
      );

  // Text theme — maps Material text styles to Vigor typography
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

  static CardThemeData _cardTheme() {
    return const CardThemeData(
      elevation: 0,
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.card),
      clipBehavior: Clip.antiAlias,
      margin: EdgeInsets.zero,
    );
  }

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
        borderSide: BorderSide(color: VigorColors.indigo, width: 2),
      ),
      errorBorder: const OutlineInputBorder(
        borderRadius: VigorRadius.input,
        borderSide: BorderSide(color: VigorColors.crimson),
      ),
      focusedErrorBorder: const OutlineInputBorder(
        borderRadius: VigorRadius.input,
        borderSide: BorderSide(color: VigorColors.crimson, width: 2),
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

  static ElevatedButtonThemeData _elevatedButtonTheme() {
    return ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: VigorColors.persimmon,
        foregroundColor: VigorColors.washi,
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: VigorSpacing.buttonPadding,
        elevation: 0,
        textStyle: VigorTypography.label,
      ),
    );
  }

  static FilledButtonThemeData _filledButtonTheme() {
    return FilledButtonThemeData(
      style: FilledButton.styleFrom(
        backgroundColor: VigorColors.persimmon,
        foregroundColor: VigorColors.washi,
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: VigorSpacing.buttonPadding,
        textStyle: VigorTypography.label,
      ),
    );
  }

  static TextButtonThemeData _textButtonTheme() {
    return TextButtonThemeData(
      style: TextButton.styleFrom(
        foregroundColor: VigorColors.indigo,
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: const EdgeInsets.symmetric(
          horizontal: VigorSpacing.md,
          vertical: VigorSpacing.sm,
        ),
        textStyle: VigorTypography.label,
      ),
    );
  }

  static OutlinedButtonThemeData _outlinedButtonTheme(Brightness brightness) {
    final borderColor = brightness == Brightness.dark
        ? VigorColors.darkBorder
        : VigorColors.lightBorder;

    return OutlinedButtonThemeData(
      style: OutlinedButton.styleFrom(
        foregroundColor: VigorColors.indigo,
        side: BorderSide(color: borderColor),
        shape: const RoundedRectangleBorder(borderRadius: VigorRadius.button),
        padding: VigorSpacing.buttonPadding,
        textStyle: VigorTypography.label,
      ),
    );
  }

  static DialogThemeData _dialogTheme() {
    return const DialogThemeData(
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.modal),
      elevation: 0,
    );
  }

  static BottomSheetThemeData _bottomSheetTheme() {
    return const BottomSheetThemeData(
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.bottomSheet),
      clipBehavior: Clip.antiAlias,
    );
  }

  static AppBarTheme _appBarTheme(Brightness brightness) {
    final isDark = brightness == Brightness.dark;
    return AppBarTheme(
      centerTitle: false,
      elevation: 0,
      scrolledUnderElevation: 0,
      backgroundColor: isDark ? VigorColors.darkBackground : VigorColors.lightBackground,
      foregroundColor: isDark ? VigorColors.darkTextPrimary : VigorColors.lightTextPrimary,
      titleTextStyle: VigorTypography.headline.copyWith(
        color: isDark ? VigorColors.darkTextPrimary : VigorColors.lightTextPrimary,
      ),
      systemOverlayStyle: isDark ? SystemUiOverlayStyle.light : SystemUiOverlayStyle.dark,
    );
  }

  static ListTileThemeData _listTileTheme() {
    return const ListTileThemeData(
      contentPadding: VigorSpacing.listTilePadding,
      shape: RoundedRectangleBorder(borderRadius: VigorRadius.radiusSm),
    );
  }

  static NavigationBarThemeData _navigationBarTheme(Brightness brightness) {
    final bgColor = brightness == Brightness.dark
        ? VigorColors.darkSurface
        : VigorColors.lightSurface;
    final accentColor = brightness == Brightness.dark
        ? VigorColors.indigoLight
        : VigorColors.indigo;
    final indicatorColor = accentColor.withValues(alpha: 0.15);

    return NavigationBarThemeData(
      height: 70,
      backgroundColor: bgColor,
      indicatorColor: indicatorColor,
      labelTextStyle: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) {
          return VigorTypography.label.copyWith(
            color: accentColor,
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
          return IconThemeData(color: accentColor, size: 24);
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

  static SwitchThemeData _switchTheme() {
    return SwitchThemeData(
      thumbColor: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) return VigorColors.washi;
        return VigorColors.stone;
      }),
      trackColor: WidgetStateProperty.resolveWith((states) {
        if (states.contains(WidgetState.selected)) return VigorColors.persimmon;
        return VigorColors.darkBorder;
      }),
      trackOutlineColor: WidgetStateProperty.all(Colors.transparent),
    );
  }

  static ChipThemeData _chipTheme(Brightness brightness) {
    return ChipThemeData(
      backgroundColor: brightness == Brightness.dark
          ? VigorColors.darkSurface
          : VigorColors.lightSurface,
      selectedColor: VigorColors.indigo.withValues(alpha: 0.15),
      labelStyle: VigorTypography.label,
      shape: const RoundedRectangleBorder(borderRadius: VigorRadius.chip),
      side: BorderSide(
        color: brightness == Brightness.dark
            ? VigorColors.darkBorder
            : VigorColors.lightBorder,
      ),
    );
  }

  static ProgressIndicatorThemeData _progressIndicatorTheme() {
    return const ProgressIndicatorThemeData(
      color: VigorColors.indigo,
      linearTrackColor: VigorColors.darkBorder,
      circularTrackColor: VigorColors.darkBorder,
    );
  }
}

/// Extension for quick access to Vigor colors from context
extension VigorThemeExtension on BuildContext {
  bool get isDarkMode => Theme.of(this).brightness == Brightness.dark;

  // Brand colors
  Color get vigorPersimmon => VigorColors.persimmon;
  Color get vigorGold => VigorColors.gold;
  Color get vigorIndigo => VigorColors.indigo;
  Color get vigorCrimson => VigorColors.crimson;

  // Semantic colors
  Color get vigorSuccess => VigorColors.success;
  Color get vigorWarning => VigorColors.warning;
  Color get vigorError => VigorColors.error;

  // Surface colors
  Color get background => VigorColors.background(this);
  Color get surface => VigorColors.surface(this);
  Color get surfaceElevated => VigorColors.surfaceElevated(this);
  Color get borderColor => VigorColors.border(this);

  // Text colors
  Color get textPrimary => VigorColors.textPrimary(this);
  Color get textSecondary => VigorColors.textSecondary(this);
  Color get textMuted => VigorColors.textMuted(this);
}

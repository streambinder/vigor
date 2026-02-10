import 'package:flutter/material.dart';

/// Vigor design system tokens
/// Based on IDENTITY.md: Japanese Aesthetics × Data Science
/// Philosophy: Kanso (simplicity), Fukinsei (asymmetry), Kintsugi (golden repair), Seijaku (stillness)

// =============================================================================
// Colors — "Elements & Heat" palette from IDENTITY.md section 3.2
// =============================================================================

class VigorColors {
  // Structure colors
  static const Color sumi = Color(0xFF0F1115); // deep charcoal, OLED optimized
  static const Color washi = Color(0xFFF2F0EB); // off-white, natural texture
  static const Color stone = Color(0xFF888C94); // secondary text, inactive

  // Status/Data colors — thermal dynamics
  static const Color gold = Color(0xFFD4AF37); // kintsugi: achievements, high proficiency
  static const Color indigo = Color(0xFF2B4C5D); // cool: fully recovered, low intensity
  static const Color indigoLight = Color(0xFF5A9ABF); // lighter variant for dark mode contrast
  static const Color persimmon = Color(0xFFE65D38); // warm: active muscle, building heat
  static const Color crimson = Color(0xFF8F1D21); // hot: overload, high stress, need recovery

  // Semantic aliases
  static const Color success = Color(0xFF22C55E);
  static const Color warning = persimmon;
  static const Color error = crimson;
  static const Color info = indigo;

  // Dark mode surfaces (Sumi-based)
  static const Color darkBackground = sumi;
  static const Color darkSurface = Color(0xFF1A1A1E);
  static const Color darkSurfaceElevated = Color(0xFF262629);
  static const Color darkBorder = Color(0xFF333338);

  // Light mode surfaces (Washi-based)
  static const Color lightBackground = washi;
  static const Color lightSurface = Color(0xFFFFFFFF);
  static const Color lightSurfaceElevated = Color(0xFFFFFFFF);
  static const Color lightBorder = Color(0xFFE5E5E0);

  // Text colors — dark mode
  static const Color darkTextPrimary = washi;
  static const Color darkTextSecondary = stone;
  static const Color darkTextMuted = Color(0xFF6B6B73);

  // Text colors — light mode
  static const Color lightTextPrimary = sumi;
  static const Color lightTextSecondary = Color(0xFF525256);
  static const Color lightTextMuted = stone;

  // Glass effect parameters (iOS-style blur)
  static const double glassBlur = 24.0;
  static const double glassOpacityDark = 0.85;
  static const double glassOpacityLight = 0.75;
  static const double glassBorderOpacity = 0.1;

  // Theme-aware getters
  static Color background(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? darkBackground : lightBackground;

  static Color surface(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? darkSurface : lightSurface;

  static Color surfaceElevated(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? darkSurfaceElevated : lightSurfaceElevated;

  static Color border(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? darkBorder : lightBorder;

  static Color textPrimary(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? darkTextPrimary : lightTextPrimary;

  static Color textSecondary(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? darkTextSecondary : lightTextSecondary;

  static Color textMuted(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? darkTextMuted : lightTextMuted;

  static Color indigoAdaptive(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? indigoLight : indigo;
}

// =============================================================================
// Data Visualization — Heat levels for muscle/proficiency states
// =============================================================================

class VigorHeat {
  static const Color cool = VigorColors.indigo; // recovered, low intensity
  static const Color warm = VigorColors.persimmon; // active, building heat
  static const Color hot = VigorColors.crimson; // overload, needs recovery
  static const Color mastery = VigorColors.gold; // high proficiency, achievements

  /// returns heat color based on 0.0-1.0 intensity value
  static Color fromIntensity(double intensity) {
    if (intensity <= 0.30) return cool;
    if (intensity <= 0.80) return warm;
    return hot;
  }
}

// =============================================================================
// Spacing
// =============================================================================

class VigorSpacing {
  static const double xs = 4.0;
  static const double sm = 8.0;
  static const double md = 16.0;
  static const double lg = 24.0;
  static const double xl = 32.0;
  static const double xxl = 48.0;

  static const EdgeInsets paddingXs = EdgeInsets.all(xs);
  static const EdgeInsets paddingSm = EdgeInsets.all(sm);
  static const EdgeInsets paddingMd = EdgeInsets.all(md);
  static const EdgeInsets paddingLg = EdgeInsets.all(lg);
  static const EdgeInsets paddingXl = EdgeInsets.all(xl);
  static const EdgeInsets paddingXxl = EdgeInsets.all(xxl);

  static const EdgeInsets paddingHorizontalSm = EdgeInsets.symmetric(horizontal: sm);
  static const EdgeInsets paddingHorizontalMd = EdgeInsets.symmetric(horizontal: md);
  static const EdgeInsets paddingHorizontalLg = EdgeInsets.symmetric(horizontal: lg);

  static const EdgeInsets paddingVerticalSm = EdgeInsets.symmetric(vertical: sm);
  static const EdgeInsets paddingVerticalMd = EdgeInsets.symmetric(vertical: md);
  static const EdgeInsets paddingVerticalLg = EdgeInsets.symmetric(vertical: lg);

  static const EdgeInsets buttonPadding = EdgeInsets.symmetric(horizontal: lg, vertical: 14.0);
  static const EdgeInsets inputPadding = EdgeInsets.symmetric(horizontal: md, vertical: md);
  static const EdgeInsets listTilePadding = EdgeInsets.symmetric(horizontal: md, vertical: sm);
}

// =============================================================================
// Border Radius
// =============================================================================

class VigorRadius {
  static const double xs = 4.0;
  static const double sm = 8.0;
  static const double md = 12.0;
  static const double lg = 16.0;
  static const double xl = 24.0;
  static const double full = 9999.0;

  static const BorderRadius radiusXs = BorderRadius.all(Radius.circular(xs));
  static const BorderRadius radiusSm = BorderRadius.all(Radius.circular(sm));
  static const BorderRadius radiusMd = BorderRadius.all(Radius.circular(md));
  static const BorderRadius radiusLg = BorderRadius.all(Radius.circular(lg));
  static const BorderRadius radiusXl = BorderRadius.all(Radius.circular(xl));
  static const BorderRadius radiusFull = BorderRadius.all(Radius.circular(full));

  static const BorderRadius button = radiusSm;
  static const BorderRadius input = radiusSm;
  static const BorderRadius card = radiusMd;
  static const BorderRadius modal = radiusLg;
  static const BorderRadius navigationBar = radiusXl;
  static const BorderRadius chip = radiusXs;
  static const BorderRadius avatar = radiusFull;

  static const BorderRadius bottomSheet = BorderRadius.vertical(top: Radius.circular(lg));
}

// =============================================================================
// Typography — per IDENTITY.md section 3.3
// =============================================================================

class VigorTypography {
  // font families
  static const String fontFamilyDisplay = 'SpaceGrotesk'; // headlines, branding
  static const String fontFamilyBody = 'Inter'; // content, instructions
  static const String fontFamilyMono = 'JetBrainsMono'; // data, timers, scores

  // Display — Space Grotesk, geometric with quirks
  static const TextStyle display = TextStyle(
    fontFamily: fontFamilyDisplay,
    fontSize: 32,
    fontWeight: FontWeight.w700,
    letterSpacing: -0.5,
    height: 1.2,
  );

  // Title — Space Grotesk
  static const TextStyle title = TextStyle(
    fontFamily: fontFamilyDisplay,
    fontSize: 24,
    fontWeight: FontWeight.w600,
    letterSpacing: -0.3,
    height: 1.3,
  );

  // Headline — Space Grotesk
  static const TextStyle headline = TextStyle(
    fontFamily: fontFamilyDisplay,
    fontSize: 20,
    fontWeight: FontWeight.w600,
    letterSpacing: -0.2,
    height: 1.4,
  );

  // Body Large — Inter
  static const TextStyle bodyLarge = TextStyle(
    fontFamily: fontFamilyBody,
    fontSize: 17,
    fontWeight: FontWeight.w400,
    letterSpacing: -0.4,
    height: 1.5,
  );

  // Body — Inter
  static const TextStyle body = TextStyle(
    fontFamily: fontFamilyBody,
    fontSize: 15,
    fontWeight: FontWeight.w400,
    letterSpacing: -0.2,
    height: 1.5,
  );

  // Label — Inter
  static const TextStyle label = TextStyle(
    fontFamily: fontFamilyBody,
    fontSize: 13,
    fontWeight: FontWeight.w500,
    letterSpacing: 0,
    height: 1.4,
  );

  // Caption — Inter
  static const TextStyle caption = TextStyle(
    fontFamily: fontFamilyBody,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    letterSpacing: 0.1,
    height: 1.4,
  );

  // Data — JetBrains Mono (calibration scores, rep counts, timers)
  static const TextStyle data = TextStyle(
    fontFamily: fontFamilyMono,
    fontSize: 14,
    fontWeight: FontWeight.w400,
    letterSpacing: 0,
    height: 1.4,
  );

  // Data Large — JetBrains Mono (prominent numbers)
  static const TextStyle dataLarge = TextStyle(
    fontFamily: fontFamilyMono,
    fontSize: 24,
    fontWeight: FontWeight.w500,
    letterSpacing: -0.5,
    height: 1.2,
  );

  // Data Display — JetBrains Mono (timers, big stats)
  static const TextStyle dataDisplay = TextStyle(
    fontFamily: fontFamilyMono,
    fontSize: 48,
    fontWeight: FontWeight.w600,
    letterSpacing: -1.0,
    height: 1.1,
  );

  static TextStyle withColor(TextStyle style, Color color) => style.copyWith(color: color);

  static TextStyle displayColored(BuildContext context) =>
      display.copyWith(color: VigorColors.textPrimary(context));

  static TextStyle titleColored(BuildContext context) =>
      title.copyWith(color: VigorColors.textPrimary(context));

  static TextStyle headlineColored(BuildContext context) =>
      headline.copyWith(color: VigorColors.textPrimary(context));

  static TextStyle bodyLargeColored(BuildContext context) =>
      bodyLarge.copyWith(color: VigorColors.textPrimary(context));

  static TextStyle bodyColored(BuildContext context) =>
      body.copyWith(color: VigorColors.textPrimary(context));

  static TextStyle labelColored(BuildContext context) =>
      label.copyWith(color: VigorColors.textPrimary(context));

  static TextStyle captionColored(BuildContext context) =>
      caption.copyWith(color: VigorColors.textSecondary(context));

  static TextStyle dataColored(BuildContext context) =>
      data.copyWith(color: VigorColors.textPrimary(context));
}

// =============================================================================
// Elevation & Shadows
// =============================================================================

class VigorShadows {
  static List<BoxShadow> elevationLevel1Dark = [
    BoxShadow(color: Colors.black.withValues(alpha: 0.3), blurRadius: 8, offset: const Offset(0, 2)),
  ];

  static List<BoxShadow> elevationLevel2Dark = [
    BoxShadow(color: Colors.black.withValues(alpha: 0.4), blurRadius: 16, offset: const Offset(0, 4)),
  ];

  static List<BoxShadow> elevationLevel3Dark = [
    BoxShadow(color: Colors.black.withValues(alpha: 0.5), blurRadius: 32, offset: const Offset(0, 8)),
  ];

  static List<BoxShadow> elevationLevel1Light = [
    BoxShadow(color: Colors.black.withValues(alpha: 0.08), blurRadius: 8, offset: const Offset(0, 2)),
    BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 4, offset: const Offset(0, 1)),
  ];

  static List<BoxShadow> elevationLevel2Light = [
    BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 16, offset: const Offset(0, 4)),
    BoxShadow(color: Colors.black.withValues(alpha: 0.05), blurRadius: 8, offset: const Offset(0, 2)),
  ];

  static List<BoxShadow> elevationLevel3Light = [
    BoxShadow(color: Colors.black.withValues(alpha: 0.12), blurRadius: 32, offset: const Offset(0, 8)),
    BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 16, offset: const Offset(0, 4)),
  ];

  // Glow shadows — using brand colors
  static List<BoxShadow> persimmonGlow = [
    BoxShadow(color: VigorColors.persimmon.withValues(alpha: 0.3), blurRadius: 20, offset: const Offset(0, 4)),
  ];

  static List<BoxShadow> goldGlow = [
    BoxShadow(color: VigorColors.gold.withValues(alpha: 0.3), blurRadius: 20, offset: const Offset(0, 4)),
  ];

  static List<BoxShadow> indigoGlow = [
    BoxShadow(color: VigorColors.indigo.withValues(alpha: 0.3), blurRadius: 20, offset: const Offset(0, 4)),
  ];

  static List<BoxShadow> elevation1(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? elevationLevel1Dark : elevationLevel1Light;

  static List<BoxShadow> elevation2(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? elevationLevel2Dark : elevationLevel2Light;

  static List<BoxShadow> elevation3(BuildContext context) =>
      Theme.of(context).brightness == Brightness.dark ? elevationLevel3Dark : elevationLevel3Light;
}

// =============================================================================
// Animation — per IDENTITY.md section 4.3 "Seijaku" (stillness in motion)
// =============================================================================

class VigorAnimation {
  static const Duration fast = Duration(milliseconds: 150);
  static const Duration medium = Duration(milliseconds: 300);
  static const Duration slow = Duration(milliseconds: 500);

  // calm, non-erratic curves per "Seijaku" principle
  static const Curve defaultCurve = Curves.easeOut;
  static const Curve entranceCurve = Curves.easeOutCubic;
  static const Curve exitCurve = Curves.easeInCubic;
  // no bounce curve — violates Seijaku
}

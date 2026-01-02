import 'package:flutter/material.dart';

/// Vigor design system tokens
/// Single source of truth for all design values

// =============================================================================
// Colors
// =============================================================================

class VigorColors {
  // Primary colors
  static const Color orange = Color(0xFFFF6B35);
  static const Color electricBlue = Color(0xFF0EA5E9);
  static const Color deepCharcoal = Color(0xFF0F0F0F);
  static const Color softBlack = Color(0xFF1A1A1A);

  // Accent colors
  static const Color success = Color(0xFF22C55E);
  static const Color warning = Color(0xFFF59E0B);
  static const Color error = Color(0xFFEF4444);
  static const Color info = Color(0xFF3B82F6);

  // Dark mode surfaces
  static const Color darkBackground = deepCharcoal;
  static const Color darkSurface = softBlack;
  static const Color darkSurfaceElevated = Color(0xFF262626);
  static const Color darkBorder = Color(0xFF333333);

  // Light mode surfaces
  static const Color lightBackground = Color(0xFFFAFAFA);
  static const Color lightSurface = Color(0xFFFFFFFF);
  static const Color lightSurfaceElevated = Color(0xFFFFFFFF);
  static const Color lightBorder = Color(0xFFE5E5E5);

  // Text colors
  static const Color darkTextPrimary = Color(0xFFFFFFFF);
  static const Color darkTextSecondary = Color(0xFFA3A3A3);
  static const Color darkTextMuted = Color(0xFF737373);

  static const Color lightTextPrimary = Color(0xFF0F0F0F);
  static const Color lightTextSecondary = Color(0xFF525252);
  static const Color lightTextMuted = Color(0xFF737373);

  // Glass effect parameters (iOS)
  static const double glassBlur = 24.0;
  static const double glassOpacityDark = 0.85;
  static const double glassOpacityLight = 0.75;
  static const double glassBorderOpacity = 0.1;

  // Helper getters for theme-aware colors
  static Color background(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? darkBackground
        : lightBackground;
  }

  static Color surface(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? darkSurface
        : lightSurface;
  }

  static Color surfaceElevated(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? darkSurfaceElevated
        : lightSurfaceElevated;
  }

  static Color border(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? darkBorder
        : lightBorder;
  }

  static Color textPrimary(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? darkTextPrimary
        : lightTextPrimary;
  }

  static Color textSecondary(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? darkTextSecondary
        : lightTextSecondary;
  }

  static Color textMuted(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? darkTextMuted
        : lightTextMuted;
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

  // Common padding presets
  static const EdgeInsets paddingXs = EdgeInsets.all(xs);
  static const EdgeInsets paddingSm = EdgeInsets.all(sm);
  static const EdgeInsets paddingMd = EdgeInsets.all(md);
  static const EdgeInsets paddingLg = EdgeInsets.all(lg);
  static const EdgeInsets paddingXl = EdgeInsets.all(xl);
  static const EdgeInsets paddingXxl = EdgeInsets.all(xxl);

  // Horizontal padding
  static const EdgeInsets paddingHorizontalSm = EdgeInsets.symmetric(horizontal: sm);
  static const EdgeInsets paddingHorizontalMd = EdgeInsets.symmetric(horizontal: md);
  static const EdgeInsets paddingHorizontalLg = EdgeInsets.symmetric(horizontal: lg);

  // Vertical padding
  static const EdgeInsets paddingVerticalSm = EdgeInsets.symmetric(vertical: sm);
  static const EdgeInsets paddingVerticalMd = EdgeInsets.symmetric(vertical: md);
  static const EdgeInsets paddingVerticalLg = EdgeInsets.symmetric(vertical: lg);

  // Button padding
  static const EdgeInsets buttonPadding = EdgeInsets.symmetric(
    horizontal: lg,
    vertical: 14.0,
  );

  // Input padding
  static const EdgeInsets inputPadding = EdgeInsets.symmetric(
    horizontal: md,
    vertical: md,
  );

  // List tile padding
  static const EdgeInsets listTilePadding = EdgeInsets.symmetric(
    horizontal: md,
    vertical: sm,
  );
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

  // BorderRadius presets
  static const BorderRadius radiusXs = BorderRadius.all(Radius.circular(xs));
  static const BorderRadius radiusSm = BorderRadius.all(Radius.circular(sm));
  static const BorderRadius radiusMd = BorderRadius.all(Radius.circular(md));
  static const BorderRadius radiusLg = BorderRadius.all(Radius.circular(lg));
  static const BorderRadius radiusXl = BorderRadius.all(Radius.circular(xl));
  static const BorderRadius radiusFull = BorderRadius.all(Radius.circular(full));

  // Component-specific radius
  static const BorderRadius button = radiusSm;
  static const BorderRadius input = radiusSm;
  static const BorderRadius card = radiusMd;
  static const BorderRadius modal = radiusLg;
  static const BorderRadius navigationBar = radiusXl;
  static const BorderRadius chip = radiusXs;
  static const BorderRadius avatar = radiusFull;

  // Bottom sheet (top only)
  static const BorderRadius bottomSheet = BorderRadius.vertical(
    top: Radius.circular(lg),
  );
}

// =============================================================================
// Typography
// =============================================================================

class VigorTypography {
  static const String fontFamily = 'Barlow';
  static const String fontFamilyCondensed = 'Barlow Condensed';

  // Display - 32px, Bold (-0.5px)
  static const TextStyle display = TextStyle(
    fontFamily: fontFamilyCondensed,
    fontSize: 32,
    fontWeight: FontWeight.w700,
    letterSpacing: -0.5,
    height: 1.2,
  );

  // Title - 24px, SemiBold (-0.3px)
  static const TextStyle title = TextStyle(
    fontFamily: fontFamily,
    fontSize: 24,
    fontWeight: FontWeight.w600,
    letterSpacing: -0.3,
    height: 1.3,
  );

  // Headline - 20px, SemiBold (-0.2px)
  static const TextStyle headline = TextStyle(
    fontFamily: fontFamily,
    fontSize: 20,
    fontWeight: FontWeight.w600,
    letterSpacing: -0.2,
    height: 1.4,
  );

  // Body Large - 17px, Regular (-0.4px)
  static const TextStyle bodyLarge = TextStyle(
    fontFamily: fontFamily,
    fontSize: 17,
    fontWeight: FontWeight.w400,
    letterSpacing: -0.4,
    height: 1.5,
  );

  // Body - 15px, Regular (-0.2px)
  static const TextStyle body = TextStyle(
    fontFamily: fontFamily,
    fontSize: 15,
    fontWeight: FontWeight.w400,
    letterSpacing: -0.2,
    height: 1.5,
  );

  // Label - 13px, Medium (0px)
  static const TextStyle label = TextStyle(
    fontFamily: fontFamily,
    fontSize: 13,
    fontWeight: FontWeight.w500,
    letterSpacing: 0,
    height: 1.4,
  );

  // Caption - 12px, Regular (0.1px)
  static const TextStyle caption = TextStyle(
    fontFamily: fontFamily,
    fontSize: 12,
    fontWeight: FontWeight.w400,
    letterSpacing: 0.1,
    height: 1.4,
  );

  // Apply color to text style
  static TextStyle withColor(TextStyle style, Color color) {
    return style.copyWith(color: color);
  }

  // Get theme-aware text styles
  static TextStyle displayColored(BuildContext context) {
    return display.copyWith(color: VigorColors.textPrimary(context));
  }

  static TextStyle titleColored(BuildContext context) {
    return title.copyWith(color: VigorColors.textPrimary(context));
  }

  static TextStyle headlineColored(BuildContext context) {
    return headline.copyWith(color: VigorColors.textPrimary(context));
  }

  static TextStyle bodyLargeColored(BuildContext context) {
    return bodyLarge.copyWith(color: VigorColors.textPrimary(context));
  }

  static TextStyle bodyColored(BuildContext context) {
    return body.copyWith(color: VigorColors.textPrimary(context));
  }

  static TextStyle labelColored(BuildContext context) {
    return label.copyWith(color: VigorColors.textPrimary(context));
  }

  static TextStyle captionColored(BuildContext context) {
    return caption.copyWith(color: VigorColors.textSecondary(context));
  }
}

// =============================================================================
// Elevation & Shadows
// =============================================================================

class VigorShadows {
  // Dark mode shadows
  static List<BoxShadow> elevationLevel1Dark = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.3),
      blurRadius: 8,
      offset: const Offset(0, 2),
    ),
  ];

  static List<BoxShadow> elevationLevel2Dark = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.4),
      blurRadius: 16,
      offset: const Offset(0, 4),
    ),
  ];

  static List<BoxShadow> elevationLevel3Dark = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.5),
      blurRadius: 32,
      offset: const Offset(0, 8),
    ),
  ];

  // Light mode shadows (softer)
  static List<BoxShadow> elevationLevel1Light = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.08),
      blurRadius: 8,
      offset: const Offset(0, 2),
    ),
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.04),
      blurRadius: 4,
      offset: const Offset(0, 1),
    ),
  ];

  static List<BoxShadow> elevationLevel2Light = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.1),
      blurRadius: 16,
      offset: const Offset(0, 4),
    ),
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.05),
      blurRadius: 8,
      offset: const Offset(0, 2),
    ),
  ];

  static List<BoxShadow> elevationLevel3Light = [
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.12),
      blurRadius: 32,
      offset: const Offset(0, 8),
    ),
    BoxShadow(
      color: Colors.black.withValues(alpha: 0.06),
      blurRadius: 16,
      offset: const Offset(0, 4),
    ),
  ];

  // Glow shadows (for CTAs)
  static List<BoxShadow> orangeGlow = [
    BoxShadow(
      color: VigorColors.orange.withValues(alpha: 0.3),
      blurRadius: 20,
      offset: const Offset(0, 4),
    ),
  ];

  static List<BoxShadow> blueGlow = [
    BoxShadow(
      color: VigorColors.electricBlue.withValues(alpha: 0.3),
      blurRadius: 20,
      offset: const Offset(0, 4),
    ),
  ];

  // Get theme-aware shadows
  static List<BoxShadow> elevation1(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? elevationLevel1Dark
        : elevationLevel1Light;
  }

  static List<BoxShadow> elevation2(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? elevationLevel2Dark
        : elevationLevel2Light;
  }

  static List<BoxShadow> elevation3(BuildContext context) {
    return Theme.of(context).brightness == Brightness.dark
        ? elevationLevel3Dark
        : elevationLevel3Light;
  }
}

// =============================================================================
// Animation
// =============================================================================

class VigorAnimation {
  static const Duration fast = Duration(milliseconds: 150);
  static const Duration medium = Duration(milliseconds: 300);
  static const Duration slow = Duration(milliseconds: 500);

  static const Curve defaultCurve = Curves.easeOut;
  static const Curve entranceCurve = Curves.easeOutCubic;
  static const Curve exitCurve = Curves.easeInCubic;
  static const Curve bounceCurve = Curves.elasticOut;
}
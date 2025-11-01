// Stub for web_only.dart to allow compilation on non-web platforms
// These are never actually called on mobile since we check kIsWeb

import 'package:flutter/widgets.dart';

// Stub function that throws if somehow called on non-web
Widget renderButton({GSIButtonConfiguration? configuration}) {
  throw UnsupportedError('renderButton is only available on web');
}

// Stub classes for button configuration
class GSIButtonConfiguration {
  const GSIButtonConfiguration({
    this.type,
    this.theme,
    this.size,
    this.text,
    this.shape,
    this.logoAlignment,
  });

  final GSIButtonType? type;
  final GSIButtonTheme? theme;
  final GSIButtonSize? size;
  final GSIButtonText? text;
  final GSIButtonShape? shape;
  final GSIButtonLogoAlignment? logoAlignment;
}

enum GSIButtonType { standard, icon }

enum GSIButtonTheme { outline, filledBlue, filledBlack }

enum GSIButtonSize { large, medium, small }

enum GSIButtonText { signinWith, signupWith, continueWith, signin }

enum GSIButtonShape { rectangular, pill }

enum GSIButtonLogoAlignment { left, center }

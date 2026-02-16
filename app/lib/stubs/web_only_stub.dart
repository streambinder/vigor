// stub for web_only.dart — works on all platforms.
// on web, delegates to the registered GoogleSignInPlugin via dynamic call
// since conditional imports (dart.library.js_interop) don't resolve on web DDC.
// on mobile, the kIsWeb guard in google_auth_screen prevents renderButton from being called.

import 'package:flutter/widgets.dart';
import 'package:google_sign_in_platform_interface/google_sign_in_platform_interface.dart';

Widget renderButton({GSIButtonConfiguration? configuration}) {
  // GoogleSignInPlugin is already registered as the platform instance on web.
  // call with null config (uses plugin defaults) since our stub types
  // aren't the same as google_sign_in_web's types.
  final dynamic plugin = GoogleSignInPlatform.instance;
  return plugin.renderButton() as Widget;
}

// stub types so call sites compile on all platforms
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

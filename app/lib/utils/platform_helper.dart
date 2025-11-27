import 'dart:io';
import 'package:flutter/foundation.dart';

/// Platform detection utility for adaptive design
class PlatformHelper {
  /// Check if running on iOS (native only, not web)
  static bool get isIOS {
    if (kIsWeb) return false;
    return Platform.isIOS;
  }

  /// Check if running on Android (native only, not web)
  static bool get isAndroid {
    if (kIsWeb) return false;
    return Platform.isAndroid;
  }

  /// Check if running on web
  static bool get isWeb => kIsWeb;

  /// Check if running on macOS
  static bool get isMacOS {
    if (kIsWeb) return false;
    return Platform.isMacOS;
  }

  /// Check if should use Liquid Glass design (iOS only)
  static bool get useLiquidGlass => isIOS;

  /// Check if should use Material You design (Android, Web, and all other platforms)
  static bool get useMaterialYou => !isIOS;
}

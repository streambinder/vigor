import 'dart:js_util' as js_util;
import 'dart:html' as html;

/// API configuration for Vigor server
class ApiConfig {
  // Private static variable to cache the base URL
  static String? _cachedBaseUrl;

  /// API URL with multi-platform support:
  /// - Web: Reads from window.ENV.API_URL (injected at Docker runtime via config.js)
  /// - Mobile: Uses --dart-define=API_URL at build time
  /// - Default: http://localhost:8000 for local development
  ///
  /// Web usage (Docker): docker run -e API_URL=https://api.example.com
  /// Mobile usage: flutter run --dart-define=API_URL=https://backend.domain.my
  /// Android emulator: Use http://10.0.2.2:8000
  /// Physical device: Use http://<your-local-ip>:8000
  static String get baseUrl {
    if (_cachedBaseUrl != null) {
      return _cachedBaseUrl!;
    }

    String? source;

    // Try to get from runtime config (web/Docker)
    try {
      final env = js_util.getProperty(html.window, 'ENV');
      if (env != null) {
        final apiUrl = js_util.getProperty(env, 'API_URL');
        if (apiUrl != null && apiUrl.toString().isNotEmpty) {
          _cachedBaseUrl = apiUrl.toString();
          source = 'window.ENV';
          print('[ApiConfig] Using API URL from $source: $_cachedBaseUrl');
          return _cachedBaseUrl!;
        }
      }
    } catch (e) {
      // Not on web platform or config.js not loaded, fall through
    }

    // Try compile-time environment variable (mobile)
    const compileTimeUrl = String.fromEnvironment('API_URL');
    if (compileTimeUrl.isNotEmpty) {
      _cachedBaseUrl = compileTimeUrl;
      source = '--dart-define';
      print('[ApiConfig] Using API URL from $source: $_cachedBaseUrl');
      return _cachedBaseUrl!;
    }

    // Default for local development
    _cachedBaseUrl = 'http://localhost:8000';
    source = 'default';
    print('[ApiConfig] Using API URL from $source: $_cachedBaseUrl');
    return _cachedBaseUrl!;
  }

  // Authentication endpoints
  static const String googleAuthEndpoint = '/auth/google';
  static const String refreshEndpoint = '/refresh';
  static const String logoutEndpoint = '/logout';

  // User endpoints
  static const String userEndpoint = '/user';
  static const String updateUserEndpoint = '/user/update';
  static const String unregisterEndpoint = '/unregister';

  // Token configuration
  static const Duration accessTokenExpiry = Duration(minutes: 15);
  static const Duration refreshTokenExpiry = Duration(days: 7);

  // HTTP headers
  static const String authorizationHeader = 'Authorization';
  static const String contentTypeHeader = 'Content-Type';
  static const String contentTypeJson = 'application/json';
}

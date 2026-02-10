import 'window_env_stub.dart'
    if (dart.library.js_interop) 'window_env_web.dart';

class ApiConfig {
  static String? _cachedBaseUrl;

  /// API URL with multi-platform support:
  /// - Web: window.ENV.API_URL (injected at Docker runtime via config.js)
  /// - Mobile: --dart-define=API_URL at build time
  /// - Default: http://localhost:8000
  static String get baseUrl {
    if (_cachedBaseUrl != null) return _cachedBaseUrl!;

    final windowUrl = getApiUrlFromWindow();
    const compileUrl = String.fromEnvironment('API_URL');

    _cachedBaseUrl = windowUrl ??
                     (compileUrl.isNotEmpty ? compileUrl : 'http://localhost:8000');

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

  // Avatar endpoints
  static const String avatarEndpoint = '/user/avatar';
  static String avatarUrl(String userId) => '$baseUrl/user/avatar/$userId';

  // Token configuration
  static const Duration accessTokenExpiry = Duration(hours: 2);
  static const Duration refreshTokenExpiry = Duration(days: 7);

  // HTTP headers
  static const String authorizationHeader = 'Authorization';
  static const String contentTypeHeader = 'Content-Type';
  static const String contentTypeJson = 'application/json';
}

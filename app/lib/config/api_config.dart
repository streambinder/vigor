/// API configuration for Vigor server
class ApiConfig {
  // API URL from environment variable or fallback to local development
  // Set via: flutter run --dart-define=API_URL=https://backend.domain.my
  // Note: For Android emulator, use http://10.0.2.2:8000
  // For physical device, use http://<your-local-ip>:8000
  static const String baseUrl = String.fromEnvironment(
    'API_URL',
    defaultValue: 'http://localhost:8000',
  );

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

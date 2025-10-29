/// API configuration for Vigor server
class ApiConfig {
  // Change this to your server URL
  // For iOS simulator: http://localhost:8080
  // For Android emulator: http://10.0.2.2:8080
  // For physical device: http://<your-local-ip>:8080
  static const String baseUrl = 'http://localhost:8080';
  // static const String baseUrl = 'https://backend.vigor.davidepucci.it';

  // Authentication endpoints
  static const String registerEndpoint = '/register';
  static const String loginEndpoint = '/login';
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

import 'package:flutter/foundation.dart';

/// API configuration for Vigor server
class ApiConfig {
  // Automatically switches between production and local URLs
  // - Production (release mode): https://backend.vigor.davidepucci.it
  // - Local (debug mode): http://localhost:8080
  // Note: For Android emulator, use http://10.0.2.2:8080
  // For physical device, use http://<your-local-ip>:8080
  static const String baseUrl = kReleaseMode
      ? 'https://backend.vigor.davidepucci.it'
      : 'http://localhost:8080';

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

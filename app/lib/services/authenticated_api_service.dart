import 'package:flutter_timezone/flutter_timezone.dart';

import '../config/api_config.dart';
import '../models/api_response.dart';
import '../models/auth_tokens.dart';
import 'api_service.dart';
import 'app_logger.dart';
import 'secure_storage_service.dart';

/// HTTP interceptor that handles authentication and automatic token refresh.
/// Wraps ApiService to add auth headers and retry on 401 with refreshed tokens.
class AuthenticatedApiService {
  final ApiService _apiService;
  final SecureStorageService _storageService;
  bool _isRefreshing = false;

  AuthenticatedApiService({
    ApiService? apiService,
    SecureStorageService? storageService,
  })  : _apiService = apiService ?? ApiService(),
        _storageService = storageService ?? SecureStorageService();

  Future<ApiResponse<Map<String, dynamic>>> get(String endpoint) async {
    return _authenticatedRequest(() => _doGet(endpoint));
  }

  Future<ApiResponse<Map<String, dynamic>>> post(
    String endpoint, {
    Map<String, dynamic>? body,
  }) async {
    return _authenticatedRequest(() => _doPost(endpoint, body: body));
  }

  Future<ApiResponse<Map<String, dynamic>>> put(
    String endpoint, {
    Map<String, dynamic>? body,
  }) async {
    return _authenticatedRequest(() => _doPut(endpoint, body: body));
  }

  Future<ApiResponse<Map<String, dynamic>>> delete(String endpoint) async {
    return _authenticatedRequest(() => _doDelete(endpoint));
  }

  Future<ApiResponse<Map<String, dynamic>>> postMultipart(
    String endpoint, {
    required List<int> bytes,
    required String fieldName,
    required String filename,
  }) async {
    return _authenticatedRequest(() => _doPostMultipart(endpoint, bytes: bytes, fieldName: fieldName, filename: filename));
  }

  /// Wraps a request with 401 interception and token refresh logic
  Future<ApiResponse<Map<String, dynamic>>> _authenticatedRequest(
    Future<ApiResponse<Map<String, dynamic>>> Function() request,
  ) async {
    final response = await request();

    if (response.statusCode != 401) {
      return response;
    }

    // got 401, try refreshing tokens
    final refreshed = await _refreshTokens();
    if (!refreshed) {
      return response; // refresh failed, return original 401
    }

    // retry with new token
    return request();
  }

  Future<bool> _refreshTokens() async {
    if (_isRefreshing) {
      return false; // prevent concurrent refresh attempts
    }

    _isRefreshing = true;
    try {
      final refreshToken = await _storageService.getRefreshToken();
      if (refreshToken == null) {
        AppLogger.warning('[AuthenticatedApi] no refresh token available');
        return false;
      }

      AppLogger.debug('[AuthenticatedApi] refreshing tokens');
      final response = await _apiService.post(
        ApiConfig.refreshEndpoint,
        body: {'refresh_token': refreshToken},
      );

      if (response.isSuccess && response.data != null) {
        try {
          final tokens = AuthTokens.fromJson(response.data!);
          await _storageService.saveTokens(
            accessToken: tokens.accessToken,
            refreshToken: tokens.refreshToken,
          );
          AppLogger.info('[AuthenticatedApi] tokens refreshed successfully');
          return true;
        } catch (e) {
          AppLogger.error('[AuthenticatedApi] failed to parse refreshed tokens', e);
          return false;
        }
      } else {
        AppLogger.error('[AuthenticatedApi] token refresh failed: ${response.error}');
        await _storageService.deleteTokens();
        return false;
      }
    } finally {
      _isRefreshing = false;
    }
  }

  Future<Map<String, String>?> _getAuthHeaders() async {
    final accessToken = await _storageService.getAccessToken();
    if (accessToken == null) {
      return null;
    }
    return {
      ApiConfig.authorizationHeader: 'Bearer $accessToken',
      'X-Timezone': await FlutterTimezone.getLocalTimezone(),
    };
  }

  Future<ApiResponse<Map<String, dynamic>>> _doGet(String endpoint) async {
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }
    return _apiService.get(endpoint, headers: headers);
  }

  Future<ApiResponse<Map<String, dynamic>>> _doPost(
    String endpoint, {
    Map<String, dynamic>? body,
  }) async {
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }
    return _apiService.post(endpoint, headers: headers, body: body);
  }

  Future<ApiResponse<Map<String, dynamic>>> _doPut(
    String endpoint, {
    Map<String, dynamic>? body,
  }) async {
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }
    return _apiService.put(endpoint, headers: headers, body: body);
  }

  Future<ApiResponse<Map<String, dynamic>>> _doDelete(String endpoint) async {
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }
    return _apiService.delete(endpoint, headers: headers);
  }

  Future<ApiResponse<Map<String, dynamic>>> _doPostMultipart(
    String endpoint, {
    required List<int> bytes,
    required String fieldName,
    required String filename,
  }) async {
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }
    return _apiService.postMultipart(endpoint, bytes: bytes, fieldName: fieldName, filename: filename, headers: headers);
  }
}

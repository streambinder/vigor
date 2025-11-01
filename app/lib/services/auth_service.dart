import 'package:logger/logger.dart';

import '../config/api_config.dart';
import '../models/api_response.dart';
import '../models/auth_tokens.dart';
import '../models/user.dart';
import 'api_service.dart';
import 'app_logger.dart';
import 'secure_storage_service.dart';

/// Authentication service for user authentication operations
class AuthService {
  final ApiService _apiService;
  final SecureStorageService _storageService;
  final Logger _log = AppLogger.getLogger('AuthService');

  AuthService({
    ApiService? apiService,
    SecureStorageService? storageService,
  })  : _apiService = apiService ?? ApiService(),
        _storageService = storageService ?? SecureStorageService();

  /// Login with Google ID token
  /// Returns tokens on success, stores them securely
  Future<ApiResponse<AuthTokens>> loginWithGoogle({
    required String idToken,
  }) async {
    _log.d('POST ${ApiConfig.googleAuthEndpoint} (token_len=${idToken.length})');

    final response = await _apiService.post(
      ApiConfig.googleAuthEndpoint,
      body: {
        'id_token': idToken,
      },
    );

    _log.d('Response: status=${response.statusCode} success=${response.isSuccess}');

    if (response.isSuccess && response.data != null) {
      try {
        final tokens = AuthTokens.fromJson(response.data!);
        _log.i('Tokens parsed: access=${tokens.accessToken.length}b refresh=${tokens.refreshToken.length}b');

        // Store tokens securely
        try {
          await _storageService.saveTokens(
            accessToken: tokens.accessToken,
            refreshToken: tokens.refreshToken,
          );
          _log.d('Tokens stored successfully');
        } catch (e) {
          _log.w('Token storage failed, continuing with in-memory tokens', e);
          // Don't fail the login just because storage failed
          // The tokens are still in the response and will be passed to getCurrentUser
        }

        return ApiResponse.success(tokens, response.statusCode);
      } catch (e) {
        _log.e('Failed to parse auth tokens', e);
        return ApiResponse.error('Failed to parse tokens', response.statusCode);
      }
    } else {
      _log.e('Google login failed: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Google login failed',
        response.statusCode,
      );
    }
  }

  /// Refresh access token using refresh token
  /// Returns new tokens on success, stores them securely
  Future<ApiResponse<AuthTokens>> refreshToken() async {
    final refreshToken = await _storageService.getRefreshToken();

    if (refreshToken == null) {
      return ApiResponse.error('No refresh token found', 401);
    }

    final response = await _apiService.post(
      ApiConfig.refreshEndpoint,
      body: {
        'refresh_token': refreshToken,
      },
    );

    if (response.isSuccess && response.data != null) {
      try {
        final tokens = AuthTokens.fromJson(response.data!);

        // Store new tokens securely
        await _storageService.saveTokens(
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
        );

        return ApiResponse.success(tokens, response.statusCode);
      } catch (e) {
        return ApiResponse.error('Failed to parse tokens', response.statusCode);
      }
    } else {
      // If refresh fails, clear stored tokens
      await _storageService.deleteTokens();
      return ApiResponse.error(
        response.error ?? 'Token refresh failed',
        response.statusCode,
      );
    }
  }

  /// Logout user
  /// Revokes refresh token on server and clears local storage
  Future<ApiResponse<String>> logout() async {
    final refreshToken = await _storageService.getRefreshToken();

    if (refreshToken != null) {
      // Try to revoke token on server
      await _apiService.post(
        ApiConfig.logoutEndpoint,
        body: {
          'refresh_token': refreshToken,
        },
      );
    }

    // Always clear local storage, even if server request fails
    await _storageService.deleteTokens();

    return ApiResponse.success('Logged out successfully', 200);
  }

  /// Get current user profile
  /// Requires valid access token
  /// If [useToken] is provided, it will be used instead of reading from storage
  /// [_refreshAttempts] tracks recursion depth to prevent infinite loops
  Future<ApiResponse<User>> getCurrentUser(
      {String? useToken, int refreshAttempts = 0}) async {
    _log.d('getCurrentUser (attempt=$refreshAttempts)');

    // Prevent infinite refresh loops
    if (refreshAttempts >= 3) {
      _log.e('Max refresh attempts reached, session expired');
      return ApiResponse.error('Session expired - max refresh attempts', 401);
    }

    String? accessToken = useToken;
    if (accessToken == null) {
      try {
        accessToken = await _storageService.getAccessToken();
        _log.d('Token retrieved from storage (len=${accessToken?.length ?? 0})');
      } catch (e) {
        _log.e('Failed to read token from storage', e);
        return ApiResponse.error('Failed to read authentication token', 500);
      }
    } else {
      _log.d('Using provided token (len=${accessToken.length})');
    }

    if (accessToken == null) {
      _log.w('No access token available');
      return ApiResponse.error('Not authenticated', 401);
    }

    _log.d('GET ${ApiConfig.userEndpoint}');

    final response = await _apiService.get(
      ApiConfig.userEndpoint,
      headers: {
        ApiConfig.authorizationHeader: 'Bearer $accessToken',
      },
    );

    _log.d('Response: status=${response.statusCode} success=${response.isSuccess}');

    if (response.isSuccess && response.data != null) {
      try {
        final user = User.fromJson(response.data!);
        _log.i('User retrieved: id=${user.id} email=${user.email}');
        return ApiResponse.success(user, response.statusCode);
      } catch (e) {
        _log.e('Failed to parse user data', e);
        return ApiResponse.error(
            'Failed to parse user data', response.statusCode);
      }
    } else if (response.statusCode == 401) {
      _log.w('Token expired (401), attempting refresh');
      // Token expired, try to refresh
      final refreshResponse = await refreshToken();
      if (refreshResponse.isSuccess) {
        _log.i('Token refresh successful, retrying getCurrentUser');
        // Retry with new token (don't pass useToken to force reading from storage)
        // Increment refresh attempts to prevent infinite loops
        return getCurrentUser(refreshAttempts: refreshAttempts + 1);
      } else {
        _log.e('Token refresh failed');
        return ApiResponse.error('Session expired', 401);
      }
    } else {
      _log.e('Get user failed: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to get user',
        response.statusCode,
      );
    }
  }

  /// Update user profile
  Future<ApiResponse<String>> updateProfile({
    DateTime? birthdate,
    String? language,
    double? height,
    double? weight,
    Map<String, dynamic>? data,
  }) async {
    final accessToken = await _storageService.getAccessToken();

    if (accessToken == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    // Build update body with only provided fields
    final Map<String, dynamic> body = {};

    if (birthdate != null) {
      // Format as DD/MM/YYYY as expected by server
      final day = birthdate.day.toString().padLeft(2, '0');
      final month = birthdate.month.toString().padLeft(2, '0');
      final year = birthdate.year.toString();
      body['birthdate'] = '$day/$month/$year';
    }
    if (language != null) body['language'] = language;
    if (height != null) body['height'] = height;
    if (weight != null) body['weight'] = weight;
    if (data != null) body['data'] = data;

    final response = await _apiService.post(
      ApiConfig.updateUserEndpoint,
      headers: {
        ApiConfig.authorizationHeader: 'Bearer $accessToken',
      },
      body: body,
    );

    if (response.isSuccess) {
      final message = response.data?['message'] as String? ?? 'Profile updated';
      return ApiResponse.success(message, response.statusCode);
    } else if (response.statusCode == 401) {
      // Token expired, try to refresh
      final refreshResponse = await refreshToken();
      if (refreshResponse.isSuccess) {
        // Retry with new token
        return updateProfile(
          birthdate: birthdate,
          language: language,
          height: height,
          weight: weight,
          data: data,
        );
      } else {
        return ApiResponse.error('Session expired', 401);
      }
    } else {
      return ApiResponse.error(
        response.error ?? 'Failed to update profile',
        response.statusCode,
      );
    }
  }

  /// Delete user account
  Future<ApiResponse<String>> deleteAccount() async {
    String? accessToken;

    try {
      accessToken = await _storageService.getAccessToken();
    } catch (e) {
      _log.e('Failed to read token for account deletion', e);
      // If we can't read the token, clear local storage and return error
      try {
        await _storageService.clearAll();
      } catch (clearError) {
        _log.w('Failed to clear storage after token read error', clearError);
      }
      return ApiResponse.error(
        'Storage error. Please log out and try again.',
        500,
      );
    }

    if (accessToken == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    _log.i('Deleting user account');
    final response = await _apiService.post(
      ApiConfig.unregisterEndpoint,
      headers: {
        ApiConfig.authorizationHeader: 'Bearer $accessToken',
      },
    );

    if (response.isSuccess) {
      // Clear local storage
      try {
        await _storageService.deleteTokens();
        _log.d('Tokens cleared after account deletion');
      } catch (e) {
        _log.w('Failed to clear tokens after deletion', e);
        // Still return success since server deletion succeeded
      }
      final message = response.data?['message'] as String? ?? 'Account deleted';
      return ApiResponse.success(message, response.statusCode);
    } else {
      _log.e('Account deletion failed: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to delete account',
        response.statusCode,
      );
    }
  }

  /// Check if user is authenticated (has valid tokens)
  Future<bool> isAuthenticated() async {
    return await _storageService.hasTokens();
  }

  /// Get stored access token
  Future<String?> getAccessToken() async {
    return await _storageService.getAccessToken();
  }

  /// Get stored refresh token
  Future<String?> getRefreshToken() async {
    return await _storageService.getRefreshToken();
  }

  /// Clear all tokens (used before login to prevent stale token issues)
  Future<void> clearTokens() async {
    await _storageService.deleteTokens();
  }
}

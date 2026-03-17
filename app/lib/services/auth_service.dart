import '../config/api_config.dart';
import '../models/api_response.dart';
import '../models/auth_tokens.dart';
import '../models/user.dart';
import 'api_service.dart';
import 'app_logger.dart';
import 'secure_storage_service.dart';

class AuthService {
  final ApiService _apiService;
  final SecureStorageService _storageService;

  AuthService({
    ApiService? apiService,
    SecureStorageService? storageService,
  })  : _apiService = apiService ?? ApiService(),
        _storageService = storageService ?? SecureStorageService() {
    // Ensure storage is initialized if using default instance
    if (storageService == null && !_storageService.isHealthy) {
      throw StateError(
        'SecureStorageService must be initialized before creating AuthService. '
        'Either pass an initialized instance or call storage.initialize() first.',
      );
    }
  }

  /// Login with Google ID token
  /// Returns tokens on success, stores them securely
  Future<ApiResponse<AuthTokens>> loginWithGoogle({
    required String idToken,
  }) async {
    AppLogger.debug('[AuthService] POST ${ApiConfig.googleAuthEndpoint} (token_len=${idToken.length})');

    final response = await _apiService.post(
      ApiConfig.googleAuthEndpoint,
      body: {
        'id_token': idToken,
      },
    );

    AppLogger.debug('[AuthService] Response: status=${response.statusCode} success=${response.isSuccess}');

    if (response.isSuccess && response.data != null) {
      try {
        final tokens = AuthTokens.fromJson(response.data!);
        AppLogger.info('[AuthService] Tokens parsed: access=${tokens.accessToken.length}b refresh=${tokens.refreshToken.length}b');

        // Store tokens securely
        try {
          await _storageService.saveTokens(
            accessToken: tokens.accessToken,
            refreshToken: tokens.refreshToken,
          );
          AppLogger.debug('[AuthService] Tokens stored successfully');
        } catch (e) {
          AppLogger.warning('[AuthService] Token storage failed, continuing with in-memory tokens', e);
          // Don't fail the login just because storage failed
          // The tokens are still in the response and will be passed to getCurrentUser
        }

        return ApiResponse.success(tokens, response.statusCode);
      } catch (e) {
        AppLogger.error('[AuthService] Failed to parse auth tokens', e);
        return ApiResponse.error('Failed to parse tokens', response.statusCode);
      }
    } else {
      AppLogger.error('[AuthService] Google login failed: ${response.error}');
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
    AppLogger.debug('[AuthService] getCurrentUser (attempt=$refreshAttempts)');

    // Prevent infinite refresh loops
    if (refreshAttempts >= 3) {
      AppLogger.error('[AuthService] Max refresh attempts reached, session expired');
      return ApiResponse.error('Session expired - max refresh attempts', 401);
    }

    String? accessToken = useToken;
    if (accessToken == null) {
      try {
        accessToken = await _storageService.getAccessToken();
        AppLogger.debug('[AuthService] Token retrieved from storage (len=${accessToken?.length ?? 0})');
      } catch (e) {
        AppLogger.error('[AuthService] Failed to read token from storage', e);
        return ApiResponse.error('Failed to read authentication token', 500);
      }
    } else {
      AppLogger.debug('[AuthService] Using provided token (len=${accessToken.length})');
    }

    if (accessToken == null) {
      AppLogger.warning('[AuthService] No access token available');
      return ApiResponse.error('Not authenticated', 401);
    }

    AppLogger.debug('[AuthService] GET ${ApiConfig.userEndpoint}');

    final response = await _apiService.get(
      ApiConfig.userEndpoint,
      headers: {
        ApiConfig.authorizationHeader: 'Bearer $accessToken',
      },
    );

    AppLogger.debug('[AuthService] Response: status=${response.statusCode} success=${response.isSuccess}');

    if (response.isSuccess && response.data != null) {
      try {
        final user = User.fromJson(response.data!);
        AppLogger.info('[AuthService] User retrieved: id=${user.id} email=${user.email}');
        return ApiResponse.success(user, response.statusCode);
      } catch (e) {
        AppLogger.error('[AuthService] Failed to parse user data', e);
        return ApiResponse.error(
            'Failed to parse user data', response.statusCode);
      }
    } else if (response.statusCode == 401) {
      AppLogger.warning('[AuthService] Token expired (401), attempting refresh');
      // Token expired, try to refresh
      final refreshResponse = await refreshToken();
      if (refreshResponse.isSuccess) {
        AppLogger.info('[AuthService] Token refresh successful, retrying getCurrentUser');
        // Retry with new token (don't pass useToken to force reading from storage)
        // Increment refresh attempts to prevent infinite loops
        return getCurrentUser(refreshAttempts: refreshAttempts + 1);
      } else {
        AppLogger.error('[AuthService] Token refresh failed');
        return ApiResponse.error('Session expired', 401);
      }
    } else {
      AppLogger.error('[AuthService] Get user failed: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to get user',
        response.statusCode,
      );
    }
  }

  /// Update user profile
  Future<ApiResponse<String>> updateProfile({
    String? firstName,
    String? lastName,
    DateTime? birthdate,
    String? gender,
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

    if (firstName != null) body['first_name'] = firstName;
    if (lastName != null) body['last_name'] = lastName;
    if (birthdate != null) body['birthdate'] = birthdate.toIso8601String().split('T')[0];
    if (gender != null) body['gender'] = gender;
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
          firstName: firstName,
          lastName: lastName,
          birthdate: birthdate,
          gender: gender,
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
      AppLogger.error('[AuthService] Failed to read token for account deletion', e);
      // If we can't read the token, clear local storage and return error
      try {
        await _storageService.clearAll();
      } catch (clearError) {
        AppLogger.warning('[AuthService] Failed to clear storage after token read error', clearError);
      }
      return ApiResponse.error(
        'Storage error. Please log out and try again.',
        500,
      );
    }

    if (accessToken == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    AppLogger.info('[AuthService] Deleting user account');
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
        AppLogger.debug('[AuthService] Tokens cleared after account deletion');
      } catch (e) {
        AppLogger.warning('[AuthService] Failed to clear tokens after deletion', e);
        // Still return success since server deletion succeeded
      }
      final message = response.data?['message'] as String? ?? 'Account deleted';
      return ApiResponse.success(message, response.statusCode);
    } else {
      AppLogger.error('[AuthService] Account deletion failed: ${response.error}');
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

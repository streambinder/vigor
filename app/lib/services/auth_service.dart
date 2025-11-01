import '../config/api_config.dart';
import '../models/api_response.dart';
import '../models/auth_tokens.dart';
import '../models/user.dart';
import 'api_service.dart';
import 'secure_storage_service.dart';

/// Authentication service for user authentication operations
class AuthService {
  final ApiService _apiService;
  final SecureStorageService _storageService;

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
    print(
        '[AuthService] Sending login request to ${ApiConfig.googleAuthEndpoint}');
    print('[AuthService] Token length: ${idToken.length}');

    final response = await _apiService.post(
      ApiConfig.googleAuthEndpoint,
      body: {
        'id_token': idToken,
      },
    );

    print('[AuthService] Response status: ${response.statusCode}');
    print('[AuthService] Response success: ${response.isSuccess}');
    print('[AuthService] Response data: ${response.data}');

    if (response.isSuccess && response.data != null) {
      try {
        print('[AuthService] Parsing tokens from response...');
        final tokens = AuthTokens.fromJson(response.data!);
        print(
            '[AuthService] Successfully parsed tokens - Access: ${tokens.accessToken.length} chars, Refresh: ${tokens.refreshToken.length} chars');

        // Store tokens securely
        print('[AuthService] Storing tokens securely...');
        try {
          await _storageService.saveTokens(
            accessToken: tokens.accessToken,
            refreshToken: tokens.refreshToken,
          );
          print('[AuthService] Tokens stored successfully');
        } catch (e) {
          print(
              '[AuthService] WARNING: Failed to store tokens (continuing anyway): $e');
          // Don't fail the login just because storage failed
          // The tokens are still in the response and will be passed to getCurrentUser
        }

        return ApiResponse.success(tokens, response.statusCode);
      } catch (e) {
        print('[AuthService] ERROR: Failed to parse tokens: $e');
        return ApiResponse.error('Failed to parse tokens', response.statusCode);
      }
    } else {
      print('[AuthService] ERROR: Login failed - ${response.error}');
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
    print(
        '[AuthService] Getting current user... (refresh attempts: $refreshAttempts)');

    // Prevent infinite refresh loops
    if (refreshAttempts >= 3) {
      print(
          '[AuthService] ERROR: Max refresh attempts reached, stopping recursion');
      return ApiResponse.error('Session expired - max refresh attempts', 401);
    }

    String? accessToken = useToken;
    if (accessToken == null) {
      print('[AuthService] No token provided, reading from storage...');
      try {
        accessToken = await _storageService.getAccessToken();
      } catch (e) {
        print('[AuthService] ERROR reading token from storage: $e');
        return ApiResponse.error('Failed to read authentication token', 500);
      }
    } else {
      print(
          '[AuthService] Using provided token, length: ${accessToken.length}');
    }

    if (accessToken == null) {
      print('[AuthService] ERROR: No access token found');
      return ApiResponse.error('Not authenticated', 401);
    }

    print('[AuthService] Access token ready, length: ${accessToken.length}');
    print('[AuthService] Fetching user from ${ApiConfig.userEndpoint}');

    final response = await _apiService.get(
      ApiConfig.userEndpoint,
      headers: {
        ApiConfig.authorizationHeader: 'Bearer $accessToken',
      },
    );

    print('[AuthService] Get user response status: ${response.statusCode}');
    print('[AuthService] Get user response success: ${response.isSuccess}');
    print('[AuthService] Get user response data: ${response.data}');

    if (response.isSuccess && response.data != null) {
      try {
        print('[AuthService] Parsing user data...');
        final user = User.fromJson(response.data!);
        print(
            '[AuthService] Successfully parsed user - ID: ${user.id}, Email: ${user.email}');
        return ApiResponse.success(user, response.statusCode);
      } catch (e) {
        print('[AuthService] ERROR: Failed to parse user data: $e');
        return ApiResponse.error(
            'Failed to parse user data', response.statusCode);
      }
    } else if (response.statusCode == 401) {
      print('[AuthService] Token expired (401), attempting refresh...');
      // Token expired, try to refresh
      final refreshResponse = await refreshToken();
      if (refreshResponse.isSuccess) {
        print(
            '[AuthService] Token refresh successful, retrying getCurrentUser...');
        // Retry with new token (don't pass useToken to force reading from storage)
        // Increment refresh attempts to prevent infinite loops
        return getCurrentUser(refreshAttempts: refreshAttempts + 1);
      } else {
        print('[AuthService] ERROR: Token refresh failed');
        return ApiResponse.error('Session expired', 401);
      }
    } else {
      print('[AuthService] ERROR: Failed to get user - ${response.error}');
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
      print('[AuthService] ERROR: Failed to read token during account deletion: $e');
      // If we can't read the token, clear local storage and return error
      try {
        await _storageService.clearAll();
      } catch (clearError) {
        print('[AuthService] WARNING: Failed to clear storage: $clearError');
      }
      return ApiResponse.error(
        'Storage error. Please log out and try again.',
        500,
      );
    }

    if (accessToken == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

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
      } catch (e) {
        print('[AuthService] WARNING: Failed to clear tokens after deletion: $e');
        // Still return success since server deletion succeeded
      }
      final message = response.data?['message'] as String? ?? 'Account deleted';
      return ApiResponse.success(message, response.statusCode);
    } else {
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

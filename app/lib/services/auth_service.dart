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

  /// Register a new user
  /// Returns success message or error
  Future<ApiResponse<String>> register({
    required String email,
    required String password,
  }) async {
    final response = await _apiService.post(
      ApiConfig.registerEndpoint,
      body: {
        'email': email,
        'password': password,
      },
    );

    if (response.isSuccess) {
      final message = response.data?['message'] as String? ?? 'User created';
      return ApiResponse.success(message, response.statusCode);
    } else {
      return ApiResponse.error(
        response.error ?? 'Registration failed',
        response.statusCode,
      );
    }
  }

  /// Login with email and password
  /// Returns tokens on success, stores them securely
  Future<ApiResponse<AuthTokens>> login({
    required String email,
    required String password,
  }) async {
    final response = await _apiService.post(
      ApiConfig.loginEndpoint,
      body: {
        'email': email,
        'password': password,
      },
    );

    if (response.isSuccess && response.data != null) {
      try {
        final tokens = AuthTokens.fromJson(response.data!);

        // Store tokens securely
        await _storageService.saveTokens(
          accessToken: tokens.accessToken,
          refreshToken: tokens.refreshToken,
        );

        return ApiResponse.success(tokens, response.statusCode);
      } catch (e) {
        return ApiResponse.error('Failed to parse tokens', response.statusCode);
      }
    } else {
      return ApiResponse.error(
        response.error ?? 'Login failed',
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
  Future<ApiResponse<User>> getCurrentUser() async {
    final accessToken = await _storageService.getAccessToken();

    if (accessToken == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    final response = await _apiService.get(
      ApiConfig.userEndpoint,
      headers: {
        ApiConfig.authorizationHeader: 'Bearer $accessToken',
      },
    );

    if (response.isSuccess && response.data != null) {
      try {
        final user = User.fromJson(response.data!);
        return ApiResponse.success(user, response.statusCode);
      } catch (e) {
        return ApiResponse.error('Failed to parse user data', response.statusCode);
      }
    } else if (response.statusCode == 401) {
      // Token expired, try to refresh
      final refreshResponse = await refreshToken();
      if (refreshResponse.isSuccess) {
        // Retry with new token
        return getCurrentUser();
      } else {
        return ApiResponse.error('Session expired', 401);
      }
    } else {
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
    final accessToken = await _storageService.getAccessToken();

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
      await _storageService.deleteTokens();
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
}

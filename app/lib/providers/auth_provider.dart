import 'package:flutter/foundation.dart';
import 'package:logger/logger.dart';

import '../models/user.dart';
import '../services/app_logger.dart';
import '../services/auth_service.dart';

/// Authentication state
enum AuthState {
  initial,
  authenticated,
  unauthenticated,
  loading,
}

/// Authentication provider for managing auth state
class AuthProvider with ChangeNotifier {
  final AuthService _authService;
  final Logger _log = AppLogger.getLogger('AuthProvider');

  AuthState _state = AuthState.initial;
  User? _currentUser;
  String? _errorMessage;

  AuthProvider({AuthService? authService})
      : _authService = authService ?? AuthService();

  // Getters
  AuthState get state => _state;
  User? get currentUser => _currentUser;
  String? get errorMessage => _errorMessage;
  bool get isAuthenticated => _state == AuthState.authenticated;
  bool get isLoading => _state == AuthState.loading;

  /// Initialize authentication state
  /// Checks if user has stored tokens and attempts to refresh them
  Future<void> initialize() async {
    _setState(AuthState.loading);
    _errorMessage = null;

    try {
      final hasTokens = await _authService.isAuthenticated();

      if (hasTokens) {
        // Try to refresh token to ensure it's still valid
        final refreshResponse = await _authService.refreshToken();

        if (refreshResponse.isSuccess) {
          // Load user data
          final userResponse = await _authService.getCurrentUser();

          if (userResponse.isSuccess && userResponse.data != null) {
            _currentUser = userResponse.data;
            _setState(AuthState.authenticated);
            return;
          }
        }
      }

      // No valid session found
      _setState(AuthState.unauthenticated);
    } catch (e) {
      _errorMessage = 'Failed to initialize: ${e.toString()}';
      _setState(AuthState.unauthenticated);
    }
  }

  /// Login with Google ID token
  Future<bool> loginWithGoogle({
    required String idToken,
  }) async {
    _setState(AuthState.loading);
    _errorMessage = null;

    try {
      // Clear any old tokens first to prevent nil UUID issues
      _log.d('Clearing stale tokens before login');
      await _authService.clearTokens();

      final response = await _authService.loginWithGoogle(
        idToken: idToken,
      );

      if (response.isSuccess) {
        // Load user data using the access token we just received
        // This avoids reading from storage immediately after writing (web timing issue)
        final accessToken = response.data?.accessToken;
        _log.d('Using token from response for getCurrentUser (len=${accessToken?.length ?? 0})');

        final userResponse = await _authService.getCurrentUser(
          useToken: accessToken,
        );

        if (userResponse.isSuccess && userResponse.data != null) {
          _currentUser = userResponse.data;
          _setState(AuthState.authenticated);
          _log.i('Login successful: user=${_currentUser?.email}');
          return true;
        } else {
          _errorMessage = 'Failed to load user data: ${userResponse.error}';
          _setState(AuthState.unauthenticated);
          _log.e('Failed to load user after login: ${userResponse.error}');
          return false;
        }
      } else {
        _errorMessage = response.error ?? 'Google login failed';
        _setState(AuthState.unauthenticated);
        _log.e('Login failed: $_errorMessage');
        return false;
      }
    } catch (e) {
      _errorMessage = 'Google login error: ${e.toString()}';
      _setState(AuthState.unauthenticated);
      _log.e('Exception during login', error: e);
      return false;
    }
  }

  /// Logout user
  Future<void> logout() async {
    _setState(AuthState.loading);
    _errorMessage = null;

    try {
      await _authService.logout();
      _currentUser = null;
      _setState(AuthState.unauthenticated);
    } catch (e) {
      _errorMessage = 'Logout error: ${e.toString()}';
      // Still log out locally even if server request fails
      _currentUser = null;
      _setState(AuthState.unauthenticated);
    }
  }

  /// Refresh current user data
  Future<bool> refreshUserData() async {
    if (_state != AuthState.authenticated) {
      return false;
    }

    try {
      final userResponse = await _authService.getCurrentUser();

      if (userResponse.isSuccess && userResponse.data != null) {
        _currentUser = userResponse.data;
        notifyListeners();
        return true;
      } else {
        _errorMessage = 'Failed to refresh user data';
        return false;
      }
    } catch (e) {
      _errorMessage = 'Refresh error: ${e.toString()}';
      return false;
    }
  }

  /// Update user profile
  Future<bool> updateProfile({
    DateTime? birthdate,
    String? language,
    double? height,
    double? weight,
    Map<String, dynamic>? data,
  }) async {
    if (_state != AuthState.authenticated) {
      _errorMessage = 'Not authenticated';
      return false;
    }

    try {
      final response = await _authService.updateProfile(
        birthdate: birthdate,
        language: language,
        height: height,
        weight: weight,
        data: data,
      );

      if (response.isSuccess) {
        // Refresh user data to get updated profile
        await refreshUserData();
        return true;
      } else {
        _errorMessage = response.error ?? 'Failed to update profile';
        return false;
      }
    } catch (e) {
      _errorMessage = 'Update error: ${e.toString()}';
      return false;
    }
  }

  /// Delete user account
  Future<bool> deleteAccount() async {
    if (_state != AuthState.authenticated) {
      _errorMessage = 'Not authenticated';
      return false;
    }

    try {
      final response = await _authService.deleteAccount();

      if (response.isSuccess) {
        _currentUser = null;
        _setState(AuthState.unauthenticated);
        return true;
      } else {
        _errorMessage = response.error ?? 'Failed to delete account';
        return false;
      }
    } catch (e) {
      _errorMessage = 'Delete error: ${e.toString()}';
      return false;
    }
  }

  /// Clear error message
  void clearError() {
    _errorMessage = null;
    notifyListeners();
  }

  /// Set authentication state and notify listeners
  void _setState(AuthState newState) {
    _state = newState;
    notifyListeners();
  }
}

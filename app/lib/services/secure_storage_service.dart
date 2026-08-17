import 'package:flutter_secure_storage/flutter_secure_storage.dart';

import 'app_logger.dart';

class StorageUnavailableException implements Exception {
  final String message;
  final Object? originalError;

  StorageUnavailableException(this.message, [this.originalError]);

  @override
  String toString() => 'StorageUnavailableException: $message${originalError != null ? ' ($originalError)' : ''}';
}

class SecureStorageService {
  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';
  static const String _healthCheckKey = '__storage_health_check__';

  final FlutterSecureStorage _storage;
  bool _isHealthy = false;

  SecureStorageService({FlutterSecureStorage? storage})
      : _storage = storage ??
            const FlutterSecureStorage(
              aOptions: AndroidOptions(),
              iOptions: IOSOptions(
                accessibility: KeychainAccessibility.first_unlock,
              ),
              webOptions: WebOptions(
                dbName: 'vigor_secure_storage',
                publicKey: 'vigor_public_key',
              ),
            );

  /// Check if secure storage is available and working
  /// This must be called at app startup before any other storage operations
  /// Throws [StorageUnavailableException] if storage is not available
  Future<void> initialize() async {
    try {
      AppLogger.debug('[SecureStorage] Initializing secure storage...');

      // Test write operation
      await _storage.write(
        key: _healthCheckKey,
        value: DateTime.now().toIso8601String(),
      );

      // Test read operation
      final value = await _storage.read(key: _healthCheckKey);
      if (value == null) {
        throw StorageUnavailableException(
          'Storage write succeeded but read returned null',
        );
      }

      // Test delete operation
      await _storage.delete(key: _healthCheckKey);

      _isHealthy = true;
      AppLogger.info('[SecureStorage] Secure storage initialized successfully');
    } catch (e) {
      _isHealthy = false;
      AppLogger.error('[SecureStorage] Secure storage initialization failed', e);
      throw StorageUnavailableException(
        'Secure storage is not available on this platform/browser. '
        'This app requires secure storage to function. '
        'Please ensure you are not in incognito/private mode and that '
        'your browser allows IndexedDB/localStorage.',
        e,
      );
    }
  }

  /// Ensure storage is initialized before any operation
  void _ensureHealthy() {
    if (!_isHealthy) {
      throw StorageUnavailableException(
        'Storage has not been initialized. Call initialize() first.',
      );
    }
  }

  /// Save access token
  Future<void> saveAccessToken(String token) async {
    _ensureHealthy();
    try {
      await _storage.write(key: _accessTokenKey, value: token);
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to save access token', e);
      throw StorageUnavailableException('Failed to save access token', e);
    }
  }

  /// Get access token
  Future<String?> getAccessToken() async {
    _ensureHealthy();
    try {
      final token = await _storage.read(key: _accessTokenKey);
      AppLogger.debug('[SecureStorage] Read access token: ${token != null ? "found (${token.length}b)" : "none"}');
      return token;
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to read access token', e);
      throw StorageUnavailableException('Failed to read access token', e);
    }
  }

  /// Save refresh token
  Future<void> saveRefreshToken(String token) async {
    _ensureHealthy();
    try {
      await _storage.write(key: _refreshTokenKey, value: token);
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to save refresh token', e);
      throw StorageUnavailableException('Failed to save refresh token', e);
    }
  }

  /// Get refresh token
  Future<String?> getRefreshToken() async {
    _ensureHealthy();
    try {
      final token = await _storage.read(key: _refreshTokenKey);
      AppLogger.debug('[SecureStorage] Read refresh token: ${token != null ? "found (${token.length}b)" : "none"}');
      return token;
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to read refresh token', e);
      throw StorageUnavailableException('Failed to read refresh token', e);
    }
  }

  /// Save both tokens
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    _ensureHealthy();
    try {
      await Future.wait([
        saveAccessToken(accessToken),
        saveRefreshToken(refreshToken),
      ]);
      AppLogger.debug('[SecureStorage] Tokens saved (access=${accessToken.length}b refresh=${refreshToken.length}b)');
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to save tokens', e);
      rethrow;
    }
  }

  /// Delete all tokens
  Future<void> deleteTokens() async {
    _ensureHealthy();
    try {
      await Future.wait([
        _storage.delete(key: _accessTokenKey),
        _storage.delete(key: _refreshTokenKey),
      ]);
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to delete tokens', e);
      throw StorageUnavailableException('Failed to delete tokens', e);
    }
  }

  /// Check if user has valid tokens stored
  Future<bool> hasTokens() async {
    _ensureHealthy();
    try {
      final accessToken = await getAccessToken();
      final refreshToken = await getRefreshToken();
      final hasTokens = accessToken != null && refreshToken != null;
      AppLogger.debug('[SecureStorage] Token check: $hasTokens');
      return hasTokens;
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to check tokens', e);
      rethrow;
    }
  }

  /// Clear all stored data
  Future<void> clearAll() async {
    _ensureHealthy();
    try {
      await _storage.deleteAll();
    } catch (e) {
      AppLogger.error('[SecureStorage] Failed to clear all storage', e);
      throw StorageUnavailableException('Failed to clear storage', e);
    }
  }

  /// Get storage health status
  bool get isHealthy => _isHealthy;
}

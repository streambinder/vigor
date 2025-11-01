import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:logger/logger.dart';

import 'app_logger.dart';

/// Service for securely storing sensitive data like tokens
class SecureStorageService {
  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';

  final FlutterSecureStorage _storage;
  final Logger _log = AppLogger.getLogger('SecureStorage');

  SecureStorageService({FlutterSecureStorage? storage})
      : _storage = storage ??
            const FlutterSecureStorage(
              webOptions: WebOptions(
                dbName: 'vigor_secure_storage',
                publicKey: 'vigor_public_key',
              ),
            );

  /// Save access token
  Future<void> saveAccessToken(String token) async {
    await _storage.write(key: _accessTokenKey, value: token);
  }

  /// Get access token
  Future<String?> getAccessToken() async {
    try {
      final token = await _storage.read(key: _accessTokenKey);
      _log.d('Read access token: ${token != null ? "found (${token.length}b)" : "none"}');
      return token;
    } catch (e) {
      _log.e('Failed to read access token', error: e);
      rethrow;
    }
  }

  /// Save refresh token
  Future<void> saveRefreshToken(String token) async {
    await _storage.write(key: _refreshTokenKey, value: token);
  }

  /// Get refresh token
  Future<String?> getRefreshToken() async {
    try {
      final token = await _storage.read(key: _refreshTokenKey);
      _log.d('Read refresh token: ${token != null ? "found (${token.length}b)" : "none"}');
      return token;
    } catch (e) {
      _log.e('Failed to read refresh token', error: e);
      rethrow;
    }
  }

  /// Save both tokens
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    try {
      await Future.wait([
        saveAccessToken(accessToken),
        saveRefreshToken(refreshToken),
      ]);
      _log.d('Tokens saved (access=${accessToken.length}b refresh=${refreshToken.length}b)');
    } catch (e) {
      _log.e('Failed to save tokens', error: e);
      rethrow;
    }
  }

  /// Delete all tokens
  Future<void> deleteTokens() async {
    await Future.wait([
      _storage.delete(key: _accessTokenKey),
      _storage.delete(key: _refreshTokenKey),
    ]);
  }

  /// Check if user has valid tokens stored
  Future<bool> hasTokens() async {
    try {
      final accessToken = await getAccessToken();
      final refreshToken = await getRefreshToken();
      final hasTokens = accessToken != null && refreshToken != null;
      _log.d('Token check: $hasTokens');
      return hasTokens;
    } catch (e) {
      _log.e('Failed to check tokens', error: e);
      return false;
    }
  }

  /// Clear all stored data
  Future<void> clearAll() async {
    await _storage.deleteAll();
  }
}

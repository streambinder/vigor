import 'package:flutter_secure_storage/flutter_secure_storage.dart';

/// Service for securely storing sensitive data like tokens
class SecureStorageService {
  static const String _accessTokenKey = 'access_token';
  static const String _refreshTokenKey = 'refresh_token';

  final FlutterSecureStorage _storage;

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
      print('[SecureStorage] Reading access token...');
      final token = await _storage.read(key: _accessTokenKey);
      print('[SecureStorage] Access token read - ${token != null ? "found (${token.length} chars)" : "not found"}');
      return token;
    } catch (e) {
      print('[SecureStorage] ERROR reading access token: $e');
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
      print('[SecureStorage] Reading refresh token...');
      final token = await _storage.read(key: _refreshTokenKey);
      print('[SecureStorage] Refresh token read - ${token != null ? "found (${token.length} chars)" : "not found"}');
      return token;
    } catch (e) {
      print('[SecureStorage] ERROR reading refresh token: $e');
      rethrow;
    }
  }

  /// Save both tokens
  Future<void> saveTokens({
    required String accessToken,
    required String refreshToken,
  }) async {
    try {
      print('[SecureStorage] Saving tokens...');
      await Future.wait([
        saveAccessToken(accessToken),
        saveRefreshToken(refreshToken),
      ]);
      print('[SecureStorage] Tokens saved successfully');
    } catch (e) {
      print('[SecureStorage] ERROR saving tokens: $e');
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
      return accessToken != null && refreshToken != null;
    } catch (e) {
      print('[SecureStorage] ERROR checking tokens: $e');
      return false;
    }
  }

  /// Clear all stored data
  Future<void> clearAll() async {
    await _storage.deleteAll();
  }
}

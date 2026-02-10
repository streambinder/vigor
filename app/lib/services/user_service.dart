import '../config/api_config.dart';
import '../models/api_response.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';

class UserInfo {
  final String id;
  final String firstName;
  final String lastName;

  const UserInfo({required this.id, required this.firstName, required this.lastName});

  String get displayName {
    if (firstName.isEmpty && lastName.isEmpty) return 'Unknown';
    return '$firstName $lastName'.trim();
  }

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['user_id'] as String? ?? '',
      firstName: json['first_name'] as String? ?? '',
      lastName: json['last_name'] as String? ?? '',
    );
  }
}

class UserService {
  final AuthenticatedApiService _apiService;

  UserService({
    AuthenticatedApiService? apiService,
    SecureStorageService? storageService,
  }) : _apiService = apiService ??
            AuthenticatedApiService(storageService: storageService);

  Future<ApiResponse<List<UserInfo>>> getUsers() async {
    AppLogger.debug('[UserService] Fetching users');

    final response = await _apiService.get('/users');

    if (response.isSuccess && response.data != null) {
      try {
        final usersJson = response.data!['users'] as List;
        final users = usersJson.map((json) => UserInfo.fromJson(json)).toList();
        AppLogger.info('[UserService] Fetched ${users.length} users');
        return ApiResponse.success(users, response.statusCode);
      } catch (e) {
        AppLogger.error('[UserService] failed to parse users', e);
        return ApiResponse.error('Failed to parse users', response.statusCode);
      }
    } else {
      AppLogger.error('[UserService] Failed to fetch users: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch users',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Map<String, dynamic>>> uploadAvatar(List<int> bytes, String filename) async {
    AppLogger.debug('[UserService] Uploading avatar');
    return _apiService.postMultipart(
      ApiConfig.avatarEndpoint,
      bytes: bytes,
      fieldName: 'avatar',
      filename: filename,
    );
  }
}

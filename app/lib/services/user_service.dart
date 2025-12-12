import '../models/api_response.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';

class UserInfo {
  final String id;
  final String email;

  const UserInfo({required this.id, required this.email});

  factory UserInfo.fromJson(Map<String, dynamic> json) {
    return UserInfo(
      id: json['id'] as String? ?? '',
      email: json['email'] as String? ?? '',
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
}

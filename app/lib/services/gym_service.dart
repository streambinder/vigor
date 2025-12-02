import '../models/api_response.dart';
import '../models/gym.dart';
import 'api_service.dart';
import 'app_logger.dart';
import 'secure_storage_service.dart';

class GymService {
  final ApiService _apiService;
  final SecureStorageService _storageService;

  GymService({
    ApiService? apiService,
    SecureStorageService? storageService,
  })  : _apiService = apiService ?? ApiService(),
        _storageService = storageService ?? SecureStorageService();

  Future<Map<String, String>?> _getAuthHeaders() async {
    final accessToken = await _storageService.getAccessToken();
    if (accessToken == null) {
      return null;
    }
    return {
      'Authorization': 'Bearer $accessToken',
    };
  }

  Future<ApiResponse<List<Gym>>> getGyms() async {
    AppLogger.debug('[GymService] Fetching gyms');
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    final response = await _apiService.get(
      '/gym',
      headers: headers,
    );

    if (response.isSuccess && response.data != null) {
      try {
        final gymsJson = response.data!['gyms'] as List;
        final gyms = gymsJson.map((json) => Gym.fromJson(json)).toList();
        AppLogger.info('[GymService] Fetched ${gyms.length} gyms');
        return ApiResponse.success(gyms, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse gyms', e);
        return ApiResponse.error('Failed to parse gyms', response.statusCode);
      }
    } else {
      AppLogger.error('[GymService] Failed to fetch gyms: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch gyms',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Gym>> getGym(String name) async {
    AppLogger.debug('[GymService] Fetching gym: $name');
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    final response = await _apiService.get(
      '/gym/$name',
      headers: headers,
    );

    if (response.isSuccess && response.data != null) {
      try {
        final gym = Gym.fromJson(response.data!);
        AppLogger.info('[GymService] Fetched gym: ${gym.name}');
        return ApiResponse.success(gym, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse gym', e);
        return ApiResponse.error('Failed to parse gym', response.statusCode);
      }
    } else {
      AppLogger.error('[GymService] Failed to fetch gym: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch gym',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Gym>> createGym({
    required String name,
    required List<String> equipment,
  }) async {
    AppLogger.debug('[GymService] Creating gym: $name');
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    final response = await _apiService.post(
      '/gym',
      headers: headers,
      body: {
        'name': name,
        'equipment': equipment,
      },
    );

    if (response.isSuccess && response.data != null) {
      try {
        final gym = Gym.fromJson(response.data!['gym']);
        AppLogger.info('[GymService] Created gym: ${gym.name}');
        return ApiResponse.success(gym, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse created gym', e);
        return ApiResponse.error('Failed to parse created gym', response.statusCode);
      }
    } else {
      AppLogger.error('[GymService] Failed to create gym: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to create gym',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Gym>> updateGym({
    required String currentName,
    String? newName,
    List<String>? equipment,
  }) async {
    AppLogger.debug('[GymService] Updating gym: $currentName');
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    final body = <String, dynamic>{};
    if (newName != null) body['name'] = newName;
    if (equipment != null) body['equipment'] = equipment;

    final response = await _apiService.put(
      '/gym/$currentName',
      headers: headers,
      body: body,
    );

    if (response.isSuccess && response.data != null) {
      try {
        final gym = Gym.fromJson(response.data!['gym']);
        AppLogger.info('[GymService] Updated gym: ${gym.name}');
        return ApiResponse.success(gym, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse updated gym', e);
        return ApiResponse.error('Failed to parse updated gym', response.statusCode);
      }
    } else {
      AppLogger.error('[GymService] Failed to update gym: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to update gym',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<String>> deleteGym(String name) async {
    AppLogger.debug('[GymService] Deleting gym: $name');
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    final response = await _apiService.delete(
      '/gym/$name',
      headers: headers,
    );

    if (response.isSuccess) {
      final message = response.data?['message'] as String? ?? 'Gym deleted';
      AppLogger.info('[GymService] Deleted gym: $name');
      return ApiResponse.success(message, response.statusCode);
    } else {
      AppLogger.error('[GymService] Failed to delete gym: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to delete gym',
        response.statusCode,
      );
    }
  }
}

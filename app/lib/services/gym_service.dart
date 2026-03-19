import '../models/api_response.dart';
import '../models/equipment_info.dart';
import '../models/gym.dart';
import 'app_event.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';
import 'api_service.dart';

class GymService {
  final AuthenticatedApiService _apiService;
  final ApiService _publicApiService;
  final void Function(AppEvent)? emitEvent;

  GymService({
    AuthenticatedApiService? apiService,
    SecureStorageService? storageService,
    ApiService? publicApiService,
    this.emitEvent,
  })  : _apiService = apiService ??
            AuthenticatedApiService(storageService: storageService),
        _publicApiService = publicApiService ?? ApiService();

  Future<ApiResponse<List<EquipmentInfo>>> getAvailableEquipment() async {
    AppLogger.debug('[GymService] Fetching available equipment');
    final response = await _publicApiService.get('/equipment');
    if (response.isSuccess && response.data != null) {
      try {
        final equipment = (response.data!['equipment'] as List)
            .map((e) => EquipmentInfo.fromJson(e as Map<String, dynamic>))
            .toList();
        AppLogger.info('[GymService] Fetched ${equipment.length} equipment');
        return ApiResponse.success(equipment, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse equipment', e);
        return ApiResponse.error('Failed to parse equipment', response.statusCode);
      }
    }
    AppLogger.error('[GymService] Failed to fetch equipment: ${response.error}');
    return ApiResponse.error(
      response.error ?? 'Failed to fetch equipment',
      response.statusCode,
    );
  }

  Future<ApiResponse<List<String>>> getAvailableGoals() async {
    AppLogger.debug('[GymService] Fetching available goals');
    final response = await _publicApiService.get('/goals');
    if (response.isSuccess && response.data != null) {
      try {
        final goals = (response.data!['goals'] as List).cast<String>();
        AppLogger.info('[GymService] Fetched ${goals.length} goals');
        return ApiResponse.success(goals, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse goals', e);
        return ApiResponse.error('Failed to parse goals', response.statusCode);
      }
    }
    AppLogger.error('[GymService] Failed to fetch goals: ${response.error}');
    return ApiResponse.error(
      response.error ?? 'Failed to fetch goals',
      response.statusCode,
    );
  }

  Future<ApiResponse<List<String>>> getAvailableMuscles() async {
    AppLogger.debug('[GymService] Fetching available muscles');
    final response = await _publicApiService.get('/muscles');
    if (response.isSuccess && response.data != null) {
      try {
        final muscles = (response.data!['muscles'] as List).cast<String>();
        AppLogger.info('[GymService] Fetched ${muscles.length} muscles');
        return ApiResponse.success(muscles, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse muscles', e);
        return ApiResponse.error('Failed to parse muscles', response.statusCode);
      }
    }
    AppLogger.error('[GymService] Failed to fetch muscles: ${response.error}');
    return ApiResponse.error(
      response.error ?? 'Failed to fetch muscles',
      response.statusCode,
    );
  }

  Future<ApiResponse<List<String>>> getAvailableMovementFamilies() async {
    AppLogger.debug('[GymService] Fetching available movement families');
    final response = await _publicApiService.get('/movement-families');
    if (response.isSuccess && response.data != null) {
      try {
        final families = (response.data!['movement_families'] as List).cast<String>();
        AppLogger.info('[GymService] Fetched ${families.length} movement families');
        return ApiResponse.success(families, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse movement families', e);
        return ApiResponse.error('Failed to parse movement families', response.statusCode);
      }
    }
    AppLogger.error('[GymService] Failed to fetch movement families: ${response.error}');
    return ApiResponse.error(
      response.error ?? 'Failed to fetch movement families',
      response.statusCode,
    );
  }

  Future<ApiResponse<List<String>>> getAvailableMethodologies() async {
    AppLogger.debug('[GymService] Fetching available methodologies');
    final response = await _publicApiService.get('/methodologies');
    if (response.isSuccess && response.data != null) {
      try {
        final methodologies = (response.data!['methodologies'] as List).cast<String>();
        AppLogger.info('[GymService] Fetched ${methodologies.length} methodologies');
        return ApiResponse.success(methodologies, response.statusCode);
      } catch (e) {
        AppLogger.error('[GymService] failed to parse methodologies', e);
        return ApiResponse.error('Failed to parse methodologies', response.statusCode);
      }
    }
    AppLogger.error('[GymService] Failed to fetch methodologies: ${response.error}');
    return ApiResponse.error(
      response.error ?? 'Failed to fetch methodologies',
      response.statusCode,
    );
  }

  Future<ApiResponse<List<Gym>>> getGyms() async {
    AppLogger.debug('[GymService] Fetching gyms');

    final response = await _apiService.get('/gym');

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

  Future<ApiResponse<Gym>> getGym(String id) async {
    AppLogger.debug('[GymService] Fetching gym: $id');

    final response = await _apiService.get('/gym/$id');

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
    Map<String, List<double>>? modifierVariants,
  }) async {
    AppLogger.debug('[GymService] Creating gym: $name');

    final body = <String, dynamic>{
      'name': name,
      'equipment': equipment,
    };
    if (modifierVariants != null && modifierVariants.isNotEmpty) {
      body['modifier_variants'] = modifierVariants;
    }

    final response = await _apiService.post('/gym', body: body);

    if (response.isSuccess && response.data != null) {
      try {
        final gym = Gym.fromJson(response.data!['gym']);
        AppLogger.info('[GymService] Created gym: ${gym.name}');
        emitEvent?.call(GymListChanged());
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
    required String id,
    String? name,
    List<String>? equipment,
    Map<String, List<double>>? modifierVariants,
  }) async {
    AppLogger.debug('[GymService] Updating gym: $id');

    final body = <String, dynamic>{};
    if (name != null) body['name'] = name;
    if (equipment != null) body['equipment'] = equipment;
    if (modifierVariants != null) body['modifier_variants'] = modifierVariants;

    final response = await _apiService.put('/gym/$id', body: body);

    if (response.isSuccess && response.data != null) {
      try {
        final gym = Gym.fromJson(response.data!['gym']);
        AppLogger.info('[GymService] Updated gym: ${gym.name}');
        emitEvent?.call(GymListChanged());
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

  Future<ApiResponse<String>> deleteGym(String id) async {
    AppLogger.debug('[GymService] Deleting gym: $id');

    final response = await _apiService.delete('/gym/$id');

    if (response.isSuccess) {
      final message = response.data?['message'] as String? ?? 'Gym deleted';
      AppLogger.info('[GymService] Deleted gym: $id');
      emitEvent?.call(GymListChanged());
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

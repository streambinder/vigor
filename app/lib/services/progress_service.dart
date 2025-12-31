import '../models/api_response.dart';
import '../models/progress.dart';
import '../models/family_progress.dart';
import '../models/muscle_impact.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';

class ProgressService {
  final AuthenticatedApiService _apiService;

  ProgressService({
    AuthenticatedApiService? apiService,
    SecureStorageService? storageService,
  }) : _apiService = apiService ??
            AuthenticatedApiService(storageService: storageService);

  Future<ApiResponse<Progress>> getProgress() async {
    AppLogger.debug('[ProgressService] Fetching progress');

    final response = await _apiService.get('/progress');

    if (response.isSuccess && response.data != null) {
      try {
        final progress = Progress.fromJson(response.data!);
        AppLogger.info('[ProgressService] Fetched progress: ${progress.trainings} trainings');
        return ApiResponse.success(progress, response.statusCode);
      } catch (e) {
        AppLogger.error('[ProgressService] failed to parse progress', e);
        return ApiResponse.error('Failed to parse progress', response.statusCode);
      }
    } else {
      AppLogger.error('[ProgressService] Failed to fetch progress: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch progress',
        response.statusCode,
      );
    }
  }

  /// Parse family progress map from dynamic JSON
  static Map<String, FamilyProgress> parseFamilies(Map<String, dynamic> families) {
    return families.map((key, value) => MapEntry(
      key,
      FamilyProgress.fromJson(value as Map<String, dynamic>),
    ));
  }

  /// Parse muscle impact map from dynamic JSON
  static Map<String, MuscleImpact> parseMuscles(Map<String, dynamic> muscles) {
    return muscles.map((key, value) => MapEntry(
      key,
      MuscleImpact.fromJson(value as Map<String, dynamic>),
    ));
  }
}

import '../models/api_response.dart';
import '../models/training.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';

class TrainingService {
  final AuthenticatedApiService _apiService;

  TrainingService({
    AuthenticatedApiService? apiService,
    SecureStorageService? storageService,
  }) : _apiService = apiService ??
            AuthenticatedApiService(storageService: storageService);

  Future<ApiResponse<Training>> generateTraining({
    required int duration,
    required String gym,
    String? prompt,
    List<String>? equipment,
  }) async {
    AppLogger.debug('[TrainingService] Generating training with duration: $duration, gym: $gym');

    final body = <String, dynamic>{
      'duration': duration,
      'gym': gym,
    };
    if (prompt != null && prompt.isNotEmpty) {
      body['prompt'] = prompt;
    }
    if (equipment != null && equipment.isNotEmpty) {
      body['equipment'] = equipment;
    }

    final response = await _apiService.post('/training', body: body);

    if (response.isSuccess && response.data != null) {
      try {
        final training = Training.fromJson(response.data!);
        AppLogger.info('[TrainingService] Generated training: ${training.id}');
        return ApiResponse.success(training, response.statusCode);
      } catch (e) {
        AppLogger.error('[TrainingService] failed to parse training', e);
        return ApiResponse.error('Failed to parse training', response.statusCode);
      }
    } else {
      AppLogger.error('[TrainingService] Failed to generate training: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to generate training',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<List<Training>>> getTrainings() async {
    AppLogger.debug('[TrainingService] Fetching trainings');

    final response = await _apiService.get('/training');

    if (response.isSuccess && response.data != null) {
      try {
        final trainingsJson = response.data!['trainings'] as List;
        final trainings = trainingsJson.map((json) => Training.fromJson(json)).toList();
        AppLogger.info('[TrainingService] Fetched ${trainings.length} trainings');
        return ApiResponse.success(trainings, response.statusCode);
      } catch (e) {
        AppLogger.error('[TrainingService] failed to parse trainings', e);
        return ApiResponse.error('Failed to parse trainings', response.statusCode);
      }
    } else {
      AppLogger.error('[TrainingService] Failed to fetch trainings: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch trainings',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<String>> deleteTraining(String trainingId) async {
    AppLogger.debug('[TrainingService] Deleting training: $trainingId');

    final response = await _apiService.delete('/training/$trainingId');

    if (response.isSuccess) {
      final message = response.data?['message'] as String? ?? 'Training deleted';
      AppLogger.info('[TrainingService] Deleted training: $trainingId');
      return ApiResponse.success(message, response.statusCode);
    } else {
      AppLogger.error('[TrainingService] Failed to delete training: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to delete training',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Training>> completeTraining(String trainingId) async {
    AppLogger.debug('[TrainingService] Completing training: $trainingId');

    final response = await _apiService.post('/training/complete/$trainingId');

    if (response.isSuccess && response.data != null) {
      try {
        final training = Training.fromJson(response.data!['training']);
        AppLogger.info('[TrainingService] Completed training: $trainingId');
        return ApiResponse.success(training, response.statusCode);
      } catch (e) {
        AppLogger.error('[TrainingService] failed to parse completed training', e);
        return ApiResponse.error('Failed to parse completed training', response.statusCode);
      }
    } else {
      AppLogger.error('[TrainingService] Failed to complete training: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to complete training',
        response.statusCode,
      );
    }
  }
}

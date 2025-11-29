import 'package:logger/logger.dart';

import '../models/api_response.dart';
import '../models/training.dart';
import 'api_service.dart';
import 'app_logger.dart';
import 'secure_storage_service.dart';

class TrainingService {
  final ApiService _apiService;
  final SecureStorageService _storageService;
  final Logger _log = AppLogger.getLogger('TrainingService');

  TrainingService({
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

  Future<ApiResponse<Training>> generateTraining({
    required int duration,
    required String gym,
    String? prompt,
    List<String>? equipment,
  }) async {
    _log.d('Generating training with duration: $duration, gym: $gym');
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

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

    final response = await _apiService.post(
      '/training',
      headers: headers,
      body: body,
    );

    if (response.isSuccess && response.data != null) {
      try {
        final training = Training.fromJson(response.data!);
        _log.i('Generated training: ${training.id}');
        return ApiResponse.success(training, response.statusCode);
      } catch (e) {
        _log.e('Failed to parse training', error: e);
        return ApiResponse.error('Failed to parse training', response.statusCode);
      }
    } else {
      _log.e('Failed to generate training: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to generate training',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<List<Training>>> getTrainings() async {
    _log.d('Fetching trainings');
    final headers = await _getAuthHeaders();
    if (headers == null) {
      return ApiResponse.error('Not authenticated', 401);
    }

    final response = await _apiService.get(
      '/training',
      headers: headers,
    );

    if (response.isSuccess && response.data != null) {
      try {
        final trainingsJson = response.data!['trainings'] as List;
        final trainings = trainingsJson.map((json) => Training.fromJson(json)).toList();
        _log.i('Fetched ${trainings.length} trainings');
        return ApiResponse.success(trainings, response.statusCode);
      } catch (e) {
        _log.e('Failed to parse trainings', error: e);
        return ApiResponse.error('Failed to parse trainings', response.statusCode);
      }
    } else {
      _log.e('Failed to fetch trainings: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch trainings',
        response.statusCode,
      );
    }
  }
}

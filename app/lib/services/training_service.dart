import 'dart:ui';
import '../models/activity.dart';
import '../models/api_response.dart';
import '../models/partner.dart';
import '../models/training.dart';
import '../models/training_feedback.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';

class TrainingService {
  final AuthenticatedApiService _apiService;
  final VoidCallback? onDataChanged;

  TrainingService({
    AuthenticatedApiService? apiService,
    SecureStorageService? storageService,
    this.onDataChanged,
  }) : _apiService = apiService ??
            AuthenticatedApiService(storageService: storageService);

  Future<ApiResponse<Training>> generateTraining({
    required int duration,
    required String gym,
    String? prompt,
    List<String>? equipment,
    List<String>? partners,
    bool? skipWarmupCooldown,
    String? methodology,
    List<String>? goals,
    List<String>? muscles,
    void Function(int attempt)? onRetry,
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
    if (partners != null && partners.isNotEmpty) {
      body['partners'] = partners;
    }
    if (skipWarmupCooldown == true) {
      body['skipWarmupCooldown'] = true;
    }
    if (methodology != null && methodology.isNotEmpty) {
      body['methodology'] = methodology;
    }
    if (goals != null && goals.isNotEmpty) {
      body['goals'] = goals;
    }
    if (muscles != null && muscles.isNotEmpty) {
      body['muscles'] = muscles;
    }

    const maxRetries = 2;
    const retryableStatusCodes = [500, 502, 503, 504];
    for (var attempt = 0; attempt <= maxRetries; attempt++) {
      final response = await _apiService.post('/training', body: body);

      // retry on 5xx errors (transient server issues) up to maxRetries times
      if (retryableStatusCodes.contains(response.statusCode) && attempt < maxRetries) {
        AppLogger.warning('[TrainingService] Got ${response.statusCode}, retrying (attempt ${attempt + 1})');
        onRetry?.call(attempt + 1);
        await Future.delayed(const Duration(seconds: 3));
        continue;
      }

      if (response.isSuccess && response.data != null) {
        try {
          final training = Training.fromJson(response.data!);
          AppLogger.info('[TrainingService] Generated training: ${training.id}');
          onDataChanged?.call();
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

    // should not reach here, but just in case
    return ApiResponse.error('Failed to generate training after retries', 500);
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
      onDataChanged?.call();
      return ApiResponse.success(message, response.statusCode);
    } else {
      AppLogger.error('[TrainingService] Failed to delete training: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to delete training',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Training>> completeTraining(
    String trainingId, {
    TrainingFeedback? feedback,
    Map<String, String>? activityFeedback,
    List<String>? activityReports,
    int? completedIn,
  }) async {
    AppLogger.debug('[TrainingService] Completing training: $trainingId');

    final body = <String, dynamic>{};
    if (feedback != null) {
      body['feedback'] = feedback.toJson();
    }
    if (activityFeedback != null && activityFeedback.isNotEmpty) {
      body['activityFeedback'] = activityFeedback;
    }
    if (activityReports != null && activityReports.isNotEmpty) {
      body['activityReports'] = activityReports;
    }
    if (completedIn != null) {
      body['completedIn'] = completedIn;
    }

    final response = await _apiService.post('/training/complete/$trainingId', body: body.isNotEmpty ? body : null);

    if (response.isSuccess && response.data != null) {
      try {
        final training = Training.fromJson(response.data!['training']);
        AppLogger.info('[TrainingService] Completed training: $trainingId');
        onDataChanged?.call();
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

  Future<ApiResponse<Training>> updateFeedback(
    String trainingId, {
    TrainingFeedback? feedback,
    Map<String, String>? activityFeedback,
    int? completedIn,
  }) async {
    AppLogger.debug('[TrainingService] Updating feedback for training: $trainingId');

    final body = <String, dynamic>{};
    if (feedback != null) {
      body['feedback'] = feedback.toJson();
    }
    if (activityFeedback != null) {
      body['activityFeedback'] = activityFeedback;
    }
    if (completedIn != null) {
      body['completedIn'] = completedIn;
    }

    final response = await _apiService.put('/training/feedback/$trainingId', body: body);

    if (response.isSuccess && response.data != null) {
      try {
        final training = Training.fromJson(response.data!['training']);
        AppLogger.info('[TrainingService] Updated feedback for training: $trainingId');
        return ApiResponse.success(training, response.statusCode);
      } catch (e) {
        AppLogger.error('[TrainingService] failed to parse updated training', e);
        return ApiResponse.error('Failed to parse updated training', response.statusCode);
      }
    } else {
      AppLogger.error('[TrainingService] Failed to update feedback: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to update feedback',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<String>> addPartner(String trainingId, String partner) async {
    AppLogger.debug('[TrainingService] Adding partner to training: $trainingId');

    final response = await _apiService.post(
      '/training/partner/$trainingId',
      body: {'partner': partner},
    );

    if (response.isSuccess) {
      final message = response.data?['message'] as String? ?? 'Partner added';
      AppLogger.info('[TrainingService] Added partner to training: $trainingId');
      return ApiResponse.success(message, response.statusCode);
    } else {
      AppLogger.error('[TrainingService] Failed to add partner: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to add partner',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Training>> copyTraining(String trainingId, String target) async {
    AppLogger.debug('[TrainingService] Copying training: $trainingId to $target');

    final response = await _apiService.post(
      '/training/copy/$trainingId',
      body: {'target': target},
    );

    if (response.isSuccess && response.data != null) {
      try {
        final training = Training.fromJson(response.data!);
        AppLogger.info('[TrainingService] Copied training: $trainingId');
        onDataChanged?.call();
        return ApiResponse.success(training, response.statusCode);
      } catch (e) {
        AppLogger.error('[TrainingService] failed to parse copied training', e);
        return ApiResponse.error('Failed to parse copied training', response.statusCode);
      }
    } else {
      AppLogger.error('[TrainingService] Failed to copy training: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to copy training',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<List<Partner>>> getPartners(String trainingId) async {
    AppLogger.debug('[TrainingService] Fetching partners for training: $trainingId');

    final response = await _apiService.get('/training/partners/$trainingId');

    if (response.isSuccess && response.data != null) {
      try {
        final partnersJson = response.data!['partners'] as List;
        final partners = partnersJson.map((json) => Partner.fromJson(json)).toList();
        AppLogger.info('[TrainingService] Fetched ${partners.length} partners');
        return ApiResponse.success(partners, response.statusCode);
      } catch (e) {
        AppLogger.error('[TrainingService] failed to parse partners', e);
        return ApiResponse.error('Failed to parse partners', response.statusCode);
      }
    } else {
      AppLogger.error('[TrainingService] Failed to fetch partners: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch partners',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<void>> createReport(String trainingId, String content) async {
    AppLogger.debug('[TrainingService] Creating report for training: $trainingId');

    final response = await _apiService.post('/report', body: {
      'training_id': trainingId,
      'content': content,
    });

    if (response.isSuccess) {
      AppLogger.info('[TrainingService] Created report for training: $trainingId');
      return ApiResponse.success(null, response.statusCode);
    } else {
      AppLogger.error('[TrainingService] Failed to create report: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to create report',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Activity>> shuffleActivity(String activityId) async {
    AppLogger.debug('[TrainingService] Shuffling activity: $activityId');

    final response = await _apiService.post('/activity/shuffle/$activityId');

    if (response.isSuccess && response.data != null) {
      try {
        final activity = Activity.fromJson(response.data!);
        AppLogger.info('[TrainingService] Shuffled activity: $activityId');
        return ApiResponse.success(activity, response.statusCode);
      } catch (e) {
        AppLogger.error('[TrainingService] failed to parse shuffled activity', e);
        return ApiResponse.error('Failed to parse shuffled activity', response.statusCode);
      }
    } else {
      AppLogger.error('[TrainingService] Failed to shuffle activity: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to shuffle activity',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<Map<String, String>>> shareTraining(String trainingId) async {
    final response = await _apiService.post('/training/share/$trainingId');
    if (response.isSuccess && response.data != null) {
      try {
        return ApiResponse.success({
          'token': response.data!['token'] as String,
          'url': response.data!['url'] as String,
        }, response.statusCode);
      } catch (e) {
        return ApiResponse.error('Failed to parse share link', response.statusCode);
      }
    }
    return ApiResponse.error(response.error ?? 'Failed to share training', response.statusCode);
  }

  Future<ApiResponse<Training>> claimSharedTraining(String token) async {
    final response = await _apiService.post('/training/shared/$token/claim');
    if (response.isSuccess && response.data != null) {
      try {
        final training = Training.fromJson(response.data!);
        onDataChanged?.call();
        return ApiResponse.success(training, response.statusCode);
      } catch (e) {
        return ApiResponse.error('Failed to parse claimed training', response.statusCode);
      }
    }
    return ApiResponse.error(response.error ?? 'Failed to claim training', response.statusCode);
  }
}

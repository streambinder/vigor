import 'dart:async';
import '../models/api_response.dart';
import '../models/activity.dart';
import '../dto/partner_info.dart';
import '../models/training.dart';
import '../models/training_feedback.dart';
import 'app_event.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';

class TrainingService {
  final AuthenticatedApiService _apiService;
  final void Function(AppEvent)? emitEvent;

  TrainingService({
    AuthenticatedApiService? apiService,
    SecureStorageService? storageService,
    this.emitEvent,
  }) : _apiService = apiService ??
            AuthenticatedApiService(storageService: storageService);

  Future<ApiResponse<Training>> generateTraining({
    int duration = 0,
    String gym = '',
    String? prompt,
    String? freeText,
    List<String>? equipment,
    List<String>? partners,
    bool? skipWarmupCooldown,
    String? methodology,
    List<String>? goals,
    List<String>? muscles,
    void Function(int attempt)? onRetry,
    void Function(String step)? onStep,
  }) async {
    // free text mode asks the backend to derive every tuning parameter from the
    // program text itself, so it carries freeText and nothing else
    final body = <String, dynamic>{};
    if (freeText != null && freeText.isNotEmpty) {
      body['freeText'] = freeText;
    } else {
      body['duration'] = duration;
      body['gym'] = gym;
      if (prompt != null && prompt.isNotEmpty) body['prompt'] = prompt;
      if (equipment != null && equipment.isNotEmpty) body['equipment'] = equipment;
      if (partners != null && partners.isNotEmpty) body['partners'] = partners;
      if (skipWarmupCooldown == true) body['skipWarmupCooldown'] = true;
      if (methodology != null && methodology.isNotEmpty) body['methodology'] = methodology;
      if (goals != null && goals.isNotEmpty) body['goals'] = goals;
      if (muscles != null && muscles.isNotEmpty) body['muscles'] = muscles;
    }
    AppLogger.debug('[TrainingService] Generating training: ${body.keys.join(', ')}');

    const maxRetries = 2;

    for (var attempt = 0; attempt <= maxRetries; attempt++) {
      try {
        final result = await _generateViaSSE(body, onStep: onStep);
        if (result != null) {
          emitEvent?.call(TrainingListChanged());
          return ApiResponse.success(result, 200);
        }
        // SSE returned an error event — treat as retryable
      } catch (e) {
        AppLogger.warning('[TrainingService] SSE attempt failed: $e');
      }

      if (attempt < maxRetries) {
        AppLogger.warning('[TrainingService] Retrying (attempt ${attempt + 1})');
        onRetry?.call(attempt + 1);
        await Future.delayed(const Duration(seconds: 3));
      }
    }

    return ApiResponse.error('Failed to generate training after retries', 500);
  }

  /// streams SSE events from POST /training, calls onStep for each step event,
  /// returns the training on "done" or null on "error".
  Future<Training?> _generateViaSSE(
    Map<String, dynamic> body, {
    void Function(String step)? onStep,
  }) async {
    await for (final event in _apiService.postSSE('/training', body: body)) {
      switch (event.event) {
        case 'step':
          final step = event.data['step'] as String?;
          if (step != null) {
            AppLogger.debug('[TrainingService] SSE step: $step');
            onStep?.call(step);
          }
        case 'done':
          final training = Training.fromJson(event.data);
          AppLogger.info('[TrainingService] Generated training: ${training.id}');
          return training;
        case 'error':
          AppLogger.error('[TrainingService] SSE error: ${event.data['error']}');
          return null;
      }
    }
    return null;
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
      emitEvent?.call(TrainingListChanged());
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
      body['quality'] = feedback.quality;
      body['qualityReason'] = feedback.qualityReason;
      body['message'] = feedback.message;
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
        emitEvent?.call(TrainingCompleted(trainingId));
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
      body['quality'] = feedback.quality;
      body['qualityReason'] = feedback.qualityReason;
      body['message'] = feedback.message;
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
        emitEvent?.call(FeedbackSubmitted(trainingId));
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

  Future<ApiResponse<TrainingFeedback?>> getUserFeedback(String trainingId) async {
    final response = await _apiService.get('/training/feedback/$trainingId');

    if (response.isSuccess && response.data != null) {
      try {
        return ApiResponse.success(TrainingFeedback.fromJson(response.data!), response.statusCode);
      } catch (e) {
        return ApiResponse.error('Failed to parse feedback', response.statusCode);
      }
    }
    // 404 = no feedback yet, not an error
    if (response.statusCode == 404) {
      return ApiResponse.success(null, response.statusCode);
    }
    return ApiResponse.error(response.error ?? 'Failed to fetch feedback', response.statusCode);
  }

  /// fetch linked health exercise session for a training (HR samples, avg/max, zones)
  Future<ApiResponse<Map<String, dynamic>?>> getHealthSession(String trainingId) async {
    final response = await _apiService.get('/health/session/$trainingId');
    if (response.isSuccess && response.data != null) {
      return ApiResponse.success(response.data, response.statusCode);
    }
    // 404 = no linked session
    if (response.statusCode == 404) {
      return ApiResponse.success(null, response.statusCode);
    }
    return ApiResponse.error(response.error ?? 'Failed to fetch health session', response.statusCode);
  }

  /// fetch last 7 days of health metrics + unlinked exercise sessions
  Future<ApiResponse<Map<String, dynamic>>> getHealthDaily() async {
    final response = await _apiService.get('/health/daily');
    if (response.isSuccess && response.data != null) {
      return ApiResponse.success(response.data!, response.statusCode);
    }
    return ApiResponse.error(response.error ?? 'Failed to fetch health daily', response.statusCode);
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
        emitEvent?.call(TrainingListChanged());
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

  Future<ApiResponse<List<PartnerInfo>>> getPartners(String trainingId) async {
    AppLogger.debug('[TrainingService] Fetching partners for training: $trainingId');

    final response = await _apiService.get('/training/partners/$trainingId');

    if (response.isSuccess && response.data != null) {
      try {
        final partnersJson = response.data!['partners'] as List;
        final partners = partnersJson.map((json) => PartnerInfo.fromJson(json)).toList();
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
        emitEvent?.call(TrainingListChanged());
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
        emitEvent?.call(TrainingListChanged());
        return ApiResponse.success(training, response.statusCode);
      } catch (e) {
        return ApiResponse.error('Failed to parse claimed training', response.statusCode);
      }
    }
    return ApiResponse.error(response.error ?? 'Failed to claim training', response.statusCode);
  }
}

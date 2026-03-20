import '../models/api_response.dart';
import '../models/flow_session.dart';
import 'app_event.dart';
import 'app_logger.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';

class FlowService {
  final AuthenticatedApiService _apiService;
  final void Function(AppEvent)? emitEvent;

  FlowService({
    AuthenticatedApiService? apiService,
    SecureStorageService? storageService,
    this.emitEvent,
  }) : _apiService = apiService ??
            AuthenticatedApiService(storageService: storageService);

  Future<ApiResponse<FlowSession>> generateFlow({
    required int duration,
    List<String>? muscles,
    String? prompt,
    void Function(int attempt)? onRetry,
  }) async {
    AppLogger.debug('[FlowService] Generating flow with duration: $duration');

    final body = <String, dynamic>{
      'duration': duration,
    };
    if (muscles != null && muscles.isNotEmpty) {
      body['muscles'] = muscles;
    }
    if (prompt != null && prompt.isNotEmpty) {
      body['prompt'] = prompt;
    }

    const maxRetries = 2;
    const retryableStatusCodes = [500, 502, 503, 504];
    for (var attempt = 0; attempt <= maxRetries; attempt++) {
      final response = await _apiService.post('/flow', body: body);

      if (retryableStatusCodes.contains(response.statusCode) && attempt < maxRetries) {
        AppLogger.warning('[FlowService] Got ${response.statusCode}, retrying (attempt ${attempt + 1})');
        onRetry?.call(attempt + 1);
        await Future.delayed(const Duration(seconds: 3));
        continue;
      }

      if (response.isSuccess && response.data != null) {
        try {
          final session = FlowSession.fromJson(response.data!);
          AppLogger.info('[FlowService] Generated flow: ${session.id}');
          emitEvent?.call(FlowSessionListChanged());
          return ApiResponse.success(session, response.statusCode);
        } catch (e) {
          AppLogger.error('[FlowService] failed to parse flow session', e);
          return ApiResponse.error('Failed to parse flow session', response.statusCode);
        }
      } else {
        AppLogger.error('[FlowService] Failed to generate flow: ${response.error}');
        return ApiResponse.error(
          response.error ?? 'Failed to generate flow',
          response.statusCode,
        );
      }
    }

    return ApiResponse.error('Failed to generate flow after retries', 500);
  }

  Future<ApiResponse<List<FlowSession>>> getFlowSessions() async {
    AppLogger.debug('[FlowService] Fetching flow sessions');

    final response = await _apiService.get('/flow');

    if (response.isSuccess && response.data != null) {
      try {
        final sessionsJson = response.data!['sessions'] as List;
        final sessions = sessionsJson.map((json) => FlowSession.fromJson(json)).toList();
        AppLogger.info('[FlowService] Fetched ${sessions.length} flow sessions');
        return ApiResponse.success(sessions, response.statusCode);
      } catch (e) {
        AppLogger.error('[FlowService] failed to parse flow sessions', e);
        return ApiResponse.error('Failed to parse flow sessions', response.statusCode);
      }
    } else {
      AppLogger.error('[FlowService] Failed to fetch flow sessions: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to fetch flow sessions',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<FlowSession>> completeFlow(String id) async {
    AppLogger.debug('[FlowService] Completing flow: $id');

    final response = await _apiService.post('/flow/complete/$id');

    if (response.isSuccess && response.data != null) {
      try {
        final session = FlowSession.fromJson(response.data!['session'] as Map<String, dynamic>);
        AppLogger.info('[FlowService] Completed flow: $id');
        emitEvent?.call(FlowSessionListChanged());
        return ApiResponse.success(session, response.statusCode);
      } catch (e) {
        AppLogger.error('[FlowService] failed to parse completed flow', e);
        return ApiResponse.error('Failed to parse completed flow', response.statusCode);
      }
    } else {
      AppLogger.error('[FlowService] Failed to complete flow: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to complete flow',
        response.statusCode,
      );
    }
  }

  Future<ApiResponse<String>> deleteFlow(String id) async {
    AppLogger.debug('[FlowService] Deleting flow: $id');

    final response = await _apiService.delete('/flow/$id');

    if (response.isSuccess) {
      final message = response.data?['message'] as String? ?? 'Flow deleted';
      AppLogger.info('[FlowService] Deleted flow: $id');
      emitEvent?.call(FlowSessionListChanged());
      return ApiResponse.success(message, response.statusCode);
    } else {
      AppLogger.error('[FlowService] Failed to delete flow: ${response.error}');
      return ApiResponse.error(
        response.error ?? 'Failed to delete flow',
        response.statusCode,
      );
    }
  }
}

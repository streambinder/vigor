import 'dart:async';

import 'package:flutter_test/flutter_test.dart';
import 'package:vigor/services/api_service.dart';
import 'package:vigor/services/authenticated_api_service.dart';
import 'package:vigor/services/training_service.dart';

/// An api service whose SSE stream hangs forever: the generation stays
/// in flight until the test closes the controller or the service
/// cancels its subscription.
class _HangingApiService extends AuthenticatedApiService {
  int sseCalls = 0;
  final StreamController<SSEEvent> controller = StreamController<SSEEvent>();

  @override
  Stream<SSEEvent> postSSE(String endpoint, {Map<String, dynamic>? body}) {
    sseCalls++;
    return controller.stream;
  }
}

void main() {
  test(
    'cancelGeneration aborts the in-flight generation without retrying',
    () async {
      final api = _HangingApiService();
      final service = TrainingService(apiService: api);

      final future = service.generateTraining(duration: 30);
      await Future<void>.delayed(Duration.zero);
      expect(api.sseCalls, 1);

      service.cancelGeneration();
      final response = await future;

      expect(response.isSuccess, isFalse);
      expect(response.statusCode, TrainingService.cancelledStatusCode);
      expect(response.error, 'Generation cancelled');
      // a user cancel is final: the retry loop must not fire another attempt
      expect(api.sseCalls, 1);

      await api.controller.close();
    },
  );

  test('cancelGeneration with nothing in flight is a no-op', () {
    final service = TrainingService(apiService: _HangingApiService());
    service.cancelGeneration();
  });

  test('a backend cancelled error event is not retried either', () async {
    final api = _HangingApiService();
    final service = TrainingService(apiService: api);

    final future = service.generateTraining(duration: 30);
    await Future<void>.delayed(Duration.zero);
    api.controller.add(
      SSEEvent('error', {'error': 'generation cancelled', 'code': 'cancelled'}),
    );
    final response = await future;

    expect(response.isSuccess, isFalse);
    expect(response.statusCode, TrainingService.cancelledStatusCode);
    expect(api.sseCalls, 1);

    await api.controller.close();
  });
}

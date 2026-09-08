import 'package:flutter_test/flutter_test.dart';
import 'package:vigor/services/training_service.dart';

void main() {
  test('absent code keeps the legacy always-retry behavior', () {
    expect(TrainingService.isRetryableErrorCode(null), isTrue);
    expect(TrainingService.isRetryableErrorCode(''), isTrue);
  });

  test('malformed training is retryable', () {
    expect(TrainingService.isRetryableErrorCode('malformed_training'), isTrue);
  });

  test('deterministic rejections are not retried', () {
    expect(
      TrainingService.isRetryableErrorCode('calibration_auto_only'),
      isFalse,
    );
    expect(TrainingService.isRetryableErrorCode('some_future_code'), isFalse);
  });

  test('non-retryable exception carries its message', () {
    const e = TrainingGenerationException(
      'non-Auto training generation is blocked during calibration',
      retryable: false,
    );
    expect(e.retryable, isFalse);
    expect(
      e.message,
      'non-Auto training generation is blocked during calibration',
    );
  });
}

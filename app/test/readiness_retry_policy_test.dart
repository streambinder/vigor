import 'package:flutter_test/flutter_test.dart';
import 'package:vigor/services/readiness_retry.dart';

void main() {
  test('delays double on every attempt', () {
    final policy = ReadinessRetryPolicy(maxAttempts: 5, baseDelay: const Duration(seconds: 30));

    expect(policy.nextDelay(), const Duration(seconds: 30));
    expect(policy.nextDelay(), const Duration(seconds: 60));
    expect(policy.nextDelay(), const Duration(seconds: 120));
    expect(policy.nextDelay(), const Duration(seconds: 240));
    expect(policy.nextDelay(), const Duration(seconds: 480));
  });

  test('returns null once the attempts are exhausted', () {
    final policy = ReadinessRetryPolicy(maxAttempts: 2, baseDelay: const Duration(seconds: 1));

    expect(policy.nextDelay(), isNotNull);
    expect(policy.nextDelay(), isNotNull);
    expect(policy.nextDelay(), isNull);
    expect(policy.nextDelay(), isNull);
  });

  test('reset restarts the delay sequence', () {
    final policy = ReadinessRetryPolicy(maxAttempts: 1, baseDelay: const Duration(seconds: 5));

    expect(policy.nextDelay(), const Duration(seconds: 5));
    expect(policy.nextDelay(), isNull);
    policy.reset();
    expect(policy.nextDelay(), const Duration(seconds: 5));
  });
}

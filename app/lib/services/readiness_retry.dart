/// Bounded retry policy for the daily readiness hint.
///
/// When the backend answers 404 because this morning's sleep has not synced
/// yet, the homepage schedules a few automatic retries with growing delays
/// instead of leaving the user with a manual pull-to-refresh as the only
/// way to see the hint.
class ReadinessRetryPolicy {
  /// Maximum number of automatic retries per homepage load.
  final int maxAttempts;

  /// Delay before the first retry; doubles on every subsequent attempt.
  final Duration baseDelay;

  int _attempts = 0;

  ReadinessRetryPolicy({this.maxAttempts = 5, this.baseDelay = const Duration(seconds: 30)});

  /// Delay before the next retry, or null when the attempts are exhausted.
  Duration? nextDelay() {
    if (_attempts >= maxAttempts) return null;
    final delay = baseDelay * (1 << _attempts);
    _attempts++;
    return delay;
  }

  /// Restart the attempt counter, e.g. on a new homepage load.
  void reset() => _attempts = 0;
}

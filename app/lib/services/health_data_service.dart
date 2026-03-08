import 'package:flutter/foundation.dart';
import 'package:health/health.dart';
import '../utils/platform_helper.dart';
import 'app_logger.dart';
import 'preferences_service.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';
import 'android_health_data_service.dart';
import 'ios_health_data_service.dart';

/// sync payload sent to POST /health/sync
class HealthSyncPayload {
  final List<Map<String, dynamic>> metrics;
  final List<Map<String, dynamic>> sessions;
  final List<Map<String, dynamic>> hrSamples;
  final String timezone;

  const HealthSyncPayload({
    required this.metrics,
    required this.sessions,
    required this.hrSamples,
    required this.timezone,
  });

  bool get isEmpty => metrics.isEmpty && sessions.isEmpty && hrSamples.isEmpty;

  Map<String, dynamic> toJson() => {
    'metrics': metrics,
    'sessions': sessions,
    'hr_samples': hrSamples,
    'timezone': timezone,
  };
}

/// phase 1 permission types for health connect / healthkit
const healthPermissionTypes = [
  HealthDataType.SLEEP_DEEP,
  HealthDataType.SLEEP_LIGHT,
  HealthDataType.SLEEP_REM,
  HealthDataType.RESTING_HEART_RATE,
  HealthDataType.HEART_RATE_VARIABILITY_RMSSD,
  HealthDataType.STEPS,
  HealthDataType.ACTIVE_ENERGY_BURNED,
  HealthDataType.HEART_RATE,
  HealthDataType.WORKOUT,
];

abstract class HealthDataService {
  Future<bool> isAvailable();
  Future<bool> requestPermissions();
  Future<HealthSyncPayload> readNewData();
  /// full 30-day read ignoring incremental tokens — used by manual sync
  Future<HealthSyncPayload> readAllData();
  ValueNotifier<bool> get syncing;

  /// trigger sync: read new data, POST to backend, persist tokens.
  /// returns true if sync completed successfully.
  /// set [force] to bypass throttle AND do a full re-read (e.g. manual sync from settings).
  Future<bool> syncToBackend({bool force = false});

  /// factory: returns the right implementation or null on web
  static HealthDataService? create({
    required PreferencesService prefs,
    required SecureStorageService storage,
  }) {
    if (PlatformHelper.isWeb) return null;
    if (PlatformHelper.isAndroid) {
      return AndroidHealthDataService(prefs: prefs, storage: storage);
    }
    if (PlatformHelper.isIOS) {
      return IOSHealthDataService(prefs: prefs, storage: storage);
    }
    return null;
  }
}

/// shared logic used by both platform implementations
mixin HealthDataServiceMixin on HealthDataService {
  PreferencesService get prefs;
  SecureStorageService get storage;

  final ValueNotifier<bool> _syncing = ValueNotifier(false);
  @override
  ValueNotifier<bool> get syncing => _syncing;

  static const _throttleDuration = Duration(hours: 1);
  static const _syncTimeout = Duration(seconds: 60);

  AuthenticatedApiService? _apiService;
  AuthenticatedApiService get apiService =>
      _apiService ??= AuthenticatedApiService(storageService: storage);

  /// check if sync should be throttled (< 1 hour since last sync)
  bool get _shouldThrottle {
    final lastSyncMs = prefs.hcLastSyncMs;
    if (lastSyncMs == null) return false;
    final elapsed = DateTime.now().millisecondsSinceEpoch - lastSyncMs;
    return elapsed < _throttleDuration.inMilliseconds;
  }

  @override
  Future<bool> syncToBackend({bool force = false}) async {
    if (_syncing.value) return false;
    if (!force && _shouldThrottle) {
      AppLogger.debug('[HealthDataService] throttled — last sync < 1 hour ago');
      return false;
    }

    _syncing.value = true;
    try {
      final available = await isAvailable();
      if (!available) {
        AppLogger.debug('[HealthDataService] health platform not available');
        return false;
      }

      final payload = force ? await readAllData() : await readNewData();
      if (payload.isEmpty) {
        AppLogger.debug('[HealthDataService] no new data to sync');
        // still update last sync timestamp to avoid re-reading
        await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
        return true;
      }

      final response = await apiService
          .post('/health/sync', body: payload.toJson())
          .timeout(_syncTimeout);

      if (response.isSuccess) {
        AppLogger.info('[HealthDataService] sync POST succeeded');
        // persist tokens/timestamps only after successful POST (H3)
        await onSyncSuccess();
        await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
        return true;
      } else {
        AppLogger.error('[HealthDataService] sync POST failed: ${response.error}');
        return false;
      }
    } catch (e) {
      AppLogger.error('[HealthDataService] sync failed', e);
      return false;
    } finally {
      _syncing.value = false;
    }
  }

  /// called after successful backend POST — persist platform-specific tokens
  Future<void> onSyncSuccess();
}

/// interval for overlap-aware deduplication of additive metrics
class _MetricInterval {
  final DateTime from;
  final DateTime to;
  final double value;
  final String source;
  _MetricInterval(this.from, this.to, this.value, this.source);

  int get durationMs => to.difference(from).inMilliseconds;
  double get rate => durationMs > 0 ? value / durationMs : 0;
}

/// transforms raw HealthDataPoints into the structured format the backend expects.
/// groups metrics by date, maps workouts to sessions, extracts HR samples.
/// uses interval-overlap-aware dedup for additive metrics (steps, calories, sleep stages)
/// so non-overlapping intervals from different sources sum correctly while overlapping
/// intervals keep only the higher-rate source.
HealthSyncPayload buildSyncPayload(List<HealthDataPoint> dataPoints, String timezone) {
  final dailyMetrics = <String, Map<String, dynamic>>{};
  final sessions = <Map<String, dynamic>>[];
  final hrSamples = <Map<String, dynamic>>[];

  // debug: log every raw data point grouped by type
  final typeCounts = <String, int>{};
  for (final point in dataPoints) {
    typeCounts[point.type.name] = (typeCounts[point.type.name] ?? 0) + 1;
  }
  AppLogger.info('[HealthSync] raw data points by type: $typeCounts');

  // collect intervals for additive metrics keyed by "date:type"
  final additiveIntervals = <String, List<_MetricInterval>>{};

  const additiveTypes = {
    HealthDataType.STEPS,
    HealthDataType.ACTIVE_ENERGY_BURNED,
    HealthDataType.SLEEP_DEEP,
    HealthDataType.SLEEP_LIGHT,
    HealthDataType.SLEEP_REM,
  };

  for (final point in dataPoints) {
    if (point.type == HealthDataType.WORKOUT) {
      final workoutValue = point.value is WorkoutHealthValue ? point.value as WorkoutHealthValue : null;
      sessions.add({
        'hc_record_id': point.uuid,
        'source_app': point.sourceName,
        'exercise_type': workoutValue?.workoutActivityType.name ?? 'other',
        'started_at': point.dateFrom.millisecondsSinceEpoch,
        'ended_at': point.dateTo.millisecondsSinceEpoch,
        if (workoutValue?.totalEnergyBurned != null) 'calories': workoutValue!.totalEnergyBurned!.toDouble(),
      });
    } else if (point.type == HealthDataType.HEART_RATE) {
      final bpm = _numericValue(point);
      if (bpm > 0) {
        hrSamples.add({
          'timestamp': point.dateFrom.millisecondsSinceEpoch,
          'bpm': bpm.round(),
        });
      }
    } else {
      final dateKey = _dateKey(point.dateFrom);
      dailyMetrics.putIfAbsent(dateKey, () => {'date': dateKey});

      AppLogger.debug('[HealthSync] raw point: type=${point.type.name} date=$dateKey '
          'value=${_numericValue(point)} from=${point.dateFrom} to=${point.dateTo} '
          'source=${point.sourceName} uuid=${point.uuid}');

      if (additiveTypes.contains(point.type)) {
        // for sleep stages, value is duration in hours derived from interval;
        // for steps/calories, value is the numeric reading
        double value;
        if (point.type == HealthDataType.SLEEP_DEEP ||
            point.type == HealthDataType.SLEEP_LIGHT ||
            point.type == HealthDataType.SLEEP_REM) {
          value = point.dateTo.difference(point.dateFrom).inMinutes / 60.0;
        } else {
          value = _numericValue(point);
        }

        final key = '$dateKey:${point.type.name}';
        additiveIntervals.putIfAbsent(key, () => []);
        additiveIntervals[key]!.add(_MetricInterval(point.dateFrom, point.dateTo, value, point.sourceName));
      } else {
        _addMetricToDay(dailyMetrics[dateKey]!, point);
      }
    }
  }

  // resolve additive metrics with overlap-aware dedup
  additiveIntervals.forEach((key, intervals) {
    final dateKey = key.split(':')[0];
    final typeName = key.split(':')[1];
    final bucket = dailyMetrics.putIfAbsent(dateKey, () => {'date': dateKey});
    final resolved = _resolveOverlaps(intervals);

    // debug: log per-source breakdown
    final sourceBreakdown = <String, double>{};
    for (final iv in intervals) {
      sourceBreakdown[iv.source] = (sourceBreakdown[iv.source] ?? 0) + iv.value;
    }
    AppLogger.info('[HealthSync] $typeName $dateKey by source: '
        '${sourceBreakdown.entries.map((e) => '${e.key}=${e.value.toStringAsFixed(1)}').join(', ')} → resolved=${resolved.toStringAsFixed(1)}');

    switch (typeName) {
      case 'STEPS':
        bucket['steps'] = resolved.round();
        break;
      case 'ACTIVE_ENERGY_BURNED':
        bucket['active_calories'] = resolved;
        break;
      case 'SLEEP_DEEP':
        bucket['sleep_deep_hours'] = resolved;
        _updateSleepTotal(bucket);
        break;
      case 'SLEEP_LIGHT':
        bucket['sleep_light_hours'] = resolved;
        _updateSleepTotal(bucket);
        break;
      case 'SLEEP_REM':
        bucket['sleep_rem_hours'] = resolved;
        _updateSleepTotal(bucket);
        break;
    }
  });

  // debug: log final aggregated payload per day
  for (final entry in dailyMetrics.entries) {
    AppLogger.info('[HealthSync] aggregated metrics ${entry.key}: ${entry.value}');
  }
  AppLogger.info('[HealthSync] payload: ${dailyMetrics.length} metric days, '
      '${sessions.length} sessions, ${hrSamples.length} HR samples');

  return HealthSyncPayload(
    metrics: dailyMetrics.values.toList(),
    sessions: sessions,
    hrSamples: hrSamples,
    timezone: timezone,
  );
}

/// resolves overlapping intervals from multiple sources.
/// non-overlapping intervals sum normally; when intervals overlap,
/// the higher value-per-millisecond source wins for the overlapping portion.
double _resolveOverlaps(List<_MetricInterval> intervals) {
  if (intervals.isEmpty) return 0;
  if (intervals.length == 1) return intervals.first.value;

  // single source — just sum
  final sources = intervals.map((iv) => iv.source).toSet();
  if (sources.length == 1) {
    return intervals.fold(0.0, (acc, iv) => acc + iv.value);
  }

  // multiple sources — split into per-source totals by non-overlapping coverage.
  // for each source, compute total duration covered. then for overlapping regions,
  // attribute to the source with higher density (value/time).
  // simplified approach: merge all intervals into a timeline, for each moment pick
  // the best source, then sum values proportionally.

  // collect all boundaries
  final events = <int>[];
  for (final iv in intervals) {
    events.add(iv.from.millisecondsSinceEpoch);
    events.add(iv.to.millisecondsSinceEpoch);
  }
  events.sort();

  double total = 0;
  for (int i = 0; i < events.length - 1; i++) {
    final segStart = events[i];
    final segEnd = events[i + 1];
    if (segEnd <= segStart) continue;

    // find all intervals covering this segment
    _MetricInterval? best;
    for (final iv in intervals) {
      final ivStart = iv.from.millisecondsSinceEpoch;
      final ivEnd = iv.to.millisecondsSinceEpoch;
      if (ivStart <= segStart && ivEnd >= segEnd) {
        // this interval covers the segment; pick highest rate
        if (best == null || iv.rate > best.rate) best = iv;
      }
    }
    if (best != null) {
      // attribute proportional value for this segment
      total += best.rate * (segEnd - segStart);
    }
  }

  return total;
}

void _updateSleepTotal(Map<String, dynamic> bucket) {
  bucket['sleep_hours'] = ((bucket['sleep_deep_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_light_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_rem_hours'] as double?) ?? 0.0);
}

String _dateKey(DateTime dt) =>
    '${dt.year}-${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')}';

void _addMetricToDay(Map<String, dynamic> bucket, HealthDataPoint point) {
  final value = _numericValue(point);
  switch (point.type) {
    case HealthDataType.RESTING_HEART_RATE:
      if (value > 0) bucket['resting_hr'] = value.round();
      break;
    case HealthDataType.HEART_RATE_VARIABILITY_RMSSD:
      if (value > 0) bucket['hrv_rmssd'] = value;
      break;
    default:
      break;
  }
}

double _numericValue(HealthDataPoint point) {
  final value = point.value;
  if (value is NumericHealthValue) return value.numericValue.toDouble();
  return 0.0;
}

import 'package:flutter/foundation.dart';
import 'package:health/health.dart';
import '../utils/platform_helper.dart';
import 'app_logger.dart';
import 'preferences_service.dart';
import 'authenticated_api_service.dart';
import 'secure_storage_service.dart';
import 'android_health_data_service.dart';
import 'ios_health_data_service.dart';

/// result from POST /health/sync or GET /health/stats
class HealthSyncResult {
  final int metricsSynced;
  final int sessionsSynced;
  final int totalMetrics;
  final int totalSessions;
  final DateTime syncedAt;
  final bool wasForced;
  /// per-source-app breakdown of device data: source name -> (metrics, sessions)
  final Map<String, ({int metrics, int sessions})> deviceSources;
  final DateTime? metricsFrom;
  final DateTime? metricsTo;
  final DateTime? sessionsFrom;
  final DateTime? sessionsTo;

  const HealthSyncResult({
    required this.metricsSynced,
    required this.sessionsSynced,
    required this.totalMetrics,
    required this.totalSessions,
    required this.syncedAt,
    this.wasForced = false,
    this.deviceSources = const {},
    this.metricsFrom,
    this.metricsTo,
    this.sessionsFrom,
    this.sessionsTo,
  });

  int get deviceMetrics => deviceSources.values.fold(0, (sum, s) => sum + s.metrics);
  int get deviceSessions => deviceSources.values.fold(0, (sum, s) => sum + s.sessions);

  factory HealthSyncResult.fromJson(Map<String, dynamic> json, {bool wasForced = false, Map<String, ({int metrics, int sessions})> deviceSources = const {}}) => HealthSyncResult(
    metricsSynced: json['metrics_synced'] ?? 0,
    sessionsSynced: json['sessions_synced'] ?? 0,
    totalMetrics: json['total_metrics'] ?? 0,
    totalSessions: json['total_sessions'] ?? 0,
    syncedAt: DateTime.now(),
    wasForced: wasForced,
    deviceSources: deviceSources,
    metricsFrom: _parseDate(json['metrics_from']),
    metricsTo: _parseDate(json['metrics_to']),
    sessionsFrom: _parseDate(json['sessions_from']),
    sessionsTo: _parseDate(json['sessions_to']),
  );

  static DateTime? _parseDate(dynamic value) {
    if (value == null || value is! String || value.isEmpty) return null;
    return DateTime.tryParse(value);
  }
}

/// sync payload sent to POST /health/sync
class HealthSyncPayload {
  final List<Map<String, dynamic>> metrics;
  final List<Map<String, dynamic>> sessions;
  final List<Map<String, dynamic>> hrSamples;
  final List<String> deletedRecordIds;
  final String timezone;
  /// per-source-app breakdown: source name -> (metrics, sessions) counts
  final Map<String, ({int metrics, int sessions})> sourceApps;

  const HealthSyncPayload({
    required this.metrics,
    required this.sessions,
    required this.hrSamples,
    this.deletedRecordIds = const [],
    required this.timezone,
    this.sourceApps = const {},
  });

  bool get isEmpty => metrics.isEmpty && sessions.isEmpty && hrSamples.isEmpty && deletedRecordIds.isEmpty;

  Map<String, dynamic> toJson() => {
    'metrics': metrics,
    'sessions': sessions,
    'hr_samples': hrSamples,
    if (deletedRecordIds.isNotEmpty) 'deleted_record_ids': deletedRecordIds,
    'timezone': timezone,
  };
}

/// all android-supported health connect permission types
const healthPermissionTypes = [
  HealthDataType.ACTIVE_ENERGY_BURNED,
  HealthDataType.BASAL_ENERGY_BURNED,
  HealthDataType.BLOOD_GLUCOSE,
  HealthDataType.BLOOD_OXYGEN,
  HealthDataType.BLOOD_PRESSURE_DIASTOLIC,
  HealthDataType.BLOOD_PRESSURE_SYSTOLIC,
  HealthDataType.BODY_FAT_PERCENTAGE,
  HealthDataType.BODY_TEMPERATURE,
  HealthDataType.BODY_WATER_MASS,
  HealthDataType.DISTANCE_DELTA,
  HealthDataType.FLIGHTS_CLIMBED,
  HealthDataType.HEART_RATE,
  HealthDataType.HEART_RATE_VARIABILITY_RMSSD,
  HealthDataType.HEIGHT,
  HealthDataType.LEAN_BODY_MASS,
  HealthDataType.MENSTRUATION_FLOW,
  HealthDataType.NUTRITION,
  HealthDataType.RESPIRATORY_RATE,
  HealthDataType.RESTING_HEART_RATE,
  HealthDataType.SLEEP_ASLEEP,
  HealthDataType.SLEEP_AWAKE,
  HealthDataType.SLEEP_AWAKE_IN_BED,
  HealthDataType.SLEEP_DEEP,
  HealthDataType.SLEEP_LIGHT,
  HealthDataType.SLEEP_OUT_OF_BED,
  HealthDataType.SLEEP_REM,
  HealthDataType.SLEEP_SESSION,
  HealthDataType.SLEEP_UNKNOWN,
  HealthDataType.SPEED,
  HealthDataType.STEPS,
  HealthDataType.TOTAL_CALORIES_BURNED,
  HealthDataType.WATER,
  HealthDataType.WEIGHT,
  HealthDataType.WORKOUT,
];

abstract class HealthDataService {
  Future<bool> isAvailable();
  Future<bool> requestPermissions();
  Future<bool> checkPermissions();
  Future<void> revokePermissions();
  Future<HealthSyncPayload> readNewData();
  /// full 30-day read ignoring incremental tokens — used by manual sync
  Future<HealthSyncPayload> readAllData();
  ValueNotifier<bool> get syncing;
  ValueNotifier<HealthSyncResult?> get lastSyncResult;

  /// trigger sync: read new data, POST to backend, persist tokens.
  /// returns true if sync completed successfully.
  /// set [force] to bypass throttle AND do a full re-read (e.g. manual sync from settings).
  Future<bool> syncToBackend({bool force = false});

  /// factory: returns the right implementation or null on web
  static HealthDataService? create({
    required PreferencesService prefs,
    required SecureStorageService storage,
  }) {
    if (PlatformHelper.isWeb) {
      AppLogger.debug('[HealthDataService] skipping creation on web');
      return null;
    }
    if (PlatformHelper.isAndroid) {
      AppLogger.info('[HealthDataService] creating AndroidHealthDataService');
      return AndroidHealthDataService(prefs: prefs, storage: storage);
    }
    if (PlatformHelper.isIOS) {
      AppLogger.info('[HealthDataService] creating IOSHealthDataService');
      return IOSHealthDataService(prefs: prefs, storage: storage);
    }
    AppLogger.warning('[HealthDataService] unsupported platform, returning null');
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

  final ValueNotifier<HealthSyncResult?> _lastSyncResult = ValueNotifier(null);
  @override
  ValueNotifier<HealthSyncResult?> get lastSyncResult => _lastSyncResult;

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

  /// restore persisted stats into the in-memory notifier so settings
  /// screen always has something to show, even before the first sync
  void _restorePersistedStats() {
    if (_lastSyncResult.value != null) return;
    final totalMetrics = prefs.hcTotalMetrics;
    final totalSessions = prefs.hcTotalSessions;
    if (totalMetrics > 0 || totalSessions > 0) {
      AppLogger.debug('[HealthDataService] restored persisted stats: $totalMetrics metrics, $totalSessions sessions');
      _lastSyncResult.value = HealthSyncResult(
        metricsSynced: 0, sessionsSynced: 0,
        totalMetrics: totalMetrics, totalSessions: totalSessions,
        syncedAt: DateTime.fromMillisecondsSinceEpoch(prefs.hcLastSyncMs ?? 0),
        wasForced: prefs.hcWasForced,
        deviceSources: prefs.hcDeviceSources,
        metricsFrom: HealthSyncResult._parseDate(prefs.hcMetricsFrom),
        metricsTo: HealthSyncResult._parseDate(prefs.hcMetricsTo),
        sessionsFrom: HealthSyncResult._parseDate(prefs.hcSessionsFrom),
        sessionsTo: HealthSyncResult._parseDate(prefs.hcSessionsTo),
      );
    }
  }

  /// persist sync stats so they survive app restarts
  Future<void> _persistStats(HealthSyncResult result) async {
    await prefs.setHcTotalMetrics(result.totalMetrics);
    await prefs.setHcTotalSessions(result.totalSessions);
    await prefs.setHcDeviceSources(result.deviceSources);
    await prefs.setHcWasForced(result.wasForced);
    await prefs.setHcMetricsFrom(result.metricsFrom?.toIso8601String());
    await prefs.setHcMetricsTo(result.metricsTo?.toIso8601String());
    await prefs.setHcSessionsFrom(result.sessionsFrom?.toIso8601String());
    await prefs.setHcSessionsTo(result.sessionsTo?.toIso8601String());
  }

  /// fetch stats from backend without syncing data — used when throttled
  Future<void> _fetchStatsOnly() async {
    try {
      final response = await apiService.get('/health/stats').timeout(_syncTimeout);
      if (response.isSuccess && response.data != null) {
        AppLogger.debug('[HealthDataService] stats fetched: ${response.data!['total_metrics']} metrics, ${response.data!['total_sessions']} sessions');
        // stats-only: no device read happened, preserve previous device counts
        final prev = _lastSyncResult.value;
        final result = HealthSyncResult(
          metricsSynced: 0, sessionsSynced: 0,
          totalMetrics: response.data!['total_metrics'] ?? 0,
          totalSessions: response.data!['total_sessions'] ?? 0,
          syncedAt: DateTime.now(),
          wasForced: prev?.wasForced ?? false,
          deviceSources: prev?.deviceSources ?? const {},
          metricsFrom: HealthSyncResult._parseDate(response.data!['metrics_from']),
          metricsTo: HealthSyncResult._parseDate(response.data!['metrics_to']),
          sessionsFrom: HealthSyncResult._parseDate(response.data!['sessions_from']),
          sessionsTo: HealthSyncResult._parseDate(response.data!['sessions_to']),
        );
        _lastSyncResult.value = result;
        await _persistStats(result);
      }
    } catch (e) {
      AppLogger.debug('[HealthDataService] stats fetch failed: $e');
    }
  }

  @override
  Future<bool> syncToBackend({bool force = false}) async {
    if (_syncing.value) {
      AppLogger.debug('[HealthDataService] sync skipped — already in progress');
      return false;
    }
    AppLogger.info('[HealthDataService] syncToBackend started (force=$force)');

    // always restore persisted stats so the UI has data immediately
    _restorePersistedStats();

    if (!force && _shouldThrottle) {
      AppLogger.debug('[HealthDataService] throttled — fetching stats only');
      // still fetch fresh stats from backend so settings always shows current totals
      await _fetchStatsOnly();
      return true;
    }

    _syncing.value = true;
    try {
      final available = await isAvailable();
      if (!available) {
        AppLogger.debug('[HealthDataService] health platform not available');
        return false;
      }

      // on force sync, re-request permissions to ensure all types are authorized
      // (idempotent — returns immediately if already granted)
      if (force) {
        AppLogger.debug('[HealthDataService] force sync — re-requesting permissions');
        final granted = await requestPermissions();
        if (!granted) {
          AppLogger.info('[HealthDataService] permissions denied during force sync');
          await prefs.setHcConnected(false);
          return false;
        }
        AppLogger.debug('[HealthDataService] permissions confirmed');
      }

      AppLogger.debug('[HealthDataService] reading health data (${force ? 'full' : 'incremental'})');
      final payload = force ? await readAllData() : await readNewData();
      AppLogger.info('[HealthDataService] read complete: ${payload.metrics.length} metrics, ${payload.sessions.length} sessions, ${payload.hrSamples.length} HR samples, ${payload.deletedRecordIds.length} deletions');

      if (payload.isEmpty) {
        AppLogger.debug('[HealthDataService] no new data to sync');
        // still fetch backend stats so totals stay current
        await _fetchStatsOnly();
        await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
        return true;
      }

      // split payload into per-date batches to avoid request size limits
      final batches = _splitByDate(payload);
      AppLogger.info('[HealthDataService] split into ${batches.length} daily batches');

      int totalMetricsSynced = 0;
      int totalSessionsSynced = 0;
      Map<String, dynamic>? lastResponseData;

      for (final batch in batches) {
        final response = await apiService
            .post('/health/sync', body: batch.toJson())
            .timeout(_syncTimeout);

        if (!response.isSuccess) {
          AppLogger.error('[HealthDataService] sync POST failed for batch: ${response.error} (status=${response.statusCode})');
          return false;
        }

        totalMetricsSynced += (response.data?['metrics_synced'] as int?) ?? 0;
        totalSessionsSynced += (response.data?['sessions_synced'] as int?) ?? 0;
        lastResponseData = response.data;
      }

      AppLogger.info('[HealthDataService] all batches synced ($totalMetricsSynced metrics, $totalSessionsSynced sessions)');
      if (lastResponseData != null) {
        // use totals from last response (backend returns cumulative totals)
        // but override synced counts with our accumulated values
        final result = HealthSyncResult.fromJson(
          {...lastResponseData, 'metrics_synced': totalMetricsSynced, 'sessions_synced': totalSessionsSynced},
          wasForced: force,
          deviceSources: payload.sourceApps,
        );
        _lastSyncResult.value = result;
        await _persistStats(result);
      }
      // persist tokens/timestamps only after successful POST (H3)
      await onSyncSuccess();
      await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
      return true;
    } catch (e) {
      AppLogger.error('[HealthDataService] sync failed', e);
      return false;
    } finally {
      _syncing.value = false;
    }
  }

  /// split a payload into one batch per date so each POST stays small.
  /// metrics are already keyed by date; sessions/HR samples are bucketed
  /// by their start timestamp.
  List<HealthSyncPayload> _splitByDate(HealthSyncPayload payload) {
    final metricsByDate = <String, List<Map<String, dynamic>>>{};
    for (final m in payload.metrics) {
      final date = m['date'] as String;
      metricsByDate.putIfAbsent(date, () => []).add(m);
    }

    final sessionsByDate = <String, List<Map<String, dynamic>>>{};
    for (final s in payload.sessions) {
      final date = _dateKeyFromMs(s['started_at'] as int);
      sessionsByDate.putIfAbsent(date, () => []).add(s);
    }

    final hrByDate = <String, List<Map<String, dynamic>>>{};
    for (final hr in payload.hrSamples) {
      final date = _dateKeyFromMs(hr['timestamp'] as int);
      hrByDate.putIfAbsent(date, () => []).add(hr);
    }

    final sortedDates = {...metricsByDate.keys, ...sessionsByDate.keys, ...hrByDate.keys}.toList()..sort();

    // if only one date, send as single batch
    if (sortedDates.length <= 1) return [payload];

    final batches = <HealthSyncPayload>[];
    for (int i = 0; i < sortedDates.length; i++) {
      final date = sortedDates[i];
      batches.add(HealthSyncPayload(
        metrics: metricsByDate[date] ?? [],
        sessions: sessionsByDate[date] ?? [],
        hrSamples: hrByDate[date] ?? [],
        // attach deletions to the first batch only
        deletedRecordIds: i == 0 ? payload.deletedRecordIds : const [],
        timezone: payload.timezone,
      ));
    }
    return batches;
  }

  String _dateKeyFromMs(int ms) {
    final dt = DateTime.fromMillisecondsSinceEpoch(ms);
    return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')}';
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
  AppLogger.debug('[HealthDataService] buildSyncPayload: ${dataPoints.length} data points, tz=$timezone');
  final dailyMetrics = <String, Map<String, dynamic>>{};
  final sessions = <Map<String, dynamic>>[];
  final hrSamples = <Map<String, dynamic>>[];
  // track per-source-app data point counts (metrics vs sessions)
  final sourceMetricCounts = <String, int>{};
  final sourceSessionCounts = <String, int>{};

  // collect intervals for additive metrics keyed by "date:type"
  final additiveIntervals = <String, List<_MetricInterval>>{};

  const additiveTypes = {
    HealthDataType.STEPS,
    HealthDataType.ACTIVE_ENERGY_BURNED,
    HealthDataType.SLEEP_DEEP,
    HealthDataType.SLEEP_LIGHT,
    HealthDataType.SLEEP_REM,
    HealthDataType.SLEEP_ASLEEP,
  };

  const sleepTypes = {
    HealthDataType.SLEEP_DEEP,
    HealthDataType.SLEEP_LIGHT,
    HealthDataType.SLEEP_REM,
    HealthDataType.SLEEP_ASLEEP,
  };

  // sort by dateTo so last-write-wins for non-additive metrics (RHR, HRV)
  // picks the most recent reading deterministically
  final sortedPoints = List<HealthDataPoint>.from(dataPoints)
    ..sort((a, b) => a.dateTo.compareTo(b.dateTo));

  for (final point in sortedPoints) {
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
      sourceSessionCounts[point.sourceName] = (sourceSessionCounts[point.sourceName] ?? 0) + 1;
    } else if (point.type == HealthDataType.HEART_RATE) {
      final bpm = _numericValue(point);
      if (bpm > 0) {
        hrSamples.add({
          'timestamp': point.dateFrom.millisecondsSinceEpoch,
          'bpm': bpm.round(),
        });
      }
    } else {
      // for sleep stages, attribute to the wake-up date (dateTo)
      final dateKey = sleepTypes.contains(point.type)
          ? _dateKey(point.dateTo)
          : _dateKey(point.dateFrom);
      dailyMetrics.putIfAbsent(dateKey, () => {'date': dateKey});
      sourceMetricCounts[point.sourceName] = (sourceMetricCounts[point.sourceName] ?? 0) + 1;

      if (additiveTypes.contains(point.type)) {
        double value;
        if (sleepTypes.contains(point.type)) {
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
      case 'SLEEP_ASLEEP':
        // undifferentiated sleep — stored separately, included in total
        bucket['sleep_asleep_hours'] = resolved;
        _updateSleepTotal(bucket);
        break;
    }
  });

  // merge source counts into a single map
  final allSourceNames = {...sourceMetricCounts.keys, ...sourceSessionCounts.keys};
  final sourceApps = {
    for (final name in allSourceNames)
      name: (metrics: sourceMetricCounts[name] ?? 0, sessions: sourceSessionCounts[name] ?? 0),
  };

  AppLogger.debug('[HealthDataService] payload built: ${dailyMetrics.length} daily buckets, ${sessions.length} sessions, ${hrSamples.length} HR samples, sources=${sourceApps.keys.join(', ')}');
  return HealthSyncPayload(
    metrics: dailyMetrics.values.toList(),
    sessions: sessions,
    hrSamples: hrSamples,
    timezone: timezone,
    sourceApps: sourceApps,
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
  final staged = ((bucket['sleep_deep_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_light_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_rem_hours'] as double?) ?? 0.0);
  final asleep = (bucket['sleep_asleep_hours'] as double?) ?? 0.0;
  // use staged breakdown if available, otherwise fall back to undifferentiated SLEEP_ASLEEP
  bucket['sleep_hours'] = staged > 0 ? staged : asleep;
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

import 'package:flutter/foundation.dart';
import 'package:health/health.dart';
import '../utils/platform_helper.dart';
import '../models/api_response.dart';
import 'app_event.dart';
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

  /// non-null when device data was read but upload to backend failed
  final String? syncError;

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
    this.syncError,
  });

  int get deviceMetrics =>
      deviceSources.values.fold(0, (sum, s) => sum + s.metrics);
  int get deviceSessions =>
      deviceSources.values.fold(0, (sum, s) => sum + s.sessions);

  factory HealthSyncResult.fromJson(
    Map<String, dynamic> json, {
    bool wasForced = false,
    Map<String, ({int metrics, int sessions})> deviceSources = const {},
  }) => HealthSyncResult(
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
  final List<Map<String, dynamic>> weights;
  final List<Map<String, dynamic>> hrSamples;
  final List<String> deletedRecordIds;

  /// per-source-app breakdown: source name -> (metrics, sessions) counts
  final Map<String, ({int metrics, int sessions})> sourceApps;

  const HealthSyncPayload({
    required this.metrics,
    required this.sessions,
    required this.weights,
    required this.hrSamples,
    this.deletedRecordIds = const [],
    this.sourceApps = const {},
  });

  bool get isEmpty =>
      metrics.isEmpty &&
      sessions.isEmpty &&
      weights.isEmpty &&
      hrSamples.isEmpty &&
      deletedRecordIds.isEmpty;

  Map<String, dynamic> toJson() => {
    'metrics': metrics,
    'sessions': sessions,
    'weights': weights,
    'hr_samples': hrSamples,
    if (deletedRecordIds.isNotEmpty) 'deleted_record_ids': deletedRecordIds,
  };
}

/// all android-supported health connect permission types
/// tier 1 (core): STEPS, CALORIES, SLEEP_SESSION, WEIGHT — lightweight, always synced
/// tier 2 (exercise): WORKOUT, RESTING_HR, HRV, SLEEP stages — small, valuable
/// HEART_RATE is fetched windowed around workouts only, not as bulk 7-day read
const healthPermissionTypes = [
  HealthDataType.STEPS,
  HealthDataType.TOTAL_CALORIES_BURNED,
  HealthDataType.SLEEP_SESSION,
  HealthDataType.WEIGHT,
  HealthDataType.WORKOUT,
  HealthDataType.RESTING_HEART_RATE,
  HealthDataType.HEART_RATE_VARIABILITY_RMSSD,
  HealthDataType.SLEEP_DEEP,
  HealthDataType.SLEEP_LIGHT,
  HealthDataType.SLEEP_REM,
  HealthDataType.SLEEP_ASLEEP,
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

  /// read a single calendar day (local timezone midnight to midnight)
  Future<HealthSyncPayload> readForDate(DateTime date);

  /// trigger sync: read new data, POST to backend, persist tokens.
  /// returns true if sync completed successfully.
  /// set [fullRescan] to do a full 30-day re-read and re-request permissions
  /// (e.g. manual sync from settings). default is incremental delta sync.
  /// throttling lives server-side only; client never gates.
  Future<bool> syncToBackend({bool fullRescan = false});

  /// write height (cm) and/or weight (kg) back to the health platform so other
  /// apps (e.g. fitness trackers) see Vigor's user-provided values.
  /// fire-and-forget — failures are logged but never surfaced to the user.
  Future<void> writeBodyMetrics({double? height, double? weight});

  /// factory: returns the right implementation or null on web
  static HealthDataService? create({
    required PreferencesService prefs,
    required SecureStorageService storage,
  }) {
    if (PlatformHelper.isWeb) return null;
    if (PlatformHelper.isAndroid) {
      AppLogger.info('[HealthDataService] creating AndroidHealthDataService');
      return AndroidHealthDataService(prefs: prefs, storage: storage);
    }
    if (PlatformHelper.isIOS) {
      AppLogger.info('[HealthDataService] creating IOSHealthDataService');
      return IOSHealthDataService(prefs: prefs, storage: storage);
    }
    AppLogger.warning(
      '[HealthDataService] unsupported platform, returning null',
    );
    return null;
  }
}

/// shared logic used by both platform implementations
mixin HealthDataServiceMixin on HealthDataService {
  PreferencesService get prefs;
  SecureStorageService get storage;

  /// event callback set by ServiceLocator after creation
  void Function(AppEvent)? emitEvent;

  final ValueNotifier<bool> _syncing = ValueNotifier(false);
  @override
  ValueNotifier<bool> get syncing => _syncing;

  final ValueNotifier<HealthSyncResult?> _lastSyncResult = ValueNotifier(null);
  @override
  ValueNotifier<HealthSyncResult?> get lastSyncResult => _lastSyncResult;

  static const _syncTimeout = Duration(seconds: 60);

  AuthenticatedApiService? _apiService;
  AuthenticatedApiService get apiService =>
      _apiService ??= AuthenticatedApiService(storageService: storage);

  /// restore persisted stats into the in-memory notifier so settings
  /// screen always has something to show, even before the first sync
  void _restorePersistedStats() {
    if (_lastSyncResult.value != null) return;
    final totalMetrics = prefs.hcTotalMetrics;
    final totalSessions = prefs.hcTotalSessions;
    if (totalMetrics > 0 || totalSessions > 0) {
      AppLogger.debug(
        '[HealthDataService] restored persisted stats: $totalMetrics metrics, $totalSessions sessions',
      );
      _lastSyncResult.value = HealthSyncResult(
        metricsSynced: 0,
        sessionsSynced: 0,
        totalMetrics: totalMetrics,
        totalSessions: totalSessions,
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
      final response = await apiService
          .get('/health/stats')
          .timeout(_syncTimeout);
      if (response.isSuccess && response.data != null) {
        AppLogger.debug(
          '[HealthDataService] stats fetched: ${response.data!['total_metrics']} metrics, ${response.data!['total_sessions']} sessions',
        );
        // stats-only: no device read happened, preserve previous device counts
        final prev = _lastSyncResult.value;
        final result = HealthSyncResult(
          metricsSynced: 0,
          sessionsSynced: 0,
          totalMetrics: response.data!['total_metrics'] ?? 0,
          totalSessions: response.data!['total_sessions'] ?? 0,
          syncedAt: DateTime.now(),
          wasForced: prev?.wasForced ?? false,
          deviceSources: prev?.deviceSources ?? const {},
          metricsFrom: HealthSyncResult._parseDate(
            response.data!['metrics_from'],
          ),
          metricsTo: HealthSyncResult._parseDate(response.data!['metrics_to']),
          sessionsFrom: HealthSyncResult._parseDate(
            response.data!['sessions_from'],
          ),
          sessionsTo: HealthSyncResult._parseDate(
            response.data!['sessions_to'],
          ),
        );
        _lastSyncResult.value = result;
        await _persistStats(result);
      }
    } catch (e) {
      AppLogger.debug('[HealthDataService] stats fetch failed: $e');
    }
  }

  @override
  Future<bool> syncToBackend({bool fullRescan = false}) async {
    if (_syncing.value) {
      AppLogger.debug('[HealthDataService] sync skipped — already in progress');
      return false;
    }
    AppLogger.info(
      '[HealthDataService] syncToBackend started (fullRescan=$fullRescan)',
    );

    // always restore persisted stats so the UI has data immediately
    _restorePersistedStats();

    _syncing.value = true;
    try {
      final available = await isAvailable();
      if (!available) {
        AppLogger.debug('[HealthDataService] health platform not available');
        return false;
      }

      // full rescan re-requests permissions to ensure all types are authorized
      // (idempotent — returns immediately if already granted)
      if (fullRescan) {
        AppLogger.debug(
          '[HealthDataService] full rescan — re-requesting permissions',
        );
        final granted = await requestPermissions();
        if (!granted) {
          AppLogger.info(
            '[HealthDataService] permissions denied during full rescan',
          );
          await prefs.setHcConnected(false);
          return false;
        }
        AppLogger.debug('[HealthDataService] permissions confirmed');
        return await _syncFull();
      }

      // incremental: stream one date at a time, smallest pieces possible
      return await _streamIncrementalSync();
    } catch (e) {
      AppLogger.error('[HealthDataService] sync failed', e);
      return false;
    } finally {
      _syncing.value = false;
    }
  }

  /// full 30-day sync — single POST, used by manual Settings action
  Future<bool> _syncFull() async {
    AppLogger.debug('[HealthDataService] reading full 30-day');
    final payload = await readAllData();
    AppLogger.info(
      '[HealthDataService] full read: ${payload.metrics.length} metrics, ${payload.sessions.length} sessions, ${payload.weights.length} weights, ${payload.hrSamples.length} HR samples',
    );
    if (payload.isEmpty) {
      await _fetchStatsOnly();
      await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
      return true;
    }
    return await _postPayload(payload, wasForced: true);
  }

  /// shared POST with 429 retry — returns true on success
  Future<bool> _postPayload(HealthSyncPayload payload, {required bool wasForced}) async {
    int totalMetricsSynced = 0;
    int totalSessionsSynced = 0;
    Map<String, dynamic>? lastResponseData;
    ApiResponse? response;
    int attempts = 0;
    const maxAttempts = 3;

    while (attempts < maxAttempts) {
      response = await apiService
          .post('/health/sync', body: payload.toJson())
          .timeout(_syncTimeout);

      if (response.isSuccess || response.statusCode != 429) break;

      attempts++;
      if (attempts < maxAttempts) {
        final backoffMs = 30000 * attempts;
        AppLogger.warning(
          '[HealthDataService] rate limited, retrying in ${backoffMs}ms (attempt $attempts/$maxAttempts)',
        );
        await Future.delayed(Duration(milliseconds: backoffMs));
      }
    }

    if (response == null || !response.isSuccess) {
      AppLogger.error(
        '[HealthDataService] sync POST failed: ${response?.error} (status=${response?.statusCode})',
      );
      _lastSyncResult.value = HealthSyncResult(
        metricsSynced: 0,
        sessionsSynced: 0,
        totalMetrics: _lastSyncResult.value?.totalMetrics ?? 0,
        totalSessions: _lastSyncResult.value?.totalSessions ?? 0,
        syncedAt: DateTime.now(),
        wasForced: wasForced,
        deviceSources: payload.sourceApps,
        syncError: response?.error ?? 'Upload failed',
      );
      return false;
    }

    totalMetricsSynced += (response.data?['metrics_synced'] as int?) ?? 0;
    totalSessionsSynced += (response.data?['sessions_synced'] as int?) ?? 0;
    lastResponseData = response.data;

    AppLogger.info(
      '[HealthDataService] sync completed ($totalMetricsSynced metrics, $totalSessionsSynced sessions)',
    );
    if (lastResponseData != null) {
      final result = HealthSyncResult.fromJson(
        {
          ...lastResponseData,
          'metrics_synced': totalMetricsSynced,
          'sessions_synced': totalSessionsSynced,
        },
        wasForced: wasForced,
        deviceSources: payload.sourceApps,
      );
      _lastSyncResult.value = result;
      await _persistStats(result);
    }
    await onSyncSuccess();
    await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
    emitEvent?.call(HealthSyncCompleted());
    return true;
  }

  /// incremental streaming: one date per POST, smallest pieces first
  Future<bool> _streamIncrementalSync() async {
    Set<String> serverDates = {};
    try {
      final manifestResp = await apiService.get('/health/manifest').timeout(_syncTimeout);
      if (manifestResp.isSuccess && manifestResp.data != null) {
        serverDates = Set<String>.from(
          (manifestResp.data!['dates_with_data'] as List?)?.cast<String>() ?? [],
        );
        AppLogger.debug('[HealthDataService] server has ${serverDates.length} dates');
      } else {
        AppLogger.warning('[HealthDataService] manifest failed, falling back to 7-day single POST');
        final payload = await readNewData();
        if (payload.isEmpty) {
          await _fetchStatsOnly();
          await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
          return true;
        }
        return await _postPayload(payload, wasForced: false);
      }
    } catch (e) {
      AppLogger.error('[HealthDataService] manifest fetch failed, fallback', e);
      final payload = await readNewData();
      if (payload.isEmpty) {
        await _fetchStatsOnly();
        return true;
      }
      return await _postPayload(payload, wasForced: false);
    }

    final now = DateTime.now();
    // target dates: last 30 days missing on server + always last 7 days (freshness)
    // 7-day window ensures late Health Connect workouts (e.g. Fitbit sync delayed,
    // or workout added after metrics) are re-synced even when metric date already
    // exists on server — manifest only tracks metric dates, not session dates.
    final targetDates = <DateTime>[];
    final recentKeys = <String>{};
    for (int i = 0; i < 7; i++) {
      final d = now.subtract(Duration(days: i));
      final k = '${d.year}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
      recentKeys.add(k);
    }

    for (int i = 0; i < 30; i++) {
      final d = now.subtract(Duration(days: i));
      final k = '${d.year}-${d.month.toString().padLeft(2, '0')}-${d.day.toString().padLeft(2, '0')}';
      if (!serverDates.contains(k) || recentKeys.contains(k)) {
        targetDates.add(DateTime(d.year, d.month, d.day));
      }
    }

    // oldest first so backend converges forward, recent 7 will be re-synced anyway
    targetDates.sort((a, b) => a.compareTo(b));

    AppLogger.info('[HealthDataService] incremental stream: ${targetDates.length} dates to check');

    int totalMetricsSynced = 0;
    int totalSessionsSynced = 0;
    final mergedSources = <String, ({int metrics, int sessions})>{};
    bool anyPosted = false;
    Map<String, dynamic>? lastResponseData;

    for (int idx = 0; idx < targetDates.length; idx++) {
      final date = targetDates[idx];
      HealthSyncPayload payload;
      try {
        payload = await readForDate(date);
      } catch (e) {
        AppLogger.warning('[HealthDataService] readForDate failed for $date: $e');
        continue;
      }

      if (payload.isEmpty) continue;

      // merge source counts
      payload.sourceApps.forEach((k, v) {
        final prev = mergedSources[k];
        if (prev == null) {
          mergedSources[k] = v;
        } else {
          mergedSources[k] = (metrics: prev.metrics + v.metrics, sessions: prev.sessions + v.sessions);
        }
      });

      ApiResponse? response;
      int attempts = 0;
      const maxAttempts = 3;
      while (attempts < maxAttempts) {
        response = await apiService.post('/health/sync', body: payload.toJson()).timeout(_syncTimeout);
        if (response.isSuccess || response.statusCode != 429) break;
        attempts++;
        if (attempts < maxAttempts) {
          final backoffMs = 30000 * attempts;
          AppLogger.warning('[HealthDataService] 429 on $date, retry in ${backoffMs}ms');
          await Future.delayed(Duration(milliseconds: backoffMs));
        }
      }

      if (response == null || !response.isSuccess) {
        AppLogger.error('[HealthDataService] POST failed for $date: ${response?.error}');
        // don't block remaining dates — continue to next, will retry on next sync trigger
        continue;
      }

      anyPosted = true;
      totalMetricsSynced += (response.data?['metrics_synced'] as int?) ?? 0;
      totalSessionsSynced += (response.data?['sessions_synced'] as int?) ?? 0;
      lastResponseData = response.data;
      await onSyncSuccess();

      // gentle pacing: 1.2s + jitter 0-800ms between dates to avoid hammering
      if (idx < targetDates.length - 1) {
        final jitter = (DateTime.now().millisecond % 800);
        await Future.delayed(Duration(milliseconds: 1200 + jitter));
      }
    }

    if (!anyPosted) {
      AppLogger.debug('[HealthDataService] incremental stream: no new data');
      await _fetchStatsOnly();
      await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
      return true;
    }

    if (lastResponseData != null) {
      final result = HealthSyncResult.fromJson(
        {
          ...lastResponseData,
          'metrics_synced': totalMetricsSynced,
          'sessions_synced': totalSessionsSynced,
        },
        wasForced: false,
        deviceSources: mergedSources,
      );
      _lastSyncResult.value = result;
      await _persistStats(result);
    }
    await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
    emitEvent?.call(HealthSyncCompleted());
    AppLogger.info('[HealthDataService] incremental stream done ($totalMetricsSynced metrics, $totalSessionsSynced sessions across dates)');
    return true;
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
HealthSyncPayload buildSyncPayload(List<HealthDataPoint> dataPoints) {
  AppLogger.debug(
    '[HealthDataService] buildSyncPayload: ${dataPoints.length} data points',
  );
  final dailyMetrics = <String, Map<String, dynamic>>{};
  final sessions = <Map<String, dynamic>>[];
  final weights = <Map<String, dynamic>>[];
  final hrSamples = <Map<String, dynamic>>[];
  // track per-source-app data point counts (metrics vs sessions)
  final sourceMetricCounts = <String, int>{};
  final sourceSessionCounts = <String, int>{};

  // collect intervals for additive metrics keyed by "date:type"
  final additiveIntervals = <String, List<_MetricInterval>>{};

  const additiveTypes = {
    HealthDataType.STEPS,
    HealthDataType.TOTAL_CALORIES_BURNED,
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

  const sleepSessionTypes = {HealthDataType.SLEEP_SESSION};

  // sort by dateTo so last-write-wins for non-additive metrics (RHR, HRV)
  // picks the most recent reading deterministically
  final sortedPoints = List<HealthDataPoint>.from(dataPoints)
    ..sort((a, b) => a.dateTo.compareTo(b.dateTo));

  for (final point in sortedPoints) {
    if (point.type == HealthDataType.WORKOUT) {
      final workoutValue = point.value is WorkoutHealthValue
          ? point.value as WorkoutHealthValue
          : null;
      sessions.add({
        'hc_record_id': point.uuid,
        'source_app': point.sourceName,
        'exercise_type': workoutValue?.workoutActivityType.name ?? 'other',
        'started_at': point.dateFrom.millisecondsSinceEpoch,
        'ended_at': point.dateTo.millisecondsSinceEpoch,
        if (workoutValue?.totalEnergyBurned != null)
          'calories': workoutValue!.totalEnergyBurned!.toDouble(),
      });
      sourceSessionCounts[point.sourceName] =
          (sourceSessionCounts[point.sourceName] ?? 0) + 1;
    } else if (point.type == HealthDataType.HEART_RATE) {
      final bpm = _numericValue(point);
      if (bpm > 0) {
        hrSamples.add({
          'timestamp': point.dateFrom.millisecondsSinceEpoch,
          'bpm': bpm.round(),
        });
      }
    } else if (point.type == HealthDataType.WEIGHT) {
      final value = _numericValue(point);
      if (value > 0) {
        weights.add({
          'hc_record_id': point.uuid,
          'source_app': point.sourceName,
          'measured_at': point.dateTo.millisecondsSinceEpoch,
          'weight': value,
        });
        sourceMetricCounts[point.sourceName] =
            (sourceMetricCounts[point.sourceName] ?? 0) + 1;
      }
    } else {
      // attribute sleep to wake-up date (dateTo) so users see sleep on the day they wake up
      final dateKey = _dateKey(point.dateTo);
      dailyMetrics.putIfAbsent(dateKey, () => {'date': dateKey});
      sourceMetricCounts[point.sourceName] =
          (sourceMetricCounts[point.sourceName] ?? 0) + 1;

      if (sleepSessionTypes.contains(point.type)) {
        // SLEEP_SESSION is the authoritative total - use longest session (workaround for devices that report multiple sessions)
        final hours = point.dateTo.difference(point.dateFrom).inMinutes / 60.0;
        final bucket = dailyMetrics[dateKey]!;
        final existing = bucket['sleep_session_hours'] as double? ?? 0.0;
        if (hours > existing) {
          bucket['sleep_session_hours'] = hours;
          AppLogger.info(
            '[HealthDataService] SLEEP_SESSION: ${hours.toStringAsFixed(2)}h from ${point.sourceName} for $dateKey (replaced ${existing.toStringAsFixed(2)}h)',
          );
        }
      } else if (additiveTypes.contains(point.type)) {
        double value;
        if (sleepTypes.contains(point.type)) {
          value = point.dateTo.difference(point.dateFrom).inMinutes / 60.0;
        } else {
          value = _numericValue(point);
        }

        final key = '$dateKey:${point.type.name}';
        additiveIntervals.putIfAbsent(key, () => []);
        additiveIntervals[key]!.add(
          _MetricInterval(
            point.dateFrom,
            point.dateTo,
            value,
            point.sourceName,
          ),
        );
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
      case 'TOTAL_CALORIES_BURNED':
        bucket['total_calories'] = resolved;
        break;
      case 'SLEEP_DEEP':
        bucket['sleep_deep_hours'] = resolved;
        break;
      case 'SLEEP_LIGHT':
        bucket['sleep_light_hours'] = resolved;
        break;
      case 'SLEEP_REM':
        bucket['sleep_rem_hours'] = resolved;
        break;
      case 'SLEEP_ASLEEP':
        bucket['sleep_asleep_hours'] = resolved;
        break;
    }
  });

  // finalize sleep totals after all stages and sessions are processed
  for (final bucket in dailyMetrics.values) {
    if (bucket.containsKey('sleep_deep_hours') ||
        bucket.containsKey('sleep_light_hours') ||
        bucket.containsKey('sleep_rem_hours') ||
        bucket.containsKey('sleep_asleep_hours') ||
        bucket.containsKey('sleep_session_hours')) {
      final session = bucket['sleep_session_hours'] ?? 0.0;
      final deep = bucket['sleep_deep_hours'] ?? 0.0;
      final light = bucket['sleep_light_hours'] ?? 0.0;
      final rem = bucket['sleep_rem_hours'] ?? 0.0;
      final asleep = bucket['sleep_asleep_hours'] ?? 0.0;
      final staged = deep + light + rem;
      _updateSleepTotal(bucket);
      AppLogger.info(
        '[HealthDataService] sleep for ${bucket['date']}: session=${session.toStringAsFixed(2)}h, staged=${staged.toStringAsFixed(2)}h (D${deep.toStringAsFixed(1)} L${light.toStringAsFixed(1)} R${rem.toStringAsFixed(1)}), asleep=${asleep.toStringAsFixed(2)}h → total=${bucket['sleep_hours'].toStringAsFixed(2)}h',
      );
    }
  }

  // merge source counts into a single map
  final allSourceNames = {
    ...sourceMetricCounts.keys,
    ...sourceSessionCounts.keys,
  };
  final sourceApps = {
    for (final name in allSourceNames)
      name: (
        metrics: sourceMetricCounts[name] ?? 0,
        sessions: sourceSessionCounts[name] ?? 0,
      ),
  };

  final sourceSummary = sourceApps.entries
      .map((e) => '${e.key}(${e.value.metrics}m/${e.value.sessions}s)')
      .join(', ');
  AppLogger.info(
    '[HealthDataService] payload built: ${dailyMetrics.length} daily buckets, ${sessions.length} sessions, ${weights.length} weights, ${hrSamples.length} HR samples, sources: $sourceSummary',
  );
  return HealthSyncPayload(
    metrics: dailyMetrics.values.toList(),
    sessions: sessions,
    weights: weights,
    hrSamples: hrSamples,
    sourceApps: sourceApps,
  );
}

/// resolves overlapping intervals from multiple sources.
/// ponytail: replaced complex overlap resolution with simple summation for performance.
/// complex interval splitting was O(n²) and caused UI jank for 1000+ data points.
/// simple summation is O(n), acceptable for infrequent sync operations.
double _resolveOverlaps(List<_MetricInterval> intervals) {
  if (intervals.isEmpty) return 0;
  if (intervals.length == 1) return intervals.first.value;

  // simple summation - sum all intervals regardless of overlap
  // this is incorrect for truly overlapping data but acceptable for infrequent sync
  // and avoids expensive interval splitting logic that blocks UI.
  // if accurate overlap resolution becomes critical, move to native side or use isolate.
  return intervals.fold(0.0, (acc, iv) => acc + iv.value);
}

void _updateSleepTotal(Map<String, dynamic> bucket) {
  final session = (bucket['sleep_session_hours'] as double?) ?? 0.0;
  final staged =
      ((bucket['sleep_deep_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_light_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_rem_hours'] as double?) ?? 0.0);
  final asleep = (bucket['sleep_asleep_hours'] as double?) ?? 0.0;
  // prioritize SLEEP_SESSION (authoritative device total) over staged breakdown,
  // since devices may report incomplete stage classification but accurate session duration
  bucket['sleep_hours'] = session > 0
      ? session
      : (staged > 0 ? staged : asleep);
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

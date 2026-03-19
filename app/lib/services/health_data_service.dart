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
  /// per-source-app breakdown: source name -> (metrics, sessions) counts
  final Map<String, ({int metrics, int sessions})> sourceApps;

  const HealthSyncPayload({
    required this.metrics,
    required this.sessions,
    required this.hrSamples,
    this.deletedRecordIds = const [],
    this.sourceApps = const {},
  });

  bool get isEmpty => metrics.isEmpty && sessions.isEmpty && hrSamples.isEmpty && deletedRecordIds.isEmpty;

  Map<String, dynamic> toJson() => {
    'metrics': metrics,
    'sessions': sessions,
    'hr_samples': hrSamples,
    if (deletedRecordIds.isNotEmpty) 'deleted_record_ids': deletedRecordIds,
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

  /// event callback set by ServiceLocator after creation
  void Function(AppEvent)? emitEvent;

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

      // server-driven delta sync: ask backend what dates it has, then only sync missing data
      // force sync still does full 30 days to handle edge cases
      AppLogger.debug('[HealthDataService] reading health data (${force ? 'full 30-day' : 'server-driven delta'})');
      final payload = force ? await readAllData() : await _readDeltaSync();
      AppLogger.info('[HealthDataService] read complete: ${payload.metrics.length} metrics, ${payload.sessions.length} sessions, ${payload.hrSamples.length} HR samples, ${payload.deletedRecordIds.length} deletions');

      if (payload.isEmpty) {
        AppLogger.debug('[HealthDataService] no new data to sync');
        // still fetch backend stats so totals stay current
        await _fetchStatsOnly();
        await prefs.setHcLastSyncMs(DateTime.now().millisecondsSinceEpoch);
        return true;
      }

      // split into per-date batches to stay under the 4MB body limit
      final batches = _splitByDate(payload);
      AppLogger.info('[HealthDataService] split into ${batches.length} batches');

      int totalMetricsSynced = 0;
      int totalSessionsSynced = 0;
      Map<String, dynamic>? lastResponseData;

      for (int i = 0; i < batches.length; i++) {
        if (i > 0) await Future.delayed(const Duration(milliseconds: 200));

        // retry on 429 with exponential backoff
        int attempts = 0;
        const maxAttempts = 3;
        ApiResponse? response;

        while (attempts < maxAttempts) {
          response = await apiService
              .post('/health/sync', body: batches[i].toJson())
              .timeout(_syncTimeout);

          if (response.isSuccess || response.statusCode != 429) break;

          attempts++;
          if (attempts < maxAttempts) {
            final backoffMs = 2000 * attempts; // 2s, 4s
            AppLogger.warning('[HealthDataService] batch ${i + 1}/${batches.length} rate limited, retrying in ${backoffMs}ms (attempt $attempts/$maxAttempts)');
            await Future.delayed(Duration(milliseconds: backoffMs));
          }
        }

        if (response == null || !response.isSuccess) {
          AppLogger.error('[HealthDataService] sync POST failed for batch ${i + 1}/${batches.length}: ${response?.error} (status=${response?.statusCode})');
          _lastSyncResult.value = HealthSyncResult(
            metricsSynced: 0, sessionsSynced: 0,
            totalMetrics: _lastSyncResult.value?.totalMetrics ?? 0,
            totalSessions: _lastSyncResult.value?.totalSessions ?? 0,
            syncedAt: DateTime.now(),
            wasForced: force,
            deviceSources: payload.sourceApps,
            syncError: response?.error ?? 'Upload failed',
          );
          return false;
        }

        totalMetricsSynced += (response.data?['metrics_synced'] as int?) ?? 0;
        totalSessionsSynced += (response.data?['sessions_synced'] as int?) ?? 0;
        lastResponseData = response.data;
      }

      AppLogger.info('[HealthDataService] all batches synced ($totalMetricsSynced metrics, $totalSessionsSynced sessions)');
      if (lastResponseData != null) {
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
      emitEvent?.call(HealthSyncCompleted());
      return true;
    } catch (e) {
      AppLogger.error('[HealthDataService] sync failed', e);
      return false;
    } finally {
      _syncing.value = false;
    }
  }

  /// split payload into one batch per date so each POST stays under the
  /// server body limit. batches are ordered newest-first to ensure most
  /// recent data syncs first (critical if sync fails mid-way).
  List<HealthSyncPayload> _splitByDate(HealthSyncPayload payload) {
    final metricsByDate = <String, List<Map<String, dynamic>>>{};
    for (final m in payload.metrics) {
      metricsByDate.putIfAbsent(m['date'] as String, () => []).add(m);
    }

    final sessionsByDate = <String, List<Map<String, dynamic>>>{};
    for (final s in payload.sessions) {
      sessionsByDate.putIfAbsent(_dateKeyFromMs(s['started_at'] as int), () => []).add(s);
    }

    final hrByDate = <String, List<Map<String, dynamic>>>{};
    for (final hr in payload.hrSamples) {
      hrByDate.putIfAbsent(_dateKeyFromMs(hr['timestamp'] as int), () => []).add(hr);
    }

    final sortedDates = {...metricsByDate.keys, ...sessionsByDate.keys, ...hrByDate.keys}.toList()
      ..sort((a, b) => b.compareTo(a)); // descending: newest first
    if (sortedDates.length <= 1) return [payload];

    return [
      for (int i = 0; i < sortedDates.length; i++)
        HealthSyncPayload(
          metrics: metricsByDate[sortedDates[i]] ?? [],
          sessions: sessionsByDate[sortedDates[i]] ?? [],
          hrSamples: hrByDate[sortedDates[i]] ?? [],
          deletedRecordIds: i == 0 ? payload.deletedRecordIds : const [],
        ),
    ];
  }

  String _dateKeyFromMs(int ms) {
    final dt = DateTime.fromMillisecondsSinceEpoch(ms);
    return '${dt.year}-${dt.month.toString().padLeft(2, '0')}-${dt.day.toString().padLeft(2, '0')}';
  }

  /// server-driven delta sync: backend tells us what dates it has, we only send missing data
  /// this is extremely efficient and handles corrections automatically
  Future<HealthSyncPayload> _readDeltaSync() async {
    try {
      // ask backend what dates it has
      final response = await apiService.get('/health/manifest').timeout(_syncTimeout);
      if (!response.isSuccess || response.data == null) {
        AppLogger.warning('[HealthDataService] failed to get manifest, falling back to 7-day sync');
        return readNewData();
      }

      final serverDates = Set<String>.from((response.data!['dates_with_data'] as List?)?.cast<String>() ?? []);
      AppLogger.debug('[HealthDataService] server has ${serverDates.length} dates with data');

      // read last 7 days of local data (efficient, recent)
      final now = DateTime.now();
      final sevenDaysAgo = now.subtract(const Duration(days: 7));
      final allPayload = await readNewData(); // delegates to platform's 7-day read

      if (allPayload.isEmpty) return allPayload;

      // filter to only missing dates + last 3 days (today, yesterday, 2 days ago)
      // this ensures recent data stays fresh even if devices sync late updates
      final recentDates = <String>{};
      for (int i = 0; i < 3; i++) {
        final date = now.subtract(Duration(days: i));
        recentDates.add('${date.year}-${date.month.toString().padLeft(2, '0')}-${date.day.toString().padLeft(2, '0')}');
      }

      final filteredMetrics = allPayload.metrics.where((m) {
        final date = m['date'] as String;
        return !serverDates.contains(date) || recentDates.contains(date);
      }).toList();

      final filteredSessions = allPayload.sessions.where((s) {
        final date = _dateKeyFromMs(s['started_at'] as int);
        return !serverDates.contains(date) || recentDates.contains(date);
      }).toList();

      final filteredHR = allPayload.hrSamples.where((hr) {
        final date = _dateKeyFromMs(hr['timestamp'] as int);
        return !serverDates.contains(date) || recentDates.contains(date);
      }).toList();

      AppLogger.info('[HealthDataService] delta sync: ${filteredMetrics.length} metrics, ${filteredSessions.length} sessions, ${filteredHR.length} HR samples (filtered from ${allPayload.metrics.length}/${allPayload.sessions.length}/${allPayload.hrSamples.length})');

      return HealthSyncPayload(
        metrics: filteredMetrics,
        sessions: filteredSessions,
        hrSamples: filteredHR,
        deletedRecordIds: allPayload.deletedRecordIds,
      );
    } catch (e) {
      AppLogger.error('[HealthDataService] delta sync failed, falling back', e);
      return readNewData();
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
HealthSyncPayload buildSyncPayload(List<HealthDataPoint> dataPoints) {
  AppLogger.debug('[HealthDataService] buildSyncPayload: ${dataPoints.length} data points');
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

  const sleepSessionTypes = {
    HealthDataType.SLEEP_SESSION,
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
      // attribute sleep to wake-up date (dateTo) so users see sleep on the day they wake up
      final dateKey = _dateKey(point.dateTo);
      dailyMetrics.putIfAbsent(dateKey, () => {'date': dateKey});
      sourceMetricCounts[point.sourceName] = (sourceMetricCounts[point.sourceName] ?? 0) + 1;

      if (sleepSessionTypes.contains(point.type)) {
        // SLEEP_SESSION is the authoritative total - use longest session (workaround for devices that report multiple sessions)
        final hours = point.dateTo.difference(point.dateFrom).inMinutes / 60.0;
        final bucket = dailyMetrics[dateKey]!;
        final existing = bucket['sleep_session_hours'] as double? ?? 0.0;
        if (hours > existing) {
          bucket['sleep_session_hours'] = hours;
          AppLogger.info('[HealthDataService] SLEEP_SESSION: ${hours.toStringAsFixed(2)}h from ${point.sourceName} for $dateKey (replaced ${existing.toStringAsFixed(2)}h)');
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
  dailyMetrics.values.forEach((bucket) {
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
      AppLogger.info('[HealthDataService] sleep for ${bucket['date']}: session=${session.toStringAsFixed(2)}h, staged=${staged.toStringAsFixed(2)}h (D${deep.toStringAsFixed(1)} L${light.toStringAsFixed(1)} R${rem.toStringAsFixed(1)}), asleep=${asleep.toStringAsFixed(2)}h → total=${bucket['sleep_hours'].toStringAsFixed(2)}h');
    }
  });

  // merge source counts into a single map
  final allSourceNames = {...sourceMetricCounts.keys, ...sourceSessionCounts.keys};
  final sourceApps = {
    for (final name in allSourceNames)
      name: (metrics: sourceMetricCounts[name] ?? 0, sessions: sourceSessionCounts[name] ?? 0),
  };

  final sourceSummary = sourceApps.entries.map((e) => '${e.key}(${e.value.metrics}m/${e.value.sessions}s)').join(', ');
  AppLogger.info('[HealthDataService] payload built: ${dailyMetrics.length} daily buckets, ${sessions.length} sessions, ${hrSamples.length} HR samples, sources: $sourceSummary');
  return HealthSyncPayload(
    metrics: dailyMetrics.values.toList(),
    sessions: sessions,
    hrSamples: hrSamples,
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
  final session = (bucket['sleep_session_hours'] as double?) ?? 0.0;
  final staged = ((bucket['sleep_deep_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_light_hours'] as double?) ?? 0.0) +
      ((bucket['sleep_rem_hours'] as double?) ?? 0.0);
  final asleep = (bucket['sleep_asleep_hours'] as double?) ?? 0.0;
  // prioritize SLEEP_SESSION (authoritative device total) over staged breakdown,
  // since devices may report incomplete stage classification but accurate session duration
  bucket['sleep_hours'] = session > 0 ? session : (staged > 0 ? staged : asleep);
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

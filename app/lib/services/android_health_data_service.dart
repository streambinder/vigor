import 'package:health/health.dart';
import 'app_logger.dart';
import 'health_data_service.dart';
import 'preferences_service.dart';
import 'secure_storage_service.dart';

class AndroidHealthDataService extends HealthDataService
    with HealthDataServiceMixin {
  @override
  final PreferencesService prefs;
  @override
  final SecureStorageService storage;

  final Health _health = Health();
  bool _configured = false;

  AndroidHealthDataService({required this.prefs, required this.storage});

  Future<void> _ensureConfigured() async {
    if (!_configured) {
      AppLogger.debug('[AndroidHealth] configuring Health Connect SDK');
      await _health.configure();
      _configured = true;
      AppLogger.info('[AndroidHealth] Health Connect SDK configured');
    }
  }

  /// raw SDK status for UI checks (install/update prompts)
  Future<HealthConnectSdkStatus?> getSdkStatus() async {
    try {
      await _ensureConfigured();
      final status = await _health.getHealthConnectSdkStatus();
      AppLogger.debug('[AndroidHealth] SDK status: $status');
      return status;
    } catch (e) {
      AppLogger.error('[AndroidHealth] getSdkStatus failed', e);
      return null;
    }
  }

  @override
  Future<bool> isAvailable() async {
    try {
      await _ensureConfigured();
      final status = await _health.getHealthConnectSdkStatus();
      AppLogger.debug('[AndroidHealth] isAvailable check: status=$status');
      return status == HealthConnectSdkStatus.sdkAvailable;
    } catch (e) {
      AppLogger.error('[AndroidHealth] availability check failed', e);
      return false;
    }
  }

  @override
  Future<bool> requestPermissions() async {
    try {
      AppLogger.info('[AndroidHealth] requesting permissions for ${healthPermissionTypes.length} types');
      await _ensureConfigured();
      // build a parallel permissions list: READ for all types, READ_WRITE for height/weight
      final writeSet = {HealthDataType.HEIGHT, HealthDataType.WEIGHT};
      final permissions = healthPermissionTypes
          .map((t) => writeSet.contains(t) ? HealthDataAccess.READ_WRITE : HealthDataAccess.READ)
          .toList().cast<HealthDataAccess>();
      final granted = await _health.requestAuthorization(healthPermissionTypes, permissions: permissions);
      AppLogger.info('[AndroidHealth] permissions ${granted ? 'granted' : 'denied'}');
      return granted;
    } catch (e) {
      AppLogger.error('[AndroidHealth] permission request failed', e);
      return false;
    }
  }

  @override
  Future<bool> checkPermissions() async {
    try {
      // hasPermissions returns null when it can't determine status (common for
      // read-only permissions on android) — treat null as granted
      final result = await _health.hasPermissions(healthPermissionTypes);
      AppLogger.debug('[AndroidHealth] checkPermissions: raw=$result, effective=${result ?? true}');
      return result ?? true;
    } catch (e) {
      AppLogger.error('[AndroidHealth] checkPermissions failed', e);
      return true;
    }
  }

  @override
  Future<void> revokePermissions() async {
    try {
      AppLogger.info('[AndroidHealth] revoking permissions');
      await _health.revokePermissions();
      AppLogger.info('[AndroidHealth] permissions revoked');
    } catch (e) {
      AppLogger.error('[AndroidHealth] revokePermissions failed', e);
    }
  }

  @override
  Future<HealthSyncPayload> readAllData() async {
    AppLogger.info('[AndroidHealth] readAllData — full 30-day read');
    await _ensureConfigured();
    return _doFullRead();
  }

  @override
  Future<HealthSyncPayload> readForDate(DateTime date) async {
    await _ensureConfigured();
    final start = DateTime(date.year, date.month, date.day, 0, 0, 0);
    final end = DateTime(date.year, date.month, date.day, 23, 59, 59, 999);
    AppLogger.debug('[AndroidHealth] readForDate $date: ${start.toIso8601String()} → ${end.toIso8601String()}');
    final dataPoints = await _health.getHealthDataFromTypes(
      types: healthPermissionTypes,
      startTime: start,
      endTime: end,
    );
    final deduped = Health().removeDuplicates(dataPoints);
    final hrPoints = await _fetchCorrelatedHR(deduped, start, end);
    final combined = hrPoints.isEmpty ? deduped : [...deduped, ...hrPoints];
    return buildSyncPayload(Health().removeDuplicates(combined));
  }

  @override
  Future<HealthSyncPayload> readNewData() async {
    await _ensureConfigured();

    // read last 7 days for server-driven delta sync (mixin filters by date)
    final now = DateTime.now();
    final sevenDaysAgo = now.subtract(const Duration(days: 7));

    AppLogger.info('[AndroidHealth] reading 7 days: ${sevenDaysAgo.toIso8601String()} to ${now.toIso8601String()}');
    final dataPoints = await _health.getHealthDataFromTypes(
      types: healthPermissionTypes,
      startTime: sevenDaysAgo,
      endTime: now,
    );
    AppLogger.info('[AndroidHealth] read returned ${dataPoints.length} data points');

    final deduped = Health().removeDuplicates(dataPoints);
    AppLogger.debug('[AndroidHealth] after dedup: ${deduped.length} data points (removed ${dataPoints.length - deduped.length})');

    // tier 2: fetch HR only around workout windows to avoid 200k bulk read
    final hrPoints = await _fetchCorrelatedHR(deduped, sevenDaysAgo, now);
    final combined = hrPoints.isEmpty ? deduped : [...deduped, ...hrPoints];

    return buildSyncPayload(Health().removeDuplicates(combined));
  }

  /// full 30-day read for manual/force sync
  Future<HealthSyncPayload> _doFullRead() async {
    final now = DateTime.now();
    final thirtyDaysAgo = now.subtract(const Duration(days: 30));

    AppLogger.debug('[AndroidHealth] full read: ${thirtyDaysAgo.toIso8601String()} to ${now.toIso8601String()}');
    final dataPoints = await _health.getHealthDataFromTypes(
      types: healthPermissionTypes,
      startTime: thirtyDaysAgo,
      endTime: now,
    );
    AppLogger.info('[AndroidHealth] full read returned ${dataPoints.length} data points');

    // deduplicate overlapping data points from multiple sources (e.g. Fitbit + RingConn)
    final deduped = Health().removeDuplicates(dataPoints);
    AppLogger.debug('[AndroidHealth] after dedup: ${deduped.length} data points (removed ${dataPoints.length - deduped.length})');

    final hrPoints = await _fetchCorrelatedHR(deduped, thirtyDaysAgo, now);
    final combined = hrPoints.isEmpty ? deduped : [...deduped, ...hrPoints];

    return buildSyncPayload(Health().removeDuplicates(combined));
  }

  Future<List<HealthDataPoint>> _fetchCorrelatedHR(
    List<HealthDataPoint> allPoints,
    DateTime rangeStart,
    DateTime rangeEnd,
  ) async {
    final workouts = allPoints.where((p) => p.type == HealthDataType.WORKOUT).toList();
    if (workouts.isEmpty) return [];

    final hrPoints = <HealthDataPoint>[];
    // cap to 20 most recent workouts to avoid excessive HC calls
    final recentWorkouts = workouts.length > 20 ? workouts.sublist(workouts.length - 20) : workouts;

    for (final w in recentWorkouts) {
      try {
        final start = w.dateFrom.subtract(const Duration(minutes: 5)).isBefore(rangeStart)
            ? rangeStart
            : w.dateFrom.subtract(const Duration(minutes: 5));
        final end = w.dateTo.add(const Duration(minutes: 5)).isAfter(rangeEnd) ? rangeEnd : w.dateTo.add(const Duration(minutes: 5));

        final points = await _health.getHealthDataFromTypes(
          types: [HealthDataType.HEART_RATE],
          startTime: start,
          endTime: end,
        );
        if (points.isNotEmpty) hrPoints.addAll(points);
      } catch (e) {
        AppLogger.debug('[AndroidHealth] HR window fetch failed for workout ${w.uuid}: $e');
      }
    }

    if (hrPoints.isNotEmpty) {
      AppLogger.info('[AndroidHealth] correlated HR fetched: ${hrPoints.length} samples for ${recentWorkouts.length} workouts');
    }
    return hrPoints;
  }

  @override
  Future<void> writeBodyMetrics({double? height, double? weight}) async {
    if (height == null && weight == null) return;
    try {
      await _ensureConfigured();
      final now = DateTime.now();
      if (height != null) {
        // health connect expects meters; profile stores cm
        final ok = await _health.writeHealthData(
          value: height / 100.0,
          type: HealthDataType.HEIGHT,
          startTime: now,
        );
        AppLogger.info('[AndroidHealth] writeHeight ${height}cm → ${ok ? 'ok' : 'failed'}');
      }
      if (weight != null) {
        final ok = await _health.writeHealthData(
          value: weight,
          type: HealthDataType.WEIGHT,
          startTime: now,
        );
        AppLogger.info('[AndroidHealth] writeWeight ${weight}kg → ${ok ? 'ok' : 'failed'}');
      }
    } catch (e) {
      AppLogger.error('[AndroidHealth] writeBodyMetrics failed', e);
    }
  }

  @override
  Future<void> onSyncSuccess() async {
    // no tokens to persist with rolling window approach
  }
}

import 'package:flutter_timezone/flutter_timezone.dart';
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
  // hold new token in memory until backend POST succeeds (H3)
  String? _pendingChangesToken;
  bool _configured = false;

  AndroidHealthDataService({required this.prefs, required this.storage});

  Future<void> _ensureConfigured() async {
    if (!_configured) {
      await _health.configure();
      _configured = true;
    }
  }

  /// raw SDK status for UI checks (install/update prompts)
  Future<HealthConnectSdkStatus?> getSdkStatus() async {
    try {
      await _ensureConfigured();
      return await _health.getHealthConnectSdkStatus();
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
      return status == HealthConnectSdkStatus.sdkAvailable;
    } catch (e) {
      AppLogger.error('[AndroidHealth] availability check failed', e);
      return false;
    }
  }

  @override
  Future<bool> requestPermissions() async {
    try {
      await _ensureConfigured();
      return await _health.requestAuthorization(healthPermissionTypes);
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
      return result ?? true;
    } catch (e) {
      AppLogger.error('[AndroidHealth] checkPermissions failed', e);
      return true;
    }
  }

  @override
  Future<void> revokePermissions() async {
    try {
      await _health.revokePermissions();
    } catch (e) {
      AppLogger.error('[AndroidHealth] revokePermissions failed', e);
    }
  }

  @override
  Future<HealthSyncPayload> readAllData() async {
    await _ensureConfigured();
    final timezone = await FlutterTimezone.getLocalTimezone();
    return _doFullRead(timezone);
  }

  // sleep stage types to fetch explicitly when changes include sleep data
  static const _sleepStageTypes = [
    HealthDataType.SLEEP_ASLEEP,
    HealthDataType.SLEEP_DEEP,
    HealthDataType.SLEEP_LIGHT,
    HealthDataType.SLEEP_REM,
    HealthDataType.SLEEP_AWAKE,
    HealthDataType.SLEEP_AWAKE_IN_BED,
  ];

  static const _sleepChangeTypes = {
    HealthDataType.SLEEP_SESSION,
    HealthDataType.SLEEP_ASLEEP,
    HealthDataType.SLEEP_DEEP,
    HealthDataType.SLEEP_LIGHT,
    HealthDataType.SLEEP_REM,
  };

  @override
  Future<HealthSyncPayload> readNewData() async {
    await _ensureConfigured();
    final timezone = await FlutterTimezone.getLocalTimezone();
    final token = prefs.hcChangesToken;

    // first sync — no existing token, do full 30-day read
    if (token == null) {
      AppLogger.info('[AndroidHealth] first sync — doing full 30-day read');
      return _doFullRead(timezone);
    }

    final allDataPoints = <HealthDataPoint>[];
    final deletedRecordIds = <String>[];
    var currentToken = token;
    var hasMore = true;
    var completedSuccessfully = true;

    while (hasMore) {
      final changesResponse = await _health.getChanges(
        changesToken: currentToken,
      );

      if (changesResponse == null) {
        AppLogger.error('[AndroidHealth] getChanges returned null');
        completedSuccessfully = false;
        break;
      }

      // token expired — fall back to full re-read
      if (changesResponse.changesTokenExpired) {
        AppLogger.warning('[AndroidHealth] changes token expired, doing full re-read');
        return _doFullRead(timezone);
      }

      allDataPoints.addAll(changesResponse.upsertedDataPoints);
      deletedRecordIds.addAll(changesResponse.deletedRecordIds);
      currentToken = changesResponse.nextChangesToken;
      hasMore = changesResponse.hasMore;
    }

    // only advance the token if the loop completed without errors
    if (completedSuccessfully) {
      _pendingChangesToken = currentToken;
    }

    // the changes API delivers SLEEP_SESSION parent records but not the
    // individual stage sub-records (SLEEP_DEEP, SLEEP_LIGHT, SLEEP_REM, etc.)
    // so when sleep changes are detected, do a supplemental read of stage types
    final hasSleepChanges = allDataPoints.any((p) => _sleepChangeTypes.contains(p.type));
    if (hasSleepChanges) {
      AppLogger.info('[AndroidHealth] sleep changes detected — fetching stage breakdown');
      DateTime? earliest;
      DateTime? latest;
      for (final p in allDataPoints.where((p) => _sleepChangeTypes.contains(p.type))) {
        if (earliest == null || p.dateFrom.isBefore(earliest)) earliest = p.dateFrom;
        if (latest == null || p.dateTo.isAfter(latest)) latest = p.dateTo;
      }
      if (earliest != null && latest != null) {
        // widen window slightly to catch full sessions spanning midnight
        final sleepStart = earliest.subtract(const Duration(hours: 12));
        final sleepEnd = latest.add(const Duration(hours: 1));
        final stagePoints = await _health.getHealthDataFromTypes(
          types: _sleepStageTypes,
          startTime: sleepStart,
          endTime: sleepEnd,
        );
        allDataPoints.addAll(stagePoints);
      }
    }

    final payload = buildSyncPayload(Health().removeDuplicates(allDataPoints), timezone);
    if (deletedRecordIds.isEmpty) return payload;
    return HealthSyncPayload(
      metrics: payload.metrics,
      sessions: payload.sessions,
      hrSamples: payload.hrSamples,
      deletedRecordIds: deletedRecordIds,
      timezone: payload.timezone,
    );
  }

  /// full re-read on first sync or expired token
  Future<HealthSyncPayload> _doFullRead(String timezone) async {
    final now = DateTime.now();
    final thirtyDaysAgo = now.subtract(const Duration(days: 30));

    // get a fresh token for next time
    final newToken = await _health.getChangesToken(types: healthPermissionTypes);
    _pendingChangesToken = newToken;

    final dataPoints = await _health.getHealthDataFromTypes(
      types: healthPermissionTypes,
      startTime: thirtyDaysAgo,
      endTime: now,
    );

    // deduplicate overlapping data points from multiple sources (e.g. Fitbit + RingConn)
    final deduped = Health().removeDuplicates(dataPoints);

    return buildSyncPayload(deduped, timezone);
  }

  @override
  Future<void> onSyncSuccess() async {
    if (_pendingChangesToken != null) {
      await prefs.setHcChangesToken(_pendingChangesToken);
      _pendingChangesToken = null;
    }
  }
}

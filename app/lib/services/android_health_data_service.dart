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

  AndroidHealthDataService({required this.prefs, required this.storage});

  @override
  Future<bool> isAvailable() async {
    try {
      await _health.configure();
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
      await _health.configure();
      return await _health.requestAuthorization(healthPermissionTypes);
    } catch (e) {
      AppLogger.error('[AndroidHealth] permission request failed', e);
      return false;
    }
  }

  @override
  Future<HealthSyncPayload> readAllData() async {
    await _health.configure();
    final timezone = await FlutterTimezone.getLocalTimezone();
    return _doFullRead(timezone);
  }

  @override
  Future<HealthSyncPayload> readNewData() async {
    await _health.configure();
    final timezone = await FlutterTimezone.getLocalTimezone();
    final token = prefs.hcChangesToken;

    // first sync — no existing token, do full 30-day read
    if (token == null) {
      AppLogger.info('[AndroidHealth] first sync — doing full 30-day read');
      return _doFullRead(timezone);
    }

    final allDataPoints = <HealthDataPoint>[];
    var currentToken = token;
    var hasMore = true;

    while (hasMore) {
      final changesResponse = await _health.getChanges(
        changesToken: currentToken,
      );

      if (changesResponse == null) {
        AppLogger.error('[AndroidHealth] getChanges returned null');
        break;
      }

      // token expired — fall back to full re-read
      if (changesResponse.changesTokenExpired) {
        AppLogger.warning('[AndroidHealth] changes token expired, doing full re-read');
        return _doFullRead(timezone);
      }

      allDataPoints.addAll(changesResponse.upsertedDataPoints);
      currentToken = changesResponse.nextChangesToken;
      hasMore = changesResponse.hasMore;
    }

    // store token in memory only — persisted after successful POST (H3)
    _pendingChangesToken = currentToken;
    return buildSyncPayload(Health().removeDuplicates(allDataPoints), timezone);
  }

  /// full re-read on first sync or expired token
  Future<HealthSyncPayload> _doFullRead(String timezone) async {
    final now = DateTime.now();
    final thirtyDaysAgo = now.subtract(const Duration(days: 30));

    // check workout permission status
    final hasWorkoutPerm = await _health.hasPermissions(
      [HealthDataType.WORKOUT],
    );
    AppLogger.info('[AndroidHealth] WORKOUT permission check: $hasWorkoutPerm');

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

    // fetch workouts separately — getHealthDataFromTypes can silently skip them
    // if the runtime permission check is stale
    try {
      final workouts = await _health.getHealthDataFromTypes(
        types: [HealthDataType.WORKOUT],
        startTime: thirtyDaysAgo,
        endTime: now,
      );
      AppLogger.info('[AndroidHealth] separate workout fetch: ${workouts.length} points');
      if (workouts.isNotEmpty) {
        // only add workouts not already in the main batch
        final existingUuids = deduped.where((p) => p.type == HealthDataType.WORKOUT).map((p) => p.uuid).toSet();
        for (final w in workouts) {
          if (!existingUuids.contains(w.uuid)) deduped.add(w);
        }
      }
    } catch (e) {
      AppLogger.error('[AndroidHealth] separate workout fetch failed', e);
    }

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

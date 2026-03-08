import 'package:flutter_timezone/flutter_timezone.dart';
import 'package:health/health.dart';
import 'app_logger.dart';
import 'health_data_service.dart';
import 'preferences_service.dart';
import 'secure_storage_service.dart';

class IOSHealthDataService extends HealthDataService
    with HealthDataServiceMixin {
  @override
  final PreferencesService prefs;
  @override
  final SecureStorageService storage;

  final Health _health = Health();

  IOSHealthDataService({required this.prefs, required this.storage});

  @override
  Future<bool> isAvailable() async {
    try {
      await _health.configure();
      // healthkit is available on all iOS devices
      return true;
    } catch (e) {
      AppLogger.error('[IOSHealth] availability check failed', e);
      return false;
    }
  }

  @override
  Future<bool> requestPermissions() async {
    try {
      await _health.configure();
      return await _health.requestAuthorization(healthPermissionTypes);
    } catch (e) {
      AppLogger.error('[IOSHealth] permission request failed', e);
      return false;
    }
  }

  @override
  Future<HealthSyncPayload> readAllData() async {
    await _health.configure();
    final timezone = await FlutterTimezone.getLocalTimezone();
    final now = DateTime.now();
    final thirtyDaysAgo = now.subtract(const Duration(days: 30));

    final dataPoints = await _health.getHealthDataFromTypes(
      types: healthPermissionTypes,
      startTime: thirtyDaysAgo,
      endTime: now,
    );

    return buildSyncPayload(dataPoints, timezone);
  }

  @override
  Future<HealthSyncPayload> readNewData() async {
    await _health.configure();
    final timezone = await FlutterTimezone.getLocalTimezone();
    final now = DateTime.now();
    final maxLookback = now.subtract(const Duration(days: 30));

    // use last sync timestamp or fall back to 30 days ago
    final lastSyncMs = prefs.hcLastSyncMs;
    DateTime startTime;
    if (lastSyncMs != null) {
      startTime = DateTime.fromMillisecondsSinceEpoch(lastSyncMs);
      // cap to 30 days back
      if (startTime.isBefore(maxLookback)) startTime = maxLookback;
    } else {
      startTime = maxLookback;
    }

    final dataPoints = await _health.getHealthDataFromTypes(
      types: healthPermissionTypes,
      startTime: startTime,
      endTime: now,
    );

    // ios can't verify read grants — just use whatever data arrives
    return buildSyncPayload(dataPoints, timezone);
  }

  @override
  Future<void> onSyncSuccess() async {
    // ios doesn't use changes tokens — timestamp is persisted by the mixin
  }
}

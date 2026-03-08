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

  // filter to iOS-supported types only
  static final _iosTypes = healthPermissionTypes
      .where((t) => dataTypeKeysIOS.contains(t))
      .toList();

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
      return await _health.requestAuthorization(_iosTypes);
    } catch (e) {
      AppLogger.error('[IOSHealth] permission request failed', e);
      return false;
    }
  }

  @override
  Future<bool> checkPermissions() async {
    // ios doesn't expose a reliable way to check read permission grants
    return true;
  }

  @override
  Future<void> revokePermissions() async {
    // no-op on iOS — permissions are managed in system Settings
  }

  @override
  Future<HealthSyncPayload> readAllData() async {
    await _health.configure();
    final timezone = await FlutterTimezone.getLocalTimezone();
    final now = DateTime.now();
    final thirtyDaysAgo = now.subtract(const Duration(days: 30));

    final dataPoints = await _health.getHealthDataFromTypes(
      types: _iosTypes,
      startTime: thirtyDaysAgo,
      endTime: now,
    );

    return buildSyncPayload(Health().removeDuplicates(dataPoints), timezone);
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
      types: _iosTypes,
      startTime: startTime,
      endTime: now,
    );

    return buildSyncPayload(Health().removeDuplicates(dataPoints), timezone);
  }

  @override
  Future<void> onSyncSuccess() async {
    // ios doesn't use changes tokens — timestamp is persisted by the mixin
  }
}

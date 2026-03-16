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
      AppLogger.debug('[IOSHealth] configuring HealthKit');
      await _health.configure();
      AppLogger.debug('[IOSHealth] HealthKit configured, available=true');
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
      AppLogger.info('[IOSHealth] requesting permissions for ${_iosTypes.length} types');
      await _health.configure();
      final granted = await _health.requestAuthorization(_iosTypes);
      AppLogger.info('[IOSHealth] permissions ${granted ? 'granted' : 'denied'}');
      return granted;
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
    AppLogger.info('[IOSHealth] readAllData — full 30-day read');
    await _health.configure();
    final now = DateTime.now();
    final thirtyDaysAgo = now.subtract(const Duration(days: 30));

    AppLogger.debug('[IOSHealth] reading ${thirtyDaysAgo.toIso8601String()} to ${now.toIso8601String()}');
    final dataPoints = await _health.getHealthDataFromTypes(
      types: _iosTypes,
      startTime: thirtyDaysAgo,
      endTime: now,
    );
    AppLogger.info('[IOSHealth] full read returned ${dataPoints.length} data points');

    final deduped = Health().removeDuplicates(dataPoints);
    AppLogger.debug('[IOSHealth] after dedup: ${deduped.length} data points (removed ${dataPoints.length - deduped.length})');
    return buildSyncPayload(deduped);
  }

  @override
  Future<HealthSyncPayload> readNewData() async {
    await _health.configure();
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

    AppLogger.debug('[IOSHealth] readNewData: ${startTime.toIso8601String()} to ${now.toIso8601String()} (lastSyncMs=$lastSyncMs)');
    final dataPoints = await _health.getHealthDataFromTypes(
      types: _iosTypes,
      startTime: startTime,
      endTime: now,
    );
    AppLogger.info('[IOSHealth] incremental read returned ${dataPoints.length} data points');

    final deduped = Health().removeDuplicates(dataPoints);
    AppLogger.debug('[IOSHealth] after dedup: ${deduped.length} data points (removed ${dataPoints.length - deduped.length})');
    return buildSyncPayload(deduped);
  }

  @override
  Future<void> onSyncSuccess() async {
    // ios doesn't use changes tokens — timestamp is persisted by the mixin
  }
}

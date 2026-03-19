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
      // READ_WRITE for height/weight, READ for everything else
      final writeSet = {HealthDataType.HEIGHT, HealthDataType.WEIGHT};
      final permissions = _iosTypes
          .map((t) => writeSet.contains(t) ? HealthDataAccess.READ_WRITE : HealthDataAccess.READ)
          .toList().cast<HealthDataAccess>();
      final granted = await _health.requestAuthorization(_iosTypes, permissions: permissions);
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

    // read last 7 days for server-driven delta sync (mixin filters by date)
    final now = DateTime.now();
    final sevenDaysAgo = now.subtract(const Duration(days: 7));

    AppLogger.info('[IOSHealth] reading 7 days: ${sevenDaysAgo.toIso8601String()} to ${now.toIso8601String()}');
    final dataPoints = await _health.getHealthDataFromTypes(
      types: _iosTypes,
      startTime: sevenDaysAgo,
      endTime: now,
    );
    AppLogger.info('[IOSHealth] read returned ${dataPoints.length} data points');

    final deduped = Health().removeDuplicates(dataPoints);
    AppLogger.debug('[IOSHealth] after dedup: ${deduped.length} data points (removed ${dataPoints.length - deduped.length})');
    return buildSyncPayload(deduped);
  }

  @override
  Future<void> writeBodyMetrics({double? height, double? weight}) async {
    if (height == null && weight == null) return;
    try {
      await _health.configure();
      final now = DateTime.now();
      if (height != null) {
        // healthkit expects meters; profile stores cm
        final ok = await _health.writeHealthData(
          value: height / 100.0,
          type: HealthDataType.HEIGHT,
          startTime: now,
        );
        AppLogger.info('[IOSHealth] writeHeight ${height}cm → ${ok ? 'ok' : 'failed'}');
      }
      if (weight != null) {
        final ok = await _health.writeHealthData(
          value: weight,
          type: HealthDataType.WEIGHT,
          startTime: now,
        );
        AppLogger.info('[IOSHealth] writeWeight ${weight}kg → ${ok ? 'ok' : 'failed'}');
      }
    } catch (e) {
      AppLogger.error('[IOSHealth] writeBodyMetrics failed', e);
    }
  }

  @override
  Future<void> onSyncSuccess() async {
    // ios doesn't use changes tokens — timestamp is persisted by the mixin
  }
}

import 'package:flutter/foundation.dart';
import '../models/family_progress.dart';
import '../models/gym.dart';
import '../models/progress.dart';
import '../models/training.dart';
import '../models/weekly_target.dart';
import 'training_service.dart';
import 'gym_service.dart';
import 'progress_service.dart';
import 'user_service.dart';
import 'secure_storage_service.dart';
import 'preferences_service.dart';
import 'app_logger.dart';
import 'health_data_service.dart';

/// Centralized service locator that caches service instances.
/// Avoids recreating services on every screen mount, reducing memory churn
/// and improving initial frame times.
class ServiceLocator extends ChangeNotifier {
  final SecureStorageService _storage;
  final PreferencesService _prefs;

  TrainingService? _trainingService;
  GymService? _gymService;
  ProgressService? _progressService;
  UserService? _userService;
  HealthDataService? _healthDataService;

  // shared observable state for cross-screen sync
  final ValueNotifier<List<Gym>?> gymsNotifier = ValueNotifier(null);
  final ValueNotifier<List<Training>?> trainingsNotifier = ValueNotifier(null);
  final ValueNotifier<bool> isCalibratingNotifier = ValueNotifier(false);
  final ValueNotifier<Map<String, dynamic>?> healthDailyNotifier = ValueNotifier(null);

  // pending share token to process after login completes
  String? pendingShareToken;
  bool pendingShareAutoClaim = false;

  // pre-loaded homepage data so splash can stay until ready
  Progress? initialProgress;
  WeeklyTarget? initialWeeklyTarget;
  bool initialDataLoaded = false;

  // concurrency guards to prevent duplicate refresh requests
  bool _isRefreshingGyms = false;
  bool _isRefreshingTrainings = false;

  ServiceLocator(this._storage, this._prefs);

  TrainingService get trainingService => _trainingService ??= TrainingService(
        storageService: _storage,
        onDataChanged: refreshTrainings,
      );

  GymService get gymService => _gymService ??= GymService(
        storageService: _storage,
        onDataChanged: refreshGyms,
      );

  ProgressService get progressService =>
      _progressService ??= ProgressService(storageService: _storage);

  UserService get userService =>
      _userService ??= UserService(storageService: _storage);

  HealthDataService? get healthDataService => _healthDataService ??=
      HealthDataService.create(prefs: _prefs, storage: _storage);

  Future<void> refreshGyms() async {
    if (_isRefreshingGyms) return;
    _isRefreshingGyms = true;
    try {
      final response = await gymService.getGyms();
      if (response.isSuccess) {
        gymsNotifier.value = response.data;
      }
    } finally {
      _isRefreshingGyms = false;
    }
  }

  Future<void> refreshTrainings() async {
    if (_isRefreshingTrainings) return;
    _isRefreshingTrainings = true;
    try {
      final response = await trainingService.getTrainings();
      if (response.isSuccess) {
        trainingsNotifier.value = response.data;
      }
    } finally {
      _isRefreshingTrainings = false;
    }
  }

  Future<void> refreshHealthDaily() async {
    final response = await trainingService.getHealthDaily();
    if (response.isSuccess) {
      healthDailyNotifier.value = response.data;
    }
  }

  /// Pre-load homepage data so splash stays until everything is ready
  Future<void> loadInitialData() async {
    final results = await Future.wait([
      progressService.getProgress(),
      progressService.getWeeklyTarget(),
      trainingService.getHealthDaily(),
    ]);
    if (results[0].isSuccess) {
      initialProgress = results[0].data as Progress?;
      _updateCalibrationState();
    }
    if (results[1].isSuccess) initialWeeklyTarget = results[1].data as WeeklyTarget?;
    if (results[2].isSuccess) {
      healthDailyNotifier.value = results[2].data as Map<String, dynamic>?;
    }
    initialDataLoaded = true;
  }

  void _updateCalibrationState() {
    if (initialProgress == null) return;
    final families = ProgressService.parseFamilies(initialProgress!.families);
    isCalibratingNotifier.value = families.values.any((fp) => fp.calibration < 100.0);
  }

  /// Update calibration state from fresh progress data
  void updateCalibrationFromProgress(Map<String, FamilyProgress> families) {
    isCalibratingNotifier.value = families.values.any((fp) => fp.calibration < 100.0);
  }

  /// Clear cached services (e.g., on logout)
  void clearServices() {
    _trainingService = null;
    _gymService = null;
    _progressService = null;
    _userService = null;
    _healthDataService = null;
    gymsNotifier.value = null;
    trainingsNotifier.value = null;
    isCalibratingNotifier.value = false;
    healthDailyNotifier.value = null;
    pendingShareToken = null;
    pendingShareAutoClaim = false;
    initialProgress = null;
    initialWeeklyTarget = null;
    initialDataLoaded = false;
    // clear health sync tokens on logout (H2 — multi-user scoping)
    AppLogger.info('[ServiceLocator] logout — clearing health data and tokens');
    _prefs.clearHealthData();
  }

  @override
  void dispose() {
    gymsNotifier.dispose();
    trainingsNotifier.dispose();
    isCalibratingNotifier.dispose();
    healthDailyNotifier.dispose();
    super.dispose();
  }
}

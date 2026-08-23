import 'dart:async';
import 'package:flutter/foundation.dart';
import '../models/family_progress.dart';
import '../models/gym.dart';
import '../models/progress.dart';
import '../models/flow_session.dart';
import '../models/training.dart';
import '../models/weekly_target.dart';
import 'app_event.dart';
import 'flow_service.dart';
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
  FlowService? _flowService;

  // shared observable state for cross-screen sync
  final ValueNotifier<List<Gym>?> gymsNotifier = ValueNotifier(null);
  final ValueNotifier<List<Training>?> trainingsNotifier = ValueNotifier(null);
  final ValueNotifier<bool> isCalibratingNotifier = ValueNotifier(false);
  final ValueNotifier<Map<String, dynamic>?> healthDailyNotifier = ValueNotifier(null);
  final ValueNotifier<Map<String, dynamic>?> readinessNotifier = ValueNotifier(null);
  final ValueNotifier<List<FlowSession>?> flowSessionsNotifier = ValueNotifier(null);

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
  bool _isRefreshingFlowSessions = false;

  // typed event bus
  final _eventController = StreamController<AppEvent>.broadcast();
  StreamSubscription<AppEvent>? _eventSub;
  Stream<AppEvent> get events => _eventController.stream;
  void emit(AppEvent event) => _eventController.add(event);

  ServiceLocator(this._storage, this._prefs) {
    _eventSub = _eventController.stream.listen(_onEvent);
  }

  void _onEvent(AppEvent event) {
    switch (event) {
      case TrainingListChanged():
        refreshTrainings();
      case TrainingCompleted():
        refreshTrainings();
        refreshHealthDaily();
      case GymListChanged():
        refreshGyms();
      case HealthSyncCompleted():
        refreshHealthDaily();
        refreshTrainings();
      case FlowSessionListChanged():
        refreshFlowSessions();
      case FeedbackSubmitted():
      case ProfileUpdated():
        break; // handled at screen level / by AuthProvider
    }
  }

  TrainingService get trainingService => _trainingService ??= TrainingService(
        storageService: _storage,
        emitEvent: emit,
      );

  GymService get gymService => _gymService ??= GymService(
        storageService: _storage,
        emitEvent: emit,
      );

  ProgressService get progressService =>
      _progressService ??= ProgressService(storageService: _storage);

  UserService get userService =>
      _userService ??= UserService(storageService: _storage);

  HealthDataService? get healthDataService {
    if (_healthDataService != null) return _healthDataService;
    _healthDataService = HealthDataService.create(prefs: _prefs, storage: _storage);
    if (_healthDataService is HealthDataServiceMixin) {
      (_healthDataService as HealthDataServiceMixin).emitEvent = emit;
    }
    return _healthDataService;
  }

  FlowService get flowService => _flowService ??= FlowService(
        storageService: _storage,
        emitEvent: emit,
      );

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

  Future<void> refreshFlowSessions() async {
    if (_isRefreshingFlowSessions) return;
    _isRefreshingFlowSessions = true;
    try {
      final response = await flowService.getFlowSessions();
      if (response.isSuccess) {
        flowSessionsNotifier.value = response.data;
      }
    } finally {
      _isRefreshingFlowSessions = false;
    }
  }

  Future<void> refreshHealthDaily() async {
    // non-blocking health daily fetch — triggers backend sync
    trainingService.getHealthDaily().then((response) {
      if (response.isSuccess && response.data != null) {
        healthDailyNotifier.value = response.data;
      }
    });
  }

  /// daily readiness hint. the probe runs at most once per calendar day per
  /// device — after that the cached value from prefs is served until the day
  /// rolls over. force (pull-to-refresh) asks the backend to recompute.
  Future<void> refreshReadiness({bool force = false}) async {
    final now = DateTime.now();
    final today = '${now.year.toString().padLeft(4, '0')}-${now.month.toString().padLeft(2, '0')}-${now.day.toString().padLeft(2, '0')}';
    if (!force && _prefs.readinessDate == today) {
      readinessNotifier.value ??= _prefs.readinessJson;
      return;
    }
    final response = await trainingService.getReadinessToday(force: force);
    if (response.isSuccess) {
      readinessNotifier.value = response.data;
      // cache positive results only: a 404 often just means the morning
      // sleep sync has not landed yet — the next app open must retry
      if (response.data != null) {
        await _prefs.setReadiness(today, response.data);
      }
    }
  }

  /// Pre-load homepage data so splash stays until everything is ready
  /// Note: health daily fetch is intentionally NOT included here because it
  /// triggers backend sync which can block on expensive database joins.
  /// Health data is fetched separately via fire-and-forget syncToBackend().
  Future<void> loadInitialData() async {
    final results = await Future.wait([
      progressService.getProgress(),
      progressService.getWeeklyTarget(),
    ]);
    if (results[0].isSuccess) {
      initialProgress = results[0].data as Progress?;
      _updateCalibrationState();
    }
    if (results[1].isSuccess) initialWeeklyTarget = results[1].data as WeeklyTarget?;
    initialDataLoaded = true;
    // fire-and-forget health daily fetch — non-blocking
    trainingService.getHealthDaily().then((response) {
      if (response.isSuccess && response.data != null) {
        healthDailyNotifier.value = response.data;
      }
    });
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
    _flowService = null;
    gymsNotifier.value = null;
    trainingsNotifier.value = null;
    isCalibratingNotifier.value = false;
    healthDailyNotifier.value = null;
    readinessNotifier.value = null;
    flowSessionsNotifier.value = null;
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
    _eventSub?.cancel();
    _eventController.close();
    gymsNotifier.dispose();
    trainingsNotifier.dispose();
    isCalibratingNotifier.dispose();
    healthDailyNotifier.dispose();
    readinessNotifier.dispose();
    flowSessionsNotifier.dispose();
    super.dispose();
  }
}

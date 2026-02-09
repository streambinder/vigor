import 'package:flutter/foundation.dart';
import '../models/gym.dart';
import '../models/training.dart';
import 'training_service.dart';
import 'gym_service.dart';
import 'progress_service.dart';
import 'user_service.dart';
import 'secure_storage_service.dart';

/// Centralized service locator that caches service instances.
/// Avoids recreating services on every screen mount, reducing memory churn
/// and improving initial frame times.
class ServiceLocator extends ChangeNotifier {
  final SecureStorageService _storage;

  TrainingService? _trainingService;
  GymService? _gymService;
  ProgressService? _progressService;
  UserService? _userService;

  // shared observable state for cross-screen sync
  final ValueNotifier<List<Gym>?> gymsNotifier = ValueNotifier(null);
  final ValueNotifier<List<Training>?> trainingsNotifier = ValueNotifier(null);

  // concurrency guards to prevent duplicate refresh requests
  bool _isRefreshingGyms = false;
  bool _isRefreshingTrainings = false;

  ServiceLocator(this._storage);

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

  /// Clear cached services (e.g., on logout)
  void clearServices() {
    _trainingService = null;
    _gymService = null;
    _progressService = null;
    _userService = null;
    gymsNotifier.value = null;
    trainingsNotifier.value = null;
  }

  @override
  void dispose() {
    gymsNotifier.dispose();
    trainingsNotifier.dispose();
    super.dispose();
  }
}

import 'package:flutter/foundation.dart';
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

  ServiceLocator(this._storage);

  TrainingService get trainingService =>
      _trainingService ??= TrainingService(storageService: _storage);

  GymService get gymService =>
      _gymService ??= GymService(storageService: _storage);

  ProgressService get progressService =>
      _progressService ??= ProgressService(storageService: _storage);

  UserService get userService =>
      _userService ??= UserService(storageService: _storage);

  /// Clear cached services (e.g., on logout)
  void clearServices() {
    _trainingService = null;
    _gymService = null;
    _progressService = null;
    _userService = null;
  }
}

import 'dart:convert';

import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:vigor/models/api_response.dart';
import 'package:vigor/services/preferences_service.dart';
import 'package:vigor/services/secure_storage_service.dart';
import 'package:vigor/services/service_locator.dart';
import 'package:vigor/services/training_service.dart';

/// ServiceLocator that fails the test if anything touches the backend.
class _NoBackendLocator extends ServiceLocator {
  // ignore: use_super_parameters — the superclass parameters are private
  _NoBackendLocator(SecureStorageService storage, PreferencesService prefs) : super(storage, prefs);

  @override
  TrainingService get trainingService => throw StateError('backend must not be called');
}

Future<ServiceLocator> _locatorWithPrefs(Map<String, Object> values) async {
  SharedPreferences.setMockInitialValues(values);
  final prefs = PreferencesService();
  await prefs.initialize();
  return _NoBackendLocator(SecureStorageService(), prefs);
}

void main() {
  test("serveCachedReadiness serves today's cached hint without a backend call", () async {
    final locator = await _locatorWithPrefs({
      'readiness_date': ServiceLocator.readinessDayKey(),
      'readiness_json': jsonEncode({'score': 80, 'level': 'green', 'summary': 'ready to go'}),
    });

    locator.serveCachedReadiness();

    final hint = locator.readinessNotifier.value;
    expect(hint, isNotNull);
    expect(hint!['score'], 80);
    expect(hint['level'], 'green');
    locator.dispose();
  });

  test('serveCachedReadiness ignores a stale cached hint', () async {
    final yesterday = ServiceLocator.readinessDayKey(DateTime.now().subtract(const Duration(days: 1)));
    final locator = await _locatorWithPrefs({
      'readiness_date': yesterday,
      'readiness_json': jsonEncode({'score': 80}),
    });

    locator.serveCachedReadiness();

    expect(locator.readinessNotifier.value, isNull);
    locator.dispose();
  });

  test('serveCachedReadiness does nothing without a cached hint', () async {
    final locator = await _locatorWithPrefs({});

    locator.serveCachedReadiness();

    expect(locator.readinessNotifier.value, isNull);
    locator.dispose();
  });

  test('serveCachedReadiness keeps an already-set notifier value', () async {
    final locator = await _locatorWithPrefs({
      'readiness_date': ServiceLocator.readinessDayKey(),
      'readiness_json': jsonEncode({'score': 80}),
    });
    locator.readinessNotifier.value = {'score': 42};

    locator.serveCachedReadiness();

    expect(locator.readinessNotifier.value!['score'], 42);
    locator.dispose();
  });

  test('refreshReadiness serves a valid cached hint without a backend call', () async {
    final fake = _FakeTrainingService();
    final locator = await _fakeLocatorWithPrefs({
      'readiness_date': ServiceLocator.readinessDayKey(),
      'readiness_json': jsonEncode({'score': 80, 'level': 'green', 'summary': 'ready to go'}),
    }, fake);

    final ready = await locator.refreshReadiness();

    expect(ready, isTrue);
    expect(fake.readinessCalls, 0);
    expect(locator.readinessNotifier.value!['score'], 80);
    locator.dispose();
  });

  test('refreshReadiness probes the backend when the day is marked but no hint is cached', () async {
    final fake = _FakeTrainingService();
    final locator = await _fakeLocatorWithPrefs({
      'readiness_date': ServiceLocator.readinessDayKey(),
    }, fake);

    final ready = await locator.refreshReadiness();

    expect(ready, isTrue);
    expect(fake.readinessCalls, 1);
    expect(locator.readinessNotifier.value!['score'], 55);
    locator.dispose();
  });
}

/// TrainingService stub that answers readiness probes from memory.
class _FakeTrainingService extends TrainingService {
  _FakeTrainingService() : super(storageService: SecureStorageService());

  int readinessCalls = 0;

  @override
  Future<ApiResponse<Map<String, dynamic>?>> getReadinessToday({bool force = false}) async {
    readinessCalls++;
    return ApiResponse.success({'score': 55, 'level': 'yellow', 'summary': 'take it steady'}, 200);
  }
}

/// ServiceLocator whose backend is the in-memory stub above.
class _FakeBackendLocator extends ServiceLocator {
  // ignore: use_super_parameters — the superclass parameters are private
  _FakeBackendLocator(SecureStorageService storage, PreferencesService prefs, this.fake) : super(storage, prefs);

  final _FakeTrainingService fake;

  @override
  TrainingService get trainingService => fake;
}

Future<ServiceLocator> _fakeLocatorWithPrefs(Map<String, Object> values, _FakeTrainingService fake) async {
  SharedPreferences.setMockInitialValues(values);
  final prefs = PreferencesService();
  await prefs.initialize();
  return _FakeBackendLocator(SecureStorageService(), prefs, fake);
}

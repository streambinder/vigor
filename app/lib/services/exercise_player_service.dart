import 'dart:convert';
import 'dart:typed_data';

import 'authenticated_api_service.dart';
import 'preferences_service.dart';

/// One renderable model, metadata only. The bytes travel separately.
class ExercisePlayerModel {
  final String id;
  final String displayName;
  final String licence;
  final String source;
  final String md5;

  const ExercisePlayerModel({
    required this.id,
    required this.displayName,
    required this.licence,
    required this.source,
    required this.md5,
  });

  factory ExercisePlayerModel.fromJson(Map<String, dynamic> json) {
    return ExercisePlayerModel(
      id: json['id'] as String,
      displayName: json['display_name'] as String,
      licence: json['licence'] as String,
      source: json['source'] as String,
      md5: json['md5'] as String,
    );
  }
}

/// Serves the exercise player data: animated exercises with their stills,
/// animation documents, and model bytes, all from the backend API.
///
/// Fetches are memoized per app run: the stills and the model list are
/// fetched once and cached, and each model is downloaded once per id.
/// A failed fetch clears its memo so the next call retries instead of
/// caching the failure. Absent data (404) reads as null so callers fall
/// back to the GIF, never to an error screen.
class ExercisePlayerService {
  final AuthenticatedApiService _api;
  final PreferencesService _prefs;

  Future<Map<String, String>>? _animatedFuture;
  Future<List<ExercisePlayerModel>>? _modelsFuture;
  final Map<String, Uint8List> _modelBytes = {};

  ExercisePlayerService({
    AuthenticatedApiService? api,
    required PreferencesService prefs,
  })  : _api = api ?? AuthenticatedApiService(),
        _prefs = prefs;

  /// Animated exercise ids mapped to their still SVGs.
  Future<Map<String, String>> animatedExercises() {
    return _animatedFuture ??= _fetchAnimatedExercises();
  }

  Future<Map<String, String>> _fetchAnimatedExercises() async {
    try {
      final response = await _api.getBytes('/exercises/animated');
      if (!response.isSuccess || response.data == null) {
        return {};
      }
      final decoded = jsonDecode(utf8.decode(response.data!)) as List<dynamic>;
      return {
        for (final row in decoded)
          (row as Map<String, dynamic>)['exercise_id'] as String:
              row['still'] as String,
      };
    } catch (_) {
      _animatedFuture = null;
      rethrow;
    }
  }

  /// Whether the exercise has an animation (and therefore a still).
  Future<bool> isAnimated(String exerciseId) async {
    final animated = await animatedExercises();
    return animated.containsKey(exerciseId);
  }

  /// The animation document for the exercise, or null when absent or
  /// unreachable. Callers fall back to the GIF on null.
  Future<String?> animation(String exerciseId) async {
    final response = await _api.getBytes('/exercises/$exerciseId/animation');
    if (!response.isSuccess || response.data == null) {
      return null;
    }
    return utf8.decode(response.data!);
  }

  /// The model catalog, metadata only.
  Future<List<ExercisePlayerModel>> models() {
    return _modelsFuture ??= _fetchModels();
  }

  Future<List<ExercisePlayerModel>> _fetchModels() async {
    try {
      final response = await _api.getBytes('/exercises/models');
      if (!response.isSuccess || response.data == null) {
        return [];
      }
      final decoded = jsonDecode(utf8.decode(response.data!)) as List<dynamic>;
      return decoded
          .map((row) =>
              ExercisePlayerModel.fromJson(row as Map<String, dynamic>))
          .toList();
    } catch (_) {
      _modelsFuture = null;
      rethrow;
    }
  }

  /// The selected model id: the persisted choice when it is still in the
  /// catalog, otherwise the first model, otherwise null.
  Future<String?> selectedModelId() async {
    final catalog = await models();
    if (catalog.isEmpty) {
      return null;
    }
    final stored = _prefs.exercisePlayerModelId;
    if (stored != null && catalog.any((model) => model.id == stored)) {
      return stored;
    }
    return catalog.first.id;
  }

  /// Persists the model choice for the future picker UI.
  Future<void> setSelectedModelId(String id) {
    return _prefs.setExercisePlayerModelId(id);
  }

  /// The model bytes for the id, downloaded once and cached. Null when
  /// absent or unreachable, so callers fall back to the GIF.
  Future<Uint8List?> modelBytes(String modelId) async {
    final cached = _modelBytes[modelId];
    if (cached != null) {
      return cached;
    }
    final response = await _api.getBytes('/exercises/models/$modelId');
    if (!response.isSuccess || response.data == null) {
      return null;
    }
    _modelBytes[modelId] = response.data!;
    return response.data;
  }
}

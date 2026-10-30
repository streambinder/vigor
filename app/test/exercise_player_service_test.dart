import 'dart:convert';
import 'dart:typed_data';

import 'package:flutter_test/flutter_test.dart';
import 'package:http/http.dart' as http;
import 'package:http/testing.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:vigor/models/api_response.dart';
import 'package:vigor/services/api_service.dart';
import 'package:vigor/services/authenticated_api_service.dart';
import 'package:vigor/services/exercise_player_service.dart';
import 'package:vigor/services/preferences_service.dart';

/// Canned AuthenticatedApiService: answers from a table, counts calls,
/// and throws for paths in failingPaths.
class _FakeAuthenticatedApi extends AuthenticatedApiService {
  final Map<String, http.Response> responses;
  final Set<String> failingPaths = {};
  final Map<String, int> counts = {};

  _FakeAuthenticatedApi(this.responses);

  @override
  Future<ApiResponse<Uint8List>> getBytes(String endpoint) async {
    counts.update(endpoint, (v) => v + 1, ifAbsent: () => 1);
    if (failingPaths.contains(endpoint)) {
      throw Exception('boom');
    }
    final response = responses[endpoint];
    if (response == null) {
      return ApiResponse.error('not found', 404);
    }
    if (response.statusCode >= 200 && response.statusCode < 300) {
      return ApiResponse.success(response.bodyBytes, response.statusCode);
    }
    return ApiResponse.error('Request failed', response.statusCode);
  }
}

Future<PreferencesService> _prefsWith(Map<String, Object> values) async {
  SharedPreferences.setMockInitialValues(values);
  final prefs = PreferencesService();
  await prefs.initialize();
  return prefs;
}

_FakeAuthenticatedApi _api() {
  return _FakeAuthenticatedApi({
    '/exercises/animated': http.Response(
      jsonEncode([
        {'exercise_id': 'squat', 'still': '<svg></svg>'},
      ]),
      200,
    ),
    '/exercises/squat/animation': http.Response('{"displayName":"Squat"}', 200),
    '/exercises/models': http.Response(
      jsonEncode([
        {
          'id': 'quaternius-mannequin',
          'display_name': 'Quaternius Mannequin',
          'licence': 'CC0 1.0 Universal',
          'source': 'Universal Animation Library',
          'md5': 'abc',
        },
      ]),
      200,
    ),
    '/exercises/models/quaternius-mannequin':
        http.Response.bytes([1, 2, 3, 4], 200),
  });
}

void main() {
  test('animatedExercises maps ids to stills and memoizes', () async {
    final api = _api();
    final service = ExercisePlayerService(
      api: api,
      prefs: await _prefsWith({}),
    );
    final first = await service.animatedExercises();
    final second = await service.animatedExercises();
    expect(first, {'squat': '<svg></svg>'});
    expect(second, {'squat': '<svg></svg>'});
    expect(api.counts['/exercises/animated'], 1);
    expect(await service.isAnimated('squat'), isTrue);
    expect(await service.isAnimated('deadlift'), isFalse);
  });

  test('animatedExercises failure clears the memo for retry', () async {
    final api = _api();
    api.failingPaths.add('/exercises/animated');
    final service = ExercisePlayerService(
      api: api,
      prefs: await _prefsWith({}),
    );
    await expectLater(service.animatedExercises(), throwsException);
    api.failingPaths.remove('/exercises/animated');
    expect(await service.animatedExercises(), {'squat': '<svg></svg>'});
    expect(api.counts['/exercises/animated'], 2);
  });

  test('animation returns the document and null on 404', () async {
    final service = ExercisePlayerService(
      api: _api(),
      prefs: await _prefsWith({}),
    );
    expect(await service.animation('squat'), '{"displayName":"Squat"}');
    expect(await service.animation('nope'), isNull);
  });

  test('models parses metadata and selection honors the stored choice',
      () async {
    final service = ExercisePlayerService(
      api: _api(),
      prefs: await _prefsWith(
        {'exercise_player_model_id': 'quaternius-mannequin'},
      ),
    );
    final models = await service.models();
    expect(models.length, 1);
    expect(models.first.id, 'quaternius-mannequin');
    expect(models.first.licence, 'CC0 1.0 Universal');
    expect(
      await service.selectedModelId(),
      'quaternius-mannequin',
    );
    await service.setSelectedModelId('quaternius-mannequin');
    expect(
      service.selectedModelId(),
      completion('quaternius-mannequin'),
    );
  });

  test('selectedModelId falls back to the first model when stored is stale',
      () async {
    final service = ExercisePlayerService(
      api: _api(),
      prefs: await _prefsWith({'exercise_player_model_id': 'gone'}),
    );
    expect(await service.selectedModelId(), 'quaternius-mannequin');
  });

  test('selectedModelId is null for an empty catalog', () async {
    final api = _FakeAuthenticatedApi({
      '/exercises/models': http.Response('[]', 200),
    });
    final service = ExercisePlayerService(
      api: api,
      prefs: await _prefsWith({}),
    );
    expect(await service.selectedModelId(), isNull);
  });

  test('modelBytes downloads once per id and caches', () async {
    final api = _api();
    final service = ExercisePlayerService(
      api: api,
      prefs: await _prefsWith({}),
    );
    final first = await service.modelBytes('quaternius-mannequin');
    final second = await service.modelBytes('quaternius-mannequin');
    expect(first, [1, 2, 3, 4]);
    expect(second, [1, 2, 3, 4]);
    expect(api.counts['/exercises/models/quaternius-mannequin'], 1);
    expect(await service.modelBytes('nope'), isNull);
  });

  test('ApiService.getBytes returns raw bytes and surfaces failures',
      () async {
    final api = ApiService(
      client: MockClient((request) async {
        if (request.url.path.endsWith('/ok')) {
          return http.Response.bytes([9, 8, 7], 200);
        }
        return http.Response('nope', 404);
      }),
    );
    final ok = await api.getBytes('/ok');
    expect(ok.isSuccess, isTrue);
    expect(ok.data, [9, 8, 7]);
    final missing = await api.getBytes('/missing');
    expect(missing.isSuccess, isFalse);
    expect(missing.statusCode, 404);
  });
}

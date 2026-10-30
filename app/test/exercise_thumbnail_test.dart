import 'package:flutter/material.dart';
import 'package:flutter_svg/flutter_svg.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:vigor/models/exercise.dart';
import 'package:vigor/services/exercise_player_service.dart';
import 'package:vigor/services/preferences_service.dart';
import 'package:vigor/widgets/cached_exercise_image.dart';
import 'package:vigor/widgets/exercise_thumbnail.dart';

class _FakePlayerService extends ExercisePlayerService {
  final Map<String, String> stills;

  _FakePlayerService(this.stills, PreferencesService prefs)
      : super(prefs: prefs);

  @override
  Future<Map<String, String>> animatedExercises() async => stills;
}

Exercise _exercise({required String id, String reference = ''}) {
  return Exercise(
    id: id,
    name: 'Squat',
    equipment: [],
    muscles: [],
    reference: reference,
    instructions: [],
    cues: [],
    difficulty: 50,
    isMobility: false,
    mode: 'reps',
  );
}

Future<ExercisePlayerService> _service(Map<String, String> stills) async {
  SharedPreferences.setMockInitialValues({});
  final prefs = PreferencesService();
  await prefs.initialize();
  return _FakePlayerService(stills, prefs);
}

Widget _thumbnail(Exercise? exercise, ExercisePlayerService service) {
  return MaterialApp(
    home: Scaffold(
      body: ExerciseThumbnail(
        exercise: exercise,
        playerService: service,
        width: 72,
        height: 72,
        borderRadius: BorderRadius.circular(8),
      ),
    ),
  );
}

void main() {
  testWidgets('shows the still with a play badge for animated exercises',
      (tester) async {
    final service = await _service({
      'squat':
          '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 72 72"><circle cx="36" cy="36" r="10"/></svg>',
    });
    await tester.pumpWidget(_thumbnail(_exercise(id: 'squat'), service));
    await tester.pumpAndSettle();
    expect(find.byType(SvgPicture), findsOneWidget);
    expect(find.byIcon(Icons.play_arrow), findsOneWidget);
  });

  testWidgets('shows the GIF for non-animated exercises with a reference',
      (tester) async {
    final service = await _service({});
    await tester.pumpWidget(
      _thumbnail(
        _exercise(id: 'squat', reference: 'https://example.com/squat.gif'),
        service,
      ),
    );
    await tester.pump();
    await tester.pump();
    expect(find.byType(CachedExerciseImage), findsOneWidget);
  });

  testWidgets('shows the placeholder without a still or a reference',
      (tester) async {
    final service = await _service({});
    await tester.pumpWidget(_thumbnail(_exercise(id: 'squat'), service));
    await tester.pumpAndSettle();
    expect(find.byIcon(Icons.fitness_center), findsOneWidget);
  });

  testWidgets('shows the placeholder for a null exercise', (tester) async {
    final service = await _service({});
    await tester.pumpWidget(_thumbnail(null, service));
    await tester.pump();
    expect(find.byIcon(Icons.fitness_center), findsOneWidget);
  });
}

import 'package:flutter_test/flutter_test.dart';
import 'package:vigor/models/activity.dart';
import 'package:vigor/models/activity_actual.dart';
import 'package:vigor/models/block.dart';
import 'package:vigor/models/routine.dart';
import 'package:vigor/models/training.dart';
import 'package:vigor/utils/workout_summary.dart';

const labels = (warmup: 'Warmup', work: 'Work', cooldown: 'Cooldown', rest: 'Rest');

Activity activity({
  required String id,
  required String name,
  int reps = 0,
  int duration = 0,
  double weightKg = 0,
  int rest = 0,
}) =>
    Activity(
      id: id,
      blockId: 'block-1',
      exerciseId: 'ex-$id',
      name: name,
      duration: duration,
      reps: reps,
      weightKg: weightKg,
      modifiers: const [],
      rest: rest,
      detail: const {},
    );

Training trainingWith(List<Routine> routines, {int duration = 2700, int? completedIn}) => Training(
      id: 'training-1',
      name: 'Leg Day',
      description: '',
      methodology: 'strength',
      duration: duration,
      equipment: const [],
      goals: const [],
      muscles: const [],
      request: '',
      references: const [],
      factIndices: const [],
      routines: routines,
      completedIn: completedIn,
      hasHealthSession: false,
      createdAt: DateTime.utc(2026, 10, 23),
      userId: 'user-1',
    );

Routine routine(String type, List<Block> blocks) =>
    Routine(id: 'routine-$type', trainingId: 'training-1', type: type, rest: 0, blocks: blocks);

Block block(List<Activity> activities, {int repeats = 1, int rest = 0}) =>
    Block(id: 'block-1', routineId: 'routine-work', repeats: repeats, rest: rest, activities: activities);

String summarize(Training training, {Map<String, ActivityActual> actuals = const {}}) => buildWorkoutSummary(
      training,
      actuals: actuals,
      labels: labels,
      dateText: '23 Oct 2026',
      methodologyLabel: 'Strength',
    );

void main() {
  group('buildWorkoutSummary', () {
    test('header carries name, date, methodology and total duration', () {
      final text = summarize(trainingWith(const [], completedIn: 2520));
      expect(text, 'Leg Day · 23 Oct 2026\nStrength · 42m');
    });

    test('planned duration is used when the training is not completed', () {
      final text = summarize(trainingWith(const [], duration: 3600));
      expect(text, contains('Strength · 1h'));
    });

    test('prescribed sets, reps, load and rest render on one line', () {
      final training = trainingWith([
        routine('work', [
          block([activity(id: 'a1', name: 'Squat', reps: 8, weightKg: 60, rest: 120)], repeats: 3),
        ]),
      ]);
      expect(summarize(training), contains('- Squat: 3×8 · 60 kg · Rest 120s'));
    });

    test('bodyweight exercise omits the load', () {
      final training = trainingWith([
        routine('work', [
          block([activity(id: 'a1', name: 'Push-up', reps: 12)]),
        ]),
      ]);
      expect(summarize(training), contains('- Push-up: 12'));
      expect(summarize(training), isNot(contains('kg')));
    });

    test('duration-based activity renders the formatted duration', () {
      final training = trainingWith([
        routine('work', [
          block([activity(id: 'a1', name: 'Plank', duration: 90)]),
        ]),
      ]);
      expect(summarize(training), contains('- Plank: 1m 30s'));
    });

    test('multi-activity block states the rounds in a header line', () {
      final training = trainingWith([
        routine('work', [
          block(
            [
              activity(id: 'a1', name: 'Kettlebell Swing', reps: 15),
              activity(id: 'a2', name: 'Goblet Squat', reps: 10),
            ],
            repeats: 4,
            rest: 60,
          ),
        ]),
      ]);
      final text = summarize(training);
      expect(text, contains('×4 · Rest 60s:'));
      expect(text, contains('- Kettlebell Swing: 15'));
    });

    test('sections follow the training order', () {
      final training = trainingWith([
        routine('cooldown', [block([activity(id: 'a3', name: 'Stretch', duration: 60)])]),
        routine('work', [block([activity(id: 'a2', name: 'Row', reps: 10)])]),
        routine('warmup', [block([activity(id: 'a1', name: 'Mobility', duration: 60)])]),
      ]);
      final text = summarize(training);
      expect(text.indexOf('Cooldown'), lessThan(text.indexOf('Work')));
      expect(text.indexOf('Work'), lessThan(text.indexOf('Warmup')));
    });

    test('actuals replace name, reps and load', () {
      final training = trainingWith([
        routine('work', [
          block([activity(id: 'a1', name: 'Bench Press', reps: 15, weightKg: 40)]),
        ]),
      ]);
      final text = summarize(training, actuals: {
        'a1': ActivityActual(name: 'Floor Press', reps: 12, weightKg: 30),
      });
      expect(text, contains('- Floor Press: 12 · 30 kg'));
    });

    test('a skipped activity leaves the summary', () {
      final training = trainingWith([
        routine('work', [
          block([
            activity(id: 'a1', name: 'Bench Press', reps: 15, weightKg: 40),
            activity(id: 'a2', name: 'Fly', reps: 12, weightKg: 10),
          ]),
        ]),
      ]);
      final text = summarize(training, actuals: {
        'a2': ActivityActual(skipped: true),
      });
      expect(text, contains('Bench Press'));
      expect(text, isNot(contains('Fly')));
    });

    test('a routine whose activities are all skipped leaves no section', () {
      final training = trainingWith([
        routine('cooldown', [block([activity(id: 'a1', name: 'Stretch', duration: 60)])]),
      ]);
      final text = summarize(training, actuals: {
        'a1': ActivityActual(skipped: true),
      });
      expect(text, isNot(contains('Cooldown')));
    });
  });
}

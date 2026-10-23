import 'package:flutter_test/flutter_test.dart';
import 'package:vigor/models/activity.dart';
import 'package:vigor/models/activity_actual.dart';
import 'package:vigor/models/activity_actual_ext.dart';
import 'package:vigor/models/training_feedback.dart';
import 'package:vigor/models/training_feedback_ext.dart';
import 'package:vigor/utils/feedback_modal.dart';

Activity benchActivity() => Activity(
      id: 'act-1',
      blockId: 'block-1',
      exerciseId: 'bench-press',
      name: 'Bench Press',
      duration: 0,
      reps: 15,
      weightKg: 40,
      modifiers: const [],
      rest: 60,
      detail: const {},
    );

void main() {
  group('activityActualFromDraft', () {
    test('untouched draft records nothing', () {
      final actual = activityActualFromDraft(
        activity: benchActivity(),
        name: 'Bench Press',
        reps: '15',
        duration: '0',
        weight: '40',
        skipped: false,
      );
      expect(actual, isNull);
    });

    test('deviating load records only the load', () {
      final actual = activityActualFromDraft(
        activity: benchActivity(),
        name: 'Bench Press',
        reps: '15',
        duration: '0',
        weight: '30',
        skipped: false,
      );
      expect(actual, isNotNull);
      expect(actual!.weightKg, 30);
      expect(actual.reps, isNull);
      expect(actual.name, isNull);
      expect(actual.isSkipped, isFalse);
    });

    test('comma decimal weight parses', () {
      final actual = activityActualFromDraft(
        activity: benchActivity(),
        name: 'Bench Press',
        reps: '12',
        duration: '0',
        weight: '30,5',
        skipped: false,
      );
      expect(actual!.reps, 12);
      expect(actual.weightKg, 30.5);
    });

    test('rename records the new name only', () {
      final actual = activityActualFromDraft(
        activity: benchActivity(),
        name: 'Panca piana',
        reps: '15',
        duration: '0',
        weight: '40',
        skipped: false,
      );
      expect(actual!.name, 'Panca piana');
      expect(actual.weightKg, isNull);
    });

    test('cleared name records nothing', () {
      final actual = activityActualFromDraft(
        activity: benchActivity(),
        name: '  ',
        reps: '15',
        duration: '0',
        weight: '40',
        skipped: false,
      );
      expect(actual, isNull);
    });

    test('skipped records only the flag', () {
      final actual = activityActualFromDraft(
        activity: benchActivity(),
        name: 'Bench Press',
        reps: '15',
        duration: '0',
        weight: '30',
        skipped: true,
      );
      expect(actual!.isSkipped, isTrue);
      expect(actual.weightKg, isNull);
    });

    test('unparsable numbers record nothing', () {
      final actual = activityActualFromDraft(
        activity: benchActivity(),
        name: 'Bench Press',
        reps: 'many',
        duration: '0',
        weight: 'heavy',
        skipped: false,
      );
      expect(actual, isNull);
    });
  });

  group('feedback actuals parsing', () {
    test('actualsByActivityId parses recorded entries', () {
      final feedback = TrainingFeedback.fromJson({
        'id': 'fb-1',
        'trainingId': 't-1',
        'userId': 'u-1',
        'qualityReason': '',
        'message': '',
        'activityFeedback': <String, dynamic>{},
        'activityActuals': {
          'act-1': {'reps': 12, 'weightKg': 30.5, 'name': 'Panca piana'},
          'act-2': {'skipped': true},
          'act-3': 'garbage',
        },
        'createdAt': '2026-10-23T00:00:00Z',
      });
      final actuals = feedback.actualsByActivityId;
      expect(actuals.length, 2);
      expect(actuals['act-1']!.reps, 12);
      expect(actuals['act-1']!.weightKg, 30.5);
      expect(actuals['act-1']!.name, 'Panca piana');
      expect(actuals['act-2']!.isSkipped, isTrue);
    });

    test('toApiJson drops null fields', () {
      final json = ActivityActual(reps: 12, name: 'Panca piana').toApiJson();
      expect(json, {'reps': 12, 'name': 'Panca piana'});
    });

    test('formatWeightValue matches the tag format', () {
      expect(formatWeightValue(30), '30');
      expect(formatWeightValue(30.5), '30.5');
      expect(formatWeightValue(7.25), '7.3');
    });

    test('formatWeightForField keeps the exact value', () {
      expect(formatWeightForField(30), '30');
      expect(formatWeightForField(7.25), '7.25');
    });
  });
}

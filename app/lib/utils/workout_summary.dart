import '../models/activity.dart';
import '../models/activity_actual.dart';
import '../models/activity_actual_ext.dart';
import '../models/block.dart';
import '../models/routine.dart';
import '../models/training.dart';

/// Labels the summary builder needs, resolved by the caller from
/// the application localizations.
typedef WorkoutSummaryLabels = ({
  String warmup,
  String work,
  String cooldown,
  String rest,
});

/// Builds the plain text summary of a training for pasting into
/// another application. Recorded actuals replace the prescribed
/// values where they exist; a skipped activity leaves the
/// summary. The wording never claims a performance, so one
/// format serves planned and completed trainings alike.
String buildWorkoutSummary(
  Training training, {
  required Map<String, ActivityActual> actuals,
  required WorkoutSummaryLabels labels,
  required String dateText,
  required String methodologyLabel,
}) {
  final lines = <String>[
    '${training.name} · $dateText',
    '$methodologyLabel · ${_formatDuration(training.completedIn ?? training.duration)}',
  ];

  for (final routine in training.routines) {
    final section = _routineLines(routine, actuals, labels);
    if (section.isEmpty) continue;
    lines
      ..add('')
      ..add(_routineLabel(routine, labels))
      ..addAll(section);
  }

  return lines.join('\n');
}

String _routineLabel(Routine routine, WorkoutSummaryLabels labels) => switch (routine.type) {
      'warmup' => labels.warmup,
      'work' => labels.work,
      'cooldown' => labels.cooldown,
      _ => routine.type,
    };

List<String> _routineLines(Routine routine, Map<String, ActivityActual> actuals, WorkoutSummaryLabels labels) {
  final lines = <String>[];
  for (final block in routine.blocks) {
    final visible = [
      for (final activity in block.activities)
        if (!(actuals[activity.id]?.isSkipped ?? false)) activity,
    ];
    if (visible.isEmpty) continue;
    if (visible.length > 1 && block.repeats > 1) {
      final rest = block.rest > 0 ? ' · ${labels.rest} ${block.rest}s' : '';
      lines.add('×${block.repeats}$rest:');
    }
    for (final activity in visible) {
      lines.add(_activityLine(activity, block, actuals, labels, inlineSets: visible.length == 1));
    }
  }
  return lines;
}

String _activityLine(
  Activity activity,
  Block block,
  Map<String, ActivityActual> actuals,
  WorkoutSummaryLabels labels, {
  required bool inlineSets,
}) {
  final actual = actuals[activity.id];
  final name = actual?.name ?? activity.name;
  final reps = actual?.reps ?? activity.reps;
  final duration = actual?.duration ?? activity.duration;
  final weight = actual?.weightKg ?? activity.weightKg;

  final sets = inlineSets && block.repeats > 1 ? '${block.repeats}×' : '';
  final volume = reps > 0 ? '$sets$reps' : '$sets${_formatDuration(duration)}';
  final weightPart = weight > 0 ? ' · ${formatWeightValue(weight)} kg' : '';
  final rest = inlineSets && block.rest > 0 ? block.rest : activity.rest;
  final restPart = rest > 0 ? ' · ${labels.rest} ${rest}s' : '';
  return '- $name: $volume$weightPart$restPart';
}

/// Duration text in the shape the training detail uses:
/// 90s under a minute stays seconds, then minutes, then hours.
String _formatDuration(int seconds) {
  if (seconds < 60) return '${seconds}s';
  final minutes = seconds ~/ 60;
  final remainingSeconds = seconds % 60;
  if (minutes < 60) {
    return remainingSeconds == 0 ? '${minutes}m' : '${minutes}m ${remainingSeconds}s';
  }
  final hours = minutes ~/ 60;
  final remainingMinutes = minutes % 60;
  return remainingMinutes == 0 ? '${hours}h' : '${hours}h ${remainingMinutes}m';
}
